package migrator

import (
	"context"
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	schema "polystore_database/src/go/schema"
	"polystore_database/src/go/storage"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func asUUID(v interface{}) id.UUID {
	if s, ok := v.(string); ok {
		return id.UUID(s)
	}
	return ""
}

func prepareDestSchema(ctx context.Context, cfg MigrationConfig, destKind storage.StoreKind, reg *storage.Registry, typeMap map[string]string) error {
	switch destKind {
	case storage.Relational:
		db, _ := reg.MySQL()
		if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (uuid VARCHAR(255) PRIMARY KEY)", cfg.Entity)); err != nil {
			return err
		}
		for _, p := range cfg.Properties {
			alter := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", cfg.Entity, p, codec.MapToSQLType(typeMap[p]))
			if _, err := db.ExecContext(ctx, alter); err != nil && !strings.Contains(err.Error(), "Duplicate column") {
				return fmt.Errorf("alter %s.%s: %w", cfg.Entity, p, err) // 既存列以外は伝播
			}
		}
	case storage.Columnar:
		sess, _ := reg.Cassandra()
		if err := sess.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (uuid text PRIMARY KEY)`, cfg.Entity)).WithContext(ctx).Exec(); err != nil {
			return err
		}
		for _, p := range cfg.Properties {
			// 既存カラムは "already exists" を許容
			if err := sess.Query(fmt.Sprintf(`ALTER TABLE "%s" ADD "%s" %s`, cfg.Entity, p, codec.MapToCassandraType(typeMap[p]))).WithContext(ctx).Exec(); err != nil &&
				!strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	case storage.Document:
		db, _ := reg.Mongo()
		coll := db.Collection(cfg.Entity)
		// uuid の一意インデックス。無いと upsert が1件ごとに全走査になり O(N^2) で事実上終わらない
		// （Relational の PRIMARY KEY, Columnar の PRIMARY KEY に相当）。同一定義の再作成は冪等。
		if _, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true),
		}); err != nil {
			return fmt.Errorf("create uuid index on %s: %w", cfg.Entity, err)
		}
	case storage.Graph:
		// upsertGraph の MERGE/MATCH を uuid で高速化するインデックス（無いと全走査）。
		// いずれも IF NOT EXISTS で冪等。Enterprise/Community 差やスキーマ競合では
		// 移行を止めず警告のみ（索引が無くても遅くなるだけで正しさには影響しない）。
		drv, _ := reg.Neo4j()
		session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		defer session.Close(ctx)
		var stmt string
		if cfg.ObjType == plan.Relationship {
			// リレーションプロパティの range index は Community でも利用可
			stmt = fmt.Sprintf("CREATE INDEX %s_uuid_idx IF NOT EXISTS FOR ()-[r:%s]-() ON (r.uuid)", strings.ToLower(cfg.Entity), cfg.Entity)
		} else {
			// エンティティは :Entity(uuid) の一意制約（data_setup と同一スキーマ・冪等）
			stmt = "CREATE CONSTRAINT node_uuid_unique IF NOT EXISTS FOR (n:Entity) REQUIRE n.uuid IS UNIQUE"
		}
		if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			res, err := tx.Run(ctx, stmt, nil)
			if err != nil {
				return nil, err
			}
			return res.Consume(ctx)
		}); err != nil {
			fmt.Printf("⚠️ graph uuid index/constraint 準備をスキップ (%s): %v\n", cfg.Entity, err)
		}
	case storage.Kvs:
		// LevelDB はキー自体が entity+uuid+prop の複合キーで、値インデックスも upsert 時に
		// 維持される（BuildEntityKey / BuildIndexKey）。追加のスキーマ準備は不要。
	}
	return nil
}

func MigrateData(config MigrationConfig, sconfig storage.Config) (int, error) {
	mappingDictionary, err := schema.LoadMappingDictionary(config.MappingPath)
	if err != nil {
		return 0, err
	}

	srcKind, destKind, err := modeStores(config.Mode)
	if err != nil {
		return 0, err
	}
	count, err := ExecuteMigrationStream(config, sconfig, srcKind, destKind, mappingDictionary)
	if err != nil {
		return 0, err
	}
	return count, mappingDictionary.SaveMappingDictionary(config.MappingPath)
}

// ExecuteMigrationStream は移行を実行し、成功時に移行件数を返す。
func ExecuteMigrationStream(config MigrationConfig, scfg storage.Config, srcKind, destKind storage.StoreKind, md *schema.MappingDictionary) (int, error) {
	if !md.CheckDatastore(config.ObjType, config.Entity, config.Properties, srcKind.String()) {
		return 0, fmt.Errorf("data not found")
	}

	typeMap := make(map[string]string)
	for _, prop := range config.Properties {
		typeMap[prop], _ = md.GetPropertyDataType(config.ObjType, config.Entity, prop)
	}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	reg, err := storage.NewRegistryFor(mainCtx, scfg, srcKind, destKind)
	if err != nil {
		return 0, fmt.Errorf("registry init: %w", err)
	}
	defer reg.Close(mainCtx)

	// 移行先スキーマを1回・単一スレッドで準備（並行 ALTER 回避）
	if err := prepareDestSchema(mainCtx, config, destKind, reg, typeMap); err != nil {
		return 0, fmt.Errorf("prepare dest schema: %w", err)
	}

	// ===== フェーズ1: Copy（Fetch -> Upsert のみ） =====
	// Delete はこのフェーズには含めない。ソース削除は Verify とマッピング切替が
	// 成功した後（フェーズ3）にのみ、かつ DeleteSource=true のときだけ実行する。
	// いずれかのステージがエラーを返すと ctx.Done() が閉じ、全ステージが停止する。
	eg, ctx := errgroup.WithContext(mainCtx)

	fetchWorkers := fetchWorkersFor(srcKind)
	upsertWorkers := 4

	// 16GBのメモリを考慮し、バッファは1万件程度（1件1KBとしても数10MB）
	fetchToUpsertCh := make(chan DataRowStream, 10000)

	eg.Go(func() error {
		defer close(fetchToUpsertCh)
		fEg, fCtx := errgroup.WithContext(ctx)
		for i := 0; i < fetchWorkers; i++ {
			i := i
			fEg.Go(func() error {
				// 各ワーカーは uuid シャード (i, fetchWorkers) のみを取得し、重複なく分担する
				return fetchDataStream(fCtx, config, srcKind, reg, typeMap, fetchToUpsertCh, i, fetchWorkers)
			})
		}
		return fEg.Wait()
	})

	eg.Go(func() error {
		uEg, uCtx := errgroup.WithContext(ctx)
		for i := 0; i < upsertWorkers; i++ {
			uEg.Go(func() error {
				return upsertDataStream(uCtx, config, destKind, reg, fetchToUpsertCh, typeMap)
			})
		}
		return uEg.Wait()
	})

	if err := eg.Wait(); err != nil {
		return 0, err
	}

	// ===== フェーズ2: Verify（コピー成功のゲート） =====
	// ソース件数と移行先件数を独立に数え、移行先 >= ソース を満たすか確認する。
	// 満たさない場合（Fetch の取りこぼしや書き込み欠損）はマッピングを切り替えず、
	// 削除も行わずに中断する。この時点でソースは無傷なので再実行できる。
	migrated, err := VerifyMigration(mainCtx, config, srcKind, destKind, reg)
	if err != nil {
		return 0, fmt.Errorf("verify failed, mapping switch and source delete skipped: %w", err)
	}

	// ===== マッピング切替（Verify 成功時のみ） =====
	// ここで読み取り先が dest に切り替わる。以降でソースを消してもクエリは dest に向く。
	md.UpdateDatastore(config.ObjType, config.Entity, config.Properties, destKind.String())

	// ===== フェーズ3: Delete（DeleteSource=true のときのみ） =====
	if config.DeleteSource {
		if err := runDeletePhase(mainCtx, config, srcKind, reg, typeMap); err != nil {
			return 0, fmt.Errorf("delete phase failed (data already copied and mapping switched): %w", err)
		}
	}

	fmt.Println("✅ 移行が正常に完了しました。")
	return int(migrated), nil
}

// fetchWorkersFor は Fetch フェーズの並列ワーカー数をソース種別ごとに決める。
// Columnar/Document/Kvs は uuid シャードで安全に並列化できるため 4 並列。
// Graph（Neo4j は往復支配で uuid インデックス依存）と Relational（uuid 列の
// コレーション順とバイト順の不一致で範囲分割が行を取りこぼす懸念）は 1 とする。
func fetchWorkersFor(k storage.StoreKind) int {
	switch k {
	case storage.Columnar, storage.Document, storage.Kvs:
		return 4
	default:
		return 1
	}
}

// runDeletePhase はコピー・検証・マッピング切替の後に、ソースから移行済みデータを削除する。
// ソースを再走査して uuid を集め、deleteDataStream へ流す（DeleteSource=true のときのみ呼ばれる）。
func runDeletePhase(ctx context.Context, cfg MigrationConfig, srcKind storage.StoreKind, reg *storage.Registry, typeMap map[string]string) error {
	const deleteWorkers = 2
	deleteCh := make(chan DataRowStream, 10000)

	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(deleteCh)
		return fetchDataStream(gctx, cfg, srcKind, reg, typeMap, deleteCh, 0, 1)
	})
	for i := 0; i < deleteWorkers; i++ {
		eg.Go(func() error {
			return deleteDataStream(gctx, cfg, srcKind, reg, deleteCh, typeMap)
		})
	}
	return eg.Wait()
}
