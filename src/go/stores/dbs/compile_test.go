package dbs_test

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/internal/testopts"
	"polystore_database/src/go/query/ir"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/exec"
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

// lowerAll compiles every StoreQuery / ListFetch of the physical plan, leaves first, and prints them.
func lowerAll(t *testing.T, q string, params map[string]string, md *catalog.MappingDictionary) (string, error) {
	t.Helper()
	a, err := ast.Parse(q, params)
	if err != nil {
		return "", err
	}
	l, err := cypher.Translate(a)
	if err != nil {
		return "", err
	}
	cat := physical.NewPlacement(md)
	p, err := physical.Build(logical.Rewrite(logical.PlanPatterns(l, nil, lopt), lopt), cat, popt)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	var walk func(physical.Node) error
	walk = func(n physical.Node) error {
		for _, c := range n.Children() {
			if err := walk(c); err != nil {
				return err
			}
		}
		switch x := n.(type) {
		case *physical.StoreQuery:
			nq, err := exec.CompileStoreQuery(x, cat.TypeOf)
			if err != nil {
				return err
			}
			b.WriteString(nq.Render())
		case *physical.ListFetch:
			qs, err := exec.CompileList(x, cat.TypeOf)
			if err != nil {
				return err
			}
			for _, nq := range qs {
				b.WriteString("(list) " + nq.Render())
			}
		case *physical.Coord:
			b.WriteString("[coordinator] " + x.Op.Describe() + "\n")
		case *physical.StepExpand:
			nq, err := exec.CompileStoreQuery(x.Hop, cat.TypeOf)
			if err != nil {
				return err
			}
			b.WriteString("(hop) " + nq.Render())
			for _, f := range x.Fetches {
				qs, err := exec.CompileList(f, cat.TypeOf)
				if err != nil {
					return err
				}
				for _, nq := range qs {
					b.WriteString("(step fetch) " + nq.Render())
				}
			}
			b.WriteString("[coordinator] " + x.Describe() + "\n")
		}
		return nil
	}
	if err := walk(p); err != nil {
		return "", err
	}
	return b.String(), nil
}

var placements = []struct {
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

// Native queries of every workload x 5 placements, pinned by a golden file.
func TestNativeWorkloads(t *testing.T) {
	base, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		for _, pc := range placements {
			cypher, params, migs := workload.Def(n)(pc.mode, pc.target != "")
			md := base
			if pc.target != "" {
				md = place(t, base, pc.target, migs)
			}
			got, err := lowerAll(t, cypher, params, md)
			if errors.Is(err, ast.ErrUnsupported) || errors.Is(err, ast.ErrSyntax) {
				if pc.name == "graph" && strings.Contains(err.Error(), "syntax error") {
					break // fails at parsing (recorded in the physical golden)
				}
				b.WriteString("### " + n + " @" + pc.name + "\nUNSUPPORTED: " + err.Error() + "\n\n")
				continue
			}
			if err != nil {
				t.Fatalf("%s/%s: %v", n, pc.name, err)
			}
			b.WriteString("### " + n + " @" + pc.name + "\n" + got + "\n")
		}
	}
	golden(t, "workloads.golden", b.String())
}

