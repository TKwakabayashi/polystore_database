package workload_test

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"polystore_database/src/go/internal/testopts"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/logical"
	"polystore_database/src/go/query/physical"
)

var update = flag.Bool("update", false, "regenerate the golden files")

const datasetsDir = "../../../../datastore/datasets"

// datasetDir is the dataset folder of a workload schema: the folder of that name, else its first scale
// (ldbc_snb_datagen -> ldbc_snb_datagen_sf1).
func datasetDir(schema string) string {
	if _, err := os.Stat(filepath.Join(datasetsDir, schema)); err == nil {
		return filepath.Join(datasetsDir, schema)
	}
	return filepath.Join(datasetsDir, schema+"_sf1")
}

func loadMapping(t *testing.T, ds string) *catalog.MappingDictionary {
	t.Helper()
	md, err := catalog.LoadMappingDictionary(filepath.Join(datasetDir(ds), "mapping.json"))
	if err != nil {
		t.Fatalf("%s: %v", ds, err)
	}
	return md
}

// place returns a mapping with migs' columns moved to target (the original is unchanged).
func place(t *testing.T, base *catalog.MappingDictionary, target string, migs []migrate.MigrationConfig) *catalog.MappingDictionary {
	t.Helper()
	b, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var md catalog.MappingDictionary
	if err := json.Unmarshal(b, &md); err != nil {
		t.Fatal(err)
	}
	for _, m := range migs {
		for _, p := range m.Properties {
			var props map[string]*catalog.PropertyMapping
			if m.ObjType == catalog.ObjRelationship {
				props = md.Relationships[m.Entity].Properties
			} else {
				props = md.Entities[m.Entity].Properties
			}
			props[p].Databases = []string{target}
		}
	}
	return &md
}

var placements = []struct {
	name, target string
	mode         migrate.MigrationMode
}{
	{"graph", "", migrate.ModeGraphToRdb},
	{"relational", "relational", migrate.ModeGraphToRdb},
	{"document", "document", migrate.ModeGraphToDoc},
	{"columnar", "columnar", migrate.ModeGraphToCol},
	{"widecolumn", "widecolumn", migrate.ModeGraphToWide},
	{"kvs", "kvs", migrate.ModeGraphToKvs},
}

// Names do not overlap between datasets, and Lookup finds each in its dataset.
func TestDatasetWorkloadNamesUnique(t *testing.T) {
	seen := map[string]string{}
	for _, ds := range workload.DatasetNames() {
		for _, n := range workload.Names(ds) {
			if other, ok := seen[n]; ok {
				t.Errorf("%s is in both %s and %s", n, other, ds)
			}
			seen[n] = ds
			if _, got, ok := workload.Lookup(n); !ok || got != ds {
				t.Errorf("Lookup(%s) = %q, %v (want %q)", n, got, ok, ds)
			}
		}
	}
}

// experimentDatasets are the placement-experiment datasets: every dataset but ldbc_snb, whose goldens and
// checks are the query packages'.
func experimentDatasets() []string {
	var out []string
	for _, ds := range workload.DatasetNames() {
		if ds != workload.LDBC {
			out = append(out, ds)
		}
	}
	return out
}

// Migration columns exist in the dataset's mapping (catches misspellings).
func TestDatasetMigrationsExist(t *testing.T) {
	for _, ds := range experimentDatasets() {
		md := loadMapping(t, ds)
		for _, n := range workload.Names(ds) {
			_, _, migs := workload.Datasets[ds][n].Define(migrate.ModeGraphToRdb, true)
			if len(migs) == 0 {
				t.Errorf("%s/%s: no migration settings", ds, n)
			}
			for _, m := range migs {
				var props map[string]*catalog.PropertyMapping
				if m.ObjType == catalog.ObjRelationship {
					r, ok := md.Relationships[m.Entity]
					if !ok {
						t.Errorf("%s/%s: relationship %s is not in the mapping", ds, n, m.Entity)
						continue
					}
					props = r.Properties
				} else {
					e, ok := md.Entities[m.Entity]
					if !ok {
						t.Errorf("%s/%s: entity %s is not in the mapping", ds, n, m.Entity)
						continue
					}
					props = e.Properties
				}
				for _, p := range m.Properties {
					if _, ok := props[p]; !ok {
						t.Errorf("%s/%s: %s.%s is not in the mapping", ds, n, m.Entity, p)
					}
				}
			}
			if _, _, migs := workload.Datasets[ds][n].Define(migrate.ModeGraphToRdb, false); migs != nil {
				t.Errorf("%s/%s: returned migration settings with isMigration=false", ds, n)
			}
		}
	}
}

