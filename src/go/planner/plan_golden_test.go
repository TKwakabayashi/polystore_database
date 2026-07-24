// plan_golden_test は全ワークロードの Cypher を planner.ParseQuery で論理プランへ変換し、
// (1) エラー無し (2) プラン木のダンプが golden ファイルと一致 を DB 無しで検証する。
// ParseQuery は catalog/mapping.json だけで parse→plan→refine→pushdown 判定まで完結し、
// DB 接続を一切張らない（query_planner.go）。
//
//	golden 生成/更新: go test ./planner/ -run Golden -update
//	検証:            go test ./planner/
package planner_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
	"polystore_database/src/go/workload"
)

var update = flag.Bool("update", false, "golden プランファイルを再生成する")

// go test の cwd は src/go/planner。リポジトリルートは 3 段上。
const goldenMappingPath = "../../../catalog/mapping.json"

// normalizeEOL は CRLF/CR を LF に揃える（OS 間・git autocrlf 差異の吸収）。
func normalizeEOL(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

// dumpPlan はプラン木を決定的にインデント付きでダンプする（各ノード String() を深さ順に）。
func dumpPlan(n plan.PlanNode, depth int, b *strings.Builder) {
	if n == nil {
		return
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString(n.String())
	b.WriteByte('\n')
	for _, c := range n.Children() {
		dumpPlan(c, depth+1, b)
	}
}

func TestGoldenPlans(t *testing.T) {
	for _, name := range workload.AllWorkloadNames() {
		name := name
		t.Run(name, func(t *testing.T) {
			def := workload.Registry[name]
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)

			root, err := planner.ParseQuery(cypher, goldenMappingPath, params)
			if err != nil {
				t.Fatalf("ParseQuery(%s) 失敗: %v", name, err)
			}
			if root == nil {
				t.Fatalf("ParseQuery(%s) が nil プランを返した", name)
			}

			var b strings.Builder
			dumpPlan(root, 0, &b)
			got := b.String()

			goldenPath := filepath.Join("testdata", name+".plan")
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatalf("mkdir testdata: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden 読み込み失敗（-update で生成せよ）: %v", err)
			}
			// Windows で autocrlf により golden が CRLF でチェックアウトされても一致するよう正規化する。
			if normalizeEOL(got) != normalizeEOL(string(want)) {
				t.Errorf("プランが golden と不一致 (%s)\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}
		})
	}
}
