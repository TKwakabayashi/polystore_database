package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/internal/testopts"
	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/logical"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// ================================
// Expression evaluation (Cypher's null semantics)
// ================================

func col(k string) ir.Expr { return &ir.Column{Key: k} }
func lit(v any) ir.Expr    { return &ir.Lit{Val: v} }
func bin(op string, l, r ir.Expr) ir.Expr {
	return &ir.Bin{Op: op, L: l, R: r}
}

func TestEvalNullSemantics(t *testing.T) {
	t.Parallel()
	r := Row{"a": int64(1), "n": nil, "f": 1.0, "s": "x"}
	cases := []struct {
		e    ir.Expr
		want any
	}{
		{bin("=", col("a"), col("f")), true},
		{bin("=", col("a"), col("n")), nil},
		{bin("<>", col("a"), col("s")), true},
		{bin("<", col("a"), col("s")), nil},
		{bin("AND", lit(false), col("n")), false},
		{bin("AND", lit(true), col("n")), nil},
		{bin("OR", lit(true), col("n")), true},
		{bin("OR", lit(false), col("n")), nil},
		{bin("IN", col("a"), &ir.List{Elems: []ir.Expr{lit(int64(2)), lit(nil)}}), nil},
		{bin("IN", col("a"), &ir.List{Elems: []ir.Expr{lit(int64(1)), lit(nil)}}), true},
		{&ir.Un{Op: "IS NULL", X: col("n")}, true},
		{&ir.Un{Op: "NOT", X: col("n")}, nil},
		{bin("+", col("a"), lit(int64(2))), int64(3)},
		{bin("/", col("f"), lit(int64(4))), 0.25},
		{&ir.Call{Fn: "coalesce", Args: []ir.Expr{col("n"), col("s")}}, "x"},
		{&ir.Case{Whens: []ir.CaseWhen{{When: col("n"), Then: lit(int64(1))}}, Else: lit(int64(0))}, int64(0)},
	}
	for _, c := range cases {
		got, err := eval.Eval(c.e, r)
		if err != nil {
			t.Fatalf("%s: %v", c.e, err)
		}
		if got != c.want {
			t.Errorf("%s = %v (%T), want %v", c.e, got, got, c.want)
		}
	}
}

func TestEvalQuantifiers(t *testing.T) {
	t.Parallel()
	r := Row{
		"p":         &ListVal{Nodes: []id.UUID{tid("N", "n1"), tid("N", "n2"), tid("N", "n3")}},
		"p.nodes.g": []any{"m", nil, "f"},
		"q":         &ListVal{Nodes: []id.UUID{tid("N", "n1"), tid("N", "n3")}},
		"q.nodes.g": []any{"m", "f"},
	}
	q := func(kind, src string) ir.Expr {
		return &ir.Quant{Kind: kind, Var: "x",
			List: &ir.Call{Fn: "nodes", Args: []ir.Expr{col(src)}},
			Pred: bin("=", &ir.ElemProp{Var: "x", Prop: "g"}, lit("m"))}
	}
	cases := []struct {
		kind, src string
		want      any
	}{
		{"any", "p", true}, {"all", "p", false}, {"none", "p", false}, {"single", "p", nil},
		{"single", "q", true}, {"all", "q", false},
	}
	for _, c := range cases {
		got, err := eval.Eval(q(c.kind, c.src), r)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s over %s = %v, want %v", c.kind, c.src, got, c.want)
		}
	}
}

// ================================
// Coordinator operators
// ================================

func TestAggregateSemantics(t *testing.T) {
	t.Parallel()
	rows := []Row{
		{"k": "a", "v": int64(1)}, {"k": "a", "v": nil}, {"k": "a", "v": int64(1)},
		{"k": "b", "v": int64(4)},
	}
	agg := &ir.Aggregation{
		Keys: []ir.ProjItem{{Name: "k", Expr: col("k")}},
		Aggs: []ir.AggItem{
			{Name: "all", Fn: "count"},
			{Name: "cnt", Fn: "count", Arg: col("v")},
			{Name: "dc", Fn: "count", Arg: col("v"), Distinct: true},
			{Name: "sum", Fn: "sum", Arg: col("v")},
			{Name: "avg", Fn: "avg", Arg: col("v")},
			{Name: "max", Fn: "max", Arg: col("v")},
		},
	}
	out, err := runCoord(agg, [][]Row{rows})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Row{
		"a": {"k": "a", "all": int64(3), "cnt": int64(2), "dc": int64(1), "sum": int64(2), "avg": 1.0, "max": int64(1)},
		"b": {"k": "b", "all": int64(1), "cnt": int64(1), "dc": int64(1), "sum": int64(4), "avg": 4.0, "max": int64(4)},
	}
	for _, r := range out {
		w := want[r["k"].(string)]
		for k, v := range w {
			if r[k] != v {
				t.Errorf("group %v: %s = %v, want %v", r["k"], k, r[k], v)
			}
		}
	}
	// An aggregation without keys returns one row for empty input (count = 0, avg/max = null).
	empty, err := runCoord(&ir.Aggregation{Aggs: agg.Aggs}, [][]Row{nil})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 1 || empty[0]["all"] != int64(0) || empty[0]["avg"] != nil || empty[0]["sum"] != int64(0) {
		t.Errorf("empty aggregation: %v", empty)
	}
}

func TestSortLimitDistinct(t *testing.T) {
	t.Parallel()
	rows := []Row{{"v": int64(2)}, {"v": nil}, {"v": int64(1)}, {"v": int64(2)}}
	asc, _ := runCoord(&ir.Sort{Keys: []ir.SortKey{{Expr: col("v")}}}, [][]Row{rows})
	desc, _ := runCoord(&ir.Sort{Keys: []ir.SortKey{{Expr: col("v"), Desc: true}}}, [][]Row{rows})
	vals := func(rs []Row) string {
		var parts []string
		for _, r := range rs {
			parts = append(parts, fmt.Sprint(r["v"]))
		}
		return strings.Join(parts, ",")
	}
	if vals(asc) != "1,2,2,<nil>" || vals(desc) != "<nil>,2,2,1" {
		t.Errorf("sort: asc=%s desc=%s", vals(asc), vals(desc))
	}
	lim, _ := runCoord(&ir.Limit{Skip: 1, Count: 2}, [][]Row{asc})
	if vals(lim) != "2,2" {
		t.Errorf("limit: %s", vals(lim))
	}
	d, _ := runCoord(&ir.Distinct{}, [][]Row{asc})
	if vals(d) != "1,2,<nil>" {
		t.Errorf("distinct: %s", vals(d))
	}
}

func TestHashJoinAndRelUnique(t *testing.T) {
	t.Parallel()
	l := []Row{{"a": tid("N", "1")}, {"a": tid("N", "2")}}
	r := []Row{{"a": tid("N", "1"), "b": tid("E", "x")}, {"a": tid("N", "1"), "b": tid("E", "y")}}
	if got := hashJoin(l, r, []string{"a"}); len(got) != 2 {
		t.Errorf("keyed join: %d rows", len(got))
	}
	if got := hashJoin(l, r, nil); len(got) != 4 {
		t.Errorf("cartesian: %d rows", len(got))
	}
	rows := []Row{
		{"r1": tid("E", "e1"), "r2": tid("E", "e2"), "_seg_s": &ListVal{Rels: []id.UUID{tid("E", "e3")}}},
		{"r1": tid("E", "e1"), "r2": tid("E", "e2"), "_seg_s": &ListVal{Rels: []id.UUID{tid("E", "e1")}}},
		{"r1": tid("E", "e1"), "r2": tid("E", "e1"), "_seg_s": &ListVal{}},
	}
	got, _ := runCoord(&ir.RelUniqueFilter{Rels: []string{"r1", "r2"}, Segs: []string{"s"}}, [][]Row{rows})
	if len(got) != 1 {
		t.Errorf("rel uniqueness kept %d rows, want 1", len(got))
	}
}

