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

func upsertDataStream(ctx context.Context, cfg MigrationConfig, dbKind storage.StoreKind, reg *storage.Registry, inCh <-chan DataRowStream, outCh chan<- DataRowStream, typeMap map[string]string) error {
	const batchSize = 2000
	batch := make([]DataRowStream, 0, batchSize)

	// バッチ書き込み実行処理
	flush := func(rows []DataRowStream) error {
		if len(rows) == 0 {
			return nil
		}

		switch dbKind {
		case storage.Relational:
			return upsertRdb(ctx, cfg, reg, typeMap, rows)
		case storage.Document:
			return upsertDoc(ctx, cfg, reg, typeMap, rows)
		case storage.Graph:
			return upsertGraph(ctx, cfg, reg, typeMap, rows)
		case storage.Columnar:
			return upsertCol(ctx, cfg, reg, typeMap, rows)
		case storage.Kvs:
			return upsertKvs(ctx, cfg, reg, typeMap, rows)
		default:

		}

		// 全ての書き込みが物理的に成功した後、Deleteステージへ流す
		for _, row := range rows {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case outCh <- row:
			}
		}
		return nil
	}

	// メインループ
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

func upsertGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	unwindParams := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		cleanPayload := make(map[string]interface{})
		for _, p := range cfg.Properties {
			if v, ok := row.Payload[p]; ok {
				finalVal, _ := codec.PrepareForDB(v, typeMap[p], "graph")
				cleanPayload[p] = finalVal
			}
		}
		unwindParams[i] = map[string]interface{}{"uuid": row.UUID, "payload": cleanPayload}
	}

	var query string
	if cfg.ObjType == plan.Relationship {
		query = fmt.Sprintf("UNWIND $batch AS row MATCH ()-[r:%s {uuid: row.uuid}]-() SET r += row.payload", cfg.Entity)
	} else {
		query = fmt.Sprintf("UNWIND $batch AS row MERGE (n:Entity:%s {uuid: row.uuid}) SET n += row.payload", cfg.Entity)
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, map[string]interface{}{"batch": unwindParams})
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	if err != nil {
		return fmt.Errorf("neo4j bulk upsert failed: %w", err)
	}

	return nil
}

func upsertKvs(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.LevelDB()
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	levelBatch := new(leveldb.Batch)
	for _, row := range rows {
		for p, val := range row.Payload {
			// KVS専用バイナリエンコード
			valBytes, err := codec.EncodeForKVS(val, typeMap[p])
			if err != nil {
				return fmt.Errorf("kvs encode failed: %w", err)
			}
			if valBytes == nil {
				return fmt.Errorf("KVSエンコード結果がnilです (UUID: %s, Prop: %s)", row.UUID.String(), p)
			}

			entityKey := codec.BuildEntityKey(cfg.Entity, row.UUID.String(), p)
			if oldVal, err := db.Get(entityKey, nil); err == nil {
				levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, oldVal, row.UUID.String()))
			}
			levelBatch.Put(entityKey, valBytes)
			levelBatch.Put(codec.BuildIndexKey(cfg.Entity, p, valBytes, row.UUID.String()), []byte{})
		}
	}
	if err := db.Write(levelBatch, &opt.WriteOptions{Sync: true}); err != nil {
		return fmt.Errorf("leveldb write failed: %w", err)
	}

	return nil
}

func upsertDoc(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	coll := db.Collection(cfg.Entity)
	models := make([]mongo.WriteModel, len(rows))
	for i, row := range rows {
		// MongoDB用に型を整えたPayloadを再構築
		finalPayload := make(map[string]interface{})
		for p, v := range row.Payload {
			finalVal, _ := codec.PrepareForDB(v, typeMap[p], "document")
			finalPayload[p] = finalVal
		}
		models[i] = mongo.NewUpdateOneModel().
			SetFilter(bson.M{"uuid": row.UUID}).
			SetUpdate(bson.M{"$set": finalPayload}).
			SetUpsert(true)
	}
	if _, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false)); err != nil {
		return fmt.Errorf("mongo bulk write failed: %w", err)
	}

	return nil
}

func upsertCol(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("columnar store not available")
	}
	// テーブル準備
	_ = session.Query(fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (uuid text PRIMARY KEY)", cfg.Entity)).WithContext(ctx).Exec()
	for _, p := range cfg.Properties {
		_ = session.Query(fmt.Sprintf("ALTER TABLE \"%s\" ADD \"%s\" %s", cfg.Entity, p, codec.MapToCassandraType(typeMap[p]))).WithContext(ctx).Exec()
	}

	eg, gctx := errgroup.WithContext(ctx)
	quotedProps := make([]string, len(cfg.Properties))
	for i, p := range cfg.Properties {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\" (uuid, %s) VALUES (?, %s)",
		cfg.Entity, strings.Join(quotedProps, ", "), strings.Repeat("?, ", len(cfg.Properties)-1)+"?")

	for _, r := range rows {
		r := r
		eg.Go(func() error {
			args := []interface{}{r.UUID}
			for _, p := range cfg.Properties {
				// カサンドラ用に int32 / int64 を厳格化
				finalVal, _ := codec.PrepareForDB(r.Payload[p], typeMap[p], "columnar")
				args = append(args, finalVal)
			}
			return session.Query(insertQuery, args...).WithContext(gctx).Exec()
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("cassandra upsert failed: %w", err)
	}

	return nil
}

func upsertRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	// テーブルとカラムの動的準備
	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (uuid VARCHAR(255) PRIMARY KEY)", cfg.Entity)
	if _, err := db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to prepare table %s: %w", cfg.Entity, err)
	}
	for _, p := range cfg.Properties {
		sqlType := codec.MapToSQLType(typeMap[p])
		alterQuery := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", cfg.Entity, p, sqlType)
		_, _ = db.ExecContext(ctx, alterQuery)
	}

	// バルククエリ構築
	numProps := len(cfg.Properties)
	placeholderGroup := "(" + strings.Repeat("?,", numProps) + "?)"
	batchPlaceholders := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*(numProps+1))

	for i, r := range rows {
		batchPlaceholders[i] = placeholderGroup
		args = append(args, r.UUID)
		for _, p := range cfg.Properties {
			// 格納先(RDB)に合わせた最終変換
			finalVal, err := codec.PrepareForDB(r.Payload[p], typeMap[p], "relational")
			if err != nil {
				return err
			}
			args = append(args, finalVal)
		}
	}

	updateClauses := make([]string, numProps)
	quotedProps := make([]string, numProps)
	for i, p := range cfg.Properties {
		quotedProps[i] = fmt.Sprintf("`%s`", p)
		updateClauses[i] = fmt.Sprintf("`%s`=VALUES(`%s`)", p, p)
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (uuid, %s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		cfg.Entity, strings.Join(quotedProps, ", "), strings.Join(batchPlaceholders, ", "), strings.Join(updateClauses, ", "),
	)
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("mysql bulk upsert failed: %w", err)
	}

	return nil
}
