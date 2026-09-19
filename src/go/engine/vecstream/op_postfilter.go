package vecstream

import (
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// postFilterOp はクロスエイリアス比較（a.id <op> b.id）のメモリ評価。DB 往復は無く、
// 両エイリアスの束縛 UUID（スロット）を id 比較して通過行のみ返す（状態を持たず並行安全）。
// exchange 経由で呼ばれるが sess は使用しない。
type postFilterOp struct {
	p *Processor
	o *plan.PostFilter
}

func (f *postFilterOp) process(_ neo4j.SessionWithContext, in *Batch) (*Batch, error) {
	out := newBatch(len(f.o.OutputSlot.VarToSlot), in.n)
	for i := 0; i < in.n; i++ {
		if core.EvalVarCmp(f.o.Cond, core.Record{Slots: in.row(i)}, f.o.InputSlot) {
			out.appendRowVals(remap(in.row(i), f.o.InputSlot, f.o.OutputSlot), in.valsAt(i))
		}
	}
	return out, nil
}
