package core

import (
	"context"
	"database/sql"

	"polystore_database/src/go/storage"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

// Deps は3エンジン（stream / bulk / volcano）が共有するデータストア接続一式。
// storage.Registry から5ハンドルを展開する重複ボイラープレート（旧来4か所に散在）を
// Open に1本化する。未設定ストアのハンドルは nil のまま（各エンジンの nil チェック挙動を踏襲）。
type Deps struct {
	Ctx       context.Context
	Cfg       storage.Config
	Registry  *storage.Registry
	Neo       neo4j.DriverWithContext
	Mongo     *mongo.Database
	LevelDB   *leveldb.DB
	MySQL     *sql.DB
	Cassandra *gocql.Session
}

// Open は Registry を開き、設定済みの5ハンドルを展開した Deps を返す。
func Open(ctx context.Context, cfg storage.Config) (*Deps, error) {
	rg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, err
	}
	d := &Deps{Ctx: ctx, Cfg: cfg, Registry: rg}
	if h, ok := rg.Neo4j(); ok {
		d.Neo = h
	}
	if h, ok := rg.Mongo(); ok {
		d.Mongo = h
	}
	if h, ok := rg.LevelDB(); ok {
		d.LevelDB = h
	}
	if h, ok := rg.MySQL(); ok {
		d.MySQL = h
	}
	if h, ok := rg.Cassandra(); ok {
		d.Cassandra = h
	}
	return d, nil
}

// Close は配下の Registry を閉じる。
func (d *Deps) Close() error {
	if d.Registry != nil {
		return d.Registry.Close(d.Ctx)
	}
	return nil
}