// ================================
// The same result in every placement (in-memory graph)
// ================================

// socialGraph is a KNOWS graph of 6 people (with the triangle p1-p2-p3) and their locations.
func socialGraph() *memGraph {
	g := newMemGraph()
	d := func(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }
	p := []id.UUID{
		g.node("p1", "Person", map[string]any{"id": 1, "firstName": "Ann", "lastName": "Lee", "gender": "female", "birthday": d("1990-01-01")}),
		g.node("p2", "Person", map[string]any{"id": 2, "firstName": "Bob", "lastName": "Kim", "gender": "male", "birthday": d("1985-05-05")}),
		g.node("p3", "Person", map[string]any{"id": 3, "firstName": "Cid", "lastName": "Lee", "gender": "male"}),
		g.node("p4", "Person", map[string]any{"id": 4, "firstName": "Dee", "lastName": "Ono", "gender": "female", "birthday": d("2000-02-02")}),
		g.node("p5", "Person", map[string]any{"id": 5, "firstName": "Eve", "lastName": "Kim", "gender": "male"}),
		g.node("p6", "Person", map[string]any{"id": 6, "firstName": "Fay", "gender": "female"}),
	}
	city := g.node("c1", "Place", map[string]any{"id": 100, "name": "Tokyo", "type": "City"})
	city2 := g.node("c2", "Place", map[string]any{"id": 101, "name": "Kyoto", "type": "City"})
	knows := [][3]any{{0, 1, 2010}, {1, 2, 2011}, {2, 0, 2012}, {1, 3, 2009}, {3, 4, 2013}, {4, 5, 2014}, {0, 4, 2015}}
	for i, k := range knows {
		g.rel(fmt.Sprintf("k%d", i), "KNOWS", p[k[0].(int)], p[k[1].(int)], map[string]any{"creationDate": d(fmt.Sprintf("%d-01-01", k[2].(int)))})
	}
	for i, pi := range []int{0, 1, 2, 3} {
		c := city
		if i%2 == 1 {
			c = city2
		}
		g.rel(fmt.Sprintf("l%d", i), "IS_LOCATED_IN", p[pi], c, nil)
	}
	return g
}

