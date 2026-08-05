package core

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInstrRecordOpAndRoundTrips(t *testing.T) {
	in := NewInstr()
	// 並行に複数ワーカーが同一 step へ書いても合算される。
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				in.RecordOp(1, "Expand", time.Millisecond, 3)
				in.CountRoundTrip()
			}
		}()
	}
	wg.Wait()

	if got := in.RoundTrips(); got != 400 {
		t.Errorf("RoundTrips = %d, want 400", got)
	}
	steps := in.StepMetrics()
	if len(steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(steps))
	}
	m := steps[0]
	if m.Op != "Expand" || m.OutRows != 400*3 || m.InRows != -1 {
		t.Errorf("step = %+v, want Op=Expand OutRows=1200 InRows=-1", m)
	}
	if m.Duration != 400*time.Millisecond {
		t.Errorf("Duration = %v, want 400ms（合算）", m.Duration)
	}
}

func TestInstrFlowGatedByTraceFlow(t *testing.T) {
	in := NewInstr()
	// off なら収集しない。
	TraceFlow = false
	in.RecordFlow(1, "Filter", 1, 1, 10, 4, 1, time.Now(), time.Now())
	if got := in.FlowMetrics(); len(got) != 0 {
		t.Fatalf("TraceFlow=false で flow=%d 件", len(got))
	}
	// on なら累積。
	TraceFlow = true
	defer func() { TraceFlow = false }()
	t0 := time.Now()
	in.RecordFlow(2, "Expand", 1, 3, 10, 100, 1, t0, t0.Add(5*time.Millisecond))
	in.RecordFlow(2, "Expand", 1, 2, 5, 40, 1, t0.Add(time.Millisecond), t0.Add(8*time.Millisecond))
	fm := in.FlowMetrics()
	if len(fm) != 1 {
		t.Fatalf("flow = %d, want 1", len(fm))
	}
	f := fm[0]
	if f.BatchesIn != 2 || f.BatchesOut != 5 || f.RowsIn != 15 || f.RowsOut != 140 || f.Queries != 2 {
		t.Errorf("flow = %+v", f)
	}
	if f.Wall() != 8*time.Millisecond { // min t0 .. max t0+8ms
		t.Errorf("Wall = %v, want 8ms", f.Wall())
	}
	out := FormatFlow("t", fm)
	if !strings.Contains(out, "Expand") || !strings.Contains(out, "fanout") {
		t.Errorf("FormatFlow 出力:\n%s", out)
	}
}

func TestInstrReset(t *testing.T) {
	TraceFlow = true
	defer func() { TraceFlow = false }()
	in := NewInstr()
	in.RecordOp(1, "Scan", time.Millisecond, 5)
	in.CountRoundTrip()
	in.RecordFlow(1, "Scan", 0, 1, 0, 5, 1, time.Now(), time.Now())
	in.Reset()
	if in.RoundTrips() != 0 || len(in.StepMetrics()) != 0 || len(in.FlowMetrics()) != 0 {
		t.Errorf("Reset 後も残存: rt=%d steps=%d flow=%d", in.RoundTrips(), len(in.StepMetrics()), len(in.FlowMetrics()))
	}
}
