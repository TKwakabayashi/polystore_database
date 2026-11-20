//go:build integration

// list_property_test checks list-valued properties (Person.email / speaks) in every placement.
package db

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/dbs"
	"polystore_database/src/go/tests/dbtest"
)

// TestListProperties places list<...> properties (Person.email / speaks) outside Neo4j and checks that reading
// values back, predicates on lists, grouping by lists and UNWIND match the oracle (Neo4j verbatim). Only columns
// with data in the target store are moved (placeMapping).
//
//	go test -tags integration ./tests/db/ -run TestListProperties -v
func TestListProperties(t *testing.T) {
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

	queries := map[string]string{
		"read":    "MATCH (p:Person) WHERE p.firstName = 'Karan' RETURN p.id, p.email, p.speaks",
		"in":      "MATCH (p:Person) WHERE 'ml' IN p.speaks RETURN p.id, p.email",
		"size":    "MATCH (p:Person) WHERE size(p.email) > 2 RETURN p.id, p.email ORDER BY p.id LIMIT 20",
		"eq":      "MATCH (p:Person) WHERE p.speaks = ['pt', 'en'] RETURN p.id",
		"groupby": "MATCH (p:Person) RETURN p.speaks AS s, count(*) AS n",
		"unwind":  "MATCH (p:Person) UNWIND p.speaks AS l RETURN l, count(*) AS n",
	}
	migs := []migrate.MigrationConfig{{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "email", "speaks"}}}
	for _, target := range []string{"relational", "columnar", "widecolumn", "kvs", "document"} {
		for name, cypher := range queries {
			target, name, cypher := target, name, cypher
			t.Run(target+"/"+name, func(t *testing.T) {
				if k, _ := kind.ParseKind(target); !up[k] {
					t.Skipf("%s is not reachable", target)
				}
				md, moved := placeMapping(t, base, target, migs, func(kind catalog.ObjectType, label, prop string) bool {
					return populated(ctx, reg, target, kind, label, prop)
				})
				if moved == 0 {
					t.Skip("no migrated columns")
				}
				if _, _, err := query.Plan(cypher, nil, md, settings.Plan()); err != nil {
					t.Skipf("cannot be planned (verbatim): %v", err)
				}
				oracle, err := runner.RunNeo4j(ctx, *cfg.Neo4j, cypher, nil)
				if err != nil {
					t.Fatalf("Neo4j (the oracle) failed: %v", err)
				}
				qctx, cancel := context.WithTimeout(ctx, 90*time.Second)
				defer cancel()
				got, err := query.Run(qctx, be, md, cypher, nil, settings.Plan(), settings.Exec(exec.ModelBulk))
				if err != nil {
					t.Fatalf("run failed: %v", err)
				}
				want, have := dbtest.Canon(oracle.Rows), dbtest.Canon(got.Rows)
				if strings.Join(want, "\n") != strings.Join(have, "\n") {
					t.Errorf("rows differ from the oracle (oracle=%d, got=%d, moved=%d)\n%s", len(want), len(have), moved, dbtest.FirstDiff(want, have))
					return
				}
				t.Logf("rows=%d match (moved=%d)", len(have), moved)
			})
		}
	}
}
