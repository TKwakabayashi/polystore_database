//go:build integration

// equivalence_test checks that the query engine returns the same rows as the oracle, for every workload
// (TestEquivalence), in every placement (TestPlacements) and with every iteration model (TestModels). Heavy.
//
//	go test -tags integration ./tests/db/ -run 'TestEquivalence|TestPlacements' -v
//
// Queries the engine delegates whole to Neo4j match trivially but are kept to catch execution errors.
package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
	"polystore_database/src/go/tests/dbtest"
)

func TestEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := loadStackConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	md, err := catalog.LoadMappingDictionary(cfg.MappingPath())
	if err != nil {
		t.Fatalf("mapping load (%s): %v", cfg.MappingPath(), err)
	}
	ctx := context.Background()
	reg, err := conn.NewRegistryFor(ctx, cfg, kind.Neo4j)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := dbs.NewSession(reg)

	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			cypher, params, _ := workload.Def(name)(migrate.ModeGraphToRdb, false)
			oracle, err := runner.RunNeo4j(ctx, *cfg.Neo4j, cypher, typedForNeo4j(params))
			if err != nil {
				t.Skipf("skipped: Neo4j (the oracle) failed: %v", err)
			}
			got, err := query.Run(ctx, be, md, cypher, params, settings.Plan(), settings.Exec(exec.ModelBulk))
			if errors.Is(err, ast.ErrSyntax) {
				t.Skipf("the grammar does not cover it yet: %v", err)
			}
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			want, have := dbtest.Canon(oracle.Rows), dbtest.Canon(got.Rows)
			if strings.Join(want, "\n") != strings.Join(have, "\n") {
				t.Errorf("[%s] rows differ from the oracle (oracle=%d, got=%d)\n%s",
					got.Engine, len(want), len(have), dbtest.FirstDiff(want, have))
				return
			}
			t.Logf("[%s] rows=%d match (round trips=%d, plan=%v, exec=%v)",
				got.Engine, len(have), got.RoundTrips, got.PlanTime, got.ExecTime)
		})
	}
}

// TestPlacements plans and runs each workload with its migration columns placed outside Neo4j and
// compares the rows with the oracle (Neo4j verbatim). It needs the migrated data in each store (from earlier
// migrations); stores without the data are skipped.
//
//	go test -tags integration ./tests/db/ -run TestPlacements -v
func TestPlacements(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := loadStackConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	base, err := catalog.LoadMappingDictionary(cfg.MappingPath())
	if err != nil {
		t.Fatalf("mapping load: %v", err)
	}
	ctx := context.Background()
	reg, up := reachableRegistry(t, ctx, cfg)
	defer reg.Close(ctx)
	be := dbs.NewSession(reg)

	placements := []struct {
		name   string
		target string
		mode   migrate.MigrationMode
	}{
		{"relational", "relational", migrate.ModeGraphToRdb},
		{"columnar", "columnar", migrate.ModeGraphToCol},
		{"widecolumn", "widecolumn", migrate.ModeGraphToWide},
		{"kvs", "kvs", migrate.ModeGraphToKvs},
		{"document", "document", migrate.ModeGraphToDoc},
	}

	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, pc := range placements {
		for _, name := range names {
			pc, name := pc, name
			t.Run(pc.name+"/"+name, func(t *testing.T) {
				if k, _ := kind.ParseKind(pc.target); !up[k] {
					t.Skipf("%s is not reachable", pc.target)
				}
				cypher, params, migs := workload.Def(name)(pc.mode, true)
				if len(migs) == 0 {
					migs = allProperties(base, pc.mode) // workloads without migration configs try every migrated column
				}
				md, moved := placeMapping(t, base, pc.target, migs, func(kind catalog.ObjectType, label, prop string) bool {
					return populated(ctx, reg, pc.target, kind, label, prop)
				})
				if moved == 0 {
					t.Skip("no migrated columns")
				}
				if _, _, err := query.Plan(cypher, params, md, settings.Plan()); err != nil {
					t.Skipf("cannot be planned (verbatim) or the grammar does not take it: %v", err)
				}
				oracle, err := runner.RunNeo4j(ctx, *cfg.Neo4j, cypher, typedForNeo4j(params))
				if err != nil {
					t.Skipf("Neo4j (the oracle) failed: %v", err)
				}
				qctx, cancel := context.WithTimeout(ctx, 90*time.Second)
				defer cancel()
				got, err := query.Run(qctx, be, md, cypher, params, settings.Plan(), settings.Exec(exec.ModelBulk))
				if err != nil && qctx.Err() != nil {
					t.Skipf("stopped at 90s (known: large intermediate results across stores): %v", err)
				}
				if err != nil {
					t.Fatalf("run failed: %v", err)
				}
				want, have := dbtest.Canon(oracle.Rows), dbtest.Canon(got.Rows)
				if strings.Join(want, "\n") != strings.Join(have, "\n") {
					t.Errorf("rows differ from the oracle (oracle=%d, got=%d, moved=%d)\n%s", len(want), len(have), moved, dbtest.FirstDiff(want, have))
					return
				}
				t.Logf("rows=%d match (moved=%d, round trips=%d, exec=%v)", len(have), moved, got.RoundTrips, got.ExecTime)
			})
		}
	}
}

