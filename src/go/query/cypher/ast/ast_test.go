package ast_test

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query/cypher/ast"
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

func mustParse(t *testing.T, q string, params map[string]string) *ast.Query {
	t.Helper()
	got, err := ast.Parse(q, params)
	if err != nil {
		t.Fatalf("Parse(%q): %v", q, err)
	}
	return got
}

// Every workload: the AST is pinned by a golden; queries the grammar rejects fail with ErrSyntax, unrepresentable ones with ErrUnsupported.
func TestWorkloadCorpus(t *testing.T) {
	var names []string
	for _, n := range workload.Names(workload.LDBC) {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, n := range names {
		cypher, params, _ := workload.Def(n)(migrate.ModeGraphToRdb, false)
		q, err := ast.Parse(cypher, params)
		switch {
		case err == nil:
			b.WriteString("### " + n + "\n" + q.String() + "\n")
		case errors.Is(err, ast.ErrSyntax):
			b.WriteString("### " + n + "\nSYNTAX ERROR (the grammar does not cover it yet)\n\n")
		case errors.Is(err, ast.ErrUnsupported):
			b.WriteString("### " + n + "\nUNSUPPORTED\n\n")
		default:
			t.Errorf("%s: unexpected error: %v", n, err)
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
		t.Errorf("unexpected naming: %s", p)
	}
	if p.Rels[0].Dir != ast.DirLeft || p.Rels[1].Dir != ast.DirBoth {
		t.Errorf("unexpected direction: %s", p)
	}
}

func TestParamsAndItemNames(t *testing.T) {
	q := mustParse(t, "MATCH (p:Person {id: $id, name: $n}) RETURN p.id, count(*) AS c, coalesce(p.a, p.b)",
		map[string]string{"id": "42", "n": "Jan"})
	m := q.Clauses[0].(*ast.MatchClause)
	if got := m.Patterns[0].Nodes[0].Props[0].Value.String(); got != "42" {
		t.Errorf("numeric parameter: %s", got)
	}
	if got := m.Patterns[0].Nodes[0].Props[1].Value.String(); got != `"Jan"` {
		t.Errorf("string parameter: %s", got)
	}
	r := q.Clauses[1].(*ast.ProjectionClause)
	names := []string{r.Items[0].Name, r.Items[1].Name, r.Items[2].Name}
	if strings.Join(names, "|") != "p.id|c|coalesce(p.a, p.b)" {
		t.Errorf("item names: %v", names)
	}
}

func TestMissingParam(t *testing.T) {
	if _, err := ast.Parse("MATCH (p {id: $x}) RETURN p", nil); err == nil || errors.Is(err, ast.ErrUnsupported) {
		t.Errorf("an unset parameter must be a plain error: %v", err)
	}
}

func TestWithClause(t *testing.T) {
	q := mustParse(t, "MATCH (p)-[:KNOWS]->(f) WITH DISTINCT p, count(f) AS deg ORDER BY deg DESC LIMIT 5 WHERE deg > 2 RETURN p.id", nil)
	w := q.Clauses[1].(*ast.ProjectionClause)
	if w.IsReturn || !w.Distinct || len(w.OrderBy) != 1 || w.Limit == nil || *w.Limit != 5 || w.Where == nil {
		t.Errorf("unexpected WITH mapping:\n%s", q)
	}
}

func TestUnsupported(t *testing.T) {
	for _, q := range []string{
		"MATCH (a) RETURN a.b.c + 1",
		"MATCH (a) RETURN a.b.c",
	} {
		if _, err := ast.Parse(q, nil); !errors.Is(err, ast.ErrUnsupported) {
			t.Errorf("%q must be ErrUnsupported: %v", q, err)
		}
	}
}

// Quantified relationships and quantified path patterns become a step with a group; an omitted lower bound is 0.
func TestQuantifiedPatterns(t *testing.T) {
	cases := map[string]string{
		"(a)-[:R]->{1,6}(b)":          "(a) ((_n0)-[_r0:R]->(_n1)){1,6} (b)",
		"(a)<-[r:R]-{2}(b)":           "(a) ((_n0)<-[r:R]-(_n1)){2} (b)",
		"(a)-[:R]-+(b)":               "(a) ((_n0)-[_r0:R]-(_n1)){1,} (b)",
		"(a)-[:R]-*(b)":               "(a) ((_n0)-[_r0:R]-(_n1)){0,} (b)",
		"(a) ((x)-[:R]->(y)){,3} (b)": "(a) ((x)-[_r0:R]->(y)){0,3} (b)",
		"(a) ((x:L {k: 1})-[r:R]->(y WHERE y.v > 2) WHERE r.w = 3){2,} (b)": "(a) ((x:L {k: 1})-[r:R]->(y WHERE (y.v > 2)) WHERE (r.w = 3)){2,} (b)",
		"(a WHERE a.v = 1)-[r:R WHERE r.w = 2]->(b)":                        "(a WHERE (a.v = 1))-[r:R WHERE (r.w = 2)]->(b)",
	}
	for pat, want := range cases {
		q := mustParse(t, "MATCH "+pat+" RETURN b", nil)
		if got := q.Clauses[0].(*ast.MatchClause).Patterns[0].String(); got != want {
			t.Errorf("%s:\n got %s\nwant %s", pat, got, want)
		}
	}
	q := mustParse(t, "MATCH (a)-[:R]->{1,6}(b) RETURN b", nil)
	r := q.Clauses[0].(*ast.MatchClause).Patterns[0].Rels[0]
	if r.Var != "_q0" || r.Quantified == nil || *r.Quantified.Range.Min != 1 || *r.Quantified.Range.Max != 6 {
		t.Errorf("quantified step: %+v", r)
	}
	// Neo4j does not allow a variable-length relationship inside a quantified pattern either.
	if _, err := ast.Parse("MATCH (a) ((x)-[:R*1..2]->(y)){1,2} (b) RETURN b", nil); !errors.Is(err, ast.ErrSyntax) {
		t.Errorf("variable length in a group must be ErrSyntax: %v", err)
	}
}
