package physical_test

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"polystore_database/src/go/internal/testopts"
	"sort"
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

const mappingPath = "../../../../datastore/datasets/ldbc_snb_sf1/mapping.json"

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

func baseMapping(t *testing.T) *catalog.MappingDictionary {
	t.Helper()
	md, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func clone(t *testing.T, md *catalog.MappingDictionary) *catalog.MappingDictionary {
	t.Helper()
	b, err := json.Marshal(md)
	if err != nil {
		t.Fatal(err)
	}
	var out catalog.MappingDictionary
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	out.Stats = md.Stats
	return &out
}

// place moves the migrations' columns to a store in a copy of the mapping (as a migration records it).
func place(t *testing.T, base *catalog.MappingDictionary, target string, migs []migrate.MigrationConfig) *catalog.MappingDictionary {
	md := clone(t, base)
	for _, m := range migs {
		for _, p := range m.Properties {
			if m.ObjType == catalog.ObjRelationship {
				if r, ok := md.Relationships[m.Entity]; ok {
					if pm, ok := r.Properties[p]; ok {
						pm.Databases = []string{target}
					}
				}
				continue
			}
			if e, ok := md.Entities[m.Entity]; ok {
				if pm, ok := e.Properties[p]; ok {
					pm.Databases = []string{target}
				}
			}
		}
	}
	return md
}

func plan(t *testing.T, q string, params map[string]string, md *catalog.MappingDictionary) (string, error) {
	t.Helper()
	a, err := ast.Parse(q, params)
	if err != nil {
		return "", err
	}
	l, err := cypher.Translate(a)
	if err != nil {
		return "", err
	}
	p, err := physical.Build(logical.Rewrite(logical.PlanPatterns(l, nil, lopt), lopt), physical.NewPlacement(md), popt)
	if err != nil {
		return "", err
	}
	return physical.Dump(p), nil
}

var placements = []struct {
	name   string
	target string
	mode   migrate.MigrationMode
}{
	{"graph", "", ""},
	{"relational", "relational", migrate.ModeGraphToRdb},
	{"document", "document", migrate.ModeGraphToDoc},
	{"columnar", "columnar", migrate.ModeGraphToCol},
	{"widecolumn", "widecolumn", migrate.ModeGraphToWide},
	{"kvs", "kvs", migrate.ModeGraphToKvs},
}

// Physical plans of every workload x 5 placements, pinned by a golden; placements follow each workload's migrations.
func TestPhysicalWorkloads(t *testing.T) {
	golden(t, "workloads.golden", workloadPlans(t))
}

// Physical plans with DeferPredicates (fewer splits), pinned.
func TestPhysicalWorkloadsDeferred(t *testing.T) {
	prevL, prevP := lopt, popt
	lopt.DeferPredicates, popt.DeferPredicates = true, true
	defer func() { lopt, popt = prevL, prevP }()
	golden(t, "workloads_deferred.golden", workloadPlans(t))
}

func workloadPlans(t *testing.T) string {
	base := baseMapping(t)
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		for _, pc := range placements {
			mode := pc.mode
			if mode == "" {
				mode = migrate.ModeGraphToRdb
			}
			cypher, params, migs := workload.Def(n)(mode, pc.target != "")
			md := base
			if pc.target != "" {
				md = place(t, base, pc.target, migs)
			}
			got, err := plan(t, cypher, params, md)
			if errors.Is(err, ast.ErrUnsupported) || errors.Is(err, ast.ErrSyntax) {
				if pc.name == "graph" {
					b.WriteString("### " + n + "\nUNSUPPORTED: " + err.Error() + "\n\n")
				}
				break
			}
			if err != nil {
				t.Fatalf("%s/%s: %v", n, pc.name, err)
			}
			b.WriteString("### " + n + " @" + pc.name + "\n" + got + "\n")
		}
	}
	return b.String()
}

