package setup

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/stores/conn"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RunDataset, with the config's connections, reshapes the graph for dataset, prepares the uuid indexes (per
// label and type) and clears the other stores (MySQL / MongoDB / ClickHouse / Cassandra; the embedded stores
// were removed with runtime/ by PrepareRuntime, and an existing directory is cleared too). An empty dataset
// skips the reshaping (a test Neo4j outside compose, ...). The email/language CSV import is skipped for now.
func RunDataset(ctx context.Context, cfg conn.Config, dataset string) error {
	/*
		emailUrl := "file:///dynamic_csv/Person_email_EmailAddress.csv"
		langUrl := "file:///dynamic_csv/Person_speaks_Language.csv"

		// --- 1. import EmailAddressId ---
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

		// --- 2. import language ---
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
		return fmt.Errorf("no neo4j in the config")
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
	/*
		// assign uuids to nodes
		fmt.Println("Assigning UUIDs to nodes...")
		if _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := `
				CALL apoc.periodic.iterate(
					"MATCH (n) WHERE n.uuid IS NULL RETURN n",
					"SET n.uuid = 'N-' + apoc.create.uuid()",
					{batchSize: 10000, parallel: false}
				)`
			return tx.Run(ctx, query, nil)
		}); err != nil {
			return fmt.Errorf("failed to assign node UUIDs: %w", err)
		}

		// assign uuids to relationships
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
	*/
	// Reshape for the dataset first (uuid indexes are made per label present at that point).
	if prepare, ok := datasetPrepare[dataset]; ok {
		fmt.Printf("Preparing graph for dataset %s...\n", dataset)
		if err := prepare(ctx, session); err != nil {
			return fmt.Errorf("dataset %s: %w", dataset, err)
		}
	} else {
		fmt.Printf("No dataset-specific preparation for %s\n", dataset)
	}
	if err := ensureUUIDIndexes(ctx, session); err != nil {
		return err
	}

	// --- clear the other stores (all connections from cfg) ---
	if cfg.Mongo != nil {
		if err := cleanMongoDB(ctx, cfg.Mongo.URI, cfg.Mongo.DBName); err != nil {
			return fmt.Errorf("MongoDB cleanup failed: %w", err)
		}
	}
	// Clear the data directories of the key-value stores (LevelDB and RocksDB).
	for _, k := range []kind.Kind{kind.LevelDB, kind.RocksDB} {
		if path := cfg.KVPath(k); path != "" {
			if err := cleanLevelDB(path); err != nil {
				return fmt.Errorf("%s cleanup failed: %w", k, err)
			}
		}
	}
	if cfg.MySQL != nil {
		if err := cleanMySQL("mysql", cfg.MySQL.DSN); err != nil {
			return fmt.Errorf("MySQL cleanup failed: %w", err)
		}
	}
	if cfg.ClickHouse != nil {
		if err := cleanClickHouse(ctx, *cfg.ClickHouse); err != nil {
			return fmt.Errorf("ClickHouse cleanup failed: %w", err)
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

	// drop the database itself
	return client.Database(dbName).Drop(ctx)
}

// cleanClickHouse creates the DB if missing and drops every table (migrations recreate the definitions).
func cleanClickHouse(ctx context.Context, c conn.ClickHouseConfig) error {
	db, err := conn.OpenClickHouseDB(ctx, c)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, "SELECT name FROM system.tables WHERE database = currentDatabase()")
	if err != nil {
		return fmt.Errorf("failed to fetch tables: %w", err)
	}
	var tables []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, n)
	}
	rows.Close()
	for _, t := range tables {
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS `"+t+"` SYNC"); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", t, err)
		}
	}
	return nil
}

// cleanLevelDB removes the kvs data directory (LevelDB / RocksDB).
func cleanLevelDB(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("path resolution error: %w", err)
	}

	// check whether the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("[INFO] kvs path %s does not exist, skipping.", absPath)
		return nil
	}

	// remove the whole directory
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

	// check the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping error: %w", err)
	}

	// disable foreign key checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// list every table
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
		// empty each table
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}
	return nil
}

func cleanCassandra(hosts []string, keyspace string) error {
	cluster := gocql.NewCluster(hosts...)
	// timeout (adjust as needed)
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer session.Close()

	// Create the keyspace if missing (so a first setup does not fail): SimpleStrategy / RF=1 for a single
	// development node; change it for production.
	createKeyspace := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		keyspace,
	)
	if err := session.Query(createKeyspace).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace %s: %w", keyspace, err)
	}

	// read the metadata
	keyspaceMetadata, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		return fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	for tableName := range keyspaceMetadata.Tables {
		// Tables are created with quoted mixed-case names ("Message", ...); TRUNCATE must quote them too, or Cassandra
		// lowercases the name and fails with "table place does not exist".
		if err := session.Query(fmt.Sprintf("TRUNCATE %s.%q", keyspace, tableName)).Exec(); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}
	return nil
}
