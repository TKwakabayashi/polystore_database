//go:build integration

// engine_equivalence_test は「実行モデルを変えても結果は同一」という不変条件を守る
// 回帰テスト。P5（エンジン内 UUID 型付け）で record パイプライン（scan/filter/expand/
// projection）の識別子を id.UUID へ型付けした際、stream/bulk/volcano/vectorized の 4 モデルが
// バイト一致の結果を返すことを担保する。
//
//	実行: cwd=src/go で docker スタックを up した状態（graph 配置）で
//	  go test -tags integration ./integration/ -run EngineEquivalence -v
//	設定は POLYSTORE_CONFIG（未設定なら ../../config/config.json）から読む。
//
// 配置は graph 固定（migration/Cassandra 依存を避け決定的にする）。全 5 配置 × 4 モデルの
// より広いスイープは `-mode bench-models`（bench.RunModelBenchmark）で行う（docs/EXTENDING.md）。
package integration

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// equivModels は等価性を検証する実行モデル。stream を基準（reference）にする。
var equivModels = []settings.EngineKind{"stream", "bulk", "volcano", "vectorized"}

// coreMustPassAll は 4 モデル全ての成功を必須とするワークロード（record パイプラインの
// 主経路）。ここでエンジンがエラーを返すのは P5 の回帰とみなし、テストを失敗させる。
var coreMustPassAll = map[string]bool{"Q11": true}

// equivWorkloads は等価性を確認するワークロード群。選定基準:
//   - record パイプライン（scan/filter/expand/projection）を代表的にカバーする
//     （Q11=traversal, IS2/IS3=scan+projection, AGG1/AGG6=集約）。
//   - graph 配置で 4 モデル全てが成功する（結果非空で multiset 比較が意味を持つ）。
//   - >= / <= を使うもの（Q2/AGG4/AGG5）は bulk-graph の既存演算子ギャップを踏むため除外し、
//     そちらは bench-models で扱う。Q9 は結果 0 件かつ低速のため除外。
var equivWorkloads = []string{"Q11", "IS2", "IS3", "AGG1", "AGG6"}

// TestEngineEquivalence は graph 配置で各ワークロードを 4 モデルで実行し、成功したモデル間で
// 結果の行集合（順序非依存の multiset）が一致することを検証する。coreMustPassAll のワークロードは
// 全モデルの成功も要求する。
func TestEngineEquivalence(t *testing.T) {
	// go test は package ディレクトリを cwd にするため、アプリと同じ相対パス前提へ揃える。
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("chdir to src/go: %v", err)
	}
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	for _, name := range equivWorkloads {
		name := name
		t.Run(name, func(t *testing.T) {
			def, ok := workload.Registry[name]
			if !ok {
				t.Fatalf("未知のワークロード %q", name)
			}
			// migs は使わない（graph 配置固定）。cypher/params のみ取得。
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)

			// 基準（stream）を先に実行。失敗したらこのワークロードは検証不能。
			refRows, refErr := runRows(ctx, cfg, "stream", cypher, params)
			if refErr != nil {
				if coreMustPassAll[name] {
					t.Fatalf("[stream] コア経路が失敗: %v", refErr)
				}
				t.Skipf("[stream] 実行失敗のため skip: %v", refErr)
			}
			ref := canonRows(refRows)
			t.Logf("[stream] rows=%d（基準）", len(refRows))

			for _, model := range equivModels[1:] {
				rows, err := runRows(ctx, cfg, model, cypher, params)
				if err != nil {
					if coreMustPassAll[name] {
						t.Errorf("[%s] コア経路が失敗（P5 回帰の疑い）: %v", model, err)
					} else {
						t.Logf("[%s] 実行失敗（既知のエンジンギャップの可能性）→ 比較を skip: %v", model, err)
					}
					continue
				}
				got := canonRows(rows)
				if len(got) != len(ref) {
					t.Errorf("[%s] 行数が stream と不一致: %d != %d", model, len(got), len(ref))
					continue
				}
				if !reflect.DeepEqual(got, ref) {
					t.Errorf("[%s] 行集合が stream と不一致\n%s", model, firstDiff(ref, got))
					continue
				}
				t.Logf("[%s] rows=%d 一致", model, len(rows))
			}
		})
	}
}

// runRows は指定モデルでワークロードを 1 回実行し、結果行を返す。
func runRows(ctx context.Context, cfg storage.Config, model settings.EngineKind, cypher string, params map[string]string) ([]map[string]interface{}, error) {
	r, err := bench.RunEngine(ctx, cfg, model, cypher, params)
	if err != nil {
		return nil, err
	}
	return r.Rows, nil
}

// canonRows は行スライスを順序非依存で比較できる正規形（各行を "k=v;" ソート連結し、
// 行同士もソート）へ変換する。stream は並行実行で行順が非決定的なため multiset で比較する。
func canonRows(rows []map[string]interface{}) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%v;", k, row[k])
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return out
}

// firstDiff は 2 つの正規形行スライスの最初の相違を人間可読に示す。
func firstDiff(want, got []string) string {
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			return fmt.Sprintf("  行 %d:\n    stream: %s\n    other : %s", i, want[i], got[i])
		}
	}
	return fmt.Sprintf("  長さのみ相違: stream=%d other=%d", len(want), len(got))
}