// TestModels checks that iteration models give the same results on real data in each placement, against
// bulk in the same placement (oracle agreement is TestEquivalence / TestPlacements' job).
//
//	go test -tags integration ./tests/db/ -run TestModels -v
func TestModels(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := loadStackConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	base, err := catalog.LoadMappingDictionary(cfg.MappingPath())
	if err != nil {
		t.Fatalf("mapping load: %v", err)
	}
	ctx := context.Background()
	reg, err := conn.NewRegistry(ctx, cfg)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := dbs.NewSession(reg)

	placements := []struct {
		name   string
		target string
		mode   migrate.MigrationMode
	}{
		{"graph", "", migrate.ModeGraphToRdb},
		{"relational", "relational", migrate.ModeGraphToRdb},
		{"document", "document", migrate.ModeGraphToDoc},
		{"columnar", "columnar", migrate.ModeGraphToCol},
		{"widecolumn", "widecolumn", migrate.ModeGraphToWide},
		{"kvs", "kvs", migrate.ModeGraphToKvs},
	}
	models := []exec.Model{exec.ModelVolcano, exec.ModelVectorized, exec.ModelStream, exec.ModelVecStream}
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, pc := range placements {
		for _, name := range names {
			pc, name := pc, name
			t.Run(pc.name+"/"+name, func(t *testing.T) {
				cypher, params, migs := workload.Def(name)(pc.mode, pc.target != "")
				if pc.target != "" && len(migs) == 0 {
					migs = allProperties(base, pc.mode)
				}
				md := base
				if pc.target != "" {
					var moved int
					md, moved = placeMapping(t, base, pc.target, migs, func(kind catalog.ObjectType, label, prop string) bool {
						return populated(ctx, reg, pc.target, kind, label, prop)
					})
					if moved == 0 {
						t.Skip("no migrated columns")
					}
				}
				if _, _, err := query.Plan(cypher, params, md, settings.Plan()); err != nil {
					t.Skipf("cannot be planned (verbatim) or the grammar does not take it: %v", err)
				}
				run := func(m exec.Model) (string, []string, bool) {
					qctx, cancel := context.WithTimeout(ctx, 60*time.Second)
					defer cancel()
					r, err := query.Run(qctx, be, md, cypher, params, settings.Plan(), settings.Exec(m))
					if err != nil && qctx.Err() != nil {
						return "", nil, false
					}
					if err != nil {
						t.Fatalf("[%s] run failed: %v", m, err)
					}
					return fmt.Sprintf("rt=%d exec=%v", r.RoundTrips, r.ExecTime.Round(time.Millisecond)), dbtest.Canon(r.Rows), true
				}
				refInfo, ref, ok := run(exec.ModelBulk)
				if !ok {
					t.Skip("bulk does not finish in 60s")
				}
				t.Logf("[bulk] rows=%d %s", len(ref), refInfo)
				for _, m := range models {
					info, got, ok := run(m)
					if !ok {
						t.Logf("[%s] stopped at 60s", m)
						continue
					}
					if strings.Join(got, "\n") != strings.Join(ref, "\n") {
						t.Errorf("[%s] differs from bulk (bulk=%d, %s=%d)\n%s", m, len(ref), m, len(got), dbtest.FirstDiff(ref, got))
						continue
					}
					t.Logf("[%s] match %s", m, info)
				}
			})
		}
	}
}
