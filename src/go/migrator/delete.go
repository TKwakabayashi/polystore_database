package migrator

import (
	"context"
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func deleteDataStream(ctx context.Context, cfg MigrationConfig, dbKind storage.StoreKind, reg *storage.Registry, inCh <-chan DataRowStream, typeMap map[string]string) error {
	const batchSize = 2000
	batch := make([]DataRowStream, 0, batchSize)

	// MySQL の「全列 NULL 行のみ削除」で使う NULL 条件は、テーブルスキーマから一度だけ構築する
	// （旧実装はバッチごとに INFORMATION_SCHEMA を引いていた）。
	var rdbConds []string
	if dbKind == storage.Relational {
		conds, err := rdbNullConditions(ctx, reg, cfg.Entity)
		if err != nil {
			return err
		}
		rdbConds = conds
	}

	flush := func(rows []DataRowStream) error {
		if len(rows) == 0 {
			return nil
		}

		switch dbKind {
		case storage.Relational:
			return deleteRdb(ctx, cfg, reg, rdbConds, rows)
		case storage.Document:
			return deleteDoc(ctx, cfg, reg, typeMap, rows)
		case storage.Graph:
			return deleteGraph(ctx, cfg, reg, typeMap, rows)
		case storage.Kvs:
			return deleteKvs(ctx, cfg, reg, typeMap, rows)
		case storage.Columnar:
			return deleteCol(ctx, cfg, reg, typeMap, rows)
		default:
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case row, ok := <-inCh:
			if !ok {
				return flush(batch)
			}
			batch = append(batch, row)
			if len(batch) >= batchSize {
				if err := flush(batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}
	}
}

func deleteGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	uuids := make([]string, len(rows))
	for i, r := range rows {
		uuids[i] = r.UUID.String()
	}

	var query string
	if cfg.ObjType == plan.Relationship {
		// Relationshipのプロパティ削除 (APOCを使用)
		query = fmt.Sprintf(`
					MATCH ()-[t:%s]-() WHERE t.uuid IN $uuids 
					CALL apoc.create.setRelProperties(t, $props, [x IN $props | null]) 
					YIELD rel RETURN count(*)`, cfg.Entity)
	} else {
		// Nodeのプロパティ削除
		query = fmt.Sprintf(`
					MATCH (t:%s) WHERE t.uuid IN $uuids 
					CALL apoc.create.removeProperties(t, $props) 
					YIELD node RETURN count(*)`, cfg.Entity)
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{"uuids": uuids, "props": cfg.Properties})
	})
	if err != nil {
		return fmt.Errorf("neo4j properties removal failed: %w", err)
	}

	return nil
}

func deleteKvs(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.LevelDB()
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	levelBatch := new(leveldb.Batch)
	for _, r := range rows {
		for _, p := range cfg.Properties {
			entityKey := codec.BuildEntityKey(cfg.Entity, r.UUID, p)
			// json(複雑値)は転置索引を張っていないため、エンティティキーのみ削除する
			complex := strings.EqualFold(typeMap[p], "json")
			if val, err := db.Get(entityKey, nil); err == nil {
				levelBatch.Delete(entityKey)
				if !complex {
					levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, val, r.UUID))
				}
			}
		}
	}
	// 物理ディスクへの書き込みを待機
	if err := db.Write(levelBatch, &opt.WriteOptions{Sync: true}); err != nil {
		return fmt.Errorf("leveldb delete failed: %w", err)
	}

	return nil
}

func deleteDoc(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	coll := db.Collection(cfg.Entity)

	unsetMap := bson.M{}
	for _, p := range cfg.Properties {
		unsetMap[p] = ""
	}

	var models []mongo.WriteModel
	for _, r := range rows {
		// プロパティの削除
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"uuid": r.UUID.String()}).
			SetUpdate(bson.M{"$unset": unsetMap}))

		// uuidと_id以外にフィールドがなければドキュメントごと削除
		deleteFilter := bson.M{
			"uuid":  r.UUID.String(),
			"$expr": bson.M{"$lte": bson.A{bson.M{"$size": bson.M{"$objectToArray": "$$ROOT"}}, 2}},
		}
		models = append(models, mongo.NewDeleteOneModel().SetFilter(deleteFilter))
	}
	// 順序を守って実行
	if _, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(true)); err != nil {
		return fmt.Errorf("mongodb bulk delete failed: %w", err)
	}

	return nil
}

func deleteCol(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("columnar store not available")
	}
	quotedProps := make([]string, len(cfg.Properties))
	for i, p := range cfg.Properties {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	// カラム単位の削除
	deleteQuery := fmt.Sprintf("DELETE %s FROM \"%s\" WHERE uuid = ?", strings.Join(quotedProps, ", "), cfg.Entity)

	eg, gctx := errgroup.WithContext(ctx)
	// 行ごとの goroutine 無制限生成による接続枯渇を防ぐため上限を設ける
	eg.SetLimit(32)
	for _, r := range rows {
		r := r
		eg.Go(func() error {
			return session.Query(deleteQuery, r.UUID.String()).WithContext(gctx).Exec()
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("cassandra column delete failed: %w", err)
	}

	return nil
}

// rdbNullConditions はテーブルの uuid 以外の全列について "`col` IS NULL" 条件を作る。
// Delete フェーズ開始時に一度だけ呼び、各バッチの deleteRdb で使い回す。
func rdbNullConditions(ctx context.Context, reg *storage.Registry, entity string) ([]string, error) {
	db, ok := reg.MySQL()
	if !ok {
		return nil, fmt.Errorf("relational store not available")
	}
	colRows, err := db.QueryContext(ctx,
		"SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		entity)
	if err != nil {
		return nil, fmt.Errorf("fetch columns for %s: %w", entity, err)
	}
	defer colRows.Close()

	var nullConditions []string
	for colRows.Next() {
		var c string
		if err := colRows.Scan(&c); err != nil {
			return nil, err
		}
		if c != "uuid" {
			nullConditions = append(nullConditions, fmt.Sprintf("`%s` IS NULL", c))
		}
	}
	return nullConditions, colRows.Err()
}

func deleteRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, nullConditions []string, rows []DataRowStream) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	uuids := make([]interface{}, len(rows))
	placeholders := make([]string, len(rows))
	for i, r := range rows {
		uuids[i] = r.UUID.String()
		placeholders[i] = "?"
	}
	inList := strings.Join(placeholders, ", ")

	// 1. 指定されたプロパティを NULL に更新
	setParts := make([]string, len(cfg.Properties))
	for i, p := range cfg.Properties {
		setParts[i] = fmt.Sprintf("`%s` = NULL", p)
	}
	updateQuery := fmt.Sprintf("UPDATE `%s` SET %s WHERE uuid IN (%s)",
		cfg.Entity, strings.Join(setParts, ", "), strings.Join(placeholders, ", "))

	if _, err := db.ExecContext(ctx, updateQuery, uuids...); err != nil {
		return fmt.Errorf("relational property nullify failed: %w", err)
	}

	// 2. uuid以外が全てNULLになった行を削除（未移行列が残る行は消さない）。
	//    uuid 以外の列が無ければ削除条件を作れないので何もしない。
	if len(nullConditions) == 0 {
		return nil
	}

	cleanupQuery := fmt.Sprintf("DELETE FROM `%s` WHERE uuid IN (%s) AND %s",
		cfg.Entity, inList, strings.Join(nullConditions, " AND "))
	if _, err := db.ExecContext(ctx, cleanupQuery, uuids...); err != nil {
		return fmt.Errorf("relational cleanup delete failed: %w", err)
	}

	return nil
}
