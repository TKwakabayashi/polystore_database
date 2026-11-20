//go:build integration

// uuid_format_test checks that switching the uuid storage form (id.CurrentStorage(): text / 16 bytes) gives
// the same results in every store:
//
//	load a synthetic graph into Neo4j (text) -> migrate to each store and query -> rewrite to binary (as
//	uuid-format) -> clear the other stores, migrate again and query -> rewrite back to text and query
//
// It runs only with a dedicated config (POLYSTORE_UUID_E2E_CONFIG) whose MongoDB DB name, MySQL DSN,
// ClickHouse DB name, Cassandra keyspace and datastore directory all contain "uuidtest", and only when Neo4j has fewer
// than 100k nodes. A throwaway Neo4j on the same VM as the real stack competes for memory: give it a small
// heap and stop it afterwards.
//
//	POLYSTORE_UUID_E2E_CONFIG=/path/to/uuidtest.json go test -tags integration ./tests/throwaway/ -run TestUUIDFormat -v
package throwaway

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/data/setup"
	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
	"polystore_database/src/go/tests/dbtest"
)

func TestUUIDFormat(t *testing.T) {
	path := os.Getenv("POLYSTORE_UUID_E2E_CONFIG")
	if path == "" {
		t.Skip("POLYSTORE_UUID_E2E_CONFIG (a throwaway environment's config) is not set")
	}
	cfg, err := conn.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{"mongo.dbname": cfg.Mongo.DBName, "mysql.dsn": cfg.MySQL.DSN,
		"cassandra.keyspace": cfg.Cassandra.Keyspace, "datastore": cfg.KVPath(kind.LevelDB)}
	if cfg.ClickHouse != nil {
		names["clickhouse.database"] = cfg.ClickHouse.Database
	}
	for name, v := range names {
		if !strings.Contains(v, "uuidtest") {
			t.Fatalf("%s=%q does not contain uuidtest (stopped to keep real data)", name, v)
		}
	}
	ctx := context.Background()
	prev := id.CurrentStorage()
	defer func() { id.SetStorage(prev) }()

	// The throwaway datastore's runtime/ starts from the ldbc_snb schema template every time (migrations
	// rewrite it).
	original := "../../datastore/datasets/ldbc_snb_sf1/mapping.json"
	if err := os.MkdirAll(cfg.Layout().RuntimeDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	resetMapping := func() {
		t.Helper()
		if err := copyFile(original, cfg.MappingPath()); err != nil {
			t.Fatal(err)
		}
	}
	resetMapping()

	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close(ctx)
	cypher := func(q string, p map[string]any) *neo4j.EagerResult {
		t.Helper()
		r, err := neo4j.ExecuteQuery(ctx, driver, q, p, neo4j.EagerResultTransformer)
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		return r
	}
	if n := cypher("MATCH (n) RETURN count(n) AS c", nil).Records[0].Values[0].(int64); n >= 100000 {
		t.Fatalf("Neo4j has %d nodes (not a throwaway environment)", n)
	}

	id.SetStorage(id.StorageText)
	seedUUIDGraph(t, cypher)

	for _, phase := range []struct {
		name   string
		format id.Storage
	}{{"text", id.StorageText}, {"binary", id.StorageBinary}, {"text-again", id.StorageText}} {
		t.Run(phase.name, func(t *testing.T) {
			id.SetStorage(phase.format)
			if err := setup.ConvertUUIDFormat(ctx, cfg); err != nil {
				t.Fatal(err)
			}
			checkNeo4jFormat(t, cypher, phase.format)
			if err := setup.RunDataset(ctx, cfg, ""); err != nil { // clear the four non-Neo4j stores (to remake column types; no reshaping outside compose)
				t.Fatal(err)
			}
			for _, pl := range []struct {
				name string
				mode migrate.MigrationMode
			}{{"graph", ""}, {"relational", migrate.ModeGraphToRdb}, {"columnar", migrate.ModeGraphToCol},
				{"widecolumn", migrate.ModeGraphToWide}, {"kvs", migrate.ModeGraphToKvs}, {"document", migrate.ModeGraphToDoc}} {
				resetMapping()
				for _, m := range uuidE2EMigs {
					if pl.mode == "" {
						break
					}
					m.Mode, m.MappingPath, m.MongoDbName = pl.mode, cfg.MappingPath(), cfg.Mongo.DBName
					if _, err := migrate.MigrateData(m, cfg); err != nil {
						t.Fatalf("%s: migrate %s: %v", pl.name, m.Entity, err)
					}
				}
				md, err := catalog.LoadMappingDictionary(cfg.MappingPath())
				if err != nil {
					t.Fatal(err)
				}
				reg, err := conn.NewRegistry(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				be := dbs.NewSession(reg)
				for qi, q := range uuidE2EQueries {
					oracle, err := runner.RunNeo4j(ctx, *cfg.Neo4j, q, map[string]any{})
					if err != nil {
						t.Fatalf("oracle %d: %v", qi, err)
					}
					want := dbtest.Canon(oracle.Rows)
					for _, model := range []exec.Model{exec.ModelBulk, exec.ModelStream, exec.ModelVectorized} {
						got, err := query.Run(ctx, be, md, q, nil, settings.Plan(), settings.Exec(model))
						if err != nil {
							t.Errorf("%s/%s q%d: %v", pl.name, model, qi, err)
							continue
						}
						if have := dbtest.Canon(got.Rows); strings.Join(have, "\n") != strings.Join(want, "\n") {
							t.Errorf("%s/%s q%d %q: differs from the oracle\n%s", pl.name, model, qi, q, dbtest.FirstDiff(want, have))
						}
					}
				}
				checkStoredFormat(t, ctx, reg, pl.name, phase.format)
				reg.Close(ctx)
				t.Logf("%s: %d queries x 3 models match", pl.name, len(uuidE2EQueries))
			}
		})
	}
}

// checkNeo4jFormat checks that every Neo4j uuid has format's type (string / bytes) and parses.
func checkNeo4jFormat(t *testing.T, cypher func(string, map[string]any) *neo4j.EagerResult, format id.Storage) {
	t.Helper()
	for _, q := range []string{"MATCH (n) RETURN n.uuid AS u", "MATCH ()-[r]->() RETURN r.uuid AS u"} {
		for _, rec := range cypher(q, nil).Records {
			v := rec.Values[0]
			_, isBytes := v.([]byte)
			if isBytes != (format == id.StorageBinary) || id.FromAny(v).Empty() {
				t.Fatalf("%s: a value in another form %T %v", q, v, v)
			}
		}
	}
}

// checkStoredFormat checks that Person uuids written to the destination store use format's physical form.
func checkStoredFormat(t *testing.T, ctx context.Context, reg *conn.Registry, store string, format id.Storage) {
	t.Helper()
	binary := format == id.StorageBinary
	switch store {
	case "relational":
		db, _ := reg.MySQL()
		var typ string
		if err := db.QueryRowContext(ctx, "SELECT DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'Person' AND COLUMN_NAME = 'uuid'").Scan(&typ); err != nil {
			t.Fatal(err)
		}
		if (typ == "binary") != binary {
			t.Errorf("relational: uuid column type %s", typ)
		}
	case "columnar":
		// ClickHouse uuid columns are String in both forms; the value length tells the form.
		db, _ := reg.ClickHouse()
		var n uint64
		if err := db.QueryRowContext(ctx, "SELECT length(uuid) FROM `Person` LIMIT 1").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if (n == uint64(id.Size)) != binary {
			t.Errorf("columnar: uuid length %d", n)
		}
	case "widecolumn":
		sess, _ := reg.Cassandra()
		iter := sess.Query(`SELECT uuid FROM "Person" LIMIT 1`).WithContext(ctx).Iter()
		cols := iter.Columns()
		_ = iter.Close()
		if len(cols) == 0 || (cols[0].TypeInfo.Type() == gocql.TypeBlob) != binary {
			t.Errorf("widecolumn: uuid column type %v", cols)
		}
	case "document":
		db, _ := reg.Mongo()
		var doc bson.M
		if err := db.Collection("Person").FindOne(ctx, bson.M{}).Decode(&doc); err != nil {
			t.Fatal(err)
		}
		if _, isBin := doc["uuid"].(primitive.Binary); isBin != binary {
			t.Errorf("document: uuid value %T", doc["uuid"])
		}
	case "kvs":
		kv, _ := reg.KV(kind.LevelDB)
		var first []byte
		stop := fmt.Errorf("stop")
		if err := conn.ScanPrefix(kv, []byte("Person\x00"), func(key, _ []byte) error {
			first = append([]byte{}, key...)
			return stop
		}); err != nil && err != stop {
			t.Fatal(err)
		}
		if first == nil {
			t.Fatal("kvs: no Person keys")
		}
		// ParseEntityKey locates the uuid by the storage form's length, so a key in the other form gives ok=false.
		if _, u, _, ok := codec.ParseEntityKey(first); !ok || u.Empty() {
			t.Errorf("kvs: key %q is not in the storage form", first)
		}
	}
}
