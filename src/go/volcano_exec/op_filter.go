package volcano_executor

import (
	"context"
	"fmt"
	"time"

	"polystore_database/src/go/plan"
)

// filterIterator は Filter の pull 実装。子から 1 バッチ pull し、対象ストアで
// 妥当な uuid を判定して残った行のみを出力する。空になったら次のバッチを引き続き pull。
type filterIterator struct {
	p     *Processor
	o     *plan.Filter
	child Iterator
	step  int
}

func (f *filterIterator) Open(ctx context.Context) error  { return f.child.Open(ctx) }
func (f *filterIterator) Close(ctx context.Context) error { return f.child.Close(ctx) }

func (f *filterIterator) Next(ctx context.Context) (*Batch, error) {
	for {
		in, err := f.child.Next(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, nil
		}
		start := time.Now()
		out, err := f.process(in)
		if err != nil {
			return nil, err
		}
		f.p.recordOp(f.step, "Filter", time.Since(start), out.n)
		if out.n > 0 {
			return out, nil
		}
		// 全滅したバッチは EOF と紛れないよう、次を pull する。
	}
}

func (f *filterIterator) process(in *Batch) (*Batch, error) {
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
