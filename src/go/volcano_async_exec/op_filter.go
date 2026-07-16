package volcano_async_executor

import (
	"fmt"

	"polystore_database/src/go/plan"
)

// filterOp は Filter の 1 バッチ処理。対象ストアで妥当な uuid を判定し、残った行のみ返す。
// 本体は同期版 (volcano_exec/op_filter.go) の filterIterator.process と同一で、
// pull ループと計測は asyncDriver 側へ移している。
//
// 状態を持たないため、複数ワーカーから同時に process を呼んで安全。
type filterOp struct {
	p *Processor
	o *plan.Filter
}

func (f *filterOp) process(in *Batch) (*Batch, error) {
	filterIdxIn := f.o.InputSlot.VarToSlot[f.o.Alias]

	// バッチ内 uuid のユニーク化
	idSet := make(map[string]struct{}, in.n)
	for i := 0; i < in.n; i++ {
		idSet[in.get(i, filterIdxIn)] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idSet))
	for id := range idSet {
		uniqueIDs = append(uniqueIDs, id)
	}

	valid, err := f.p.filterValid(f.o, uniqueIDs)
	if err != nil {
		return nil, err
	}

	outSlots := len(f.o.OutputSlot.VarToSlot)
	out := newBatch(outSlots, in.n)
	for i := 0; i < in.n; i++ {
		if _, ok := valid[in.get(i, filterIdxIn)]; ok {
			out.appendRow(remap(in.row(i), f.o.InputSlot, f.o.OutputSlot))
		}
	}
	return out, nil
}

// filterValid は対象ストアで uniqueIDs のうち条件を満たす uuid 集合を返す（実装は access_<store>.go）。
func (p *Processor) filterValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	switch o.DataStore {
	case "graph", "", "unknown":
		return p.filterGraphValid(o, ids)
	case "document":
		return p.filterDocValid(o, ids)
	case "kvs":
		return p.filterKvsValid(o, ids)
	case "relational":
		return p.filterRdbValid(o, ids)
	case "columnar":
		return p.filterColValid(o, ids)
	default:
		return nil, fmt.Errorf("未知のフィルタ対象ストア: %s", o.DataStore)
	}
}

// remap は InputSlot 基準の行を OutputSlot 基準へ引き継ぐ。
func remap(in []string, inSlot, outSlot plan.SlotTable) []string {
	out := make([]string, len(outSlot.VarToSlot))
	for alias, outIdx := range outSlot.VarToSlot {
		if inIdx, ok := inSlot.VarToSlot[alias]; ok && inIdx < len(in) {
			out[outIdx] = in[inIdx]
		}
	}
	return out
}