// Representative cases of merging, absorbing and path handling (mixed placements for illustration).
func TestPhysicalCases(t *testing.T) {
	base := baseMapping(t)
	mixed := clone(t, base)
	person := mixed.Entities["Person"].Properties
	person["firstName"].Databases = []string{"relational"}
	person["lastName"].Databases = []string{"relational"}
	person["gender"].Databases = []string{"document"}
	person["birthday"].Databases = []string{"widecolumn"}
	ch := clone(t, base)
	ch.Entities["Person"].Properties["birthday"].Databases = []string{"columnar"}
	ch.Entities["Person"].Properties["gender"].Databases = []string{"columnar"}

	cases := []struct {
		name string
		md   *catalog.MappingDictionary
		q    string
	}{
		{"all_graph_single_query", base, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"three_stores_chain", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"sort_limit_then_fetch", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.firstName ORDER BY f.id ASC LIMIT 10"},
		{"scan_in_document", mixed, "MATCH (p:Person) WHERE p.gender = 'male' MATCH (p)-[:KNOWS]->(f:Person) RETURN f.id"},
		{"aggregate_single_store", mixed, "MATCH (p:Person) WHERE p.gender = 'male' RETURN p.gender, count(*) AS cnt"},
		{"aggregate_after_traversal", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.gender, count(*) AS cnt ORDER BY cnt DESC LIMIT 5"},
		{"widecolumn_is_null_at_coordinator", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.birthday IS NULL RETURN f.id"},
		{"columnar_is_null_in_store", ch, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.birthday IS NULL OR NOT f.gender = 'male' RETURN f.id"},
		{"columnar_group_sort_limit", ch, "MATCH (p:Person) WHERE p.gender <> 'x' RETURN p.gender, count(DISTINCT p.birthday) AS n ORDER BY n DESC LIMIT 3"},
		{"cross_store_predicate", mixed, "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person) WHERE a.firstName = b.gender RETURN b.id"},
		{"cycle_all_graph", base, "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id"},
		{"split_graph", base, "MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 2}) RETURN x.firstName"},
		{"split_mixed", mixed, "MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 2}) RETURN x.firstName"},
		{"varlen_between_bound_graph", base, "MATCH (a:Person {id: 1})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id"},
		{"cartesian_graph", base, "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.id, t.id"},
		{"cartesian_mixed", mixed, "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.firstName, t.id"},
		{"path_preds_graph", base, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_preds_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_any_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id"},
		{"with_stage_mixed", mixed, "MATCH (p:Person)-[:KNOWS]->(f:Person) WITH p, count(f) AS deg WHERE deg > 10 RETURN p.firstName, deg ORDER BY deg DESC LIMIT 3"},
		{"quantified_graph", base, "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
		{"quantified_document", mixed, "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
	}
	var b strings.Builder
	for _, c := range cases {
		got, err := plan(t, c.q, map[string]string{"x": "1"}, c.md)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + got + "\n")
	}
	golden(t, "cases.golden", b.String())
}

// lopt and popt are the planner options the goldens are pinned to; tests may toggle them.
var lopt, popt = testopts.Logical(), testopts.Physical()

// pathStrategyCases put conditions on path elements whose values sit outside Neo4j (pathMixed), plus one
// all-graph case where every strategy plans as post.
func pathStrategyCases(t *testing.T) []struct {
	name string
	md   *catalog.MappingDictionary
	q    string
} {
	base := baseMapping(t)
	mixed := clone(t, base)
	person := mixed.Entities["Person"].Properties
	person["firstName"].Databases = []string{"relational"}
	person["lastName"].Databases = []string{"relational"}
	person["gender"].Databases = []string{"document"}
	mixed.Relationships["KNOWS"].Properties["creationDate"].Databases = []string{"relational"}
	return []struct {
		name string
		md   *catalog.MappingDictionary
		q    string
	}{
		{"all_nodes", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
		{"none_two_stores", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE none(n IN nodes(p) WHERE n.gender = 'male' AND n.lastName = 'Lee') RETURN c.id"},
		{"any_point", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE any(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id"},
		{"single_point", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE single(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id"},
		{"all_rels", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(r IN relationships(p) WHERE r.creationDate > '2012-01-01T00:00:00Z') RETURN c.id"},
		{"nodes_and_rels", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.firstName <> 'Dee') AND all(r IN relationships(p) WHERE r.creationDate > '2012-01-01T00:00:00Z') RETURN c.id"},
		{"fixed_and_var", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"unbounded", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
		{"returned_segment", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id, nodes(p) AS ns"},
		{"count_weighted", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN count(*) AS n"},
		{"all_graph", base, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
		{"quantified_nodes", mixed, "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
		{"quantified_rels", mixed, "MATCH (a:Person {id: $x})-[r:KNOWS WHERE r.creationDate > '2012-01-01T00:00:00Z']-{1,3}(c:Person) RETURN c.id"},
	}
}

// Plans of path element conditions under prefilter (id sets computed before the traversal) and stepwise
// (an expansion hop by hop that checks each hop), pinned.
func TestPhysicalPathStrategies(t *testing.T) {
	prev := popt
	defer func() { popt = prev }()
	var b strings.Builder
	for _, c := range pathStrategyCases(t) {
		for _, s := range []physical.PathStrategy{physical.PathPrefilter, physical.PathStepwise} {
			popt.PathPredicates = s
			got, err := plan(t, c.q, map[string]string{"x": "1"}, c.md)
			if err != nil {
				t.Fatalf("%s/%s: %v", c.name, s, err)
			}
			b.WriteString("### " + c.name + " @" + s.String() + "\n" + c.q + "\n" + got + "\n")
		}
	}
	golden(t, "path_strategies.golden", b.String())
}

// Path strategies are named post / prefilter / stepwise.
func TestParsePathStrategy(t *testing.T) {
	for _, s := range []physical.PathStrategy{physical.PathPost, physical.PathPrefilter, physical.PathStepwise} {
		got, err := physical.ParsePathStrategy(s.String())
		if err != nil || got != s {
			t.Errorf("%s: got %v, %v", s, got, err)
		}
	}
	if _, err := physical.ParsePathStrategy("eager"); err == nil {
		t.Error("an unknown strategy must be an error")
	}
}
