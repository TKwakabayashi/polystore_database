package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

type Registry struct {
	stores map[StoreKind]Store
}

// NewRegistry は設定済み全ストアを即時接続（all-or-nothing）。
func NewRegistry(ctx context.Context, cfg Config) (*Registry, error) {
	rg := &Registry{stores: make(map[StoreKind]Store)}
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

func NewRegistryFor(ctx context.Context, cfg Config, kinds ...StoreKind) (*Registry, error) {
	rg := &Registry{stores: make(map[StoreKind]Store)}
	for _, k := range kinds {
		if rg.Has(k) {
			continue // 重複指定は無視
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

func (rg *Registry) Has(kind StoreKind) bool {
	_, ok := rg.stores[kind]
	return ok
}

func (rg *Registry) Close(ctx context.Context) error {
	var errs []string
	for kind, s := range rg.stores {
		if err := s.Close(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", kind, err))
		}
	}
	rg.stores = make(map[StoreKind]Store)
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func openStore(ctx context.Context, cfg Config, kind StoreKind) (Store, error) {
	switch kind {
	case Graph:
		if cfg.Neo4j == nil {
			return nil, fmt.Errorf("%s not configured", kind)
		}
		return openNeo4j(ctx, *cfg.Neo4j)
	case Document:
		if cfg.Mongo == nil {
			return nil, fmt.Errorf("%s not configured", kind)
		}
		return openMongo(ctx, *cfg.Mongo)
	case Kvs:
		if cfg.LevelDB == nil {
			return nil, fmt.Errorf("%s not configured", kind)
		}
		return openLevelDB(ctx, *cfg.LevelDB)
	case Relational:
		if cfg.MySQL == nil {
			return nil, fmt.Errorf("%s not configured", kind)
		}
		return openMySQL(ctx, *cfg.MySQL)
	case Columnar:
		if cfg.Cassandra == nil {
			return nil, fmt.Errorf("%s not configured", kind)
		}
		return openCassandra(ctx, *cfg.Cassandra)
	default:
		return nil, fmt.Errorf("unknown store kind: %v", kind)
	}
}

func (cfg Config) configuredKinds() []StoreKind {
	var ks []StoreKind
	if cfg.Neo4j != nil {
		ks = append(ks, Graph)
	}
	if cfg.Mongo != nil {
		ks = append(ks, Document)
	}
	if cfg.LevelDB != nil {
		ks = append(ks, Kvs)
	}
	if cfg.MySQL != nil {
		ks = append(ks, Relational)
	}
	if cfg.Cassandra != nil {
		ks = append(ks, Columnar)
	}
	return ks
}

func (rg *Registry) Neo4j() (neo4j.DriverWithContext, bool) {
	if s, ok := rg.stores[Graph].(*Neo4jStore); ok {
		return s.Driver(), true
	}
	return nil, false
}
func (rg *Registry) Mongo() (*mongo.Database, bool) {
	if s, ok := rg.stores[Document].(*MongoStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) LevelDB() (*leveldb.DB, bool) {
	if s, ok := rg.stores[Kvs].(*LevelDBStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) MySQL() (*sql.DB, bool) {
	if s, ok := rg.stores[Relational].(*MySQLStore); ok {
		return s.DB(), true
	}
	return nil, false
}
func (rg *Registry) Cassandra() (*gocql.Session, bool) {
	if s, ok := rg.stores[Columnar].(*CassandraStore); ok {
		return s.Session(), true
	}
	return nil, false
}
