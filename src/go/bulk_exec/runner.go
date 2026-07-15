package bulk_executor

import (
	"context"
	"fmt"
	"time"

	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/storage"
)

// Result は 1 クエリ実行の結果と計測。
type Result struct {
	Rows       []map[string]interface{}
	Latency    time.Duration // Trials 平均
	RoundTrips int64         // 最終試行の DB 往復回数
	Steps      []Metrics     // 最終試行の演算子別計測
}

// RowCount は結果行数。
func (r Result) RowCount() int { return len(r.Rows) }

// RunBulk は全件マテリアライズモデルで実行する（trials 回の平均）。
func RunBulk(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, trials int) (Result, error) {
	if trials < 1 {
		trials = 1
	}
	qp, err := NewProcessor(ctx, cfg)
	if err != nil {
		return Result{}, err
	}
	defer qp.Close()

	var (
		sum  time.Duration
		last Result
	)
	for i := 0; i < trials; i++ {
		qp.Reset()
		op, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
		if err != nil {
			return Result{}, fmt.Errorf("プラン構築に失敗: %w", err)
		}
		start := time.Now()
		rows, err := qp.Run(op)
		elapsed := time.Since(start)
		if err != nil {
			return Result{}, fmt.Errorf("クエリ実行に失敗: %w", err)
		}
		sum += elapsed
		last = Result{
			Rows:       rows,
			Latency:    elapsed,
			RoundTrips: qp.RoundTrips(),
			Steps:      qp.StepMetrics(),
		}
	}
	last.Latency = sum / time.Duration(trials)
	return last, nil
}

// PrintResult は結果の要約を標準出力へ出す（演算子ごとに in/out/time）。
func PrintResult(title string, r Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - モデル            : Bulk (全件マテリアライズ)\n")
	fmt.Printf("  - 全体実行時間      : %v\n", r.Latency)
	fmt.Printf("  - DB 往復回数       : %d\n", r.RoundTrips)
	fmt.Printf("  - 最終結果数        : %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - 演算子別:")
		for _, s := range r.Steps {
			fmt.Printf("      %-16s in=%-8d out=%-8d time=%v\n", s.OpType, s.InRows, s.OutRows, s.Duration)
		}
	}
	fmt.Println()
}
