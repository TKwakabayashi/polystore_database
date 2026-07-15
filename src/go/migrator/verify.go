package migrator

import (
	"context"
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
)

// VerifyMigration はソースと移行先の件数を独立に数え、移行先 >= ソース を満たすか確認する。
// 満たさない場合はデータが取りこぼされているためエラーを返し、
// 呼び出し側はマッピング切替とソース削除を中止する（=データ消失を防ぐゲート）。
// 成功時は移行件数（=ソース件数）を返す。
func VerifyMigration(ctx context.Context, cfg MigrationConfig, srcKind, destKind storage.StoreKind, reg *storage.Registry) (int64, error) {
	fmt.Printf("🧐 移行整合性を確認中 (src=%s → dest=%s)...\n", srcKind.String(), destKind.String())

	srcCount, err := countStore(ctx, cfg, srcKind, reg)
	if err != nil {
		return 0, fmt.Errorf("source count: %w", err)
	}
	destCount, err := countStore(ctx, cfg, destKind, reg)
	if err != nil {
		return 0, fmt.Errorf("dest count: %w", err)
	}
	if destCount < srcCount {
		return 0, fmt.Errorf("%s: 移行先の件数が不足しています (ソース %d 件, 移行先 %d 件)", destKind.String(), srcCount, destCount)
	}
	fmt.Printf("   → OK: ソース %d 件, 移行先 %d 件\n", srcCount, destCount)
	return srcCount, nil
}

// CountEntity は指定ストアにおける cfg.Entity の件数（uuid 数）を返す。
// 検証・実験ハーネスから移行前後の件数を比較するための公開ヘルパ。
func CountEntity(ctx context.Context, cfg MigrationConfig, kind storage.StoreKind, reg *storage.Registry) (int64, error) {
	return countStore(ctx, cfg, kind, reg)
}

func countStore(ctx context.Context, cfg MigrationConfig, kind storage.StoreKind, reg *storage.Registry) (int64, error) {
	switch kind {
	case storage.Graph:
		return countGraph(ctx, cfg, reg)
	case storage.Kvs:
		return countKvs(ctx, cfg, reg)
	case storage.Document:
		return countDoc(ctx, cfg, reg)
	case storage.Relational:
		return countRdb(ctx, cfg, reg)
	case storage.Columnar:
		return countCol(ctx, cfg, reg)
	default:
		return 0, fmt.Errorf("unsupported data store")
	}
}

func countGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) (int64, error) {
	drv, ok := reg.Neo4j()
	if !ok {
		return 0, fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	if cfg.ObjType == plan.Relationship {
		// 有向 -> で各リレーションを1回だけ数える（無向 - だと両方向で2重計上になる）
		query = fmt.Sprintf("MATCH ()-[r:%s]->() RETURN count(r) AS c", cfg.Entity)
	} else {
		query = fmt.Sprintf("MATCH (n:%s) RETURN count(n) AS c", cfg.Entity)
	}

	res, err := session.Run(ctx, query, nil)
	if err != nil || !res.Next(ctx) {
		return 0, fmt.Errorf("Neo4j: エンティティ %s の件数取得に失敗しました (err: %v)", cfg.Entity, err)
	}
	c, _ := res.Record().Get("c")
	n, ok := c.(int64)
	if !ok {
		return 0, fmt.Errorf("Neo4j: 件数の型が不正です")
	}
	return n, nil
}

func countKvs(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) (int64, error) {
	db, ok := reg.LevelDB()
	if !ok {
		return 0, fmt.Errorf("kvs store not available")
	}
	// キーは Entity + Sep + uuid + Sep + prop。uuid ごとに複数キーがあるため重複を除いて数える。
	prefix := cfg.Entity + codec.Sep
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	var count int64
	var lastUUID id.UUID
	first := true
	for iter.Next() {
		parts := strings.Split(string(iter.Key()), codec.Sep)
		if len(parts) < 3 {
			continue
		}
		u := id.UUID(parts[1])
		if first || u != lastUUID {
			count++
			lastUUID = u
			first = false
		}
	}
	if err := iter.Error(); err != nil {
		return 0, fmt.Errorf("LevelDB: エンティティ %s の走査に失敗しました: %w", cfg.Entity, err)
	}
	return count, nil
}

func countDoc(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) (int64, error) {
	db, ok := reg.Mongo()
	if !ok {
		return 0, fmt.Errorf("document store not available")
	}
	count, err := db.Collection(cfg.Entity).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("MongoDB: コレクション %s の件数取得に失敗しました: %w", cfg.Entity, err)
	}
	return count, nil
}

func countRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) (int64, error) {
	sqlDB, ok := reg.MySQL()
	if !ok {
		return 0, fmt.Errorf("relational store not available")
	}
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", cfg.Entity)
	if err := sqlDB.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("MySQL: テーブル %s の件数取得に失敗しました: %w", cfg.Entity, err)
	}
	return count, nil
}

func countCol(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) (int64, error) {
	sess, ok := reg.Cassandra()
	if !ok {
		return 0, fmt.Errorf("columnar store not available")
	}
	// SELECT COUNT(*) は大規模テーブルでコーディネータ全走査となり timeout しやすいため、
	// uuid をページング走査してクライアント側で数える（fetchCol と同じ自動ページング）。
	iter := sess.Query(fmt.Sprintf("SELECT uuid FROM \"%s\"", cfg.Entity)).
		WithContext(ctx).PageSize(5000).Iter()
	var (
		uuid  string
		count int64
	)
	for iter.Scan(&uuid) {
		count++
	}
	if err := iter.Close(); err != nil {
		return 0, fmt.Errorf("Cassandra: テーブル %s の件数取得に失敗しました: %w", cfg.Entity, err)
	}
	return count, nil
}
