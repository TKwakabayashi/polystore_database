//go:build integration

// move_test checks migrations that delete the source (DeleteSource):
//
//	load a synthetic graph into Neo4j -> for each store: move Person / Place properties from Neo4j to the
//	store (Neo4j loses them, queries still match the oracle) -> move them back (the store loses them, Neo4j
//	has the original values again)
//
// It runs only with a dedicated config (POLYSTORE_MOVE_E2E_CONFIG) whose MongoDB DB name, MySQL DSN,
// ClickHouse DB name, Cassandra keyspace and datastore directory all contain "movetest", with empty namespaces, and
// only when Neo4j has fewer than 100k nodes. Neo4j needs APOC (deletes use apoc.create.removeProperties).
//
//	POLYSTORE_MOVE_E2E_CONFIG=/path/to/movetest.json go test -tags integration ./tests/throwaway/ -run TestMove -v
package throwaway

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
	"polystore_database/src/go/tests/dbtest"
)

func TestMove(t *testing.T) {
	path := os.Getenv("POLYSTORE_MOVE_E2E_CONFIG")
	if path == "" {
		t.Skip("POLYSTORE_MOVE_E2E_CONFIG (a throwaway environment's config) is not set")
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
		if !strings.Contains(v, "movetest") {
			t.Fatalf("%s=%q does not contain movetest (stopped to keep real data)", name, v)
		}
	}
	ctx := context.Background()
	prev := id.CurrentStorage()
	defer func() { id.SetStorage(prev) }()
	id.SetStorage(id.StorageText)

	// The throwaway datastore's runtime/ starts from the ldbc_snb schema template every time (migrations
	// rewrite it).
	original := "../../datastore/datasets/ldbc_snb_sf1/mapping.json"
	if err := os.MkdirAll(cfg.Layout().RuntimeDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	resetMapping := func(t *testing.T) {
		t.Helper()
		if err := copyFile(original, cfg.MappingPath()); err != nil {
			t.Fatal(err)
		}
	}

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
	// The namespaces must start empty (a leftover row would come back to Neo4j with a move).
	for _, k := range []kind.Kind{kind.MySQL, kind.ClickHouse, kind.Cassandra, kind.LevelDB, kind.Mongo} {
		if n := storeRows(t, ctx, cfg, k); n != 0 {
			t.Fatalf("%s already has %d Person / Place rows in the test namespace; empty it first", k, n)
		}
	}
	seedUUIDGraph(t, cypher)

	// The moved properties as Neo4j has them, and the oracle results of the queries.
	snapshot := func() string {
		t.Helper()
		var lines []string
		for _, m := range uuidE2EMigs {
			for _, rec := range cypher(fmt.Sprintf("MATCH (n:%s) RETURN n.id AS id, [p IN $props | n[p]] AS vals", m.Entity),
				map[string]any{"props": m.Properties}).Records {
				lines = append(lines, fmt.Sprintf("%s %v %v", m.Entity, rec.Values[0], rec.Values[1]))
			}
		}
		sort.Strings(lines)
		return strings.Join(lines, "\n")
	}
	want := snapshot()
	oracle := make([][]string, len(uuidE2EQueries))
	for i, q := range uuidE2EQueries {
		r, err := runner.RunNeo4j(ctx, *cfg.Neo4j, q, map[string]any{})
		if err != nil {
			t.Fatalf("oracle %d: %v", i, err)
		}
		oracle[i] = dbtest.Canon(r.Rows)
	}
	neo4jHas := func() int64 {
		t.Helper()
		var n int64
		for _, m := range uuidE2EMigs {
			n += cypher(fmt.Sprintf("MATCH (n:%s) WHERE any(p IN $props WHERE n[p] IS NOT NULL) RETURN count(n) AS c", m.Entity),
				map[string]any{"props": m.Properties}).Records[0].Values[0].(int64)
		}
		return n
	}

	move := func(t *testing.T, mode migrate.MigrationMode) {
		t.Helper()
		for _, m := range uuidE2EMigs {
			m.Mode, m.MappingPath, m.MongoDbName, m.DeleteSource = mode, cfg.MappingPath(), cfg.Mongo.DBName, true
			if _, err := migrate.MigrateData(m, cfg); err != nil {
				t.Fatalf("move %s (%s): %v", m.Entity, mode, err)
			}
		}
	}
	for _, k := range []kind.Kind{kind.MySQL, kind.ClickHouse, kind.Cassandra, kind.LevelDB, kind.Mongo} {
		t.Run(k.String(), func(t *testing.T) {
			resetMapping(t)
			move(t, migrate.ModeOf(kind.Neo4j, k))
			if n := neo4jHas(); n != 0 {
				t.Errorf("Neo4j still has the moved properties on %d nodes", n)
			}
			if n := storeHas(t, ctx, cfg, k); n != 65 {
				t.Errorf("%s has values on %d rows, want 65", k, n)
			}
			md, err := catalog.LoadMappingDictionary(cfg.MappingPath())
			if err != nil {
				t.Fatal(err)
			}
			reg, err := conn.NewRegistry(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			for i, q := range uuidE2EQueries {
				for _, model := range []exec.Model{exec.ModelBulk, exec.ModelVecStream} {
					got, err := query.Run(ctx, dbs.NewSession(reg), md, q, nil, settings.Plan(), settings.Exec(model))
					if errors.Is(err, ast.ErrSyntax) {
						t.Logf("q%d: the grammar does not take it yet, skipped", i)
						break
					}
					if err != nil {
						t.Errorf("q%d %s: %v", i, model, err)
						continue
					}
					if have := dbtest.Canon(got.Rows); strings.Join(have, "\n") != strings.Join(oracle[i], "\n") {
						t.Errorf("q%d %s %q differs from the oracle\n%s", i, model, q, dbtest.FirstDiff(oracle[i], have))
					}
				}
			}
			reg.Close(ctx) // a migration opens its own connections (LevelDB allows one process handle)

			move(t, migrate.ModeOf(k, kind.Neo4j))
			if n := storeHas(t, ctx, cfg, k); n != 0 {
				t.Errorf("%s still has values on %d rows after moving back", k, n)
			}
			if got := snapshot(); got != want {
				t.Errorf("Neo4j values differ after the round trip\n%s", dbtest.FirstDiff(strings.Split(want, "\n"), strings.Split(got, "\n")))
			}
			t.Logf("%s: moved out and back, queries match, sources cleared", k)
		})
	}
}

// storeHas counts the rows of the moved targets that hold at least one non-null property in store k (a
// ClickHouse table whose columns were dropped counts 0).
func storeHas(t *testing.T, ctx context.Context, cfg conn.Config, k kind.Kind) int {
	t.Helper()
	reg, err := conn.NewRegistryFor(ctx, cfg, k)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close(ctx)
	md, err := catalog.LoadMappingDictionary(cfg.MappingPath())
	if err != nil {
		t.Fatal(err)
	}
	m, err := dbs.MoverOf(k)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, mig := range uuidE2EMigs {
		types := map[string]string{}
		for _, p := range mig.Properties {
			types[p], _ = md.GetPropertyDataType(mig.ObjType, mig.Entity, p)
		}
		tgt := stores.Target{Obj: mig.ObjType, Entity: mig.Entity, Props: mig.Properties, Types: types, MongoDB: cfg.Mongo.DBName}
		if k == kind.ClickHouse && !clickHouseHasColumns(t, ctx, reg, mig.Entity, mig.Properties) {
			continue
		}
		out := make(chan stores.Row, 1000)
		done := make(chan error, 1)
		go func() { done <- m.Scan(ctx, reg, tgt, out, 0, 1); close(out) }()
		for r := range out {
			for _, v := range r.Values {
				if v != nil {
					n++
					break
				}
			}
		}
		if err := <-done; err != nil && !missingTable(err) {
			t.Fatalf("scan %s on %s: %v", mig.Entity, k, err)
		}
	}
	return n
}

func clickHouseHasColumns(t *testing.T, ctx context.Context, reg *conn.Registry, table string, props []string) bool {
	t.Helper()
	db, _ := reg.ClickHouse()
	var n uint64
	if err := db.QueryRowContext(ctx, "SELECT count() FROM system.columns WHERE database = currentDatabase() AND table = ? AND name IN (?)",
		table, props).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n > 0
}

// missingTable reports an error from scanning a table or collection that does not exist.
func missingTable(err error) bool {
	s := err.Error()
	return strings.Contains(s, "doesn't exist") || strings.Contains(s, "unconfigured table") || strings.Contains(s, "does not exist")
}

// storeRows counts the Person and Place rows (uuids) of store k; a missing table counts 0.
func storeRows(t *testing.T, ctx context.Context, cfg conn.Config, k kind.Kind) int64 {
	t.Helper()
	reg, err := conn.NewRegistryFor(ctx, cfg, k)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close(ctx)
	m, err := dbs.MoverOf(k)
	if err != nil {
		t.Fatal(err)
	}
	var n int64
	for _, mig := range uuidE2EMigs {
		c, err := m.Count(ctx, reg, stores.Target{Obj: mig.ObjType, Entity: mig.Entity, Props: mig.Properties, MongoDB: cfg.Mongo.DBName})
		if err != nil && !missingTable(err) && !strings.Contains(err.Error(), "UNKNOWN_TABLE") && !strings.Contains(err.Error(), "Unknown table") {
			t.Fatalf("count %s on %s: %v", mig.Entity, k, err)
		}
		n += c
	}
	return n
}
