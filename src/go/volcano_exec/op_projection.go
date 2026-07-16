package volcano_executor

import (
	"sort"
	"time"

	"polystore_database/src/go/plan"
)

// runProjection は Projection のシンク実装。子から全バッチを pull しつつ、
// バッチ単位でプロパティを取得・射影し、最後にソート・リミットを適用する。
// バッチ幅は vectorWidth に従うため、Volcano では射影のプロパティ取得も tuple-at-a-time となる。
func (p *Processor) runProjection(o *plan.Projection, child Iterator) ([]map[string]interface{}, error) {
	step := p.nextStep + 1
	p.nextStep = step
	aliasToSlot := o.InputSlot.VarToSlot

	var allResults []map[string]interface{}
	for {
		batch, err := child.Next(p.ctx)
		if err != nil {
			return nil, err
		}
		if batch == nil {
			break
		}
		start := time.Now()

		// A. バッチ内の Unit ごとの ID 集合
		unitIDMap := make(map[string]map[string]struct{})
		for _, unit := range o.Units {
			unitIDMap[unit.Alias] = make(map[string]struct{})
		}
		for i := 0; i < batch.n; i++ {
			for _, unit := range o.Units {
				slotIdx := aliasToSlot[unit.Alias]
				id := batch.get(i, slotIdx)
				if id != "" {
					unitIDMap[unit.Alias][id] = struct{}{}
				}
			}
		}

		// B. プロパティフェッチ cache[alias][id][prop] = value
		cache := make(map[string]map[string]map[string]interface{})
		for _, unit := range o.Units {
			idSet := unitIDMap[unit.Alias]
			if len(idSet) == 0 {
				continue
			}
			ids := make([]string, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			cache[unit.Alias] = make(map[string]map[string]interface{})
			for i := range unit.Fetches {
				data := p.fetchPropertiesBulk(ids, &unit, &unit.Fetches[i])
				for id, propsMap := range data {
					if _, ok := cache[unit.Alias][id]; !ok {
						cache[unit.Alias][id] = make(map[string]interface{})
					}
					for pName, pVal := range propsMap {
						cache[unit.Alias][id][pName] = pVal
					}
				}
			}
		}

		// C. 射影
		for i := 0; i < batch.n; i++ {
			allResults = append(allResults, projectRow(batch.row(i), o.Items, aliasToSlot, cache))
		}
		p.recordOp(step, "Projection", time.Since(start), batch.n)
	}

	final := applySortAndLimit(o, allResults)
	p.results = final
	return final, nil
}

// fetchPropertiesBulk は各種ストアからプロパティを一括取得する共通の入り口（実装は access_<store>.go）。
func (p *Processor) fetchPropertiesBulk(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	switch fetch.Store {
	case "graph":
		return p.fetchGraphProps(ids, unit, fetch)
	case "document":
		return p.fetchDocProps(ids, unit, fetch)
	case "kvs":
		return p.fetchKvsProps(ids, unit, fetch)
	case "columnar":
		return p.fetchColProps(ids, unit, fetch)
	case "relational":
		return p.fetchRdbProps(ids, unit, fetch)
	default:
		return nil
	}
}

// ---------- 射影・ソート・リミット ----------

func projectRow(row []string, items []plan.ReturnItem, aliasToSlot map[string]int, cache map[string]map[string]map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, item := range items {
		id := ""
		if idx, ok := aliasToSlot[item.Alias]; ok && idx < len(row) {
			id = row[idx]
		}
		entityCache := cache[item.Alias][id]

		var finalVal interface{}
		if item.IsCoalesce {
			for _, prop := range item.Props {
				if val, ok := entityCache[prop]; ok && val != nil {
					finalVal = val
					break
				}
			}
		} else if len(item.Props) > 0 {
			finalVal = entityCache[item.Props[0]]
		}
		out[item.Name] = finalVal
	}
	return out
}

func applySortAndLimit(o *plan.Projection, results []map[string]interface{}) []map[string]interface{} {
	if len(o.OrderItems) > 0 {
		sort.SliceStable(results, func(i, j int) bool {
			for _, order := range o.OrderItems {
				sortKey := order.Alias + "." + order.Prop
				res := compareValues(results[i][sortKey], results[j][sortKey])
				if res != 0 {
					if order.Direction == plan.OrderAsc {
						return res < 0
					}
					return res > 0
				}
			}
			return false
		})
	}
	if o.Limit > 0 && len(results) > o.Limit {
		results = results[:o.Limit]
	}
	return results
}

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
	switch va := a.(type) {
	case int, int32, int64:
		valA := toInt64(va)
		valB := toInt64(b)
		if valA < valB {
			return -1
		}
		if valA > valB {
			return 1
		}
		return 0
	case string:
		vb, ok := b.(string)
		if !ok {
			return 0
		}
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
		return 0
	case time.Time:
		vb, ok := b.(time.Time)
		if !ok {
			return 0
		}
		if va.Before(vb) {
			return -1
		}
		if va.After(vb) {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	default:
		return 0
	}
}
