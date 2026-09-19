package bulk

import (
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// bulkPostFilter はクロスエイリアス比較（a.id <op> b.id）をメモリ評価する（DB 往復なし）。
func bulkPostFilter(qp *Processor, o *plan.PostFilter, in []Record) ([]Record, error) {
	newSlotCount := len(o.OutputSlot.VarToSlot)
	out := make([]Record, 0, len(in))
	for _, r := range in {
		if !core.EvalVarCmp(o.Cond, r, o.InputSlot) {
			continue
		}
		nr := Record{Slots: make([]uid.UUID, newSlotCount), Vals: r.Vals}
		for alias, outIdx := range o.OutputSlot.VarToSlot {
			if inIdx, ok := o.InputSlot.VarToSlot[alias]; ok {
				nr.Slots[outIdx] = r.Slots[inIdx]
			}
		}
		out = append(out, nr)
	}
	return out, nil
}
