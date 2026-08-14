package stream

import (
	"fmt"
	"polystore_database/src/go/plan"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// executeRowStream は tail（Projection/Aggregate/Sort/Limit/Return）を row ストリームで実行する。
// Projection が record-stream と row-stream の橋渡し点。
func executeRowStream(qp *Processor, op plan.PlanNode, counter *int, wg *sync.WaitGroup) (chan []Row, error) {
	if op == nil {
		return nil, fmt.Errorf("Empty Operator Passed (row stream)")
	}

	switch o := op.(type) {
	case *plan.TailPushdown:
		// tail pushdown は bulk 専用。stream は Fallback（元 coordinator tail）を通常実行（結果等価）。
		return executeRowStream(qp, o.Fallback, counter, wg)

	case *plan.StorePushdown:
		return spawnRowOp(qp, "Pushdown["+o.Store.String()+"]", counter, wg, func(step int, out chan []Row) int {
			return streamStorePushdown(qp, o, out)
		}), nil

	case *plan.StoreFragment:
		sp := plan.StorePushdownFromFragment(o)
		return spawnRowOp(qp, "Fragment["+o.Store.String()+"]", counter, wg, func(step int, out chan []Row) int {
			return streamStorePushdown(qp, sp, out)
		}), nil

	case *plan.Projection:
		recCh, err := ExecuteOperatorStream(qp, o.Input, counter, wg)
		if err != nil {
			return nil, err
		}
		return spawnRowOp(qp, "Projection", counter, wg, func(step int, out chan []Row) int {
			return streamProjection(qp, o, step, recCh, out)
		}), nil

	case *plan.Aggregate:
		in, err := executeRowStream(qp, o.Input, counter, wg)
		if err != nil {
			return nil, err
		}
		return spawnRowOp(qp, "Aggregate", counter, wg, func(step int, out chan []Row) int {
			return streamAggregate(o, in, out)
		}), nil

	case *plan.Sort:
		in, err := executeRowStream(qp, o.Input, counter, wg)
		if err != nil {
			return nil, err
		}
		return spawnRowOp(qp, "Sort", counter, wg, func(step int, out chan []Row) int {
			return streamSort(o, in, out)
		}), nil

	case *plan.Limit:
		in, err := executeRowStream(qp, o.Input, counter, wg)
		if err != nil {
			return nil, err
		}
		return spawnRowOp(qp, "Limit", counter, wg, func(step int, out chan []Row) int {
			return streamLimit(o, in, out)
		}), nil

	case *plan.Return:
		in, err := executeRowStream(qp, o.Input, counter, wg)
		if err != nil {
			return nil, err
		}
		return spawnRowOp(qp, "Return", counter, wg, func(step int, out chan []Row) int {
			return streamReturn(o, in, out)
		}), nil

	default:
		return nil, fmt.Errorf("unexpected row operator: %T", op)
	}
}

func spawnRowOp(qp *Processor, name string, counter *int, wg *sync.WaitGroup, run func(step int, out chan []Row) int) chan []Row {
	out := make(chan []Row, 500)
	*counter++
	step := *counter
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		t0 := time.Now()
		rows := run(step, out)
		t1 := time.Now()
		qp.recordOp(step, name, t1.Sub(t0), rows)
		// Projection は自身の per-batch フローを記録する（RecordFlow）ため二重計上を避ける。
		if name != "Projection" {
			qp.recordFlow(step, name, 0, 0, 0, int64(rows), 0, t0, t1)
		}
	}()
	return out
}

// ---- Aggregate（pipeline breaker）----
func streamAggregate(o *plan.Aggregate, in <-chan []Row, out chan<- []Row) int {
	var all []Row
	for batch := range in {
		all = append(all, batch...)
	}

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

	for _, r := range all {
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
	if len(rows) > 0 {
		out <- rows
	}
	return len(rows)
}

// ---- Sort（pipeline breaker）----
func streamSort(o *plan.Sort, in <-chan []Row, out chan<- []Row) int {
	var all []Row
	for batch := range in {
		all = append(all, batch...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		for _, oi := range o.OrderItems {
			key := oi.Key
			if key == "" {
				key = oi.Alias + "." + oi.Prop
			}
			res := compareValues(all[i][key], all[j][key])
			if res != 0 {
				if oi.Direction == plan.OrderAsc {
					return res < 0
				}
				return res > 0
			}
		}
		return false
	})
	if len(all) > 0 {
		out <- all
	}
	return len(all)
}

// ---- Limit ----
func streamLimit(o *plan.Limit, in <-chan []Row, out chan<- []Row) int {
	emitted := 0
	for batch := range in {
		if emitted >= o.Count {
			continue // 上流（Sort）はブロッキングなので drain のみ
		}
		if emitted+len(batch) > o.Count {
			batch = batch[:o.Count-emitted]
		}
		emitted += len(batch)
		out <- batch
	}
	return emitted
}

// ---- Return（表示整形: AS / coalesce）----
func streamReturn(o *plan.Return, in <-chan []Row, out chan<- []Row) int {
	emitted := 0
	for batch := range in {
		shaped := make([]Row, 0, len(batch))
		for _, r := range batch {
			row := make(Row, len(o.Items))
			for _, item := range o.Items {
				row[item.Name] = shapeValue(item, r)
			}
			shaped = append(shaped, row)
		}
		emitted += len(shaped)
		out <- shaped
	}
	return emitted
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