func loadMapping(t *testing.T) *catalog.MappingDictionary {
	t.Helper()
	md, err := catalog.LoadMappingDictionary("../../../../datastore/datasets/ldbc_snb_sf1/mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func cloneMD(t *testing.T, md *catalog.MappingDictionary) *catalog.MappingDictionary {
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

func placementsFor(t *testing.T) map[string]*catalog.MappingDictionary {
	base := loadMapping(t)
	out := map[string]*catalog.MappingDictionary{"graph": base}
	mixed := cloneMD(t, base)
	pp := mixed.Entities["Person"].Properties
	pp["firstName"].Databases = []string{"relational"}
	pp["lastName"].Databases = []string{"relational"}
	pp["gender"].Databases = []string{"document"}
	pp["birthday"].Databases = []string{"columnar"}
	mixed.Relationships["KNOWS"].Properties["creationDate"].Databases = []string{"relational"}
	out["mixed"] = mixed
	for _, k := range []string{"relational", "document", "columnar", "widecolumn", "kvs"} {
		md := cloneMD(t, base)
		for _, p := range md.Entities["Person"].Properties {
			p.Databases = []string{k}
		}
		for _, p := range md.Entities["Place"].Properties {
			p.Databases = []string{k}
		}
		out["all-"+k] = md
	}
	return out
}

func canon(rows []map[string]any) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		keys := make([]string, 0, len(r))
		for k := range r {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%v;", k, r[k])
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return out
}

func runMem(t *testing.T, g *memGraph, md *catalog.MappingDictionary, q string) []map[string]any {
	t.Helper()
	p, cat, err := Plan(q, map[string]string{"x": "1"}, md, tPlan())
	if err != nil {
		t.Fatalf("Plan(%q): %v", q, err)
	}
	ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(ModelBulk))
	rows, err := ex.Execute(p)
	if err != nil {
		t.Fatalf("Execute(%q): %v", q, err)
	}
	return OutputRows(p, rows)
}

var equivalenceQueries = []struct {
	name string
	q    string
	want int // expected rows (-1: not checked)
}{
	{"chain_filter", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id", 3},
	{"two_hops_values", "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person)-[:KNOWS]->(g:Person) RETURN f.firstName, g.firstName, g.lastName", 3},
	{"sort_limit", "MATCH (p:Person) RETURN p.firstName, p.lastName ORDER BY p.id DESC LIMIT 3", 3},
	{"order_by_value_desc", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN f.firstName, f.birthday ORDER BY f.birthday DESC, f.firstName ASC", 7},
	{"aggregate_group", "MATCH (p:Person)-[:KNOWS]-(f:Person) RETURN p.lastName, count(*) AS n, count(DISTINCT f.gender) AS g ORDER BY n DESC, p.lastName ASC", -1},
	{"aggregate_no_keys", "MATCH (p:Person)-[k:KNOWS]->(f:Person) WHERE k.creationDate >= '2012-01-01T00:00:00Z' RETURN count(*) AS n, min(f.id) AS lo, max(f.id) AS hi", 1},
	{"is_null", "MATCH (p:Person) WHERE p.birthday IS NULL RETURN p.id", 3},
	{"in_list", "MATCH (p:Person) WHERE p.firstName IN ['Ann', 'Eve', 'Zed'] RETURN p.id", 2},
	{"cross_alias", "MATCH (a:Person)-[:KNOWS]-(b:Person) WHERE a.lastName = b.lastName RETURN a.id, b.id", 2},
	{"rel_property_filter", "MATCH (a:Person {id: $x})-[k:KNOWS]-(b:Person) WHERE k.creationDate < '2013-01-01T00:00:00Z' RETURN b.id, k.creationDate", 2},
	{"cycle", "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id", 1},
	{"rel_uniqueness", "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person)-[:KNOWS]-(c:Person) RETURN c.id", -1},
	{"reverse_connection", "MATCH (a:Person {id: $x}), (b:Person)-[:KNOWS]->(a) RETURN b.firstName", 1},
	{"cartesian", "MATCH (a:Person {id: $x}), (c:Place {name: 'Kyoto'}) RETURN a.firstName, c.name", 1},
	{"multi_pattern", "MATCH (p:Person)-[:KNOWS]->(f:Person), (f)-[:IS_LOCATED_IN]->(c:Place) RETURN f.firstName, c.name", -1},
	{"var_length", "MATCH (p:Person {id: $x})-[:KNOWS*1..2]->(f:Person) RETURN f.id", -1},
	{"var_length_zero", "MATCH (p:Person {id: $x})-[:KNOWS*0..1]->(f:Person) RETURN f.id", 3},
	{"path_all_nodes", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id", -1},
	{"path_any_nodes", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.lastName = 'Kim') RETURN c.id, length(p) AS len", -1},
	{"path_rels", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]->(c:Person) WHERE all(r IN relationships(p) WHERE r.creationDate > '2009-06-01T00:00:00Z') AND length(p) >= 2 RETURN c.id", -1},
	{"path_fixed_and_var", "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..2]->(c:Person) WHERE length(p) <= 2 AND none(n IN nodes(p) WHERE n.firstName = 'Cid') RETURN c.id", -1},
	{"with_aggregate_rematch", "MATCH (p:Person)-[:KNOWS]-(f:Person) WITH p, count(f) AS deg WHERE deg >= 3 MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.firstName, deg, c.name ORDER BY deg DESC", -1},
	{"with_distinct_limit", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH DISTINCT f ORDER BY f.id ASC LIMIT 2 MATCH (f)-[:KNOWS]-(g:Person) RETURN f.id, g.id", -1},
	{"with_distinct_rematch", "MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person) WHERE NOT p = f WITH DISTINCT f MATCH (f)-[:IS_LOCATED_IN]->(c:Place) RETURN f.id, f.firstName, c.name", -1},
	{"with_distinct_two_vars", "MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person)-[:IS_LOCATED_IN]->(c:Place) WITH DISTINCT f, c MATCH (f)-[:KNOWS]-(g:Person) RETURN f.id, c.name, g.id", -1},
	{"with_distinct_many_anchors", "MATCH (p:Person)-[:KNOWS]-(f:Person) WITH DISTINCT f MATCH (f)-[:IS_LOCATED_IN]->(c:Place) RETURN f.id, c.name", -1},
	{"with_distinct_fanout", "MATCH (p:Person)-[:KNOWS]-(f:Person) WITH DISTINCT f MATCH (f)-[:KNOWS]-(g:Person) RETURN f.id, g.id", -1},
	{"with_distinct_varlength_after", "MATCH (p:Person)-[:KNOWS]-(f:Person) WITH DISTINCT f MATCH (f)-[:KNOWS*1..2]-(g:Person) RETURN f.id, g.id", -1},
	{"with_distinct_count", "MATCH (p:Person)-[:KNOWS]-(f:Person) WITH DISTINCT f MATCH (f)-[:KNOWS]-(g:Person) RETURN count(*) AS n", -1},
	{"with_distinct_two_vars_many", "MATCH (p:Person)-[:KNOWS]-(f:Person)-[:IS_LOCATED_IN]->(c:Place) WITH DISTINCT f, c MATCH (f)-[:KNOWS]-(g:Person) WHERE g.id > 0 RETURN f.id, c.name, g.id", -1},
	{"with_values", "MATCH (p:Person) WITH p.lastName AS ln, count(*) AS c WHERE c > 1 RETURN ln, c ORDER BY ln ASC", 2},
	{"case_expr", "MATCH (p:Person) RETURN p.id, CASE WHEN p.gender = 'male' THEN 1 ELSE 0 END AS isMale ORDER BY p.id ASC", 6},
	{"optional_basic", "MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.id, c.name", 6},
	{"optional_where", "MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) WHERE f.gender = 'female' RETURN p.id, f.id", 6},
	{"optional_count", "MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) RETURN p.id, count(f) AS n", 6},
	{"optional_chain_null_start", "MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) OPTIONAL MATCH (c)<-[:IS_LOCATED_IN]-(o:Person) RETURN p.id, c.name, count(o) AS n", 6},
	{"optional_case_subject", "MATCH (a:Person)-[:KNOWS]->(b:Person) OPTIONAL MATCH (b)-[r:KNOWS]->(a) RETURN a.id, b.id, CASE r WHEN null THEN false ELSE true END AS back", 7},
	{"optional_into_bound", "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person) OPTIONAL MATCH (b)-[:IS_LOCATED_IN]->(c:Place)<-[:IS_LOCATED_IN]-(a) RETURN b.id, c.name", 3},
	{"optional_then_with", "MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) WITH p, count(f) AS deg RETURN deg, count(p) AS people ORDER BY deg ASC", 3},
	{"unwind_collected_nodes", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f) AS fs UNWIND fs AS g MATCH (g)-[:IS_LOCATED_IN]->(c:Place) RETURN g.id, c.name", 2},
	{"unwind_distinct_weighted", "MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person) WITH collect(DISTINCT f) AS fs UNWIND fs AS g RETURN g.id", -1},
	{"unwind_literal", "MATCH (p:Person {id: $x}) UNWIND [1, 2, 3] AS k RETURN p.id, k", 3},
	{"collect_values", "MATCH (p:Person)-[:IS_LOCATED_IN]->(c:Place) WITH c.name AS n, collect(p.id) AS ids RETURN n, size(ids) AS k", 2},
	{"in_list_column", "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f.id) AS ids MATCH (q:Person) WHERE q.id IN ids RETURN q.id", 3},
	{"pattern_prop_variable", "MATCH (a:Person {id: $x}) WITH a.lastName AS ln MATCH (b:Person {lastName: ln}) RETURN b.id", -1},
	{"split_two_anchors", "MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id", 2},
	{"split_bound_and_anchor", "MATCH (a:Person {id: $x}) WITH a MATCH (a)-[:KNOWS]-(x:Person)-[:KNOWS]-(y:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id, y.id", 1},
	{"split_cycle_three_anchors", "MATCH (a:Person {id: 1})-[:KNOWS]-(b:Person {id: 2})-[:KNOWS]-(c:Person {id: 3})-[:KNOWS]-(a) RETURN a.id", 1},
	{"varlen_between_bound", "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id", 1},
	{"order_by_output_expr", "MATCH (p:Person)-[:KNOWS]-(f:Person) RETURN p.id AS pid, count(f) AS n ORDER BY n DESC, toInteger(pid) ASC LIMIT 3", 3},
	// Partial aggregation (SplitAggregation / ShipRows): keys coarser than the owner, weighted rows, null owners, empty input, count skipping null
	{"agg_coarse_key", "MATCH (p:Person)-[:KNOWS]-(f:Person) RETURN f.lastName, count(*) AS n, avg(p.id) AS a, sum(p.id) AS s, min(p.firstName) AS lo, max(p.firstName) AS hi, count(p.birthday) AS b ORDER BY f.lastName ASC", 4},
	{"agg_weighted", "MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person) RETURN f.gender, count(*) AS n, avg(f.id) AS a, sum(f.id) AS s", 2},
	{"agg_null_owner", "MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN c.name, count(*) AS n, avg(p.id) AS a", 3},
	{"agg_empty_global", "MATCH (p:Person) WHERE p.id > 100 RETURN count(*) AS n, avg(p.id) AS a, sum(p.id) AS s, max(p.id) AS m", 1},
	{"agg_two_owners", "MATCH (p:Person)-[:KNOWS]->(f:Person) RETURN p.lastName, f.gender, count(*) AS n, max(f.id) AS m", -1},
}

func TestPlacementEquivalence(t *testing.T) {
	t.Parallel()
	g := socialGraph()
	pls := placementsFor(t)
	names := make([]string, 0, len(pls))
	for n := range pls {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, c := range equivalenceQueries {
		t.Run(c.name, func(t *testing.T) {
			ref := canon(runMem(t, g, pls["graph"], c.q))
			if c.want >= 0 && len(ref) != c.want {
				t.Errorf("graph: %d rows, want %d\n%s", len(ref), c.want, strings.Join(ref, "\n"))
			}
			for _, n := range names {
				if n == "graph" {
					continue
				}
				got := canon(runMem(t, g, pls[n], c.q))
				if strings.Join(got, "\n") != strings.Join(ref, "\n") {
					t.Errorf("%s differs from graph\n--- graph ---\n%s\n--- %s ---\n%s", n, strings.Join(ref, "\n"), n, strings.Join(got, "\n"))
				}
			}
		})
	}
}

// Hand-checked results (catch semantic mix-ups).
func TestHandCheckedResults(t *testing.T) {
	t.Parallel()
	g := socialGraph()
	md := loadMapping(t)
	cases := []struct {
		q    string
		want []string
	}{
		// p1's friends (any direction) are p2, p3, p5, all male
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WHERE f.gender = 'male' RETURN f.id",
			[]string{"f.id=2;", "f.id=3;", "f.id=5;"}},
		// the triangle p1->p2->p3->p1
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id",
			[]string{"b.id=2;c.id=3;"}},
		// relationship uniqueness: p1-p2-? does not return p1 over the same edge (via p2: p3, p4 / via p3: p2 / via p5: p4, p6)
		{"MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person)-[:KNOWS]-(c:Person) RETURN c.id",
			[]string{"c.id=2;", "c.id=3;", "c.id=4;", "c.id=4;", "c.id=6;"}},
		// zero hops include the node itself
		{"MATCH (p:Person {id: $x})-[:KNOWS*0..1]->(f:Person) RETURN f.id",
			[]string{"f.id=1;", "f.id=2;", "f.id=5;"}},
		// one row per path (destinations of 2 hops: p1->p2->p4, p1->p5->p6, ...)
		{"MATCH (p:Person {id: $x})-[:KNOWS*2]->(f:Person) RETURN f.id",
			[]string{"f.id=3;", "f.id=4;", "f.id=6;"}},
		// OPTIONAL MATCH: p5 and p6 have no location and stay with null
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.id, c.name",
			[]string{"c.name=Tokyo;p.id=1;", "c.name=Kyoto;p.id=2;", "c.name=Tokyo;p.id=3;", "c.name=Kyoto;p.id=4;", "c.name=<nil>;p.id=5;", "c.name=<nil>;p.id=6;"}},
		// count(f) skips rows without a match (f = null): out-degrees p1:2 p2:2 p3:1 p4:1 p5:1 p6:0
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) RETURN p.id, count(f) AS n",
			[]string{"n=2;p.id=1;", "n=2;p.id=2;", "n=1;p.id=3;", "n=1;p.id=4;", "n=1;p.id=5;", "n=0;p.id=6;"}},
		// WHERE is part of the match: people without a female friend stay with null (female: p1, p4, p6; p2->p4, p3->p1, p5->p6)
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) WHERE f.gender = 'female' RETURN p.id, f.id",
			[]string{"f.id=<nil>;p.id=1;", "f.id=4;p.id=2;", "f.id=1;p.id=3;", "f.id=<nil>;p.id=4;", "f.id=6;p.id=5;", "f.id=<nil>;p.id=6;"}},
		// a simple CASE with WHEN null never matches (as in Neo4j), so back is always true
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person) OPTIONAL MATCH (b)-[r:KNOWS]->(a) RETURN b.id, CASE r WHEN null THEN false ELSE true END AS back",
			[]string{"b.id=2;back=true;", "b.id=5;back=true;"}},
		// UNWIND the collected friends (p2, p3, p5) to their locations (p5 has none)
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f) AS fs UNWIND fs AS g MATCH (g)-[:IS_LOCATED_IN]->(c:Place) RETURN g.id, c.name",
			[]string{"c.name=Kyoto;g.id=2;", "c.name=Tokyo;g.id=3;"}},
		// the right side of IN is a list column
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f.id) AS ids MATCH (q:Person) WHERE q.id IN ids RETURN q.id",
			[]string{"q.id=2;", "q.id=3;", "q.id=5;"}},
		// people within 2 hops repeat over paths (weighted rows), but collect(DISTINCT) lists each once: {p2, p3, p4, p5, p6}
		// (returning to p1 would use an edge twice, which relationship uniqueness forbids)
		{"MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person) WITH collect(DISTINCT f) AS fs UNWIND fs AS g RETURN count(g) AS n",
			[]string{"n=5;"}},
		// expand from both anchors and meet at x: 1-2-4 and 1-5-4
		{"MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id",
			[]string{"x.id=2;", "x.id=5;"}},
		// from both bound a and anchor b: only 1-3-2-4 (1-?-5-4 has no common neighbor of 1 and 5)
		{"MATCH (a:Person {id: $x}) WITH a MATCH (a)-[:KNOWS]-(x:Person)-[:KNOWS]-(y:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id, y.id",
			[]string{"x.id=3;y.id=2;"}},
		// variable length between bound ends: the a->b edge is used, so b=2 gives 1-3-2 and b=5 no path
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id",
			[]string{"b.id=2;"}},
		// ORDER BY / LIMIT after an aggregation (inside WITH)
		{"MATCH (p:Person)-[:KNOWS]-(f:Person) WITH p, count(f) AS deg ORDER BY deg DESC, p.id ASC LIMIT 2 RETURN p.id, deg",
			[]string{"deg=3;p.id=1;", "deg=3;p.id=2;"}},
	}
	for _, c := range cases {
		got := canon(runMem(t, g, md, c.q))
		sort.Strings(c.want)
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("%s\n got: %v\nwant: %v", c.q, got, c.want)
		}
	}
}