// Native queries of the representative physical plan cases.
func TestNativeCases(t *testing.T) {
	base, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	mixed := clone(t, base)
	person := mixed.Entities["Person"].Properties
	person["firstName"].Databases = []string{"relational"}
	person["lastName"].Databases = []string{"relational"}
	person["gender"].Databases = []string{"document"}
	person["birthday"].Databases = []string{"widecolumn"}
	ch := clone(t, base)
	ch.Entities["Person"].Properties["birthday"].Databases = []string{"columnar"}
	ch.Entities["Person"].Properties["gender"].Databases = []string{"columnar"}
	kv := clone(t, base)
	kv.Entities["Person"].Properties["gender"].Databases = []string{"kvs"}
	kv.Entities["Person"].Properties["firstName"].Databases = []string{"kvs"}

	cases := []struct {
		name string
		md   *catalog.MappingDictionary
		q    string
	}{
		{"all_graph_single_query", base, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"three_stores_chain", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"sort_limit_then_fetch", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.firstName ORDER BY f.id ASC LIMIT 10"},
		{"scan_in_document", mixed, "MATCH (p:Person) WHERE p.gender = 'male' MATCH (p)-[:KNOWS]->(f:Person) RETURN f.id"},
		{"date_literal_graph", base, "MATCH (p:Person) WHERE p.birthday >= '1985-01-01' RETURN p.id"},
		{"date_literal_relational", mixed, "MATCH (p:Person) WHERE p.birthday >= '1985-01-01' RETURN p.id"},
		{"split_two_anchors_graph", base, "MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 2}) RETURN x.firstName"},
		{"varlen_between_bound_graph", base, "MATCH (a:Person {id: 1})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id"},
		{"aggregate_single_store", mixed, "MATCH (p:Person) WHERE p.gender = 'male' RETURN p.gender, count(*) AS cnt"},
		{"aggregate_distinct_document", mixed, "MATCH (p:Person) WHERE p.gender IN ['male', 'female'] RETURN p.gender, count(DISTINCT p.gender) AS kinds, count(*) AS cnt ORDER BY cnt DESC LIMIT 2"},
		{"widecolumn_fetch_and_filter", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.birthday IS NULL RETURN f.id"},
		{"columnar_fetch_and_filter", ch, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.birthday IS NULL OR NOT f.gender = 'male' RETURN f.id"},
		{"columnar_group_sort_limit", ch, "MATCH (p:Person) WHERE p.gender <> 'x' AND p.birthday >= '1990-01-01' RETURN p.gender, count(DISTINCT p.birthday) AS n ORDER BY n DESC LIMIT 3"},
		{"relational_scan_aggregate", mixed, "MATCH (p:Person) WHERE p.firstName = 'Jan' RETURN p.lastName, count(*) AS n ORDER BY n DESC LIMIT 5"},
		{"datetime_literal_typed", base, "MATCH (p:Person {id: $x})<-[:HAS_CREATOR]-(m:Message) WHERE m.creationDate < '2011-06-16T00:00:00.000Z' RETURN m.id"},
		{"cycle_all_graph", base, "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id"},
		{"cartesian_graph", base, "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.id, t.id"},
		{"path_preds_graph", base, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_preds_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_any_graph", base, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id, length(p) AS len"},
		{"path_any_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id"},
		{"with_stage_graph", base, "MATCH (p:Person)-[:KNOWS]->(f:Person) WITH p, count(f) AS deg WHERE deg > 10 RETURN p.firstName, deg ORDER BY deg DESC LIMIT 3"},
		{"with_distinct_order", base, "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH DISTINCT f ORDER BY f.id ASC LIMIT 3 MATCH (f)-[:KNOWS]-(g:Person) RETURN g.id"},
		{"kvs_scan_index_eq", kv, "MATCH (p:Person) WHERE p.gender = 'male' RETURN p.firstName"},
		{"quantified_graph", base, "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
		{"quantified_document", mixed, "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
	}
	var b strings.Builder
	for _, c := range cases {
		got, err := lowerAll(t, c.q, map[string]string{"x": "1"}, c.md)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + got + "\n")
	}
	golden(t, "cases.golden", b.String())
}

// Native queries of shipped boundaries (SplitAggregation + ShipRows): workloads that plan ShipAggregate on
// relational / document, binary uuid storage, and the compile-only ShipFilter / ShipProject.
func TestNativeShip(t *testing.T) {
	base, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	prevL, prevP, prevUUID := lopt, popt, id.CurrentStorage()
	defer func() { lopt, popt = prevL, prevP; id.SetStorage(prevUUID) }()
	lopt.SplitAggregation, popt.ShipRows = true, true
	popt.ShipStores = testopts.AllShipStores()

	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		for _, pc := range placements {
			cypher, params, migs := workload.Def(n)(pc.mode, pc.target != "")
			md := base
			if pc.target != "" {
				md = place(t, base, pc.target, migs)
			}
			got, err := lowerAll(t, cypher, params, md)
			if err != nil {
				continue // workloads that cannot be planned are recorded in other goldens
			}
			if !strings.Contains(got, "shipped") {
				continue
			}
			b.WriteString("### " + n + " @" + pc.name + "\n" + got + "\n")
		}
	}

	// Binary uuid storage (sent as hex and matched with UNHEX).
	id.SetStorage(id.StorageBinary)
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person)<-[:HAS_CREATOR]-(m:Message) RETURN f.firstName, count(*) AS n, avg(m.length) AS a"
	md := clone(t, base)
	md.Entities["Message"].Properties["length"].Databases = []string{"relational"}
	got, err := lowerAll(t, q, map[string]string{"x": "1"}, md)
	if err != nil {
		t.Fatal(err)
	}
	b.WriteString("### binary uuid @relational\n" + q + "\n" + got + "\n")
	id.SetStorage(prevUUID)

	// Compile-only modes: never planned, but they must compile.
	for _, mode := range []ir.ShipMode{ir.ShipFilter, ir.ShipProject} {
		for _, k := range []kind.Kind{kind.MySQL, kind.Mongo} {
			arg := &ir.Boundary{Cols: []string{"m"}, Ship: []string{"f"}, Rid: true}
			ops := ir.Node(&ir.GetProperty{Input: arg, Alias: "m", Prop: "length", Kind: ir.NodeObj, Labels: []string{"Message"}})
			ops = &ir.Filter{Input: ops, Pred: &ir.Bin{Op: ">", L: &ir.Column{Key: "m.length"}, R: &ir.Lit{Val: int64(10)}}}
			out := []string{ir.RidCol}
			if mode == ir.ShipProject {
				out = append(out, "m.length")
			}
			sq := &physical.StoreQuery{StoreQuery: ir.StoreQuery{Store: k, Ops: ops, Keys: []string{"m"}, Ship: []string{"f"}, ShipMode: mode, Outputs: out}}
			nq, err := exec.CompileStoreQuery(sq, physical.NewPlacement(base).TypeOf)
			if err != nil {
				t.Fatalf("%s @%s: %v", mode, k, err)
			}
			b.WriteString("### mode=" + mode.String() + " @" + k.String() + "\n" + nq.Render() + "\n")
		}
	}
	golden(t, "ship.golden", b.String())
}

// An absorbed `WITH DISTINCT` receives its input with UNWIND (with `IN $in` Neo4j drops the DISTINCT).
func TestDistinctInputUnwind(t *testing.T) {
	base, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	prevL, prevP := lopt, popt
	lopt.DeferPredicates, popt.DeferPredicates = true, true
	defer func() { lopt, popt = prevL, prevP }()
	md := clone(t, base)
	md.Entities["Person"].Properties["firstName"].Databases = []string{"relational"}
	got, err := lowerAll(t, "MATCH (p:Person {firstName: 'John'})-[:KNOWS*1..2]-(f:Person) WITH DISTINCT f MATCH (f)-[:IS_LOCATED_IN]->(c:City) RETURN f.id", nil, md)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "UNWIND $in AS _k MATCH (p:Person {uuid: _k})") || !strings.Contains(got, "WITH DISTINCT f AS f") {
		t.Errorf("unexpected Cypher:\n%s", got)
	}
}

// lopt and popt are the planner options the goldens are pinned to; tests may toggle them.
var lopt, popt = testopts.Logical(), testopts.Physical()

// Native queries of path element conditions under prefilter (id sets as Neo4j parameters, set plans in the
// stores that hold the values) and stepwise (one Neo4j hop query and value fetches per depth).
func TestNativePathStrategies(t *testing.T) {
	base, err := catalog.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	mixed := clone(t, base)
	person := mixed.Entities["Person"].Properties
	person["firstName"].Databases = []string{"relational"}
	person["lastName"].Databases = []string{"relational"}
	person["gender"].Databases = []string{"document"}
	mixed.Relationships["KNOWS"].Properties["creationDate"].Databases = []string{"relational"}
	cases := []struct{ name, q string }{
		{"all_nodes", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
		{"none_two_stores", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE none(n IN nodes(p) WHERE n.gender = 'male' AND n.lastName = 'Lee') RETURN c.id"},
		{"any_point", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE any(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id"},
		{"nodes_and_rels", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.firstName <> 'Dee') AND all(r IN relationships(p) WHERE r.creationDate > '2012-01-01T00:00:00Z') RETURN c.id"},
		{"quantified_nodes", "MATCH (a:Person {id: $x}) ((n:Person)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
	}
	prev := popt
	defer func() { popt = prev }()
	var b strings.Builder
	for _, c := range cases {
		for _, s := range []physical.PathStrategy{physical.PathPrefilter, physical.PathStepwise} {
			popt.PathPredicates = s
			got, err := lowerAll(t, c.q, map[string]string{"x": "1"}, mixed)
			if err != nil {
				t.Fatalf("%s/%s: %v", c.name, s, err)
			}
			b.WriteString("### " + c.name + " @" + s.String() + "\n" + c.q + "\n" + got + "\n")
		}
	}
	golden(t, "path_strategies.golden", b.String())
}
