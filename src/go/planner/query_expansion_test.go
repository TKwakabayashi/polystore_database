package planner_test

import (
	"strings"
	"testing"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
	"polystore_database/src/go/workload"
)

// query_expansion_test は WHERE 表現力（IN/IS NULL/変数間比較）・複数パターン・可変長パス全体
// フィルタが論理→物理プランへ正しく落ちることを DB 無しで検証する。ParseQuery は
// catalog/mapping.json だけで完結する（query_planner.go）。

const qeMappingPath = "../../../catalog/mapping.json"

func dumpQE(n plan.PlanNode, depth int, b *strings.Builder) {
	if n == nil {
		return
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString(n.String())
	b.WriteByte('\n')
	for _, c := range n.Children() {
		dumpQE(c, depth+1, b)
	}
}

func parseWorkload(t *testing.T, name string) string {
	t.Helper()
	def, ok := workload.Registry[name]
	if !ok {
		t.Fatalf("workload %s 未登録", name)
	}
	cypher, params, _ := def(migrator.ModeGraphToRdb, true)
	root, err := planner.ParseQuery(cypher, qeMappingPath, params)
	if err != nil {
		t.Fatalf("ParseQuery(%s) 失敗: %v", name, err)
	}
	var b strings.Builder
	dumpQE(root, 0, &b)
	return b.String()
}

func TestQueryExpansionShapes(t *testing.T) {
	cases := []struct {
		name string
		want []string // プランダンプに現れるべき部分文字列
	}{
		{"INLIST", []string{`IN [`, "Company", "University"}},
		{"ISNULL", []string{"IS NOT NULL"}},
		{"VARCMP", []string{"PostFilter", "a.id < b.id"}},
		{"MULTIPAT", []string{"Expand", "friend", "city"}},
		{"PATHFILT", []string{"VarLengthExpand", "all(r IN relationships(p)", "length(p) <= 3"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			got := parseWorkload(t, c.name)
			for _, w := range c.want {
				if !strings.Contains(got, w) {
					t.Errorf("[%s] プランに %q が含まれない\n--- plan ---\n%s", c.name, w, got)
				}
			}
		})
	}
}

// 直積（変数非共有の複数パターン）は本増分では非対象。明示的に parse エラーになること。
func TestQueryExpansionRejectsCartesian(t *testing.T) {
	q := "MATCH (a:Person {id: $x}), (b:Tag {id: $y})\nRETURN a.id, b.id"
	_, err := planner.ParseQuery(q, qeMappingPath, map[string]string{"x": "1", "y": "2"})
	if err == nil {
		t.Fatalf("直積パターンは parse エラーになるべき")
	}
}
