package ast_test

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/migrator"
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

func mustParse(t *testing.T, q string, params map[string]string) *ast.Query {
	t.Helper()
	got, err := ast.Parse(q, params)
	if err != nil {
		t.Fatalf("Parse(%q): %v", q, err)
	}
	return got
}

// 全ワークロード: 構造化できるものは AST を golden で固定し、できないものは ErrUnsupported で落ちること。
func TestWorkloadCorpus(t *testing.T) {
	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Registry[n](migrator.ModeGraphToRdb, false)
		q, err := ast.Parse(cypher, params)
		switch {
		case err == nil:
			b.WriteString("### " + n + "\n" + q.String() + "\n")
		case errors.Is(err, ast.ErrUnsupported):
			b.WriteString("### " + n + "\nUNSUPPORTED\n\n")
		default:
			t.Errorf("%s: ErrUnsupported 以外のエラー: %v", n, err)
		}
	}
	golden(t, "workloads.golden", b.String())
}

func TestRanges(t *testing.T) {
	cases := map[string]string{
		"*":     "*",
		"*2":    "*2",
		"*1..3": "*1..3",
		"*2..":  "*2..",
		"*..3":  "*..3",
		"*0..":  "*0..",
		"*0..2": "*0..2",
	}
	for lit, want := range cases {
		q := mustParse(t, "MATCH (a)-[:R"+lit+"]->(b) RETURN b", nil)
		r := q.Clauses[0].(*ast.MatchClause).Patterns[0].Rels[0]
		if r.Range == nil || r.Range.Cypher() != want {
			t.Errorf("%s: got %+v", lit, r.Range)
		}
	}
}

func TestAnonymousNamesAndDirections(t *testing.T) {
	q := mustParse(t, "MATCH (:Person)<-[:KNOWS]-(b)-[r]-(:Tag) RETURN b", nil)
	p := q.Clauses[0].(*ast.MatchClause).Patterns[0]
	if p.Nodes[0].Var != "_n0" || p.Nodes[2].Var != "_n1" || p.Rels[0].Var != "_r0" || p.Rels[1].Var != "r" {
		t.Errorf("名前付けが想定外: %s", p)
	}
	if p.Rels[0].Dir != ast.DirLeft || p.Rels[1].Dir != ast.DirBoth {
		t.Errorf("向きが想定外: %s", p)
	}
}

func TestParamsAndItemNames(t *testing.T) {
	q := mustParse(t, "MATCH (p:Person {id: $id, name: $n}) RETURN p.id, count(*) AS c, coalesce(p.a, p.b)",
		map[string]string{"id": "42", "n": "Jan"})
	m := q.Clauses[0].(*ast.MatchClause)
	if got := m.Patterns[0].Nodes[0].Props[0].Value.String(); got != "42" {
		t.Errorf("数値パラメータ: %s", got)
	}
	if got := m.Patterns[0].Nodes[0].Props[1].Value.String(); got != `"Jan"` {
		t.Errorf("文字列パラメータ: %s", got)
	}
	r := q.Clauses[1].(*ast.ProjectionClause)
	names := []string{r.Items[0].Name, r.Items[1].Name, r.Items[2].Name}
	if strings.Join(names, "|") != "p.id|c|coalesce(p.a, p.b)" {
		t.Errorf("項目名: %v", names)
	}
}

func TestMissingParam(t *testing.T) {
	if _, err := ast.Parse("MATCH (p {id: $x}) RETURN p", nil); err == nil || errors.Is(err, ast.ErrUnsupported) {
		t.Errorf("未指定パラメータは通常エラーになるべき: %v", err)
	}
}

func TestWithClause(t *testing.T) {
	q := mustParse(t, "MATCH (p)-[:KNOWS]->(f) WITH DISTINCT p, count(f) AS deg ORDER BY deg DESC LIMIT 5 WHERE deg > 2 RETURN p.id", nil)
	w := q.Clauses[1].(*ast.ProjectionClause)
	if w.IsReturn || !w.Distinct || len(w.OrderBy) != 1 || w.Limit == nil || *w.Limit != 5 || w.Where == nil {
		t.Errorf("WITH の写し取りが想定外:\n%s", q)
	}
}

func TestUnsupported(t *testing.T) {
	for _, q := range []string{
		"MATCH (a) UNWIND [1, 2] AS x RETURN x",
		"MATCH (a) RETURN a.b.c",
	} {
		if _, err := ast.Parse(q, nil); !errors.Is(err, ast.ErrUnsupported) {
			t.Errorf("%q は ErrUnsupported になるべき: %v", q, err)
		}
	}
}
