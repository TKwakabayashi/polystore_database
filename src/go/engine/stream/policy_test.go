package stream

import (
	"context"
	"testing"

	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
)

// runBatches が出力を VectorWidth 幅へ再チャンクし、演算子計測とフローを記録することを DB 非依存で検証する。
func TestRunBatchesReChunkAndFlow(t *testing.T) {
	core.TraceFlow = true
	defer func() { core.TraceFlow = false }()

	qp := &Processor{
		ctx:   context.Background(),
		instr: core.NewInstr(),
		exec:  ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 1}, VectorWidth: 4},
	}

	in := make(chan []Record, 1)
	in <- []Record{{Slots: []uid.UUID{"a"}}} // 1 バッチ・1 行
	close(in)

	out := make(chan []Record, 16)
	go func() {
		_, _ = runBatches(context.Background(), qp.exec, nil, OpExpand, qp, 1, in, out,
			noResource, closeNoResource,
			func(_ struct{}, b []Record) ([]Record, error) {
				r := make([]Record, 10) // fan-out ×10
				for i := range r {
					r[i] = Record{Slots: []uid.UUID{"x"}}
				}
				return r, nil
			})
		close(out)
	}()

	total, batches := 0, 0
	for b := range out {
		batches++
		total += len(b)
		if len(b) > qp.exec.vectorWidth() {
			t.Errorf("出力バッチ %d が VectorWidth %d 超過", len(b), qp.exec.vectorWidth())
		}
	}
	if total != 10 {
		t.Fatalf("総行数 = %d, want 10", total)
	}
	if batches != 3 { // 4,4,2
		t.Errorf("バッチ数 = %d, want 3", batches)
	}

	// フロー: 入力1バッチ1行 → 出力10行を width4 で3バッチへ、DBクエリ1。
	fm := qp.instr.FlowMetrics()
	if len(fm) != 1 {
		t.Fatalf("flow = %d, want 1", len(fm))
	}
	f := fm[0]
	if f.Op != "Expand" || f.BatchesIn != 1 || f.BatchesOut != 3 || f.RowsIn != 1 || f.RowsOut != 10 || f.Queries != 1 {
		t.Errorf("flow = %+v", f)
	}
	// 演算子計測: OutRows=10。
	sm := qp.StepMetrics()
	if len(sm) != 1 || sm[0].OutRows != 10 || sm[0].Op != "Expand" {
		t.Errorf("steps = %+v", sm)
	}
}
