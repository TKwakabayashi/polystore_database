package migrator

import (
	"context"
	"fmt"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"

	_ "github.com/go-sql-driver/mysql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func VerifyMigration(ctx context.Context, cfg MigrationConfig, destKind storage.StoreKind, reg *storage.Registry) error {
	fmt.Printf("🧐 ターゲットDB (%s) の書き込み整合性を確認中...\n", destKind.String())

	switch destKind {
	case storage.Graph:
		return verifyGraph(ctx, cfg, reg)
	case storage.Kvs:
		return verifyKvs(ctx, cfg, reg)
	case storage.Document:
		return verifyDoc(ctx, cfg, reg)
	case storage.Relational:
		return verifyRdb(ctx, cfg, reg)
	case storage.Columnar:
		return verifyCol(ctx, cfg, reg)
	default:
		return fmt.Errorf("unsupported data store")
	}
}

func verifyGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) error {
	// Neo4jにノードまたはリレーションがあるか確認
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	if cfg.ObjType == plan.Relationship {
		query = fmt.Sprintf("MATCH ()-[r:%s]-() RETURN count(r) > 0 AS exists LIMIT 1", cfg.Entity)
	} else {
		query = fmt.Sprintf("MATCH (n:%s) RETURN count(n) > 0 AS exists LIMIT 1", cfg.Entity)
	}

	res, err := session.Run(ctx, query, nil)
	if err != nil || !res.Next(ctx) {
		return fmt.Errorf("Neo4j: エンティティ %s が見つかりません", cfg.Entity)
	}
	exists, _ := res.Record().Get("exists")
	if b, ok := exists.(bool); !ok || !b {
		return fmt.Errorf("Neo4j: データが0件です")
	}

	return nil
}

func verifyKvs(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) error {
	db, ok := reg.LevelDB()
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	// Entity名 + Sep で始まるキーが1つ以上あるか確認
	prefix := cfg.Entity + "\x00"
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()
	if !iter.Next() {
		return fmt.Errorf("LevelDB: エンティティ %s のデータが見つかりません", cfg.Entity)
	}

	return nil
}

func verifyDoc(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) error {
	// コレクションに1件以上ドキュメントがあるか確認
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	coll := db.Collection(cfg.Entity)

	count, err := coll.CountDocuments(ctx, bson.M{}, options.Count().SetLimit(1))
	if err != nil || count == 0 {
		return fmt.Errorf("MongoDB: コレクション %s にデータがありません", cfg.Entity)
	}

	return nil
}

func verifyRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) error {
	// テーブルが存在し、かつ1行以上あるか確認
	sqlDB, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s` LIMIT 1", cfg.Entity)
	err := sqlDB.QueryRowContext(ctx, query).Scan(&count)
	if err != nil || count == 0 {
		return fmt.Errorf("MySQL: テーブル %s が空、または存在しません (err: %v)", cfg.Entity, err)
	}

	return nil
}

func verifyCol(ctx context.Context, cfg MigrationConfig, reg *storage.Registry) error {
	// Cassandraにデータがあるか確認
	sess, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("columnar store not available")
	}
	var uuid string
	query := fmt.Sprintf("SELECT uuid FROM \"%s\" LIMIT 1", cfg.Entity)
	err := sess.Query(query).WithContext(ctx).Scan(&uuid)
	if err != nil {
		return fmt.Errorf("Cassandra: テーブル %s からの読み取りに失敗しました (err: %v)", cfg.Entity, err)
	}

	return nil
}
