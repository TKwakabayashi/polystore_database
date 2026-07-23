package bulk

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
	engine.Register(settings.EngineBulk, func() engine.Engine { return bulkEngine{} })
}

type bulkEngine struct{}

func (bulkEngine) Name() string { return "bulk" }

func (bulkEngine) Open(ctx context.Context, cfg storage.Config) (engine.Instance, error) {
	p, err := NewProcessorWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &bulkInstance{p: p}, nil
}

type bulkInstance struct{ p *Processor }

func (in *bulkInstance) Run(op plan.PlanNode) (core.Result, error) {
	in.p.Reset()
	start := time.Now()
	rows, err := in.p.ProcessQueryBulk(op)
	elapsed := time.Since(start)
	if err != nil {
		return core.Result{}, err
	}
	return core.Result{
		Rows:     rows,
		ExecTime: elapsed,
		Steps:    toStepMetrics(in.p.StepMetrics()),
		Engine:   "bulk",
	}, nil
}

func (in *bulkInstance) Close() error { return in.p.Close() }

// toStepMetrics は bulk.Metrics を core.StepMetric へ変換する。
func toStepMetrics(ms []Metrics) []core.StepMetric {
	out := make([]core.StepMetric, len(ms))
	for i, m := range ms {
		out[i] = core.StepMetric{Step: m.StepNum, Op: m.OpType, Duration: m.Duration, InRows: m.InRows, OutRows: m.RowCount}
	}
	return out
}
