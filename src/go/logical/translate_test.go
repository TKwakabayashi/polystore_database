package logical_test

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/stats"
	"polystore_database/src/go/workload"
)

var update = flag.Bool("update", false, "golden ファイルを再生成する")

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
		t.Fatalf("golden %s が無い（-update で生成）: %v", path, err)
	}
	if string(want) != got {
		t.Errorf("golden %s と不一致\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func translate(t *testing.T, q string) string {
	t.Helper()
	a, err := ast.Parse(q, map[string]string{"x": "1"})
	if err != nil {
		t.Fatalf("Parse(%q): %v", q, err)
	}
	n, err := logical.Translate(a)
	if err != nil {
		t.Fatalf("Translate(%q): %v", q, err)
	}
	return logical.Dump(n)
}

// 全ワークロードの論理計画を golden で固定する。
func TestTranslateWorkloads(t *testing.T) {
	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Registry[n](migrator.ModeGraphToRdb, false)
		a, err := ast.Parse(cypher, params)
		if errors.Is(err, ast.ErrUnsupported) {
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		plan, err := logical.Translate(a)
		b.WriteString("### " + n + "\n")
		switch {
		case err == nil:
			b.WriteString(logical.Dump(plan))
		case errors.Is(err, ast.ErrUnsupported):
			b.WriteString("UNSUPPORTED: " + err.Error() + "\n")
		default:
			t.Errorf("%s: %v", n, err)
		}
		b.WriteString("\n")
	}
	golden(t, "workloads.golden", b.String())
}

// 統計（catalog/stats.json。ldbc_sf1 で収集）を使った計画のうち、規則の計画と違うものだけを golden で固定する。
func TestTranslateWorkloadsCost(t *testing.T) {
	st, err := stats.Load("../../../catalog/stats.json")
	if err != nil || st == nil {
		t.Skipf("catalog/stats.json が無い（-mode stats で収集）: %v", err)
	}
	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Registry[n](migrator.ModeGraphToRdb, false)
		a, err := ast.Parse(cypher, params)
		if err != nil {
			continue
		}
		rule, err := logical.Translate(a)
		if err != nil {
			continue
		}
		cost, err := logical.TranslateWith(a, st)
		if err != nil {
			t.Errorf("%s: %v", n, err)
			continue
		}
		if logical.Dump(cost) != logical.Dump(rule) {
			b.WriteString("### " + n + "\n" + logical.Dump(cost) + "\n")
		}
	}
	golden(t, "workloads_cost.golden", b.String())
}

// 代表ケース（循環・逆向き・直積・WITH・パス条件）を golden で固定する。
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

func TestTranslateUnsupported(t *testing.T) {
	for _, q := range []string{
		"MATCH (a)-[r:R*1..2]->(b) RETURN r",                             // 区間変数を値として
		"MATCH (a)-[:R]->(b) WITH a.x AS v RETURN v.y",                   // 値へのプロパティ参照
		"MATCH (a) RETURN b",                                             // 未束縛変数
		"MATCH (a)-[:R]->(b) RETURN a.x ORDER BY b.y DESC, count(*) ASC", // 集約は ORDER BY 単独不可（構文上）
	} {
		a, err := ast.Parse(q, nil)
		if err != nil {
			continue // 構文段階で落ちるものは対象外
		}
		if _, err := logical.Translate(a); !errors.Is(err, ast.ErrUnsupported) {
			t.Errorf("%q は ErrUnsupported になるべき: %v", q, err)
		}
	}
}

// 論理書き換え（遅延実体化・刈り込み）後の計画を golden で固定する。
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
		n, err := logical.Translate(a)
		if err != nil {
			t.Fatal(err)
		}
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + logical.Dump(logical.Rewrite(n)) + "\n")
	}
	golden(t, "rewrite.golden", b.String())
}
