package planner_test

import (
	"strings"
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
)

func veParse(t *testing.T, q string) plan.PlanNode {
	t.Helper()
	root, err := planner.ParseQuery(q, caseMapping, map[string]string{"x": "1"})
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	return root
}

// 算術式 RETURN（a.length + 1）は structured プランになり、参照プロパティが材料化される（T1）。
func TestValueExprArithmetic(t *testing.T) {
	q := "MATCH (m:Message {id: $x}) RETURN m.id, m.length + 1 AS plus"
	root := veParse(t, q)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	dump := b.String()
	t.Logf("\n%s", dump)
	for _, want := range []string{"length", "plus"} {
		if !strings.Contains(dump, want) {
			t.Errorf("expected %q in plan\n%s", want, dump)
		}
	}
}

// 対応関数（size）を使った RETURN は structured。
func TestValueExprSupportedFunc(t *testing.T) {
	q := "MATCH (m:Message {id: $x}) RETURN toInteger(m.length) AS li"
	root := veParse(t, q)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	if !strings.Contains(b.String(), "li") {
		t.Errorf("expected li in plan\n%s", b.String())
	}
}

// 未対応関数は verbatim へフォールバック（誤った null を返さない）。
func TestValueExprUnsupportedFuncFallsBack(t *testing.T) {
	q := "MATCH (m:Message {id: $x}) RETURN substring(m.content) AS s"
	root := veParse(t, q)
	f, ok := root.(*plan.StoreFragment)
	if !ok || f.Verbatim == "" {
		var b strings.Builder
		dumpQE(root, 0, &b)
		t.Fatalf("expected verbatim fallback, got:\n%s", b.String())
	}
}

// 単純プロパティ参照（m.id）は従来どおり材料化列（Expr ではない）。
func TestValueExprSimplePropUnchanged(t *testing.T) {
	q := "MATCH (m:Message {id: $x}) RETURN m.id"
	root := veParse(t, q)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	// Return 項目は m.id（式ダンプの CASE/演算子記号を含まない）。
	dump := b.String()
	if strings.Contains(dump, "+") || strings.Contains(dump, "CASE") {
		t.Errorf("simple prop should not be an expr\n%s", dump)
	}
}