// Cost-based plans (start, split and edge order from statistics) return the same results as rule plans.
func TestCostBasedEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	if pls["graph"].Stats == nil {
		t.Skip("no datastore/datasets/ldbc_snb_sf1/stats.json")
	}
	prev := costBased
	defer func() { costBased = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			costBased = false
			rule := canon(runMem(t, g, md, c.q))
			costBased = true
			cost := canon(runMem(t, g, md, c.q))
			if strings.Join(rule, "\n") != strings.Join(cost, "\n") {
				t.Errorf("%s/%s: cost-based differs\n--- rule ---\n%s\n--- cost ---\n%s", c.name, pname, strings.Join(rule, "\n"), strings.Join(cost, "\n"))
			}
		}
	}
}

// Plans split at the anchors and joined (splitAnchored) return the same results as plans grown from one start.
func TestSplitAnchoredEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev := splitAnchored
	defer func() { splitAnchored = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			splitAnchored = false
			off := canon(runMem(t, g, md, c.q))
			splitAnchored = true
			on := canon(runMem(t, g, md, c.q))
			if strings.Join(on, "\n") != strings.Join(off, "\n") {
				t.Errorf("%s/%s: split differs\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, strings.Join(off, "\n"), strings.Join(on, "\n"))
			}
		}
	}
}

// Plans with filters moved after the expansions (deferPredicates) return the same results as filters placed when bound.
func TestDeferPredicatesEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev := deferPredicates
	defer func() { deferPredicates = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			deferPredicates = false
			off := canon(runMem(t, g, md, c.q))
			deferPredicates = true
			on := canon(runMem(t, g, md, c.q))
			if strings.Join(on, "\n") != strings.Join(off, "\n") {
				t.Errorf("%s/%s: deferred predicates differ\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, strings.Join(off, "\n"), strings.Join(on, "\n"))
			}
		}
	}
}

// deferPredicates plans return the same results in models that split the input into one-row batches
// (duplicates of a query that absorbed `WITH DISTINCT` across batches).
func TestDeferPredicatesAcrossBatches(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev, prevV := deferPredicates, vectorSize
	defer func() { deferPredicates, vectorSize = prev, prevV }()
	vectorSize = 1
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			deferPredicates = false
			want := canon(runMem(t, g, md, c.q))
			deferPredicates = true
			p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md, tPlan())
			if err != nil {
				t.Fatalf("Plan(%q): %v", c.q, err)
			}
			for _, m := range allModels {
				rows, err := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(Model(m))).ExecuteModel(p)
				if err != nil {
					t.Fatalf("%s %s/%s: %v", m, c.name, pname, err)
				}
				got := canon(OutputRows(p, rows))
				if strings.Join(got, "\n") != strings.Join(want, "\n") {
					t.Errorf("%s/%s/%s: differs\n--- want ---\n%s\n--- got ---\n%s", c.name, pname, m, strings.Join(want, "\n"), strings.Join(got, "\n"))
				}
			}
		}
	}
}

// Plans that move filters on non-Neo4j columns after the traversal (deferOffGraph) return the same results as plans that do not.
func TestDeferOffGraphEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev := deferOffGraph
	defer func() { deferOffGraph = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			deferOffGraph = false
			off := canon(runMem(t, g, md, c.q))
			deferOffGraph = true
			on := canon(runMem(t, g, md, c.q))
			if strings.Join(on, "\n") != strings.Join(off, "\n") {
				t.Errorf("%s/%s: deferred off-graph predicates differ\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, strings.Join(off, "\n"), strings.Join(on, "\n"))
			}
		}
	}
}

