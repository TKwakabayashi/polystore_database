package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"polystore_database/src/go/id"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
)

// ================================
// 式の評価（Cypher の null 意味論）
// ================================

func col(k string) logical.Expr { return &logical.Column{Key: k} }
func lit(v any) logical.Expr    { return &logical.Lit{Val: v} }
func bin(op string, l, r logical.Expr) logical.Expr {
	return &logical.Bin{Op: op, L: l, R: r}
}

func TestEvalNullSemantics(t *testing.T) {
	r := Row{"a": int64(1), "n": nil, "f": 1.0, "s": "x"}
	cases := []struct {
		e    logical.Expr
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
		{bin("IN", col("a"), &logical.List{Elems: []logical.Expr{lit(int64(2)), lit(nil)}}), nil},
		{bin("IN", col("a"), &logical.List{Elems: []logical.Expr{lit(int64(1)), lit(nil)}}), true},
		{&logical.Un{Op: "IS NULL", X: col("n")}, true},
		{&logical.Un{Op: "NOT", X: col("n")}, nil},
		{bin("+", col("a"), lit(int64(2))), int64(3)},
		{bin("/", col("f"), lit(int64(4))), 0.25},
		{&logical.Call{Fn: "coalesce", Args: []logical.Expr{col("n"), col("s")}}, "x"},
		{&logical.Case{Whens: []logical.CaseWhen{{When: col("n"), Then: lit(int64(1))}}, Else: lit(int64(0))}, int64(0)},
	}
	for _, c := range cases {
		got, err := Eval(c.e, r)
		if err != nil {
			t.Fatalf("%s: %v", c.e, err)
		}
		if got != c.want {
			t.Errorf("%s = %v (%T), want %v", c.e, got, got, c.want)
		}
	}
}

func TestEvalQuantifiers(t *testing.T) {
	r := Row{
		"p":         &ListVal{Nodes: []id.UUID{"n1", "n2", "n3"}},
		"p.nodes.g": []any{"m", nil, "f"},
		"q":         &ListVal{Nodes: []id.UUID{"n1", "n3"}},
		"q.nodes.g": []any{"m", "f"},
	}
	q := func(kind, src string) logical.Expr {
		return &logical.Quant{Kind: kind, Var: "x",
			List: &logical.Call{Fn: "nodes", Args: []logical.Expr{col(src)}},
			Pred: bin("=", &logical.ElemProp{Var: "x", Prop: "g"}, lit("m"))}
	}
	cases := []struct {
		kind, src string
		want      any
	}{
		{"any", "p", true}, {"all", "p", false}, {"none", "p", false}, {"single", "p", nil},
		{"single", "q", true}, {"all", "q", false},
	}
	for _, c := range cases {
		got, err := Eval(q(c.kind, c.src), r)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s over %s = %v, want %v", c.kind, c.src, got, c.want)
		}
	}
}

// ================================
// コーディネータの演算子
// ================================

