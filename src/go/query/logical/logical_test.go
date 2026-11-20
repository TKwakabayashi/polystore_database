package logical_test

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/internal/testopts"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/logical"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "regenerate the golden files")

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

// plan runs the front end and plans its patterns.
func plan(q *ast.Query, st *stats.Stats) (ir.Node, error) {
	n, err := cypher.Translate(q)
	if err != nil {
		return nil, err
	}
	return logical.PlanPatterns(n, st, testopts.Logical()), nil
}

func translate(t *testing.T, q string) string {
	t.Helper()
	a, err := ast.Parse(q, map[string]string{"x": "1"})
	if err != nil {
		t.Fatalf("Parse(%q): %v", q, err)
	}
	n, err := plan(a, nil)
	if err != nil {
		t.Fatalf("Translate(%q): %v", q, err)
	}
	return ir.Dump(n)
}

// Logical plans of every workload, pinned by a golden.
func TestTranslateWorkloads(t *testing.T) {
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Def(n)(migrate.ModeGraphToRdb, false)
		a, err := ast.Parse(cypher, params)
		if errors.Is(err, ast.ErrUnsupported) || errors.Is(err, ast.ErrSyntax) {
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		plan, err := plan(a, nil)
		b.WriteString("### " + n + "\n")
		switch {
		case err == nil:
			b.WriteString(ir.Dump(plan))
		case errors.Is(err, ast.ErrUnsupported):
			b.WriteString("UNSUPPORTED: " + err.Error() + "\n")
		default:
			t.Errorf("%s: %v", n, err)
		}
		b.WriteString("\n")
	}
	golden(t, "workloads.golden", b.String())
}

// Plans with statistics (datastore/datasets/ldbc_snb_sf1/stats.json, collected on ldbc_sf1) that differ from the rule plans, pinned by a golden.
func TestTranslateWorkloadsCost(t *testing.T) {
	st, err := stats.Load("../../../../datastore/datasets/ldbc_snb_sf1/stats.json")
	if err != nil || st == nil {
		t.Skipf("no datastore/datasets/ldbc_snb_sf1/stats.json: %v", err)
	}
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Def(n)(migrate.ModeGraphToRdb, false)
		a, err := ast.Parse(cypher, params)
		if err != nil {
			continue
		}
		rule, err := plan(a, nil)
		if err != nil {
			continue
		}
		cost, err := plan(a, st)
		if err != nil {
			t.Errorf("%s: %v", n, err)
			continue
		}
		if ir.Dump(cost) != ir.Dump(rule) {
			b.WriteString("### " + n + "\n" + ir.Dump(cost) + "\n")
		}
	}
	golden(t, "workloads_cost.golden", b.String())
}

