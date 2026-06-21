package test

import (
	"fmt"
	"sort"
	"time"
)

// StepMetric は演算子1つ分の計測。bulk実行時に各演算子で埋める。
type StepMetric struct {
	Name     string        // 演算子名 (EntityScan, Expand, ...)
	Duration time.Duration // その演算子の実行時間
	Rows     int           // 中間結果の件数
}

// ExecResult は1回のクエリ実行結果。Neo4j単体・自作システム共通。
type ExecResult struct {
	Rows         []map[string]interface{} // 結果行（1件ずつ出力用）
	TotalLatency time.Duration            // 全体実行時間
	Steps        []StepMetric             // 演算子ごとの詳細（未計測なら nil）
}

func (r ExecResult) RowCount() int { return len(r.Rows) }

func (r ExecResult) SumOpTime() time.Duration {
	var s time.Duration
	for _, st := range r.Steps {
		s += st.Duration
	}
	return s
}

// PrintRows は結果行を1件ずつ出力する。
func PrintRows(r ExecResult) {
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
func PrintTiming(title string, r ExecResult) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - 全体実行時間 (Latency): %v\n", r.TotalLatency)
	fmt.Printf("  - 最終結果数: %d\n", r.RowCount())
	fmt.Println()
}

// PrintDetail は演算子ごとの実行時間・中間件数を出力する（bulk 用）。
func PrintDetail(title string, r ExecResult) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - 全体実行時間 (Latency): %v\n", r.TotalLatency)
	fmt.Printf("  - オペレータ合計時間 (Sum): %v\n", r.SumOpTime())
	fmt.Printf("  - 最終結果数: %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - 実行詳細:")
		for _, st := range r.Steps {
			fmt.Printf("      %-16s time=%-12v rows=%d\n", st.Name, st.Duration, st.Rows)
		}
	}
	fmt.Println()
}
