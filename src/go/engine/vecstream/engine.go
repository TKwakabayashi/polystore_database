package vecstream

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
	engine.Register(settings.EngineVecStream, func() engine.Engine { return vecstreamEngine{} })
}

type vecstreamEngine struct{}

func (vecstreamEngine) Name() string { return "vecstream" }

func (vecstreamEngine) Open(ctx context.Context, cfg storage.Config) (engine.Instance, error) {
	p, err := NewProcessorWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &vecstreamInstance{p: p}, nil
}

type vecstreamInstance struct{ p *Processor }

func (in *vecstreamInstance) Run(op plan.PlanNode) (core.Result, error) {
	in.p.Reset()
	start := time.Now()
	rows, err := in.p.Run(op)
	elapsed := time.Since(start)
	if err != nil {
		return core.Result{}, err
	}
	// stream と違い、往復数・ベクトル幅・演算子別計測を埋める（vectorized との対照のため）。
	return core.Result{
		Rows:        rows,
		ExecTime:    elapsed,
		Steps:       in.p.StepMetrics(),
		RoundTrips:  in.p.RoundTrips(),
		VectorWidth: in.p.VectorWidth(),
		Engine:      "vecstream",
	}, nil
}

func (in *vecstreamInstance) Close() error { return in.p.Close() }
