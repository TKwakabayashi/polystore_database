package volcano

import (
	"context"
	"time"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
)

// postFilterIterator はクロスエイリアス比較（a.id <op> b.id）をメモリ評価する pull 演算子。
// DB 往復は無く、両エイリアスの束縛 UUID（スロット）を id 比較して通過行のみ出力する。
type postFilterIterator struct {
	p     *Processor
	o     *plan.PostFilter
	child Iterator
	step  int
}

func (f *postFilterIterator) Open(ctx context.Context) error  { return f.child.Open(ctx) }
func (f *postFilterIterator) Close(ctx context.Context) error { return f.child.Close(ctx) }

func (f *postFilterIterator) Next(ctx context.Context) (*Batch, error) {
	for {
		in, err := f.child.Next(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, nil
		}
		start := time.Now()
		outSlots := len(f.o.OutputSlot.VarToSlot)
		out := newBatch(outSlots, in.n)
		for i := 0; i < in.n; i++ {
			if core.EvalVarCmp(f.o.Cond, core.Record{Slots: in.row(i)}, f.o.InputSlot) {
				out.appendRowVals(remap(in.row(i), f.o.InputSlot, f.o.OutputSlot), in.valsAt(i))
			}
		}
		f.p.recordOp(f.step, "PostFilter", time.Since(start), out.n)
		if out.n > 0 {
			return out, nil
		}
	}
}
