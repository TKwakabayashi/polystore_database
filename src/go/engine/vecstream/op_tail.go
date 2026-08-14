package vecstream

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/plan"
)

// Row は tail 演算子間を流れる行（wide row）。
// キーは "alias.prop"（取得プロパティ）/ "alias"（束縛 uuid）/ 集約出力名。
type Row = map[string]interface{}

// runRow は tail（StoreFragment/Projection/Aggregate/Sort/Limit/Return）を実行する。
// Projection が record(Batch) 経路と row 経路の橋渡し点で、そこで push 並列パイプラインを起動する。
// Aggregate/Sort/Limit/Return は pipeline breaker のためメモリ内・直列（並列化の対象は record 側）。
func (p *Processor) runRow(op plan.PlanNode) ([]Row, error) {
	if op == nil {
		return nil, fmt.Errorf("nil row operator")
	}

	switch o := op.(type) {
	case *plan.StoreFragment:
		// tail 委譲形（Plan に束縛フラグメントが入れ子）は bulk 専用。vecstream は Plan を
		// そのまま通常実行してフォールバックする（結果は等価）。
		if _, ok := plan.LowerTail(o.Plan); ok {
			return p.runRow(o.Plan)
		}
		// 全体委譲: Plan を lowering してネイティブ発行。
		start := time.Now()
		rows, err := p.runStoreFragment(o)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		step := p.newStep()
		name := "Fragment[" + o.Store.String() + "]"
		p.recordOp(step, name, now.Sub(start), len(rows))
		p.recordFlow(step, name, 0, 0, 0, int64(len(rows)), 1, start, now)
		return rows, nil

	case *plan.Projection:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		if err := child.Open(p.ctx); err != nil {
			return nil, err
		}
		defer child.Close(p.ctx)
		step := p.newStep() // record 演算子の後（葉→根）
		return p.runProjection(o, child, step)

	case *plan.Aggregate:
		in, err := p.runRow(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyAggregate(o, in)
		now := time.Now()
		step := p.newStep()
		p.recordOp(step, "Aggregate", now.Sub(start), len(rows))
		p.recordFlow(step, "Aggregate", 0, 0, int64(len(in)), int64(len(rows)), 0, start, now)
		return rows, nil

	case *plan.Sort:
		in, err := p.runRow(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applySort(o, in)
		now := time.Now()
		step := p.newStep()
		p.recordOp(step, "Sort", now.Sub(start), len(rows))
		p.recordFlow(step, "Sort", 0, 0, int64(len(in)), int64(len(rows)), 0, start, now)
		return rows, nil

	case *plan.Limit:
		in, err := p.runRow(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyLimit(o, in)
		now := time.Now()
		step := p.newStep()
		p.recordOp(step, "Limit", now.Sub(start), len(rows))
		p.recordFlow(step, "Limit", 0, 0, int64(len(in)), int64(len(rows)), 0, start, now)
		return rows, nil

	case *plan.Return:
		in, err := p.runRow(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyReturn(o, in)
		now := time.Now()
		step := p.newStep()
		p.recordOp(step, "Return", now.Sub(start), len(rows))
		p.recordFlow(step, "Return", 0, 0, int64(len(in)), int64(len(rows)), 0, start, now)
		return rows, nil

	default:
		return nil, fmt.Errorf("unexpected row operator: %T", op)
	}
}

// ---- Aggregate（pipeline breaker）----
func applyAggregate(o *plan.Aggregate, in []Row) []Row {
	type acc struct {
		groupVals map[string]interface{}
		count     int64 // count(*)
		sums      []float64
		mins      []interface{}
		maxs      []interface{}
		seen      []map[string]struct{} // DISTINCT
		counts    []int64               // 各 agg の非 null / distinct 件数
	}

	order := []string{}
	groups := map[string]*acc{}

	for _, r := range in {
		keyParts := make([]string, len(o.GroupKeys))
		gvals := make(map[string]interface{}, len(o.GroupKeys))
		for i, gk := range o.GroupKeys {
			ck := gk.Alias + "." + gk.Prop
			v := r[ck]
			keyParts[i] = fmt.Sprintf("%v", v)
			gvals[ck] = v
		}
		gkey := strings.Join(keyParts, "\x1f")

		a, ok := groups[gkey]
		if !ok {
			a = &acc{
				groupVals: gvals,
				sums:      make([]float64, len(o.Aggs)),
				mins:      make([]interface{}, len(o.Aggs)),
				maxs:      make([]interface{}, len(o.Aggs)),
				seen:      make([]map[string]struct{}, len(o.Aggs)),
				counts:    make([]int64, len(o.Aggs)),
			}
			for i := range a.seen {
				a.seen[i] = map[string]struct{}{}
			}
			groups[gkey] = a
			order = append(order, gkey)
		}
		a.count++

		for i, ag := range o.Aggs {
			var v interface{}
			present := false
			switch {
			case ag.Alias == "" && ag.Prop == "": // count(*)
				present = true
			case ag.Prop == "": // count(alias): 束縛の有無
				id, _ := r[ag.Alias].(string)
				present = id != ""
				v = id
			default:
				v = r[ag.Alias+"."+ag.Prop]
				present = v != nil
			}
			if !present {
				continue
			}
			if ag.Distinct {
				dk := fmt.Sprintf("%v", v)
				if _, dup := a.seen[i][dk]; dup {
					continue
				}
				a.seen[i][dk] = struct{}{}
			}
			a.counts[i]++
			switch ag.Func {
			case plan.AggSum, plan.AggAvg:
				a.sums[i] += toFloat64(v)
			case plan.AggMin:
				if a.mins[i] == nil || compareValues(v, a.mins[i]) < 0 {
					a.mins[i] = v
				}
			case plan.AggMax:
				if a.maxs[i] == nil || compareValues(v, a.maxs[i]) > 0 {
					a.maxs[i] = v
				}
			}
		}
	}

	rows := make([]Row, 0, len(order))
	for _, gkey := range order {
		a := groups[gkey]
		row := make(Row)
		for ck, v := range a.groupVals {
			row[ck] = v
		}
		for i, ag := range o.Aggs {
			switch ag.Func {
			case plan.AggCount:
				if ag.Alias == "" && ag.Prop == "" && !ag.Distinct {
					row[ag.OutName] = a.count
				} else {
					row[ag.OutName] = a.counts[i]
				}
			case plan.AggSum:
				row[ag.OutName] = a.sums[i]
			case plan.AggAvg:
				if a.counts[i] > 0 {
					row[ag.OutName] = a.sums[i] / float64(a.counts[i])
				} else {
					row[ag.OutName] = nil
				}
			case plan.AggMin:
				row[ag.OutName] = a.mins[i]
			case plan.AggMax:
				row[ag.OutName] = a.maxs[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// ---- Sort（pipeline breaker）----
func applySort(o *plan.Sort, in []Row) []Row {
	sort.SliceStable(in, func(i, j int) bool {
		for _, oi := range o.OrderItems {
			key := oi.Key
			if key == "" {
				key = oi.Alias + "." + oi.Prop
			}
			res := compareValues(in[i][key], in[j][key])
			if res != 0 {
				if oi.Direction == plan.OrderAsc {
					return res < 0
				}
				return res > 0
			}
		}
		return false
	})
	return in
}

// ---- Limit ----
func applyLimit(o *plan.Limit, in []Row) []Row {
	if o.Count >= 0 && len(in) > o.Count {
		return in[:o.Count]
	}
	return in
}

// ---- Return（表示整形: AS / coalesce / 集約）----
func applyReturn(o *plan.Return, in []Row) []Row {
	out := make([]Row, 0, len(in))
	for _, r := range in {
		row := make(Row, len(o.Items))
		for _, item := range o.Items {
			row[item.Name] = shapeValue(item, r)
		}
		out = append(out, row)
	}
	return out
}

func shapeValue(item plan.ReturnItem, r Row) interface{} {
	if item.IsAggregate {
		return r[item.Agg.OutName]
	}
	if item.IsCoalesce {
		for _, p := range item.Props {
			if v, ok := r[item.Alias+"."+p]; ok && v != nil {
				return v
			}
		}
		return nil
	}
	if len(item.Props) > 0 && item.Props[0] != "" {
		return r[item.Alias+"."+item.Props[0]]
	}
	return r[item.Alias]
}

// ---- 比較・数値ヘルパ ----
func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	if af, aok := asFloat(a); aok {
		if bf, bok := asFloat(b); bok {
			switch {
			case af < bf:
				return -1
			case af > bf:
				return 1
			default:
				return 0
			}
		}
	}
	switch va := a.(type) {
	case string:
		if vb, ok := b.(string); ok {
			switch {
			case va < vb:
				return -1
			case va > vb:
				return 1
			default:
				return 0
			}
		}
	case time.Time:
		if vb, ok := b.(time.Time); ok {
			switch {
			case va.Before(vb):
				return -1
			case va.After(vb):
				return 1
			default:
				return 0
			}
		}
	}
	return 0
}

func asFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	default:
		return 0, false
	}
}

func toFloat64(v interface{}) float64 {
	if f, ok := asFloat(v); ok {
		return f
	}
	if s, ok := v.(string); ok {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	return 0
}
