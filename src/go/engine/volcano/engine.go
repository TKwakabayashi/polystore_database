package volcano

import (
	"context"
	"fmt"
	"time"

	planner "polystore_database/src/go/planner"
	"polystore_database/src/go/storage"
)

// Result は 1 クエリ実行の結果と計測。
type Result struct {
	Rows        []map[string]interface{}
	Latency     time.Duration // Trials 平均
	RoundTrips  int64         // 最終試行の DB 往復回数
	Steps       []Metrics     // 最終試行の演算子別計測
	Mode        Mode
	VectorWidth int
}

// RowCount は結果行数。
func (r Result) RowCount() int { return len(r.Rows) }

// RunVolcano は Volcano(tuple-at-a-time) モードで実行する。
func RunVolcano(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, trials int) (Result, error) {
	return runModel(ctx, cfg, cypher, params, ModeVolcano, 1, trials)
}

// RunVectorized は Vectorized(batch-at-a-time) モードで実行する。vectorSize はベクトル長。
func RunVectorized(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, vectorSize, trials int) (Result, error) {
	return runModel(ctx, cfg, cypher, params, ModeVectorized, vectorSize, trials)
}

func runModel(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, mode Mode, vectorSize, trials int) (Result, error) {
	if trials < 1 {
		trials = 1
	}
	qp, err := NewProcessor(ctx, cfg, mode, vectorSize)
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
			Rows:        rows,
			Latency:     elapsed,
			RoundTrips:  qp.RoundTrips(),
			Steps:       qp.StepMetrics(),
			Mode:        qp.Mode(),
			VectorWidth: qp.VectorWidth(),
		}
	}
	last.Latency = sum / time.Duration(trials)
	return last, nil
}

// PrintResult は結果の要約を標準出力へ出す（往復回数を含む）。
func PrintResult(title string, r Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - モデル            : %s (vectorWidth=%d)\n", r.Mode, r.VectorWidth)
	fmt.Printf("  - 全体実行時間      : %v\n", r.Latency)
	fmt.Printf("  - DB 往復回数       : %d\n", r.RoundTrips)
	fmt.Printf("  - 最終結果数        : %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - 演算子別:")
		for _, s := range r.Steps {
			fmt.Printf("      %-16s time=%-12v rows=%d\n", s.OpType, s.Duration, s.RowCount)
		}
	}
	fmt.Println()
}
