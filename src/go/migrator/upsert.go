package migrator

import (
	"context"
	"errors"
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

// upsertDataStream は inCh から受け取った行をバッチで移行先へ書き込む。
// Delete への転送はフェーズ分離により廃止し、書き込み専任にした。
func upsertDataStream(ctx context.Context, cfg MigrationConfig, dbKind storage.StoreKind, reg *storage.Registry, inCh <-chan DataRowStream, typeMap map[string]string) error {
	const batchSize = 2000
	batch := make([]DataRowStream, 0, batchSize)

	flush := func(rows []DataRowStream) error {
		if len(rows) == 0 {
			return nil
		}

		var err error
		switch dbKind {
		case storage.Relational:
			err = upsertRdb(ctx, cfg, reg, typeMap, rows)
		case storage.Document:
			err = upsertDoc(ctx, cfg, reg, typeMap, rows)
		case storage.Graph:
			err = upsertGraph(ctx, cfg, reg, typeMap, rows)
		case storage.Columnar:
			err = upsertCol(ctx, cfg, reg, typeMap, rows)
		case storage.Kvs:
			err = upsertKvs(ctx, cfg, reg, typeMap, rows)
		default:
			return fmt.Errorf("unsupported dest store: %s", dbKind.String())
		}
		if err != nil {
			return err
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

func upsertGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, rows []DataRowStream) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
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
		// 境界変換: id.UUID -> string（neo4j packstream は名前付き型を扱えない）
		unwindParams[i] = map[string]interface{}{"uuid": row.UUID.String(), "payload": cleanPayload}
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
			if val == nil {
				continue
			}
			// KVS専用バイナリエンコード
			valBytes, err := codec.EncodeForKVS(val, typeMap[p])
			if err != nil {
				return fmt.Errorf("kvs encode failed: %w", err)
			}

			// json(複雑値)は保存のみ。転置索引は張らない（値検索は非対応）。
			complex := strings.EqualFold(typeMap[p], "json")

			entityKey := codec.BuildEntityKey(cfg.Entity, row.UUID, p)
			if !complex {
				if oldVal, err := db.Get(entityKey, nil); err == nil {
					levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, oldVal, row.UUID))
				}
			}
			levelBatch.Put(entityKey, valBytes)
			if !complex {
				levelBatch.Put(codec.BuildIndexKey(cfg.Entity, p, valBytes, row.UUID), []byte{})
			}
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
			SetFilter(bson.M{"uuid": row.UUID.String()}).
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
	// スキーマ準備は prepareDestSchema で1回実施済み（ここでは DDL を行わない）

	eg, gctx := errgroup.WithContext(ctx)
	// 行ごとに goroutine を無制限生成すると Cassandra の接続プールが枯渇するため上限を設ける
	eg.SetLimit(32)
	quotedProps := make([]string, len(cfg.Properties))
	for i, p := range cfg.Properties {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\" (uuid, %s) VALUES (?, %s)",
		cfg.Entity, strings.Join(quotedProps, ", "), strings.Repeat("?, ", len(cfg.Properties)-1)+"?")

	for _, r := range rows {
		r := r
		eg.Go(func() error {
			args := []interface{}{r.UUID.String()}
			for _, p := range cfg.Properties {
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
	// スキーマ準備は prepareDestSchema で1回実施済み（ここでは DDL を行わない）

	// ワーカー間でロック取得順を揃え、デッドロック頻度を下げる
	sort.Slice(rows, func(i, j int) bool { return id.Less(rows[i].UUID, rows[j].UUID) })

	numProps := len(cfg.Properties)
	placeholderGroup := "(" + strings.Repeat("?,", numProps) + "?)"
	batchPlaceholders := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*(numProps+1))

	for i, r := range rows {
		batchPlaceholders[i] = placeholderGroup
		args = append(args, r.UUID.String())
		for _, p := range cfg.Properties {
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

	// デッドロック(1213)/ロック待ち(1205)は再試行
	const maxAttempts = 5
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			lastErr = err
			if isMySQLRetryable(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(attempt+1) * 20 * time.Millisecond):
				}
				continue
			}
			return fmt.Errorf("mysql bulk upsert failed: %w", err)
		}
		return nil
	}
	return fmt.Errorf("mysql bulk upsert failed after %d attempts: %w", maxAttempts, lastErr)
}

func isMySQLRetryable(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1213 || me.Number == 1205
	}
	return false
}
