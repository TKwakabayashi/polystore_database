package cypher_test

import (
	"errors"
	"strings"
	"testing"

	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/logical"
)

// An unknown variable is an error in the query, not a form the planner cannot handle.
func TestTranslateSemanticError(t *testing.T) {
	a, err := ast.Parse("MATCH (a) RETURN b", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cypher.Translate(a); !errors.Is(err, ast.ErrSemantic) {
		t.Errorf("want ErrSemantic: %v", err)
	}
}

func TestTranslateUnsupported(t *testing.T) {
	for _, q := range []string{
		"MATCH (a)-[r:R*1..2]->(b) RETURN r",                             // a segment variable as a value
		"MATCH (a)-[:R]->(b) WITH a.x AS v RETURN v.y",                   // a property of a value
		"MATCH (a)-[:R]->(b) RETURN a.x ORDER BY b.y DESC, count(*) ASC", // an aggregate cannot stand alone in ORDER BY (grammar)
		"MATCH (a) ((x)-[:R]->(y)-[:S]->(z)){1,2} (b) RETURN b",          // a group of two relationships
		"MATCH (a) ((x)-[:R]->(y) WHERE x.v < y.v){1,2} (b) RETURN b",    // a condition across group variables
		"MATCH (a) ((x:L)-[:R]->(y)){0,2} (b) RETURN b",                  // lower bound 0 with a node condition
		"MATCH (a) ((x)-[r:R]->(y)){1,2} (b) RETURN size(r) AS n",        // a group variable outside its group
		"MATCH (a)-[r:R*1..2 WHERE r.v = 1]->(b) RETURN b",               // WHERE on a variable-length relationship
	} {
		a, err := ast.Parse(q, nil)
		if err != nil {
			continue // queries that fail to parse are out of scope
		}
		if _, err := cypher.Translate(a); !errors.Is(err, ast.ErrUnsupported) {
			t.Errorf("%q must be ErrUnsupported: %v", q, err)
		}
	}
}

// A quantified path pattern is a variable-length edge: conditions on x hold on the left node and the inner nodes,
// conditions on y on the inner nodes and the right node, conditions on r on every relationship.
func TestTranslateQuantified(t *testing.T) {
	a, err := ast.Parse("MATCH (a {id: 1}) ((x:P WHERE x.u = 1)-[r:R WHERE r.w = 2]->(y:P) WHERE y.v = 3){1,3} (b) RETURN b.id", nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := cypher.Translate(a)
	if err != nil {
		t.Fatal(err)
	}
	got := ir.Dump(logical.PlanPatterns(p, nil, logical.Options{}))
	for _, want := range []string{
		"_q0:R*1..3", "inner:P",
		"all(x IN nodes WHERE ((x.u = 1) AND (x.v = 3)))", "all(r IN rels WHERE (r.w = 2))",
		"(a.u = 1)", "(b.v = 3)", "a:P", "b:P",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}