func TestAggregateSemantics(t *testing.T) {
	rows := []Row{
		{"k": "a", "v": int64(1)}, {"k": "a", "v": nil}, {"k": "a", "v": int64(1)},
		{"k": "b", "v": int64(4)},
	}
	agg := &logical.Aggregation{
		Keys: []logical.ProjItem{{Name: "k", Expr: col("k")}},
		Aggs: []logical.AggItem{
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
	// キー無しの集約は空入力でも 1 行（count = 0、avg/max = null）。
	empty, err := runCoord(&logical.Aggregation{Aggs: agg.Aggs}, [][]Row{nil})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 1 || empty[0]["all"] != int64(0) || empty[0]["avg"] != nil || empty[0]["sum"] != int64(0) {
		t.Errorf("empty aggregation: %v", empty)
	}
}

func TestSortLimitDistinct(t *testing.T) {
	rows := []Row{{"v": int64(2)}, {"v": nil}, {"v": int64(1)}, {"v": int64(2)}}
	asc, _ := runCoord(&logical.Sort{Keys: []logical.SortKey{{Expr: col("v")}}}, [][]Row{rows})
	desc, _ := runCoord(&logical.Sort{Keys: []logical.SortKey{{Expr: col("v"), Desc: true}}}, [][]Row{rows})
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
	lim, _ := runCoord(&logical.Limit{Skip: 1, Count: 2}, [][]Row{asc})
	if vals(lim) != "2,2" {
		t.Errorf("limit: %s", vals(lim))
	}
	d, _ := runCoord(&logical.Distinct{}, [][]Row{asc})
	if vals(d) != "1,2,<nil>" {
		t.Errorf("distinct: %s", vals(d))
	}
}

func TestHashJoinAndRelUnique(t *testing.T) {
	l := []Row{{"a": id.UUID("1")}, {"a": id.UUID("2")}}
	r := []Row{{"a": id.UUID("1"), "b": id.UUID("x")}, {"a": id.UUID("1"), "b": id.UUID("y")}}
	if got := hashJoin(l, r, []string{"a"}); len(got) != 2 {
		t.Errorf("keyed join: %d rows", len(got))
	}
	if got := hashJoin(l, r, nil); len(got) != 4 {
		t.Errorf("cartesian: %d rows", len(got))
	}
	rows := []Row{
		{"r1": id.UUID("e1"), "r2": id.UUID("e2"), "_seg_s": &ListVal{Rels: []id.UUID{"e3"}}},
		{"r1": id.UUID("e1"), "r2": id.UUID("e2"), "_seg_s": &ListVal{Rels: []id.UUID{"e1"}}},
		{"r1": id.UUID("e1"), "r2": id.UUID("e1"), "_seg_s": &ListVal{}},
	}
	got, _ := runCoord(&logical.RelUniqueFilter{Rels: []string{"r1", "r2"}, Segs: []string{"s"}}, [][]Row{rows})
	if len(got) != 1 {
		t.Errorf("rel uniqueness kept %d rows, want 1", len(got))
	}
}

// ================================
// 配置によらず同じ結果になること（メモリ上のグラフ）
// ================================

// socialGraph は 6 人の KNOWS グラフ（三角形 p1-p2-p3 を含む）と所在地。
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

func loadMapping(t *testing.T) *schema.MappingDictionary {
	t.Helper()
	md, err := schema.LoadMappingDictionary("../../../catalog/mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func cloneMD(t *testing.T, md *schema.MappingDictionary) *schema.MappingDictionary {
	b, err := json.Marshal(md)
	if err != nil {
		t.Fatal(err)
	}
	var out schema.MappingDictionary
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return &out
}

func placementsFor(t *testing.T) map[string]*schema.MappingDictionary {
	base := loadMapping(t)
	out := map[string]*schema.MappingDictionary{"graph": base}
	mixed := cloneMD(t, base)
	pp := mixed.Entities["Person"].Properties
	pp["firstName"].Databases = []string{"relational"}
	pp["lastName"].Databases = []string{"relational"}
	pp["gender"].Databases = []string{"document"}
	pp["birthday"].Databases = []string{"columnar"}
	mixed.Relationships["KNOWS"].Properties["creationDate"].Databases = []string{"relational"}
	out["mixed"] = mixed
	for _, k := range []string{"relational", "document", "columnar", "kvs"} {
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

func runMem(t *testing.T, g *memGraph, md *schema.MappingDictionary, q string) []map[string]any {
	t.Helper()
	p, cat, err := Plan(q, map[string]string{"x": "1"}, md)
	if err != nil {
		t.Fatalf("Plan(%q): %v", q, err)
	}
	ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf)
	rows, err := ex.Execute(p)
	if err != nil {
		t.Fatalf("Execute(%q): %v", q, err)
	}
	return OutputRows(p, rows)
}

var equivalenceQueries = []struct {
	name string
	q    string
	want int // 期待行数（-1 は確認しない）
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
}

func TestPlacementEquivalence(t *testing.T) {
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

// 手で確かめた結果（意味論の取り違えを検出する）。
func TestHandCheckedResults(t *testing.T) {
	g := socialGraph()
	md := loadMapping(t)
	cases := []struct {
		q    string
		want []string
	}{
		// p1 の友人（向き無し）は p2, p3, p5 で、いずれも male
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WHERE f.gender = 'male' RETURN f.id",
			[]string{"f.id=2;", "f.id=3;", "f.id=5;"}},
		// 三角形 p1→p2→p3→p1
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id",
			[]string{"b.id=2;c.id=3;"}},
		// 関係一意性: p1-p2-? で同じ辺を戻る p1 は出ない（p2 経由: p3, p4 / p3 経由: p2 / p5 経由: p4, p6）
		{"MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person)-[:KNOWS]-(c:Person) RETURN c.id",
			[]string{"c.id=2;", "c.id=3;", "c.id=4;", "c.id=4;", "c.id=6;"}},
		// 0 ホップは自分自身を含む
		{"MATCH (p:Person {id: $x})-[:KNOWS*0..1]->(f:Person) RETURN f.id",
			[]string{"f.id=1;", "f.id=2;", "f.id=5;"}},
		// 経路ごとに 1 行（p1→p2→p4 と p1→p5→p6 … 2 ホップの到達先）
		{"MATCH (p:Person {id: $x})-[:KNOWS*2]->(f:Person) RETURN f.id",
			[]string{"f.id=3;", "f.id=4;", "f.id=6;"}},
		// OPTIONAL MATCH: p5, p6 は所在地が無いので null で残る
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.id, c.name",
			[]string{"c.name=Tokyo;p.id=1;", "c.name=Kyoto;p.id=2;", "c.name=Tokyo;p.id=3;", "c.name=Kyoto;p.id=4;", "c.name=<nil>;p.id=5;", "c.name=<nil>;p.id=6;"}},
		// count(f) は一致しなかった行（f = null）を数えない: 出次数 p1:2 p2:2 p3:1 p4:1 p5:1 p6:0
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) RETURN p.id, count(f) AS n",
			[]string{"n=2;p.id=1;", "n=2;p.id=2;", "n=1;p.id=3;", "n=1;p.id=4;", "n=1;p.id=5;", "n=0;p.id=6;"}},
		// WHERE は一致の条件: female の友人がいない人も null で残る（female は p1, p4, p6: p2→p4, p3→p1, p5→p6）
		{"MATCH (p:Person) OPTIONAL MATCH (p)-[:KNOWS]->(f:Person) WHERE f.gender = 'female' RETURN p.id, f.id",
			[]string{"f.id=<nil>;p.id=1;", "f.id=4;p.id=2;", "f.id=1;p.id=3;", "f.id=<nil>;p.id=4;", "f.id=6;p.id=5;", "f.id=<nil>;p.id=6;"}},
		// 単純 CASE の WHEN null は一致しない（Neo4j と同じ）ので back は常に true
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person) OPTIONAL MATCH (b)-[r:KNOWS]->(a) RETURN b.id, CASE r WHEN null THEN false ELSE true END AS back",
			[]string{"b.id=2;back=true;", "b.id=5;back=true;"}},
		// collect した友人（p2, p3, p5）を UNWIND して所在地へ（p5 は所在地なし）
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f) AS fs UNWIND fs AS g MATCH (g)-[:IS_LOCATED_IN]->(c:Place) RETURN g.id, c.name",
			[]string{"c.name=Kyoto;g.id=2;", "c.name=Tokyo;g.id=3;"}},
		// IN の右辺がリスト列
		{"MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH collect(f.id) AS ids MATCH (q:Person) WHERE q.id IN ids RETURN q.id",
			[]string{"q.id=2;", "q.id=3;", "q.id=5;"}},
		// 2 ホップ以内の到達者は経路で重複する（重み付き行）が、collect(DISTINCT) で 1 回ずつ: {p2, p3, p4, p5, p6}
		// （p1 へ戻るには同じ辺を 2 回使うので関係一意性で除かれる）
		{"MATCH (p:Person {id: $x})-[:KNOWS*1..2]-(f:Person) WITH collect(DISTINCT f) AS fs UNWIND fs AS g RETURN count(g) AS n",
			[]string{"n=5;"}},
		// 両端の起点から展開して x で合流: 1-2-4 と 1-5-4
		{"MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id",
			[]string{"x.id=2;", "x.id=5;"}},
		// 束縛済みの a と起点 b の両側から: 1-3-2-4 だけ（1-?-5-4 は 1 と 5 の共通の隣が無い）
		{"MATCH (a:Person {id: $x}) WITH a MATCH (a)-[:KNOWS]-(x:Person)-[:KNOWS]-(y:Person)-[:KNOWS]-(b:Person {id: 4}) RETURN x.id, y.id",
			[]string{"x.id=3;y.id=2;"}},
		// 両端束縛の可変長: a→b の辺は使用済みなので、b=2 は 1-3-2、b=5 は経路なし
		{"MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person), (a)-[:KNOWS*1..2]-(b) RETURN b.id",
			[]string{"b.id=2;"}},
		// 集約の後の ORDER BY / LIMIT（WITH の途中）
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

