package stream

import (
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// streamPostFilter はクロスエイリアス比較（a.id <op> b.id）をコーディネータ上でメモリ評価する。
// DB 往復は無い。両エイリアスの束縛 UUID（スロット）を id 比較し、通過行のみ下流へ流す。
func streamPostFilter(qp *Processor, o *plan.PostFilter, step int, in <-chan []Record, out chan<- []Record) (int, error) {
	newSlotCount := len(o.OutputSlot.VarToSlot)
	count := 0
	for batch := range in {
		res := make([]Record, 0, len(batch))
		for _, r := range batch {
			if !core.EvalVarCmp(o.Cond, r, o.InputSlot) {
				continue
			}
			nr := Record{Slots: make([]uid.UUID, newSlotCount), Vals: r.Vals}
			for alias, outIdx := range o.OutputSlot.VarToSlot {
				if inIdx, ok := o.InputSlot.VarToSlot[alias]; ok {
					nr.Slots[outIdx] = r.Slots[inIdx]
				}
			}
			res = append(res, nr)
			count++
		}
		if len(res) > 0 {
			out <- res
		}
	}
	return count, nil
}
