package stream

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
	engine.Register(settings.EngineStream, func() engine.Engine { return streamEngine{} })
}

type streamEngine struct{}

func (streamEngine) Name() string { return "stream" }

func (streamEngine) Open(ctx context.Context, cfg storage.Config) (engine.Instance, error) {
	p, err := NewProcessorWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &streamInstance{p: p}, nil
}

type streamInstance struct{ p *Processor }

func (in *streamInstance) Run(op plan.PlanNode) (core.Result, error) {
	in.p.Reset()
	start := time.Now()
	rows, err := in.p.ProcessQueryStream(op)
	elapsed := time.Since(start)
	if err != nil {
		return core.Result{}, err
	}
	// vecstream と揃えて往復数・ベクトル幅・演算子別計測を埋める（差は push/pull と行/列表現のみ）。
	return core.Result{
		Rows:        rows,
		ExecTime:    elapsed,
		Steps:       in.p.StepMetrics(),
		RoundTrips:  in.p.RoundTrips(),
		VectorWidth: in.p.exec.vectorWidth(),
		Engine:      "stream",
	}, nil
}

func (in *streamInstance) Close() error { return in.p.Close() }
