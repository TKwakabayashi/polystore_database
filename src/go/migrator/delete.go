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

	flush := func(rows []DataRowStream) error {
		if len(rows) == 0 {
			return nil
		}

		switch dbKind {
		case storage.Relational:
			return deleteRdb(ctx, cfg, reg, typeMap, rows)
		case storage.Document:
			return deleteDoc(ctx, cfg, reg, typeMap, rows)
		case storage.Graph:
			return deleteGraph(ctx, cfg, reg, typeMap, rows)
		case storage.Kvs:
			return deleteCol(ctx, cfg, reg, typeMap, rows)
		case storage.Columnar:
			return deleteKvs(ctx, cfg, reg, typeMap, rows)
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
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	uuids := make([]id.UUID, len(rows))
	for i, r := range rows {
		uuids[i] = r.UUID
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
			entityKey := codec.BuildEntityKey(cfg.Entity, r.UUID.String(), p)
			// インデックスも確実に消すために現在の値を取得
			if val, err := db.Get(entityKey, nil); err == nil {
				levelBatch.Delete(entityKey)
				levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, val, r.UUID.String()))
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
			SetFilter(bson.M{"uuid": r.UUID}).
			SetUpdate(bson.M{"$unset": unsetMap}))

		// uuidと_id以外にフィールドがなければドキュメントごと削除
		deleteFilter := bson.M{
			"uuid":  r.UUID,
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
	for _, r := range rows {
		r := r
		eg.Go(func() error {
			return session.Query(deleteQuery, r.UUID).WithContext(gctx).Exec()
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("cassandra column delete failed: %w", err)
	}

	return nil
}

func deleteRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	uuids := make([]interface{}, len(rows))
	placeholders := make([]string, len(rows))
	for i, r := range rows {
		uuids[i] = r.UUID
		placeholders[i] = "?"
	}

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

	// 2. uuid以外が全てNULLになった行を削除
	var nullConditions []string
	for p := range typeMap {
		if p != "uuid" {
			nullConditions = append(nullConditions, fmt.Sprintf("`%s` IS NULL", p))
		}
	}
	cleanupQuery := fmt.Sprintf("DELETE FROM `%s` WHERE uuid IN (%s) AND %s",
		cfg.Entity, strings.Join(placeholders, ", "), strings.Join(nullConditions, " AND "))

	if _, err := db.ExecContext(ctx, cleanupQuery, uuids...); err != nil {
		return fmt.Errorf("relational cleanup delete failed: %w", err)
	}

	return nil
}
