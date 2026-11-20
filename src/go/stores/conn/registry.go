package conn

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/base/kind"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
)

type Registry struct {
	stores map[kind.Kind]Store
}

// NewRegistry connects every configured store at once (all or nothing).
func NewRegistry(ctx context.Context, cfg Config) (*Registry, error) {
	rg := &Registry{stores: make(map[kind.Kind]Store)}
	for _, k := range cfg.configuredKinds() {
		s, err := openStore(ctx, cfg, k)
		if err != nil {
			_ = rg.Close(ctx)
			return nil, fmt.Errorf("connect %s: %w", k, err)
		}
		rg.stores[k] = s
	}
	return rg, nil
}

func NewRegistryFor(ctx context.Context, cfg Config, kinds ...kind.Kind) (*Registry, error) {
	rg := &Registry{stores: make(map[kind.Kind]Store)}
	for _, k := range kinds {
		if rg.Has(k) {
			continue // duplicates are ignored
		}
		s, err := openStore(ctx, cfg, k)
		if err != nil {
			_ = rg.Close(ctx)
			return nil, fmt.Errorf("connect %s: %w", k, err)
		}
		rg.stores[k] = s
	}
	return rg, nil
}

func (rg *Registry) Has(k kind.Kind) bool {
	_, ok := rg.stores[k]
	return ok
}

func (rg *Registry) Close(ctx context.Context) error {
	var errs []string
	for k, s := range rg.stores {
		if err := s.Close(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", k, err))
		}
	}
	rg.stores = make(map[kind.Kind]Store)
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func openStore(ctx context.Context, cfg Config, k kind.Kind) (Store, error) {
	switch k {
	case kind.Neo4j:
		if cfg.Neo4j == nil {
			return nil, fmt.Errorf("%s not configured", k)
		}
		return openNeo4j(ctx, *cfg.Neo4j)
	case kind.Mongo:
		if cfg.Mongo == nil {
			return nil, fmt.Errorf("%s not configured", k)
		}
		return openMongo(ctx, *cfg.Mongo)
	case kind.LevelDB, kind.RocksDB:
		return openKV(ctx, cfg, k)
	case kind.MySQL:
		if cfg.MySQL == nil {
			return nil, fmt.Errorf("%s not configured", k)
		}
		return openMySQL(ctx, *cfg.MySQL)
	case kind.ClickHouse:
		if cfg.ClickHouse == nil {
			return nil, fmt.Errorf("%s not configured", k)
		}
		return openClickHouse(ctx, *cfg.ClickHouse)
	case kind.Cassandra:
		if cfg.Cassandra == nil {
			return nil, fmt.Errorf("%s not configured", k)
		}
		return openCassandra(ctx, *cfg.Cassandra)
	default:
		return nil, fmt.Errorf("unknown store kind: %v", k)
	}
}

func (cfg Config) configuredKinds() []kind.Kind {
	var ks []kind.Kind
	if cfg.Neo4j != nil {
		ks = append(ks, kind.Neo4j)
	}
	if cfg.Mongo != nil {
		ks = append(ks, kind.Mongo)
	}
	for _, k := range []kind.Kind{kind.LevelDB, kind.RocksDB} {
		if cfg.kvConfigured(k) {
			ks = append(ks, k)
		}
	}
	if cfg.MySQL != nil {
		ks = append(ks, kind.MySQL)
	}
	if cfg.ClickHouse != nil {
		ks = append(ks, kind.ClickHouse)
	}
	if cfg.Cassandra != nil {
		ks = append(ks, kind.Cassandra)
	}
	return ks
}

func (rg *Registry) Neo4j() (neo4j.DriverWithContext, bool) {
	if s, ok := rg.stores[kind.Neo4j].(*Neo4jStore); ok {
		return s.Driver(), true
	}
	return nil, false
}
func (rg *Registry) Mongo() (*mongo.Database, bool) {
	if s, ok := rg.stores[kind.Mongo].(*MongoStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) MySQL() (*sql.DB, bool) {
	if s, ok := rg.stores[kind.MySQL].(*MySQLStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) ClickHouse() (*sql.DB, bool) {
	if s, ok := rg.stores[kind.ClickHouse].(*ClickHouseStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) Cassandra() (*gocql.Session, bool) {
	if s, ok := rg.stores[kind.Cassandra].(*CassandraStore); ok {
		return s.Session(), true
	}
	return nil, false
}
