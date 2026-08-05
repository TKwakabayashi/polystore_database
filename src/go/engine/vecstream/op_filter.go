package vecstream

import (
	"fmt"

	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// filterOp は Filter の 1 バッチ処理。対象ストアで妥当な uuid を判定し、残った行のみ返す。
// 本体は engine/volcano/op_filter.go の filterIterator.process と同一。状態を持たないため、
// 複数ワーカーから同時に process を呼んで安全。
type filterOp struct {
	p *Processor
	o *plan.Filter
}

// process はワーカーが使い回す sess を受け取る（graph フィルタのみ使用。非 graph は共有ハンドル）。
func (f *filterOp) process(sess neo4j.SessionWithContext, in *Batch) (*Batch, error) {
	filterIdxIn := f.o.InputSlot.VarToSlot[f.o.Alias]

	// バッチ内 uuid のユニーク化
	idSet := make(map[uid.UUID]struct{}, in.n)
	for i := 0; i < in.n; i++ {
		idSet[in.get(i, filterIdxIn)] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idSet))
	for id := range idSet {
		uniqueIDs = append(uniqueIDs, id.String())
	}

	valid, err := f.p.filterValid(sess, f.o, uniqueIDs)
	if err != nil {
		return nil, err
	}

	outSlots := len(f.o.OutputSlot.VarToSlot)
	out := newBatch(outSlots, in.n)
	for i := 0; i < in.n; i++ {
		if _, ok := valid[in.get(i, filterIdxIn).String()]; ok {
			out.appendRow(remap(in.row(i), f.o.InputSlot, f.o.OutputSlot))
		}
	}
	return out, nil
}

// filterValid は対象ストアで uniqueIDs のうち条件を満たす uuid 集合を返す（実装は access_<store>.go）。
// sess は graph フィルタでのみ使う（ワーカー再利用セッション）。非 graph は共有ハンドルで sess を無視。
func (p *Processor) filterValid(sess neo4j.SessionWithContext, o *plan.Filter, ids []string) (map[string]struct{}, error) {
	switch o.DataStore {
	case store.Graph:
		return p.filterGraphValid(sess, o, ids)
	case store.Document:
		return p.filterDocValid(o, ids)
	case store.Kvs:
		return p.filterKvsValid(o, ids)
	case store.Relational:
		return p.filterRdbValid(o, ids)
	case store.Columnar:
		return p.filterColValid(o, ids)
	default:
		return nil, fmt.Errorf("未知のフィルタ対象ストア: %s", o.DataStore)
	}
}

// remap は InputSlot 基準の行を OutputSlot 基準へ引き継ぐ。
func remap(in []uid.UUID, inSlot, outSlot plan.SlotTable) []uid.UUID {
	out := make([]uid.UUID, len(outSlot.VarToSlot))
	for alias, outIdx := range outSlot.VarToSlot {
		if inIdx, ok := inSlot.VarToSlot[alias]; ok && inIdx < len(in) {
			out[outIdx] = in[inIdx]
		}
	}
	return out
}
