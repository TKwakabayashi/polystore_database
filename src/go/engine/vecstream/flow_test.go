package vecstream

import (
	"context"
	"strings"
	"testing"

	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
)

// TraceFlow のフロー計測が exchange 経由で正しく集計されることを DB 非依存で検証する。
// （testProcessor / fakeSource は exchange_test.go 定義）

func TestFlowCollectsBatchesRowsQueries(t *testing.T) {
	core.TraceFlow = true
	defer func() { core.TraceFlow = false }()

	// VectorWidth=4、fan-out ×10 で「再チャンクにより batOut が増える」ことを見る。
	p := testProcessor(2, 8, 4, false)
	src := &fakeSource{batches: 10} // 1 行バッチ × 10
	ex := newExchange(p, src, OpExpand, "Expand", 1, noRes, noResClose,
		func(_ struct{}, b *Batch) (*Batch, error) {
			out := newBatch(1, b.n*10)
			for i := 0; i < b.n; i++ {
				for k := 0; k < 10; k++ {
					out.appendRow([]uid.UUID{b.get(i, 0)})
				}
			}
			return out, nil
		})
	if err := ex.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	for {
		b, err := ex.Next(context.Background())
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if b == nil {
			break
		}
	}
	ex.Close(context.Background())

	flows := p.FlowMetrics()
	if len(flows) != 1 {
		t.Fatalf("flows = %d, want 1", len(flows))
	}
	m := flows[0]
	// 入力 10 バッチ・10 行、各入力→10 行 fan-out→合計 100 行。
	if m.BatchesIn != 10 || m.RowsIn != 10 {
		t.Errorf("BatchesIn=%d RowsIn=%d, want 10/10", m.BatchesIn, m.RowsIn)
	}
	if m.RowsOut != 100 {
		t.Errorf("RowsOut=%d, want 100", m.RowsOut)
	}
	// 各入力の 10 行を width4 で split → ceil(10/4)=3 バッチ × 10 入力 = 30。
	if m.BatchesOut != 30 {
		t.Errorf("BatchesOut=%d, want 30（再チャンク）", m.BatchesOut)
	}
	// 1 入力バッチ = 1 DB クエリ。
	if m.Queries != 10 {
		t.Errorf("Queries=%d, want 10", m.Queries)
	}
	if m.Wall() < 0 {
		t.Errorf("Wall=%v, want >=0", m.Wall())
	}

	// FormatFlow が表を返す。
	out := core.FormatFlow("test", flows)
	if !strings.Contains(out, "Expand") || !strings.Contains(out, "batOut") {
		t.Errorf("FormatFlow 出力に期待列が無い:\n%s", out)
	}
}

// TraceFlow=false のときは何も集計しない（オーバーヘッド無し）。
func TestFlowDisabledCollectsNothing(t *testing.T) {
	core.TraceFlow = false
	p := testProcessor(2, 8, 1024, false)
	src := &fakeSource{batches: 20}
	ex := newExchange(p, src, OpExpand, "Expand", 1, noRes, noResClose,
		func(_ struct{}, b *Batch) (*Batch, error) { return b, nil })
	_ = ex.Open(context.Background())
	for {
		b, _ := ex.Next(context.Background())
		if b == nil {
			break
		}
	}
	ex.Close(context.Background())
	if got := p.FlowMetrics(); len(got) != 0 {
		t.Errorf("TraceFlow=false でも flow=%d 件収集された", len(got))
	}
}
