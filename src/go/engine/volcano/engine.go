package volcano

import (
	"context"
	"time"

	"polystore_database/src/go/engine"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
)

func init() {
	engine.Register(settings.EngineVolcano, func() engine.Engine { return volcanoEngine{mode: ModeVolcano, name: "volcano"} })
	engine.Register(settings.EngineVectorized, func() engine.Engine { return volcanoEngine{mode: ModeVectorized, name: "vectorized"} })
}

type volcanoEngine struct {
	mode Mode
	name string
}

func (e volcanoEngine) Name() string { return e.name }

func (e volcanoEngine) Open(ctx context.Context, cfg storage.Config) (engine.Instance, error) {
	width := 1
	if e.mode == ModeVectorized {
		width = settings.VectorSize // 実行時の設定を反映
	}
	p, err := NewProcessor(ctx, cfg, e.mode, width)
	if err != nil {
		return nil, err
	}
	return &volcanoInstance{p: p, name: e.name}, nil
}

type volcanoInstance struct {
	p    *Processor
	name string
}

func (in *volcanoInstance) Run(op plan.PlanNode) (core.Result, error) {
	in.p.Reset()
	start := time.Now()
	rows, err := in.p.Run(op)
	elapsed := time.Since(start)
	if err != nil {
		return core.Result{}, err
	}
	return core.Result{
		Rows:        rows,
		ExecTime:    elapsed,
		Steps:       toStepMetrics(in.p.StepMetrics()),
		RoundTrips:  in.p.RoundTrips(),
		VectorWidth: in.p.VectorWidth(),
		Engine:      in.name,
	}, nil
}

func (in *volcanoInstance) Close() error { return in.p.Close() }

// toStepMetrics は volcano.Metrics を core.StepMetric へ変換する（InRows は未計測 -1）。
func toStepMetrics(ms []Metrics) []core.StepMetric {
	out := make([]core.StepMetric, len(ms))
	for i, m := range ms {
		out[i] = core.StepMetric{Step: m.StepNum, Op: m.OpType, Duration: m.Duration, InRows: -1, OutRows: m.RowCount}
	}
	return out
}
