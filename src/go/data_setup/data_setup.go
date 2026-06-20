package data_setup

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"polystore_database/src/go/storage"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Run は設定ファイルの接続情報に従い、Neo4j のラベル/UUID/制約セットアップと
// 他4ストアのクリアを行う。（email/language の CSV 取り込みは現状スキップ）
func Run(ctx context.Context, cfg storage.Config) error {
	/*
		emailUrl := "file:///dynamic_csv/Person_email_EmailAddress.csv"
		langUrl := "file:///dynamic_csv/Person_speaks_Language.csv"

		// --- 1. EmailAddressId のインポート ---
		fmt.Printf("Updating EmailAddressId from %s...\n", emailUrl)
		_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {

			query := fmt.Sprintf(`
				CALL apoc.periodic.iterate(
					"LOAD CSV WITH HEADERS FROM '%s' AS row FIELDTERMINATOR '|' RETURN row",
					"MATCH (p:Person {id: toInteger(row.PersonId)})
					 SET p.email = row.EmailAddressId",
					 {batchSize: 10000, parallel: false}
				)
			`, emailUrl)
			return tx.Run(ctx, query, nil)
		})
		if err != nil {
			log.Fatalf("CRITICAL: Failed to update EmailAddressId: %v", err)
		}

		// --- 2. language のインポート ---
		fmt.Printf("Updating LanguageId from %s...\n", langUrl)
		_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := fmt.Sprintf(`
				CALL apoc.periodic.iterate(
					"LOAD CSV WITH HEADERS FROM '%s' AS row FIELDTERMINATOR '|' RETURN row",
					"MATCH (p:Person {id: toInteger(row.PersonId)})
					 SET p.speaks = row.LanguageId",
					 {batchSize: 10000, parallel: false}
				)
			`, langUrl)
			return tx.Run(ctx, query, nil)
		})
		if err != nil {
			log.Fatalf("CRITICAL: Failed to update LanguageId: %v", err)
		}

		fmt.Println("All property updates from CSV completed successfully.")
	*/
	if cfg.Neo4j == nil {
		return fmt.Errorf("neo4j 設定がありません")
	}

	driver, err := neo4j.NewDriverWithContext(
		cfg.Neo4j.URI,
		neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""),
	)
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// 1. 全ノードに共通ラベル 'Entity' を付与  インデックスポリシーに依存
	fmt.Println("Applying common label 'Entity' to all nodes...")
	if _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CALL apoc.periodic.iterate(
				"MATCH (n) WHERE NOT n:Entity RETURN n",
				"SET n:Entity",
				{batchSize: 10000, parallel: false}
			)`
		return tx.Run(ctx, query, nil)
	}); err != nil {
		return fmt.Errorf("failed to apply common label: %w", err)
	}

	// 2. ノードに UUID を付与
	fmt.Println("Assigning UUIDs to nodes...")
	if _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CALL apoc.periodic.iterate(
				"MATCH (n:Entity) WHERE n.uuid IS NULL RETURN n",
				"SET n.uuid = 'N-' + apoc.create.uuid()",
				{batchSize: 10000, parallel: false}
			)`
		return tx.Run(ctx, query, nil)
	}); err != nil {
		return fmt.Errorf("failed to assign node UUIDs: %w", err)
	}

	// 3. リレーションシップに UUID を付与
	fmt.Println("Assigning UUIDs to relationships...")
	if _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CALL apoc.periodic.iterate(
				"MATCH ()-[r]->() WHERE r.uuid IS NULL RETURN r",
				"SET r.uuid = 'E-' + apoc.create.uuid()",
				{batchSize: 10000, parallel: false}
			)`
		return tx.Run(ctx, query, nil)
	}); err != nil {
		return fmt.Errorf("failed to assign relationship UUIDs: %w", err)
	}

	// 4. Entity(uuid) の一意制約
	fmt.Println("Creating unique constraint for Entity(uuid)...")
	if _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := "CREATE CONSTRAINT node_uuid_unique IF NOT EXISTS FOR (n:Entity) REQUIRE n.uuid IS UNIQUE"
		return tx.Run(ctx, query, nil)
	}); err != nil {
		return fmt.Errorf("failed to create constraint: %w", err)
	}

	// 5. 各リレーションシップ型に uuid 一意制約
	relresult, err := session.Run(ctx, "CALL db.relationshipTypes() YIELD relationshipType RETURN relationshipType", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch relationship types: %w", err)
	}
	var relTypes []string
	for relresult.Next(ctx) {
		if r, ok := relresult.Record().Get("relationshipType"); ok {
			relTypes = append(relTypes, r.(string))
		}
	}
	for _, relType := range relTypes {
		constraintName := strings.ToLower(relType) + "_uuid_unique"
		query := fmt.Sprintf(
			"CREATE CONSTRAINT %s IF NOT EXISTS FOR ()-[r:%s]-() REQUIRE r.uuid IS UNIQUE",
			constraintName, relType,
		)
		if _, err := session.Run(ctx, query, nil); err != nil {
			fmt.Printf("Could not create constraint for relationship %s: %v\n", relType, err)
			continue
		}
	}

	// --- 他4ストアのクリア（接続情報はすべて cfg から）---
	if cfg.Mongo != nil {
		if err := cleanMongoDB(ctx, cfg.Mongo.URI, cfg.Mongo.DBName); err != nil {
			return fmt.Errorf("MongoDB cleanup failed: %w", err)
		}
	}
	if cfg.LevelDB != nil {
		if err := cleanLevelDB(cfg.LevelDB.Path); err != nil {
			return fmt.Errorf("LevelDB cleanup failed: %w", err)
		}
	}
	if cfg.MySQL != nil {
		if err := cleanMySQL("mysql", cfg.MySQL.DSN); err != nil {
			return fmt.Errorf("MySQL cleanup failed: %w", err)
		}
	}
	if cfg.Cassandra != nil {
		if err := cleanCassandra(cfg.Cassandra.Hosts, cfg.Cassandra.Keyspace); err != nil {
			return fmt.Errorf("Cassandra cleanup failed: %w", err)
		}
	}

	fmt.Println("All processes completed successfully.")
	return nil
}

func cleanMongoDB(ctx context.Context, uri, dbName string) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer client.Disconnect(ctx)

	// データベース自体を削除
	return client.Database(dbName).Drop(ctx)
}

func cleanLevelDB(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("path resolution error: %w", err)
	}

	// ディレクトリが存在するか確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("[INFO] LevelDB path %s does not exist, skipping.", absPath)
		return nil
	}

	// ディレクトリごと削除して初期化
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", absPath, err)
	}
	return nil
}

func cleanMySQL(driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer db.Close()

	// 接続確認
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping error: %w", err)
	}

	// 外部キー制約の無効化
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// 全テーブル名の取得
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return fmt.Errorf("failed to fetch tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		// 各テーブルを空にする
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}
	return nil
}

func cleanCassandra(hosts []string, keyspace string) error {
	cluster := gocql.NewCluster(hosts...)
	// タイムアウト設定（必要に応じて調整）
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer session.Close()

	// メタデータの取得
	keyspaceMetadata, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		return fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	for tableName := range keyspaceMetadata.Tables {
		if err := session.Query(fmt.Sprintf("TRUNCATE %s.%s", keyspace, tableName)).Exec(); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}
	return nil
}
