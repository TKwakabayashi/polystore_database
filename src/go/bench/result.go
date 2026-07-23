package bench

import (
	"fmt"
	"sort"
	"time"

	"polystore_database/src/go/engine/core"
)

// 結果型は core.Result / core.StepMetric に統一（旧 bench.ExecResult / StepMetric は廃止）。

// TrialResult は1ワークロード試行の集計（将来の詳細レポート用スキャフォールド）。
type TrialResult struct {
	WorkloadName string
	Mode         string
	TotalTime    time.Duration
	Steps        []core.StepMetric
}

// PrintRows は結果行を1件ずつ出力する。
func PrintRows(r core.Result) {
	fmt.Printf("\n--- 結果: %d 件 ---\n", r.RowCount())
	for i, row := range r.Rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Printf("[%d] ", i)
		for j, k := range keys {
			if j > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s=%v", k, row[k])
		}
		fmt.Println()
	}
}

// PrintTiming は全体実行時間と件数を出力する。
func PrintTiming(title string, r core.Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - 全体実行時間 (Latency): %v\n", r.TotalLatency())
	fmt.Printf("  - 最終結果数: %d\n", r.RowCount())
	fmt.Println()
}

// PrintDetail は演算子ごとの実行時間・中間件数を出力する（bulk 用）。
func PrintDetail(title string, r core.Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - 全体実行時間 (Latency): %v\n", r.TotalLatency())
	fmt.Printf("  - オペレータ合計時間 (Sum): %v\n", r.SumStepTime())
	fmt.Printf("  - 最終結果数: %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - 実行詳細:")
		for _, st := range r.Steps {
			fmt.Printf("      %-16s time=%-12v rows=%d\n", st.Op, st.Duration, st.OutRows)
		}
	}
	fmt.Println()
}