// Plans with split aggregations (splitAggregation) and with partial aggregations shipped to stores (shipRows)
// return the same results as unsplit plans. A small ShipChunkRows makes groups span chunks.
func TestSplitAggregationEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prevSplit, prevShip, prevChunk, prevStores := splitAggregation, shipRows, shipChunkRows, shipStores
	defer func() {
		splitAggregation, shipRows, shipChunkRows, shipStores = prevSplit, prevShip, prevChunk, prevStores
	}()
	// Ship to every store (Neo4j, MySQL and MongoDB aggregate natively; ClickHouse and kvs through executor emulation).
	shipStores = everyShipStore
	conds := []struct {
		name        string
		split, ship bool
		chunk       int
	}{
		{"split", true, false, 20000},
		{"ship", true, true, 20000},
		{"ship_chunk2", true, true, 2},
	}
	shipped := 0
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			splitAggregation, shipRows = true, true
			if p, _, err := Plan(c.q, map[string]string{"x": "1"}, md, tPlan()); err == nil && strings.Contains(physical.Dump(p), "mode=aggregate") {
				shipped++
			}
			splitAggregation, shipRows = false, false
			off := canon(runMem(t, g, md, c.q))
			for _, cd := range conds {
				splitAggregation, shipRows, shipChunkRows = cd.split, cd.ship, cd.chunk
				on := canon(runMem(t, g, md, c.q))
				if strings.Join(on, "\n") != strings.Join(off, "\n") {
					t.Errorf("%s/%s/%s: differs\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, cd.name, strings.Join(off, "\n"), strings.Join(on, "\n"))
				}
			}
		}
	}
	if shipped == 0 {
		t.Error("no plan shipped rows to a store (ShipAggregate is never exercised)")
	}
}

// Iteration models do not change results (every query x placement x 5 models). -race catches races in the parallel models.
func TestModelEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	models := []Model{ModelVolcano, ModelVectorized, ModelStream, ModelVecStream}
	prev := vectorSize
	vectorSize = 2 // small, so paths cross batch boundaries
	defer func() { vectorSize = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md, tPlan())
			if err != nil {
				t.Fatalf("Plan(%q): %v", c.q, err)
			}
			ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(ModelBulk))
			ref, err := ex.Execute(p)
			if err != nil {
				t.Fatalf("bulk %s/%s: %v", c.name, pname, err)
			}
			want := canon(OutputRows(p, ref))
			for _, m := range models {
				rows, err := ex.model(m).ExecuteModel(p)
				if err != nil {
					t.Fatalf("%s %s/%s: %v", m, c.name, pname, err)
				}
				got := canon(OutputRows(p, rows))
				if strings.Join(got, "\n") != strings.Join(want, "\n") {
					t.Errorf("%s %s/%s differs from bulk\n--- bulk ---\n%s\n--- %s ---\n%s",
						m, c.name, pname, strings.Join(want, "\n"), m, strings.Join(got, "\n"))
				}
			}
		}
	}
}

// A Limit after a Sort returns the first rows in sorted order, in parallel models too.
func TestModelOrderAfterSort(t *testing.T) {
	g := socialGraph()
	md := loadMapping(t)
	q := "MATCH (p:Person) RETURN p.id ORDER BY p.id DESC LIMIT 3"
	p, cat, err := Plan(q, nil, md, tPlan())
	if err != nil {
		t.Fatal(err)
	}
	prev := vectorSize
	vectorSize = 1
	defer func() { vectorSize = prev }()
	for _, m := range []Model{ModelBulk, ModelVolcano, ModelVectorized, ModelStream, ModelVecStream} {
		rows, err := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(m)).ExecuteModel(p)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, r := range OutputRows(p, rows) {
			ids = append(ids, fmt.Sprint(r["p.id"]))
		}
		if strings.Join(ids, ",") != "6,5,4" {
			t.Errorf("%s: %v", m, ids)
		}
	}
}

// failingBackend fails every store query from the n-th on.
type failingBackend struct {
	memBackend
	mu    sync.Mutex
	calls int
	after int
}

func (b *failingBackend) Run(ctx context.Context, q *stores.Query, in []Key) ([][]any, error) {
	b.mu.Lock()
	b.calls++
	fail := b.calls > b.after
	b.mu.Unlock()
	if fail {
		return nil, fmt.Errorf("injected store failure")
	}
	return b.memBackend.Run(ctx, q, in)
}

// A store failure is an error in every model (parallel models neither hang nor return partial results).
func TestModelErrorPropagation(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"
	p, cat, err := Plan(q, map[string]string{"x": "1"}, pls["mixed"], tPlan())
	if err != nil {
		t.Fatal(err)
	}
	prev := vectorSize
	vectorSize = 1
	defer func() { vectorSize = prev }()
	for _, m := range []Model{ModelBulk, ModelVolcano, ModelVectorized, ModelStream, ModelVecStream} {
		done := make(chan error, 1)
		go func() {
			_, err := New(context.Background(), &failingBackend{memBackend: memBackend{g: g}, after: 1}, cat.TypeOf, tExec(m)).ExecuteModel(p)
			done <- err
		}()
		select {
		case err := <-done:
			if err == nil || !strings.Contains(err.Error(), "injected") {
				t.Errorf("%s: want injected error, got %v", m, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: did not return after a store failure", m)
		}
	}
}

// ---- weighted rows ----

func w(r Row, n int64) Row { eval.SetWeight(r, n); return r }

func TestWeightedCoordinator(t *testing.T) {
	t.Parallel()
	rows := []Row{
		w(Row{"k": "a", "v": int64(2)}, 3),
		w(Row{"k": "a", "v": nil}, 2),
		Row{"k": "b", "v": int64(5)},
	}
	agg := &ir.Aggregation{
		Keys: []ir.ProjItem{{Name: "k", Expr: col("k")}},
		Aggs: []ir.AggItem{
			{Name: "all", Fn: "count"},
			{Name: "cnt", Fn: "count", Arg: col("v")},
			{Name: "dc", Fn: "count", Arg: col("v"), Distinct: true},
			{Name: "sum", Fn: "sum", Arg: col("v")},
			{Name: "avg", Fn: "avg", Arg: col("v")},
			{Name: "col", Fn: "collect", Arg: col("v")},
		},
	}
	out, err := runCoord(agg, [][]Row{rows})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range out {
		if r["k"] != "a" {
			continue
		}
		if r["all"] != int64(5) || r["cnt"] != int64(3) || r["dc"] != int64(1) || r["sum"] != int64(6) || r["avg"] != 2.0 || len(r["col"].([]any)) != 3 {
			t.Errorf("weighted aggregation: %v", r)
		}
	}

	// Limit counts weights and trims the boundary row's weight: skip 1 limit 3 of [a x3, b x2] -> a x2, b x1
	lim, _ := runCoord(&ir.Limit{Skip: 1, Count: 3}, [][]Row{{w(Row{"x": "a"}, 3), w(Row{"x": "b"}, 2)}})
	if len(lim) != 2 || eval.WeightOf(lim[0]) != 2 || eval.WeightOf(lim[1]) != 1 {
		t.Errorf("weighted limit: %v", lim)
	}
	// the input weights are unchanged (Limit copies before trimming)
	src := []Row{w(Row{"x": "a"}, 3)}
	runCoord(&ir.Limit{Count: 1}, [][]Row{src})
	if eval.WeightOf(src[0]) != 3 {
		t.Error("limit mutated its input")
	}

	// Distinct ignores weights and keeps one row
	d, _ := runCoord(&ir.Distinct{}, [][]Row{{w(Row{"x": "a"}, 3), Row{"x": "a"}}})
	if len(d) != 1 || eval.WeightOf(d[0]) != 1 {
		t.Errorf("weighted distinct: %v", d)
	}
	// a join multiplies weights
	j := hashJoin([]Row{w(Row{"a": tid("N", "1")}, 2)}, []Row{w(Row{"a": tid("N", "1"), "b": "x"}, 3)}, []string{"a"})
	if len(j) != 1 || eval.WeightOf(j[0]) != 6 {
		t.Errorf("weighted join: %v", j)
	}
	// expanded at the end
	p := &physical.Coord{Op: &ir.Projection{Items: []ir.ProjItem{{Name: "x", Expr: col("x")}}}}
	if got := OutputRows(p, []Row{w(Row{"x": "a"}, 3)}); len(got) != 3 {
		t.Errorf("output expansion: %d rows", len(got))
	}
}

// In placements where a Neo4j query with a variable-length expansion is weighted, results equal the unweighted plan (weightedRows=false).
func TestWeightedRowsEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	weighted := 0
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			run := func(on bool) []string {
				prev := weightedRows
				weightedRows = on
				defer func() { weightedRows = prev }()
				p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md, tPlan())
				if err != nil {
					t.Fatalf("Plan: %v", err)
				}
				if on && strings.Contains(physical.Dump(p), "weighted") {
					weighted++
				}
				rows, err := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(ModelVecStream)).ExecuteModel(p)
				if err != nil {
					t.Fatalf("%s/%s: %v", c.name, pname, err)
				}
				return canon(OutputRows(p, rows))
			}
			off, on := run(false), run(true)
			if strings.Join(on, "\n") != strings.Join(off, "\n") {
				t.Errorf("%s/%s: weighted differs\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, strings.Join(off, "\n"), strings.Join(on, "\n"))
			}
		}
	}
	if weighted == 0 {
		t.Error("no plan used weighted rows; the test does not exercise them")
	}
	t.Logf("%d plans used weighted rows", weighted)
}