// Representative cases (cycles, reverse direction, cross products, WITH, path conditions), pinned by a golden.
func TestTranslateCases(t *testing.T) {
	cases := []struct{ name, q string }{
		{"chain_with_filter", "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"cycle", "MATCH (a:Person {id: $x})-[:KNOWS]->(b)-[:KNOWS]->(c)-[:KNOWS]->(a) RETURN b.id, c.id"},
		{"reverse_connection", "MATCH (a:Person {id: $x}), (b:Person)-[:KNOWS]->(a) RETURN b.id"},
		{"cartesian", "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.id, t.id"},
		{"varlen_between_bound", "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id"},
		{"split_two_anchors", "MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 2}) RETURN x.id"},
		{"split_bound_and_anchor", "MATCH (a:Person {id: 1}) WITH a MATCH (a)-[:KNOWS]-(x)-[:KNOWS]-(y)-[:KNOWS]-(b:Person {id: 2}) RETURN x.id, y.id"},
		{"split_cycle_three_anchors", "MATCH (a:Person {id: 1})-[:KNOWS]-(b:Person {id: 2})-[:KNOWS]-(c:Person {id: 3})-[:KNOWS]-(a) RETURN a.id"},
		{"cross_alias_predicate", "MATCH (a:Person)-[:KNOWS]-(b:Person) WHERE a.id < b.id RETURN a.id, b.id"},
		{"start_choice", "MATCH (a:Person)-[:KNOWS]->(b:Person {id: $x}) RETURN a.id"},
		{"with_aggregate", "MATCH (p:Person)-[:KNOWS]->(f:Person) WITH p, count(f) AS deg WHERE deg > 10 MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.id, deg, c.name ORDER BY deg DESC LIMIT 3"},
		{"with_distinct_order_limit", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH DISTINCT f ORDER BY f.id ASC LIMIT 3 MATCH (f)-[:KNOWS]-(g:Person) RETURN g.id"},
		{"with_rename", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f AS friend RETURN friend.id"},
		{"order_by_hidden", "MATCH (p:Person) RETURN p.firstName ORDER BY p.id DESC LIMIT 10"},
		{"case_expr", "MATCH (p:Person {id: $x}) RETURN p.id, CASE WHEN p.gender = 'male' THEN 1 ELSE 0 END AS isMale"},
		{"path_length_absorbed", "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE length(p) <= 3 RETURN c.id"},
		{"path_all_nodes", "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
		{"path_none_rels", "MATCH p=(a:Person {id: $x})-[r:KNOWS]->(b:Person)-[:KNOWS*1..2]-(c:Person) WHERE none(e IN relationships(p) WHERE e.creationDate < 2010) RETURN c.id"},
		{"path_any_kept", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id"},
		{"path_length_fixed_only", "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person) WHERE length(p) > 2 RETURN b.id"},
		{"rel_uniqueness", "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person)-[:KNOWS]-(c:Person) RETURN c.id"},
	}
	var b strings.Builder
	for _, c := range cases {
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + translate(t, c.q) + "\n")
	}
	golden(t, "cases.golden", b.String())
}

// Plans after the logical rewrites (deferred materialization, pruning), pinned by a golden.
func TestRewriteCases(t *testing.T) {
	cases := []struct{ name, q string }{
		{"defer_after_limit", "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.firstName, f.lastName ORDER BY f.id ASC LIMIT 10"},
		{"defer_keeps_sort_key", "MATCH (p:Person {id: $x})<-[:HAS_CREATOR]-(m:Message) RETURN m.id, m.length ORDER BY m.length DESC, m.id ASC LIMIT 20"},
		{"defer_partial", "MATCH (p:Person {id: $x})<-[:HAS_CREATOR]-(m:Message) RETURN m.id, m.content, p.firstName ORDER BY m.creationDate DESC LIMIT 10"},
		{"no_defer_with_distinct", "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WITH DISTINCT f.firstName AS n, f.lastName AS l ORDER BY n ASC LIMIT 5 RETURN n, l"},
		{"no_defer_without_limit", "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.firstName"},
		{"with_stage_defer", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f ORDER BY f.id ASC LIMIT 3 MATCH (f)-[:KNOWS]-(g:Person) RETURN g.id"},
	}
	var b strings.Builder
	for _, c := range cases {
		a, err := ast.Parse(c.q, map[string]string{"x": "1"})
		if err != nil {
			t.Fatal(err)
		}
		n, err := plan(a, nil)
		if err != nil {
			t.Fatal(err)
		}
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + ir.Dump(logical.Rewrite(n, testopts.Logical())) + "\n")
	}
	golden(t, "rewrite.golden", b.String())
}

// DeferPredicatesWhere moves only the late columns and their Filters after the traversal; other conditions and the start's stay.
func TestDeferPredicatesWhere(t *testing.T) {
	q, err := ast.Parse("MATCH (a:Person {id: 1})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person) WHERE b.gender = 'x' AND b.firstName = 'y' RETURN c.id", nil)
	if err != nil {
		t.Fatal(err)
	}
	n, err := plan(q, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := ir.Dump(logical.DeferPredicatesWhere(n, func(g *ir.GetProperty) bool {
		return g != nil && g.Prop == "gender"
	}))
	at := func(s string) int {
		i := strings.Index(got, s)
		if i < 0 {
			t.Fatalf("%q is missing\n%s", s, got)
		}
		return i
	}
	// Dump lists from the top (executed last)
	if !(at("Filter((b.gender") < at("Expand((b)-") && at("Expand((b)-") < at("Filter((b.firstName")) {
		t.Errorf("only gender should move after the expansion above\n%s", got)
	}
	if at("Filter((b.firstName") > at("Expand((a)-") {
		t.Errorf("a condition that stays moved down\n%s", got)
	}
}

func splitPlan(t *testing.T, q string) string {
	t.Helper()
	a, err := ast.Parse(q, map[string]string{"x": "1"})
	if err != nil {
		t.Fatalf("Parse(%q): %v", q, err)
	}
	n, err := plan(a, nil)
	if err != nil {
		t.Fatalf("Translate(%q): %v", q, err)
	}
	return ir.Dump(logical.SplitAggregation(n))
}

// Shapes that split into partial aggregations and shapes that do not (DISTINCT, collect, key values used elsewhere).
func TestSplitAggregation(t *testing.T) {
	cases := []struct {
		name  string
		q     string
		split bool
		has   []string // what appears in the plan when split
	}{
		{"key_props_to_owner", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.firstName, f.lastName, count(*) AS n", true,
			[]string{"PartialAggregation[keys: f | count(*) AS _p1_0]", "sum(_p1_0) AS n"}},
		{"avg_to_sum_count", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.lastName, avg(p.id) AS a", true,
			[]string{"sum(p.id) AS _p1_0s, count(p.id) AS _p1_0c", "CASE WHEN (_p1_0c = 0) THEN null ELSE (tofloat(_p1_0s) / _p1_0c) END AS a"}},
		{"identity_key", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f, min(p.id) AS m", true,
			[]string{"PartialAggregation[keys: f | min(p.id) AS _p1_0]", "min(_p1_0) AS m"}},
		{"no_keys", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN count(*) AS n, max(f.id) AS m", true,
			[]string{"PartialAggregation[keys: ∅"}},
		{"distinct_not_split", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.lastName, count(DISTINCT p.id) AS n", false, nil},
		{"collect_not_split", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.lastName, collect(p.id) AS ids", false, nil},
		{"key_used_by_filter", "MATCH (p:Person)-[:KNOWS]->(f:Person) WHERE f.lastName <> 'Lee' RETURN f.lastName, count(*) AS n", false, nil},
		{"key_used_by_arg", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.lastName, count(f.lastName) AS n", false, nil},
	}
	for _, c := range cases {
		got := splitPlan(t, c.q)
		if strings.Contains(got, "PartialAggregation") != c.split {
			t.Errorf("%s: split=%v\n%s", c.name, !c.split, got)
			continue
		}
		for _, h := range c.has {
			if !strings.Contains(got, h) {
				t.Errorf("%s: missing %q\n%s", c.name, h, got)
			}
		}
	}
}