// 起点ごとに分割して合流する計画（settings.SplitAnchoredPatterns）と、1 つの起点から伸ばす計画が同じ結果を返す。
func TestSplitAnchoredEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	prev := settings.SplitAnchoredPatterns
	defer func() { settings.SplitAnchoredPatterns = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			settings.SplitAnchoredPatterns = false
			off := canon(runMem(t, g, md, c.q))
			settings.SplitAnchoredPatterns = true
			on := canon(runMem(t, g, md, c.q))
			if strings.Join(on, "\n") != strings.Join(off, "\n") {
				t.Errorf("%s/%s: split differs\n--- off ---\n%s\n--- on ---\n%s", c.name, pname, strings.Join(off, "\n"), strings.Join(on, "\n"))
			}
		}
	}
}

// 反復モデルを変えても結果は同じ（全クエリ × 全配置 × 5 モデル）。並列モデルの競合は -race で検出する。
func TestModelEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	models := []settings.EngineKind{settings.EngineVolcano, settings.EngineVectorized, settings.EngineStream, settings.EngineVecStream}
	prev := settings.VectorSize
	settings.VectorSize = 2 // バッチ境界をまたぐ経路を通すため小さくする
	defer func() { settings.VectorSize = prev }()
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md)
			if err != nil {
				t.Fatalf("Plan(%q): %v", c.q, err)
			}
			ex := New(context.Background(), &memBackend{g: g}, cat.TypeOf)
			ref, err := ex.Execute(p)
			if err != nil {
				t.Fatalf("bulk %s/%s: %v", c.name, pname, err)
			}
			want := canon(OutputRows(p, ref))
			for _, m := range models {
				rows, err := ex.ExecuteModel(p, m)
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

// Sort の後の Limit は、並列モデルでも並べ替えた順の先頭を返す。
func TestModelOrderAfterSort(t *testing.T) {
	g := socialGraph()
	md := loadMapping(t)
	q := "MATCH (p:Person) RETURN p.id ORDER BY p.id DESC LIMIT 3"
	p, cat, err := Plan(q, nil, md)
	if err != nil {
		t.Fatal(err)
	}
	prev := settings.VectorSize
	settings.VectorSize = 1
	defer func() { settings.VectorSize = prev }()
	for _, m := range []settings.EngineKind{settings.EngineBulk, settings.EngineVolcano, settings.EngineVectorized, settings.EngineStream, settings.EngineVecStream} {
		rows, err := New(context.Background(), &memBackend{g: g}, cat.TypeOf).ExecuteModel(p, m)
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

// failingBackend は n 回目以降のストアへの発行を失敗させる。
type failingBackend struct {
	memBackend
	mu    sync.Mutex
	calls int
	after int
}

func (b *failingBackend) RunLogical(q *physical.StoreQuery, in []Key) ([]Row, error) {
	b.mu.Lock()
	b.calls++
	fail := b.calls > b.after
	b.mu.Unlock()
	if fail {
		return nil, fmt.Errorf("injected store failure")
	}
	return b.memBackend.RunLogical(q, in)
}

// ストアの失敗はどのモデルでもエラーとして返る（並列モデルが固まったり、途中の結果を返したりしない）。
func TestModelErrorPropagation(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"
	p, cat, err := Plan(q, map[string]string{"x": "1"}, pls["mixed"])
	if err != nil {
		t.Fatal(err)
	}
	prev := settings.VectorSize
	settings.VectorSize = 1
	defer func() { settings.VectorSize = prev }()
	for _, m := range []settings.EngineKind{settings.EngineBulk, settings.EngineVolcano, settings.EngineVectorized, settings.EngineStream, settings.EngineVecStream} {
		done := make(chan error, 1)
		go func() {
			_, err := New(context.Background(), &failingBackend{memBackend: memBackend{g: g}, after: 1}, cat.TypeOf).ExecuteModel(p, m)
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

// ---- 重み付きの行 ----

func w(r Row, n int64) Row { setWeight(r, n); return r }

func TestWeightedCoordinator(t *testing.T) {
	rows := []Row{
		w(Row{"k": "a", "v": int64(2)}, 3),
		w(Row{"k": "a", "v": nil}, 2),
		Row{"k": "b", "v": int64(5)},
	}
	agg := &logical.Aggregation{
		Keys: []logical.ProjItem{{Name: "k", Expr: col("k")}},
		Aggs: []logical.AggItem{
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

	// Limit は重みを数え、境目の行は重みを削る: [a×3, b×2] の skip 1 limit 3 → a×2, b×1
	lim, _ := runCoord(&logical.Limit{Skip: 1, Count: 3}, [][]Row{{w(Row{"x": "a"}, 3), w(Row{"x": "b"}, 2)}})
	if len(lim) != 2 || weightOf(lim[0]) != 2 || weightOf(lim[1]) != 1 {
		t.Errorf("weighted limit: %v", lim)
	}
	// 入力の重みは変えない（Limit が複製してから削る）
	src := []Row{w(Row{"x": "a"}, 3)}
	runCoord(&logical.Limit{Count: 1}, [][]Row{src})
	if weightOf(src[0]) != 3 {
		t.Error("limit mutated its input")
	}

	// Distinct は重みを無視して 1 行にする
	d, _ := runCoord(&logical.Distinct{}, [][]Row{{w(Row{"x": "a"}, 3), Row{"x": "a"}}})
	if len(d) != 1 || weightOf(d[0]) != 1 {
		t.Errorf("weighted distinct: %v", d)
	}
	// 結合は重みを掛ける
	j := hashJoin([]Row{w(Row{"a": id.UUID("1")}, 2)}, []Row{w(Row{"a": id.UUID("1"), "b": "x"}, 3)}, []string{"a"})
	if len(j) != 1 || weightOf(j[0]) != 6 {
		t.Errorf("weighted join: %v", j)
	}
	// 最後に展開する
	p := &physical.Coord{Op: &logical.Projection{Items: []logical.ProjItem{{Name: "x", Expr: col("x")}}}}
	if got := OutputRows(p, []Row{w(Row{"x": "a"}, 3)}); len(got) != 3 {
		t.Errorf("output expansion: %d rows", len(got))
	}
}

// 可変長を含む graph のクエリが重み付きになる配置で、重み無し（settings.WeightedRows=false）と結果が同じ。
func TestWeightedRowsEquivalence(t *testing.T) {
	g := socialGraph()
	pls := placementsFor(t)
	weighted := 0
	for _, c := range equivalenceQueries {
		for pname, md := range pls {
			run := func(on bool) []string {
				prev := settings.WeightedRows
				settings.WeightedRows = on
				defer func() { settings.WeightedRows = prev }()
				p, cat, err := Plan(c.q, map[string]string{"x": "1"}, md)
				if err != nil {
					t.Fatalf("Plan: %v", err)
				}
				if on && strings.Contains(physical.Dump(p), "weighted") {
					weighted++
				}
				rows, err := New(context.Background(), &memBackend{g: g}, cat.TypeOf).ExecuteModel(p, settings.EngineVecStream)
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
