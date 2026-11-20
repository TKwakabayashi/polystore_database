//go:build integration

// Smoke test: a few small Person-centric queries on every placement, compared with the oracle. Each
// placement moves only properties that already have data in that store (TestSmokePrepare copies them once),
// so no migration runs per cycle.
//
//	go test -tags integration ./tests/db/ -run 'TestSmoke$' -v
package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
	"polystore_database/src/go/tests/dbtest"
)

var smokeProps = []string{"firstName", "lastName", "birthday", "gender"}

var smokeQueries = []string{
	`MATCH (p:Person {id: $id}) RETURN p.firstName AS f, p.lastName AS l, p.birthday AS b`,
	`MATCH (p:Person {id: $id})-[:KNOWS]-(f:Person) RETURN count(f) AS n, min(f.birthday) AS oldest`,
	`MATCH (p:Person {id: $id})-[:KNOWS]-(f:Person) WHERE f.gender = 'female' RETURN f.firstName AS name ORDER BY name LIMIT 5`,
}

var smokePlacements = []struct {
	target string
	mode   migrate.MigrationMode
}{
	{"relational", migrate.ModeGraphToRdb},
	{"columnar", migrate.ModeGraphToCol},
	{"widecolumn", migrate.ModeGraphToWide},
	{"kvs", migrate.ModeGraphToKvs},
	{"document", migrate.ModeGraphToDoc},
	{"rocksdb", migrate.MigrationMode("graph_to_rocksdb")}, // needs a build with -tags rocksdb
}

func smokeConfig(t *testing.T) conn.Config {
	t.Helper()
	path := os.Getenv("POLYSTORE_CONFIG")
	if path == "" {
		path = "../../config/config.json"
	}
	cfg, err := loadStackConfig(path)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	return cfg
}

// smokePerson picks a person with friends whose friends exist, so every query returns rows.
func smokePerson(t *testing.T, ctx context.Context, reg *conn.Registry) string {
	t.Helper()
	d, _ := reg.Neo4j()
	res, err := neo4j.ExecuteQuery(ctx, d,
		`MATCH (p:Person)-[:KNOWS]-(f:Person) WITH p, count(f) AS n WHERE n >= 5 AND n <= 20 RETURN p.id AS id ORDER BY id LIMIT 1`,
		nil, neo4j.EagerResultTransformer)
	if err != nil || len(res.Records) == 0 {
		t.Fatalf("pick person: %v", err)
	}
	v, _ := res.Records[0].Get("id")
	return fmt.Sprint(v)
}

func TestSmoke(t *testing.T) {
	cfg := smokeConfig(t)
	base, err := catalog.LoadMappingDictionary(cfg.MappingPath())
	if err != nil {
		t.Fatalf("mapping: %v", err)
	}
	ctx := context.Background()
	reg, up := reachableRegistry(t, ctx, cfg)
	defer reg.Close(ctx)
	be := dbs.NewSession(reg)
	params := map[string]string{"id": smokePerson(t, ctx, reg)}

	oracle := make([][]string, len(smokeQueries))
	for i, q := range smokeQueries {
		res, err := runner.RunNeo4j(ctx, *cfg.Neo4j, q, typedForNeo4j(params))
		if err != nil {
			t.Fatalf("oracle q%d: %v", i, err)
		}
		oracle[i] = dbtest.Canon(res.Rows)
	}

	migs := []migrate.MigrationConfig{{ObjType: catalog.ObjEntity, Entity: "Person", Properties: smokeProps}}
	type run struct {
		name string
		md   *catalog.MappingDictionary
	}
	runs := []run{{"graph", base}}
	for _, pc := range smokePlacements {
		if k, _ := kind.ParseKind(pc.target); !up[k] {
			t.Logf("%s unreachable, skipped", pc.target)
			continue
		}
		md, moved := placeMapping(t, base, pc.target, migs, func(kind catalog.ObjectType, label, prop string) bool {
			return populated(ctx, reg, pc.target, kind, label, prop)
		})
		if moved < len(smokeProps) {
			t.Errorf("%s: only %d of %d properties have data; run TestSmokePrepare once", pc.target, moved, len(smokeProps))
			continue
		}
		runs = append(runs, run{pc.target, md})
	}

	var total time.Duration
	perQuery := make([]time.Duration, len(smokeQueries))
	for _, r := range runs {
		for i, q := range smokeQueries {
			start := time.Now()
			got, err := query.Run(ctx, be, r.md, q, params, settings.Plan(), settings.Exec(exec.ModelBulk))
			d := time.Since(start)
			total += d
			perQuery[i] += d
			if err != nil {
				t.Errorf("%s q%d: %v", r.name, i, err)
				continue
			}
			if have := dbtest.Canon(got.Rows); strings.Join(have, "\n") != strings.Join(oracle[i], "\n") {
				t.Errorf("%s q%d: mismatch (oracle=%d rows, got=%d)\n%s", r.name, i, len(oracle[i]), len(have), dbtest.FirstDiff(oracle[i], have))
			}
		}
	}
	t.Logf("%d placements x %d queries in %v (per query: %v)", len(runs), len(smokeQueries), total, perQuery)
}

// TestSmokePrepare copies the smoke properties of Person into every reachable store, keeping
// Neo4j as is, and restores mapping.json afterwards. Run once per instance:
//
//	POLYSTORE_SMOKE_PREPARE=1 go test -tags integration ./tests/db/ -run TestSmokePrepare -v
func TestSmokePrepare(t *testing.T) {
	if os.Getenv("POLYSTORE_SMOKE_PREPARE") == "" {
		t.Skip("set POLYSTORE_SMOKE_PREPARE=1 to copy the smoke properties")
	}
	cfg := smokeConfig(t)
	snapshot, err := os.ReadFile(cfg.MappingPath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.WriteFile(cfg.MappingPath(), snapshot, 0o644); err != nil {
			t.Errorf("restore mapping: %v", err)
		}
	}()
	ctx := context.Background()
	reg, up := reachableRegistry(t, ctx, cfg)
	need := map[string]bool{}
	for _, pc := range smokePlacements {
		if k, _ := kind.ParseKind(pc.target); !up[k] {
			continue
		}
		for _, p := range smokeProps {
			if !populated(ctx, reg, pc.target, catalog.ObjEntity, "Person", p) {
				need[pc.target] = true
			}
		}
	}
	_ = reg.Close(ctx) // a migration opens its own connections (LevelDB allows one process handle)
	for _, pc := range smokePlacements {
		if k, _ := kind.ParseKind(pc.target); !up[k] {
			t.Logf("%s unreachable, skipped", pc.target)
			continue
		}
		if !need[pc.target] {
			t.Logf("%s already has the smoke properties", pc.target)
			continue
		}
		// Each copy starts from the original mapping: a migration rewrites it to name the target.
		if err := os.WriteFile(cfg.MappingPath(), snapshot, 0o644); err != nil {
			t.Fatal(err)
		}
		migs := []migrate.MigrationConfig{{ObjType: catalog.ObjEntity, Entity: "Person", Properties: smokeProps, Mode: pc.mode, DeleteSource: false}}
		res, err := migrate.Run(ctx, cfg, migs)
		if err != nil {
			t.Fatalf("%s: %v", pc.target, err)
		}
		t.Logf("%s: %+v", pc.target, res)
	}
}
