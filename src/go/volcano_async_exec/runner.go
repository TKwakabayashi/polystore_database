package volcano_async_executor

import (
	"context"
	"fmt"
	"time"

	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/storage"
)

// Result は 1 クエリ実行の結果と計測。同期版 Result に AsyncMode / Workers を足したもの。
type Result struct {
	Rows        []map[string]interface{}
	Latency     time.Duration // Trials 平均
	RoundTrips  int64         // 最終試行の DB 往復回数
	Steps       []Metrics     // 最終試行の演算子別計測
	Mode        Mode          // 処理粒度
	VectorWidth int
	Async       AsyncMode // 並行化方式
	Policy      ExecPolicy
}

// RowCount は結果行数。
func (r Result) RowCount() int { return len(r.Rows) }

// RunVolcanoAsync は Volcano(tuple-at-a-time) を非同期実行する。
// 往復数は同期版 RunVolcano と一致し、実行時間だけが並行化で縮む
// （= レイテンシ隠蔽の効果だけを取り出せる対照条件）。
func RunVolcanoAsync(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, async AsyncMode, policy ExecPolicy, trials int) (Result, error) {
	return runModel(ctx, cfg, cypher, params, ModeVolcano, 1, async, policy, trials)
}

// RunVectorizedAsync は Vectorized(batch-at-a-time) を非同期実行する。vectorSize はベクトル長。
// 「往復削減（vectorWidth）」と「レイテンシ隠蔽（workers）」の両方が効く条件。
func RunVectorizedAsync(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, vectorSize int, async AsyncMode, policy ExecPolicy, trials int) (Result, error) {
	return runModel(ctx, cfg, cypher, params, ModeVectorized, vectorSize, async, policy, trials)
}

func runModel(ctx context.Context, cfg storage.Config, cypher string, params map[string]string, mode Mode, vectorSize int, async AsyncMode, policy ExecPolicy, trials int) (Result, error) {
	if trials < 1 {
		trials = 1
	}
	qp, err := NewProcessor(ctx, cfg, mode, vectorSize, async, policy)
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
			Async:       qp.AsyncMode(),
			Policy:      qp.Policy(),
		}
	}
	last.Latency = sum / time.Duration(trials)
	return last, nil
}

// PrintResult は結果の要約を標準出力へ出す（往復回数を含む）。
func PrintResult(title string, r Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - モデル            : %s (vectorWidth=%d) / async=%s\n", r.Mode, r.VectorWidth, r.Async)
	if r.Policy.Mode == ExecDynamic {
		fmt.Printf("  - 並行度            : ExecDynamic (globalMax=%d)\n", r.Policy.globalMax())
	} else {
		fmt.Printf("  - 並行度            : ExecFixed (expand=%d, filter=%d, projection=%d)\n",
			r.Policy.For(OpExpand).workers(),
			r.Policy.For(OpFilter).workers(),
			r.Policy.For(OpProjection).workers())
	}
	fmt.Printf("  - 全体実行時間      : %v\n", r.Latency)
	fmt.Printf("  - DB 往復回数       : %d\n", r.RoundTrips)
	fmt.Printf("  - 最終結果数        : %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - 演算子別（Duration は全ワーカー合算のため実時間を超え得る）:")
		for _, s := range r.Steps {
			fmt.Printf("      %-16s time=%-12v rows=%d\n", s.OpType, s.Duration, s.RowCount)
		}
	}
	fmt.Println()
}
