package planner_test

import (
	"strings"
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
	"polystore_database/src/go/settings"
)

const msMapping = "../../../catalog/mapping.json"

func msDump(n plan.PlanNode, depth int, b *strings.Builder) {
	if n == nil {
		return
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString(n.String())
	b.WriteByte('\n')
	for _, c := range n.Children() {
		msDump(c, depth+1, b)
	}
}

func msParse(t *testing.T, q string) (plan.PlanNode, string) {
	t.Helper()
	prev := settings.MultiStage
	settings.MultiStage = true
	defer func() { settings.MultiStage = prev }()
	root, err := planner.ParseQuery(q, msMapping, map[string]string{"x": "123"})
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	var b strings.Builder
	msDump(root, 0, &b)
	return root, b.String()
}

// ステージ連鎖（bare 束縛 WITH ＋ 後続 MATCH）: RowSource で 2 ステージが連結され、
// 各ステージが独立した record パイプラインを持つこと。
func TestMultiStageChaining(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f MATCH (f)-[:KNOWS]-(g:Person) RETURN f.id, g.id"
	_, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	for _, want := range []string{"Return", "RowSource", "Expand", "EntityScan"} {
		if !strings.Contains(dump, want) {
			t.Errorf("plan missing %q\n%s", want, dump)
		}
	}
	// RowSource が現れる（ステージ境界）＝ verbatim ではなく structured 多段。
	if strings.Contains(dump, "verbatim") {
		t.Errorf("expected structured plan, got verbatim\n%s", dump)
	}
}

// WITH 束縛のあと第2 MATCH 無しでプロパティ再材料化（RETURN f.id）: RowSource＋後段 Projection。
func TestMultiStageRematerialize(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f RETURN f.id"
	root, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	if _, ok := root.(*plan.StoreFragment); ok {
		t.Fatalf("expected structured plan, got verbatim:\n%s", dump)
	}
	if !strings.Contains(dump, "RowSource") {
		t.Errorf("expected RowSource for re-materialization\n%s", dump)
	}
}

// WITH 集約＋値列＋post-WITH WHERE＋最終 RETURN（WITH 出力名のみ）: Shape B（RowSource 無し）。
// record→Projection→Aggregate→RowProject→RowFilter→Return。
func TestMultiStageAggProject(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f.id AS fid, count(*) AS c WHERE c > 1 RETURN fid, c ORDER BY c DESC"
	root, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	if _, ok := root.(*plan.StoreFragment); ok {
		t.Fatalf("expected structured plan, got verbatim:\n%s", dump)
	}
	for _, want := range []string{"Aggregate", "RowProject", "RowFilter", "Return"} {
		if !strings.Contains(dump, want) {
			t.Errorf("plan missing %q\n%s", want, dump)
		}
	}
}

// 集約後に carried 束縛のプロパティ再材料化（RETURN f.id, c）: RowSource が集約値 c を Vals で運び、
// 後段 Projection が f.id を材料化＋c を MergeVals する。
func TestMultiStageAggRematerialize(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f, count(*) AS c WHERE c > 1 RETURN f.id, c"
	root, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	if _, ok := root.(*plan.StoreFragment); ok {
		t.Fatalf("expected structured plan, got verbatim:\n%s", dump)
	}
	// post-WITH WHERE(c>1) は中間ステージの RowFilter、境界は RowSource、集約は Aggregate。
	for _, want := range []string{"RowSource", "Aggregate", "RowFilter"} {
		if !strings.Contains(dump, want) {
			t.Errorf("plan missing %q\n%s", want, dump)
		}
	}
}

// 値/集約を後続 MATCH へ持ち越す（WITH f, count(*) AS c MATCH (f)-…-(g) RETURN g.id, c）:
// RowSource が f をスロット・c を Vals で運び、後段が (f)-(g) を展開して g.id 材料化＋c passthrough。
func TestMultiStageValueIntoMatch(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) WITH f, count(*) AS c MATCH (f)-[:KNOWS]-(g:Person) RETURN g.id, c"
	root, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	if _, ok := root.(*plan.StoreFragment); ok {
		t.Fatalf("expected structured plan, got verbatim:\n%s", dump)
	}
	for _, want := range []string{"RowSource", "Aggregate", "Expand"} {
		if !strings.Contains(dump, want) {
			t.Errorf("plan missing %q\n%s", want, dump)
		}
	}
}

// WITH で CASE 計算列を作り最終 RETURN で参照する形（T2）: RowProject が式を評価して clean 名へ写し、
// verbatim ではなく structured 多段になること。RowFilter で post-WITH WHERE も式で判定する。
func TestMultiStageWithComputedColumn(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) " +
		"WITH f, CASE WHEN f.id > 10 THEN 1 ELSE 0 END AS big " +
		"WHERE big = 1 RETURN f.id, big"
	root, dump := msParse(t, q)
	t.Logf("\n%s", dump)
	if f, ok := root.(*plan.StoreFragment); ok && f.Verbatim != "" {
		t.Fatalf("expected structured plan, got verbatim: %s", f.Verbatim)
	}
	for _, want := range []string{"RowProject", "RowFilter", "big"} {
		if !strings.Contains(dump, want) {
			t.Errorf("plan missing %q\n%s", want, dump)
		}
	}
}

// 前段 WITH の値/式列を後段の集約引数にする形は未対応 → verbatim（T2 のガード）。
func TestMultiStageAggOverComputedFallsBack(t *testing.T) {
	q := "MATCH (p:Person {id: $x})-[:KNOWS]-(f:Person) " +
		"WITH f, f.id AS fid RETURN sum(fid) AS s"
	prev := settings.MultiStage
	settings.MultiStage = true
	defer func() { settings.MultiStage = prev }()
	root, err := planner.ParseQuery(q, msMapping, map[string]string{"x": "123"})
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	f, ok := root.(*plan.StoreFragment)
	if !ok || f.Verbatim == "" {
		var b strings.Builder
		msDump(root, 0, &b)
		t.Fatalf("expected verbatim fallback, got:\n%s", b.String())
	}
}
