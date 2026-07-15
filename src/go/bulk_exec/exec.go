package bulk_executor

import (
	"fmt"
	"time"

	"polystore_database/src/go/plan"
)

// Run は plan ツリーを全件マテリアライズ実行し、最終結果行を返す。
func (p *Processor) Run(op plan.PlanNode) ([]map[string]interface{}, error) {
	if op == nil {
		return nil, fmt.Errorf("nil plan node")
	}
	if proj, ok := op.(*plan.Projection); ok {
		return p.execProjection(proj) // sink
	}
	recs, err := p.execute(op)
	if err != nil {
		return nil, err
	}
	return p.materializeRaw(recs), nil
}

// execute は子を全件実行してから自演算子を全件処理し、[]Record を返す。
func (p *Processor) execute(op plan.PlanNode) ([]Record, error) {
	switch o := op.(type) {
	case *plan.EntityScan:
		return p.execScan(o)
	case *plan.Filter:
		return p.execFilter(o)
	case *plan.Expand:
		return p.execExpand(o)
	case *plan.VarLengthExpand:
		return p.execVarExpand(o)
	case *plan.Projection:
		return nil, fmt.Errorf("projection は root 以外に置けません")
	default:
		return nil, fmt.Errorf("未知の演算子: %T", op)
	}
}

// execScan は EntityScan を全件実行する。
func (p *Processor) execScan(o *plan.EntityScan) ([]Record, error) {
	p.nextStep++
	step := p.nextStep
	start := time.Now()

	slotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	ids, err := p.scanIDs(o)
	if err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(ids))
	for _, id := range ids {
		row := make(Record, slotCount)
		row[aliasIdx] = id
		out = append(out, row)
	}
	p.recordOp(step, "EntityScan", time.Since(start), 0, len(out))
	return out, nil
}

// execFilter は Filter を全件実行する（子実行は計測外、自処理のみ計測）。
func (p *Processor) execFilter(o *plan.Filter) ([]Record, error) {
	in, err := p.execute(o.Input)
	if err != nil {
		return nil, err
	}
	p.nextStep++
	step := p.nextStep
	start := time.Now()

	filterIdx := o.InputSlot.VarToSlot[o.Alias]
	ids := uniqueSlot(in, filterIdx)
	valid, err := p.filterValid(o, ids)
	if err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(in))
	for _, r := range in {
		if filterIdx >= len(r) {
			continue
		}
		if _, ok := valid[r[filterIdx]]; ok {
			out = append(out, Record(remap([]string(r), o.InputSlot, o.OutputSlot)))
		}
	}
	p.recordOp(step, "Filter", time.Since(start), len(in), len(out))
	return out, nil
}

// materializeRaw は Projection の無いプランで []Record を slot%d キーの map 化する（補助経路）。
func (p *Processor) materializeRaw(recs []Record) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(recs))
	for _, r := range recs {
		row := make(map[string]interface{}, len(r))
		for s, v := range r {
			row[fmt.Sprintf("slot%d", s)] = v
		}
		out = append(out, row)
	}
	return out
}

// uniqueSlot は in の idx スロットの値をユニーク化して（出現順で）返す。
func uniqueSlot(in []Record, idx int) []string {
	set := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, r := range in {
		if idx >= len(r) {
			continue
		}
		id := r[idx]
		if _, ok := set[id]; ok {
			continue
		}
		set[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
