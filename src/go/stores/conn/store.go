package conn

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/base/kind"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	Kind() kind.Kind
	Close(ctx context.Context) error
}

type Neo4jStore struct {
	driver neo4j.DriverWithContext
}

func openNeo4j(ctx context.Context, cfg Neo4jConfig) (*Neo4jStore, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver: %w", err)
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("neo4j connect: %w", err)
	}
	return &Neo4jStore{driver: driver}, nil
}

func (s *Neo4jStore) Kind() kind.Kind                 { return kind.Neo4j }
func (s *Neo4jStore) Driver() neo4j.DriverWithContext { return s.driver }
func (s *Neo4jStore) Close(ctx context.Context) error { return s.driver.Close(ctx) }

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
}

func openMongo(ctx context.Context, c MongoConfig) (*MongoStore, error) {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(c.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := cl.Ping(ctx, nil); err != nil {
		_ = cl.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &MongoStore{client: cl, db: cl.Database(c.DBName)}, nil
}
func (s *MongoStore) Kind() kind.Kind                 { return kind.Mongo }
func (s *MongoStore) Client() *mongo.Client           { return s.client }
func (s *MongoStore) DB() *mongo.Database             { return s.db }
func (s *MongoStore) Close(ctx context.Context) error { return s.client.Disconnect(ctx) }

type MySQLStore struct{ db *sql.DB }

func openMySQL(ctx context.Context, c MySQLConfig) (*MySQLStore, error) {
	db, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	return &MySQLStore{db: db}, nil
}
func (s *MySQLStore) Kind() kind.Kind               { return kind.MySQL }
func (s *MySQLStore) DB() *sql.DB                   { return s.db }
func (s *MySQLStore) Close(_ context.Context) error { return s.db.Close() }

type CassandraStore struct{ session *gocql.Session }

func openCassandra(_ context.Context, c CassandraConfig) (*CassandraStore, error) {
	cluster := gocql.NewCluster(c.Hosts...)
	cluster.Keyspace = c.Keyspace
	// Avoid timeouts in large bulk upserts (about 2.86M Message rows):
	// - the default 600ms is too short under heavy writes, so it is raised
	// - transient timeouts are retried
	// - writes use consistency One (a single node)
	cluster.Timeout = 30 * time.Second
	cluster.ConnectTimeout = 30 * time.Second
	cluster.Consistency = gocql.One
	cluster.NumConns = 4
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 5}
	sess, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cassandra session: %w", err)
	}
	return &CassandraStore{session: sess}, nil
}
func (s *CassandraStore) Kind() kind.Kind               { return kind.Cassandra }
func (s *CassandraStore) Session() *gocql.Session       { return s.session }
func (s *CassandraStore) Close(_ context.Context) error { s.session.Close(); return nil }