// Per dataset, the physical plans of every workload x 5 placements, pinned by testdata/<dataset>.golden.
// Workloads that cannot be planned are errors (they are useless for placement experiments).
func TestDatasetWorkloadPlans(t *testing.T) {
	for _, ds := range experimentDatasets() {
		t.Run(ds, func(t *testing.T) {
			base := loadMapping(t, ds)
			var b strings.Builder
			for _, n := range workload.Names(ds) {
				b.WriteString("## " + n + " — " + workload.Datasets[ds][n].Traits.Purpose + "\n")
				b.WriteString("traits: " + workload.Datasets[ds][n].Traits.String() + "\n\n")
				for _, pc := range placements {
					stmt, params, migs := workload.Datasets[ds][n].Define(pc.mode, pc.target != "")
					md := base
					if pc.target != "" {
						md = place(t, base, pc.target, migs)
					}
					a, err := ast.Parse(stmt, params)
					if errors.Is(err, ast.ErrUnsupported) || errors.Is(err, ast.ErrSyntax) {
						t.Errorf("%s: cannot be planned (verbatim): %v", n, err)
						break
					}
					if err != nil {
						t.Fatalf("%s: %v", n, err)
					}
					l, err := cypher.Translate(a)
					if err != nil {
						t.Fatalf("%s: %v", n, err)
					}
					p, err := physical.Build(logical.Rewrite(logical.PlanPatterns(l, nil, lopt), lopt), physical.NewPlacement(md), popt)
					if err != nil {
						t.Fatalf("%s @%s: %v", n, pc.name, err)
					}
					b.WriteString("### " + n + " @" + pc.name + "\n" + physical.Dump(p) + "\n")
				}
			}
			golden(t, ds+".golden", b.String())
		})
	}
}

// Traits agree with the Cypher: role columns appear in it, and variable length, aggregation and LIMIT match it.
func TestDatasetTraits(t *testing.T) {
	limitRe := regexp.MustCompile(`LIMIT (\d+)`)
	aggRe := regexp.MustCompile(`\b(count|sum|avg|min|max)\(`)
	varLenRe := regexp.MustCompile(`\[[^\]]*\*[^\]]*\]`)
	for _, ds := range experimentDatasets() {
		for _, n := range workload.Names(ds) {
			w := workload.Datasets[ds][n]
			tr := w.Traits
			cypher, _, _ := w.Define(migrate.ModeGraphToRdb, false)
			if tr.Purpose == "" || tr.Hops == "" || len(tr.Roles) == 0 {
				t.Errorf("%s/%s: traits (purpose, hops, roles) are empty", ds, n)
			}
			for r, cols := range tr.Roles {
				for _, c := range cols {
					if !mentions(cypher, c) {
						t.Errorf("%s/%s: column %[4]s of %[3]s is not in the Cypher", ds, n, r, c)
					}
				}
			}
			if got := varLenRe.MatchString(cypher); got != tr.VarLength {
				t.Errorf("%s/%s: VarLength=%v but the Cypher has variable length %v", ds, n, tr.VarLength, got)
			}
			if got := aggRe.MatchString(cypher); got != tr.Aggregate {
				t.Errorf("%s/%s: Aggregate=%v but the Cypher aggregates %v", ds, n, tr.Aggregate, got)
			}
			k := 0
			if m := limitRe.FindStringSubmatch(cypher); m != nil {
				k, _ = strconv.Atoi(m[1])
			}
			if k != tr.TopK {
				t.Errorf("%s/%s: TopK=%d but the Cypher has LIMIT %d", ds, n, tr.TopK, k)
			}
		}
	}
}

// mentions reports whether column "v.p" appears in the Cypher (as v.p or in a pattern (v:L {p: ...})).
func mentions(cypher, col string) bool {
	if strings.Contains(cypher, col) {
		return true
	}
	v, p, ok := strings.Cut(col, ".")
	if !ok {
		return false
	}
	re := regexp.MustCompile(`\(` + regexp.QuoteMeta(v) + `(:\w+)?\s*\{[^}]*\b` + regexp.QuoteMeta(p) + `\s*:`)
	return re.MatchString(cypher)
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s is missing (make it with -update): %v", path, err)
	}
	if string(want) != got {
		t.Errorf("golden %s differs\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

// lopt and popt are the planner options the goldens are pinned to; tests may toggle them.
var lopt, popt = testopts.Logical(), testopts.Physical()
