package planner_test

import (
	"strings"
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
)

const caseMapping = "../../../catalog/mapping.json"

func caseParse(t *testing.T, q string) plan.PlanNode {
	t.Helper()
	root, err := planner.ParseQuery(q, caseMapping, map[string]string{"x": "1"})
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	return root
}

// 単段 RETURN の CASE 式はエンジンで評価される（verbatim ではなく structured プラン）。
// Projection が参照プロパティ（m.length）を材料化し、Return が Expr を評価する。
func TestCaseInReturnStructured(t *testing.T) {
	q := "MATCH (m:Message {id: $x}) RETURN m.id, CASE WHEN m.length > 10 THEN 'long' ELSE 'short' END AS bucket"
	root := caseParse(t, q)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	dump := b.String()
	t.Logf("\n%s", dump)
	if !strings.Contains(dump, "length") {
		t.Errorf("expected m.length materialized for CASE\n%s", dump)
	}
	if !strings.Contains(dump, "bucket") {
		t.Errorf("expected RETURN bucket\n%s", dump)
	}
}

// WHEN が変数間比較（a.id = b.id）の CASE は structured プランになる（T3）。
// 両辺の参照プロパティ（a.id / b.id）が Projection で材料化され、Return が Expr を評価する。
func TestCaseVarCmpStructured(t *testing.T) {
	q := "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person) RETURN CASE WHEN a.id = b.id THEN 1 ELSE 0 END AS same"
	root := caseParse(t, q)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	dump := b.String()
	t.Logf("\n%s", dump)
	if !strings.Contains(dump, "same") {
		t.Errorf("expected RETURN same\n%s", dump)
	}
}