// ================================
// Path element conditions (physical.PathStrategy)
// ================================

// pathPredicateQueries put conditions on the elements of variable-length paths. Where a condition reads
// values outside Neo4j, prefilter and stepwise plan it differently from post.
var pathPredicateQueries = []struct{ name, q string }{
	{"seg_all_min2", "MATCH p=(a:Person {id: $x})-[:KNOWS*2..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id"},
	{"seg_all_len", "MATCH p=(a:Person {id: $x})-[:KNOWS*2..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id, length(p) AS len"},
	{"seg_none", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]->(c:Person) WHERE none(n IN nodes(p) WHERE n.birthday > '1995-01-01') RETURN c.id"},
	{"seg_null_prop", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.birthday IS NULL OR n.birthday < '1995-01-01') RETURN c.id"},
	{"seg_or_props", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male' OR n.lastName = 'Lee') RETURN c.id"},
	{"seg_and_props", "MATCH p=(a:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.lastName <> 'Ono' AND n.gender <> 'x') RETURN a.id, c.id"},
	{"seg_nodes_and_rels", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.firstName <> 'Dee') AND all(r IN relationships(p) WHERE r.creationDate > '2009-06-01T00:00:00Z') RETURN c.id"},
	{"seg_return_path", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender <> 'x') RETURN c.id, nodes(p) AS ns, relationships(p) AS rs"},
	{"seg_count", "MATCH p=(a:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'female' OR n.id = 2) RETURN count(*) AS n"},
	{"seg_unbounded", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..]->(c:Person) WHERE all(n IN nodes(p) WHERE n.gender <> 'x') RETURN c.id"},
	{"seg_with_fixed", "MATCH p=(a:Person {id: $x})-[:KNOWS]-(b:Person)-[:KNOWS*1..2]-(c:Person) WHERE all(n IN nodes(p) WHERE n.lastName <> 'Ono') RETURN b.id, c.id"},
	{"any_point", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE any(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id, length(p) AS len"},
	{"single_point", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE single(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id"},
	{"any_rels", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE any(r IN relationships(p) WHERE r.creationDate > '2014-06-01T00:00:00Z') RETURN c.id"},
	{"optional_seg", "MATCH (a:Person) OPTIONAL MATCH p=(a)-[:KNOWS*1..2]->(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN a.id, c.id"},
	{"seg_then_expand", "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') MATCH (c)-[:IS_LOCATED_IN]->(pl:Place) RETURN c.id, pl.name"},
	{"qpp_rel", "MATCH (a:Person {id: $x})-[:KNOWS]-{1,3}(c:Person) RETURN c.id"},
	{"qpp_right_cond", "MATCH (a:Person {id: $x}) ((n)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id"},
	{"qpp_left_cond", "MATCH (a:Person {id: $x}) ((n:Person WHERE n.lastName <> 'Ono')-[:KNOWS]-(m)){1,3} (c) RETURN c.id"},
	{"qpp_both_cond", "MATCH (a:Person) ((n:Person {gender: 'female'})-[:KNOWS]-(m:Person) WHERE m.birthday IS NULL OR m.firstName <> 'Bob'){1,2} (c) RETURN a.id, c.id"},
	{"qpp_rel_cond", "MATCH (a:Person {id: $x})-[r:KNOWS WHERE r.creationDate > '2010-06-01T00:00:00Z']->{1,3}(c:Person) RETURN c.id"},
	{"qpp_path_len", "MATCH p = (a:Person {id: $x}) ((n)-[k:KNOWS]-(m) WHERE m.gender = 'female' AND k.creationDate > '2009-06-01T00:00:00Z'){1,3} (c) RETURN c.id, length(p) AS len"},
	{"qpp_plus_count", "MATCH (a:Person {id: $x})-[:KNOWS]->+(c:Person) RETURN count(*) AS n"},
	{"qpp_zero", "MATCH (a:Person {id: $x})-[:KNOWS]->{0,1}(c:Person) RETURN c.id"},
	{"qpp_after_fixed", "MATCH (p:Place {name: 'Tokyo'})<-[:IS_LOCATED_IN]-(a:Person) ((n)-[:KNOWS]->(m WHERE m.gender = 'male')){1,2} (c) RETURN a.id, c.id"},
}

var pathStrategies = []physical.PathStrategy{physical.PathPost, physical.PathPrefilter, physical.PathStepwise}

// Every path strategy returns the graph placement's result in every placement and model. A vector size of 2
// splits the input into batches, so id sets and stepwise condition results are shared across batches.
func TestPathStrategyEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prevV, prevS := vectorSize, pathPredicates
	vectorSize = 2
	defer func() { vectorSize, pathPredicates = prevV, prevS }()
	qs := append([]struct{ name, q string }{}, pathPredicateQueries...)
	for _, c := range equivalenceQueries {
		qs = append(qs, struct{ name, q string }{c.name, c.q})
	}
	used := map[physical.PathStrategy]int{}
	for _, c := range qs {
		pathPredicates = physical.PathPost
		want := strings.Join(canon(runMem(t, g, pls["graph"], c.q)), "\n")
		for pname, md := range pls {
			for _, s := range pathStrategies {
				pathPredicates = s
				p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md, tPlan())
				if err != nil {
					t.Fatalf("%s %s/%s: %v", s, c.name, pname, err)
				}
				d := physical.Dump(p)
				if strings.Contains(d, "WithSets") || strings.Contains(d, "StepExpand") {
					used[s]++
				}
				ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(ModelBulk))
				for _, m := range allModels {
					rows, err := ex.model(m).ExecuteModel(p)
					if err != nil {
						t.Fatalf("%s %s %s/%s: %v\n%s", s, m, c.name, pname, err, d)
					}
					if got := strings.Join(canon(OutputRows(p, rows)), "\n"); got != want {
						t.Errorf("%s %s %s/%s differs from graph\n--- graph ---\n%s\n--- got ---\n%s\n%s", s, m, c.name, pname, want, got, d)
					}
				}
			}
		}
	}
	for _, s := range pathStrategies[1:] {
		if used[s] == 0 {
			t.Errorf("no plan used %s; the test does not exercise it", s)
		}
	}
	t.Logf("plans using the strategy: %v", used)
}

// Hand-checked results of path element conditions under every strategy (trails may revisit nodes but not
// relationships; inner nodes are checked).
func TestPathStrategyHandChecked(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev := pathPredicates
	defer func() { pathPredicates = prev }()
	cases := []struct {
		q    string
		want []string
	}{
		// paths from p1 that avoid Bob (p2): 1-3, 1-5, 1-5-4, 1-5-6
		{"MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE none(n IN nodes(p) WHERE n.firstName = 'Bob') RETURN c.id",
			[]string{"c.id=3;", "c.id=4;", "c.id=5;", "c.id=6;"}},
		// outgoing relationships created after mid 2010: 1->5 (2015), 1->5->6 (2014); 1->2 (2010) fails
		{"MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]->(c:Person) WHERE all(r IN relationships(p) WHERE r.creationDate > '2010-06-01T00:00:00Z') RETURN c.id",
			[]string{"c.id=5;", "c.id=6;"}},
		// paths through Dee (p4) within 2 hops: 1-2-4 and 1-5-4
		{"MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.firstName = 'Dee') RETURN c.id, length(p) AS len",
			[]string{"c.id=4;len=2;", "c.id=4;len=2;"}},
		// exactly 3 hops: 1-2-3-1, 1-3-2-1, 1-2-4-5, 1-3-2-4, 1-5-4-2 (p1 is reached again; 1-5-6 has no third relationship)
		{"MATCH p=(a:Person {id: $x})-[:KNOWS*3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender IN ['male', 'female']) RETURN c.id",
			[]string{"c.id=1;", "c.id=1;", "c.id=2;", "c.id=4;", "c.id=5;"}},
		// quantified: every node after p1 is male (p2, p3, p5): 1-2, 1-3, 1-5, 1-2-3, 1-3-2
		{"MATCH (a:Person {id: $x}) ((n)-[:KNOWS]-(m:Person) WHERE m.gender = 'male'){1,3} (c) RETURN c.id",
			[]string{"c.id=2;", "c.id=2;", "c.id=3;", "c.id=3;", "c.id=5;"}},
		// quantified: every node but the last is female; p1's friends are all male, so one hop only
		{"MATCH (a:Person {id: $x}) ((n:Person WHERE n.gender = 'female')-[:KNOWS]-(m)){1,3} (c) RETURN c.id",
			[]string{"c.id=2;", "c.id=3;", "c.id=5;"}},
		// + is one or more: outgoing trails from p1 end at 2, 5 | 3, 4, 6 | 1, 5 | 5, 6 | 6
		{"MATCH (a:Person {id: $x})-[:KNOWS]->+(c:Person) RETURN c.id",
			[]string{"c.id=1;", "c.id=2;", "c.id=3;", "c.id=4;", "c.id=5;", "c.id=5;", "c.id=5;", "c.id=6;", "c.id=6;", "c.id=6;"}},
		// {0,1} includes p1 itself
		{"MATCH (a:Person {id: $x})-[:KNOWS]->{0,1}(c:Person) RETURN c.id",
			[]string{"c.id=1;", "c.id=2;", "c.id=5;"}},
		// relationship condition on a quantified relationship: 1->5 (2015), 1->5->6 (2014)
		{"MATCH (a:Person {id: $x})-[r:KNOWS WHERE r.creationDate > '2010-06-01T00:00:00Z']->{1,2}(c:Person) RETURN c.id",
			[]string{"c.id=5;", "c.id=6;"}},
	}
	for _, c := range cases {
		sort.Strings(c.want)
		for _, pname := range []string{"graph", "mixed", "all-document", "all-kvs"} {
			for _, s := range pathStrategies {
				pathPredicates = s
				got := canon(runMem(t, g, pls[pname], c.q))
				if strings.Join(got, "|") != strings.Join(c.want, "|") {
					t.Errorf("%s %s: %s\n got: %v\nwant: %v", s, pname, c.q, got, c.want)
				}
			}
		}
	}
}

// A stepwise expansion issues one Neo4j query per depth and fetches each node's values once: from p1 over
// *1..3, depth 1 checks p2, p3, p5, depth 2 only the new p4, p6, and depth 3 extends nothing.
func TestStepExpandRoundTrips(t *testing.T) {
	g := socialGraph()
	md := placementsFor(t)["all-document"]
	prev := pathPredicates
	pathPredicates = physical.PathStepwise
	defer func() { pathPredicates = prev }()
	q := "MATCH p=(a:Person {id: $x})-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender <> 'x') RETURN c.id"
	p, cat, err := Plan(q, map[string]string{"x": "1"}, md, tPlan())
	if err != nil {
		t.Fatal(err)
	}
	ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf, tExec(ModelBulk))
	if _, err := ex.Execute(p); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, st := range ex.Steps() {
		if st.Op != "StepExpand" {
			continue
		}
		found = true
		if st.RoundTrips != 5 || st.Store != "neo4j" {
			t.Errorf("StepExpand: %d round trips at %q, want 5 (3 hops, 2 fetches) at neo4j\n%s", st.RoundTrips, st.Store, physical.Dump(p))
		}
	}
	if !found {
		t.Fatalf("no StepExpand in the plan\n%s", physical.Dump(p))
	}
}

// An id-set condition reads an id column or a quantifier element; a missing id is null, an unbound set an error.
func TestEvalInSet(t *testing.T) {
	t.Parallel()
	u1, u2 := tid("N", "s1"), tid("N", "s2")
	set := ir.NewIDSet([]id.UUID{u1})
	r := Row{"a": u1, "b": u2, "n": nil, "s": "x", "seg": &ListVal{Nodes: []id.UUID{u2, u1}}}
	for _, c := range []struct {
		col  string
		want any
	}{{"a", true}, {"b", false}, {"n", nil}, {"s", nil}} {
		got, err := eval.Eval(&ir.InSet{Var: c.col, Set: "set0", Bound: set}, r)
		if err != nil || got != c.want {
			t.Errorf("%s IN $set0 = %v, %v; want %v", c.col, got, err, c.want)
		}
	}
	elem := &ir.InSet{Var: "x", Elem: true, Set: "set0", Bound: set}
	for i, want := range []any{false, true, nil} {
		got, err := eval.EvalElem(elem, r, "x", "seg", ir.SegNodes, i)
		if err != nil || got != want {
			t.Errorf("element %d IN $set0 = %v, %v; want %v", i, got, err, want)
		}
	}
	if _, err := eval.Eval(&ir.InSet{Var: "a", Set: "set0"}, r); err == nil {
		t.Error("an unbound set must be an error")
	}
}

// Integers in a list parameter for delegated queries do not become float64.

// Test copies of the query package's planning entry (package query imports exec, so exec's
// tests cannot import it).

type PlanOptions struct {
	// CostBased uses the mapping's statistics to choose anchors, splits and edge order.
	CostBased bool
	Logical   logical.Options
	Physical  physical.Options
}

func Plan(stmt string, params map[string]string, md *catalog.MappingDictionary, opt PlanOptions) (physical.Node, *physical.Placement, error) {
	a, err := ast.Parse(stmt, params)
	if err != nil {
		return nil, nil, err
	}
	var st *stats.Stats
	if opt.CostBased && md != nil {
		st = md.Stats
	}
	// Unique keys are schema knowledge from the mapping, not a tuning switch.
	keys := md.UniqueKeys()
	opt.Logical.UniqueKeys, opt.Physical.UniqueKeys = keys, keys
	l, err := cypher.Translate(a)
	if err != nil {
		return nil, nil, err
	}
	cat := physical.NewPlacement(md)
	p, err := physical.Build(logical.Rewrite(logical.PlanPatterns(l, st, opt.Logical), opt.Logical), cat, opt.Physical)
	if err != nil {
		return nil, nil, err
	}
	// Check that every store query compiles, so an uncompilable shape is delegated before execution.
	if err := checkLowerable(p, cat); err != nil {
		return nil, nil, err
	}
	return p, cat, nil
}
func RenderPlan(p physical.Node, cat *physical.Placement) string {
	var b strings.Builder
	b.WriteString(physical.Dump(p))
	var walk func(physical.Node)
	walk = func(n physical.Node) {
		for _, ch := range n.Children() {
			walk(ch)
		}
		switch x := n.(type) {
		case *physical.StoreQuery:
			if nq, err := CompileStoreQuery(x, cat.TypeOf); err == nil {
				b.WriteString(nq.Render() + "\n")
			} else {
				fmt.Fprintf(&b, "(lower: %v)\n", err)
			}
		case *physical.ListFetch:
			qs, err := CompileList(x, cat.TypeOf)
			if err != nil {
				fmt.Fprintf(&b, "(lower: %v)\n", err)
			}
			for _, q := range qs {
				b.WriteString(q.Render() + "\n")
			}
		}
	}
	walk(p)
	return b.String()
}
func checkLowerable(n physical.Node, cat *physical.Placement) error {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if _, err := CompileStoreQuery(x, cat.TypeOf); err != nil {
			return err
		}
	case *physical.ListFetch:
		if _, err := CompileList(x, cat.TypeOf); err != nil {
			return err
		}
	case *physical.StepExpand:
		if _, err := CompileStoreQuery(x.Hop, cat.TypeOf); err != nil {
			return err
		}
		for _, f := range x.Fetches {
			if _, err := CompileList(f, cat.TypeOf); err != nil {
				return err
			}
		}
	}
	for _, c := range n.Children() {
		if err := checkLowerable(c, cat); err != nil {
			return err
		}
	}
	return nil
}

// echoBackend returns the input uuids one per row (to check order and round trips); a chunk containing fail fails.
type echoBackend struct {
	delay time.Duration
	fail  id.UUID
}

func (b *echoBackend) Run(_ context.Context, _ *stores.Query, in []Key) ([][]any, error) {
	time.Sleep(b.delay)
	var out [][]any
	for _, k := range in {
		if k[0] == b.fail {
			return nil, fmt.Errorf("chunk with %s failed", k[0])
		}
		out = append(out, []any{k[0]})
	}
	return out, nil
}

func (b *echoBackend) Verbatim(context.Context, string, map[string]any) ([]map[string]any, error) {
	return nil, nil
}

func TestRunChunksInParallelAndInOrder(t *testing.T) {
	prevChunk, prevPar := materializeChunk, storeFetchPar
	materializeChunk, storeFetchPar = 3, 4
	defer func() { materializeChunk, storeFetchPar = prevChunk, prevPar }()

	var keys []Key
	for i := 0; i < 10; i++ {
		keys = append(keys, Key{tid("N", fmt.Sprintf("u%02d", i))})
	}
	q := &stores.Query{Store: kind.ClickHouse, Columns: []string{"a"}}
	e := New(context.Background(), &echoBackend{delay: 50 * time.Millisecond}, nil, tExec(ModelBulk))
	e.reset(nil)
	start := time.Now()
	rows, err := e.run(nil, q, keys)
	if err != nil {
		t.Fatal(err)
	}
	if took := time.Since(start); took > 150*time.Millisecond {
		t.Errorf("4 chunks with parallelism 4 took %v (not parallel?)", took)
	}
	if e.RoundTrips() != 4 {
		t.Errorf("round trips = %d, want 4", e.RoundTrips())
	}
	for i, r := range rows {
		if r[0] != tid("N", fmt.Sprintf("u%02d", i)) {
			t.Fatalf("row %d = %v (order lost)", i, r[0])
		}
	}

	e = New(context.Background(), &echoBackend{fail: tid("N", "u07")}, nil, tExec(ModelBulk))
	e.reset(nil)
	if _, err := e.run(nil, q, keys); err == nil {
		t.Error("a failing chunk must fail the whole fetch")
	}
}

// Round trips are counted per issuing operator and store; output rows also count weighted rows expanded.
func TestStepMetrics(t *testing.T) {
	saved := materializeChunk
	materializeChunk = 3
	defer func() { materializeChunk = saved }()
	e := New(context.Background(), &echoBackend{}, nil, tExec(ModelBulk))
	q := &stores.Query{Store: kind.MySQL, Columns: []string{"a"}}
	owner := &physical.StoreQuery{StoreQuery: ir.StoreQuery{Store: kind.MySQL}}
	e.reset(owner)
	var keys []Key
	for i := 0; i < 7; i++ {
		keys = append(keys, Key{tid("N", fmt.Sprintf("u%02d", i))})
	}
	if _, err := e.run(owner, q, keys); err != nil {
		t.Fatal(err)
	}
	e.record(owner, 0, 7, []Row{{eval.WeightCol: int64(4)}, {}})
	steps := e.Steps()
	if len(steps) != 1 {
		t.Fatalf("steps = %v", steps)
	}
	st := steps[0]
	if st.RoundTrips != 3 || st.Store != "mysql" || st.InRows != 7 || st.OutRows != 2 || st.OutRowsExpanded != 5 {
		t.Errorf("step = %+v", st)
	}
	if got := e.StoreTrips(); got["mysql"] != 3 || len(got) != 1 {
		t.Errorf("StoreTrips = %v", got)
	}
}

// Test knobs. Tests toggle them (and restore them) to compare plans; they start at the values
// the goldens are pinned to.
var (
	costBased        = true
	splitAnchored    = true
	deferPredicates  = false
	deferOffGraph    = true
	splitAggregation = false
	shipRows         = false
	shipStores       = testopts.Physical().ShipStores
	weightedRows     = true
	pathPredicates   = physical.PathPost
	vectorSize       = 1024
	materializeChunk = 0
	shipChunkRows    = 20000
	storeFetchPar    = 4
	graphFetchPar    = 1
)

func tPlan() PlanOptions {
	l, p := testopts.Logical(), testopts.Physical()
	l.SplitAnchoredPatterns = splitAnchored
	l.DeferPredicates, p.DeferPredicates = deferPredicates, deferPredicates
	l.SplitAggregation = splitAggregation
	p.DeferOffGraphPredicates = deferOffGraph
	p.ShipRows, p.ShipStores = shipRows, shipStores
	p.WeightedRows = weightedRows
	p.PathPredicates = pathPredicates
	return PlanOptions{CostBased: costBased, Logical: l, Physical: p}
}

func tExec(m Model) ExecOptions {
	return ExecOptions{Model: m, VectorSize: vectorSize, MaterializeChunkSize: materializeChunk,
		ShipChunkRows: shipChunkRows, StoreFetchParallelism: storeFetchPar, GraphFetchParallelism: graphFetchPar}
}

// model switches the iteration model of an executor built once for several models.
func (e *Executor) model(m Model) *Executor {
	e.opt.Model = m
	return e
}

var everyShipStore = map[kind.Kind]bool{kind.Neo4j: true, kind.MySQL: true, kind.Mongo: true,
	kind.ClickHouse: true, kind.LevelDB: true, kind.RocksDB: true}
