package bulk_executor

import (
	"sort"
	"time"

	"polystore_database/src/go/plan"
)

// bulkProjection は Projection のシンク実装。全入力行に対して unit ごとのプロパティを
// 一括取得・射影し、最後にソート・リミットを適用して qp.results に格納する。
func bulkProjection(qp *QueryProcessor, o *plan.Projection, in []Record) error {
	aliasToSlot := o.InputSlot.VarToSlot

	// --- A. ID収集（全行に含まれる全UnitのIDをセット化） ---
	unitIDMap := make(map[string]map[string]struct{})
	for _, unit := range o.Units {
		unitIDMap[unit.Alias] = make(map[string]struct{})
	}
	for _, r := range in {
		for _, unit := range o.Units {
			slotIdx := aliasToSlot[unit.Alias]
			if slotIdx >= len(r.Slots) {
				continue
			}
			id := r.Slots[slotIdx]
			if id != "" {
				unitIDMap[unit.Alias][id] = struct{}{}
			}
		}
	}

	// --- B. プロパティフェッチ cache[alias][id][propName] = value ---
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
		for _, fetch := range unit.Fetches {
			data := FetchPropertiesBulk(qp, ids, &unit, &fetch)
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

	// --- C. 射影 ---
	var allResults []map[string]interface{}
	for _, r := range in {
		allResults = append(allResults, ProjectRow(r, o.Items, aliasToSlot, cache))
	}

	qp.results = applySortAndLimit(o, allResults)
	return nil
}

// FetchPropertiesBulk: 各種ストレージからプロパティを一括取得する共通の入り口。
func FetchPropertiesBulk(qp *QueryProcessor, ids []string, unit *plan.ProjectionUnit, plan *plan.FetchPlan) map[string]map[string]interface{} {
	switch plan.Store {
	case "graph":
		return fetchGraphPropsBulk(qp, ids, unit, plan)
	case "document":
		return fetchDocPropsBulk(qp, ids, unit, plan)
	case "kvs":
		return fetchKvsPropsBulk(qp, ids, unit, plan)
	case "columnar":
		return fetchColPropsBulk(qp, ids, unit, plan)
	case "relational":
		return fetchRdbPropsBulk(qp, ids, unit, plan)
	default:
		return nil
	}
}

// ProjectRow: 1つの Record とキャッシュから、指定された ReturnItem に基づいて Map を生成する。
func ProjectRow(r Record, items []plan.ReturnItem, aliasToSlot map[string]int, cache map[string]map[string]map[string]interface{}) map[string]interface{} {
	row := make(map[string]interface{})
	for _, item := range items {
		id := ""
		if idx, ok := aliasToSlot[item.Alias]; ok && idx < len(r.Slots) {
			id = r.Slots[idx]
		}
		entityCache := cache[item.Alias][id]

		var finalVal interface{}
		if item.IsCoalesce {
			for _, p := range item.Props {
				if val, ok := entityCache[p]; ok && val != nil {
					finalVal = val
					break
				}
			}
		} else if len(item.Props) > 0 {
			finalVal = entityCache[item.Props[0]]
		}
		row[item.Name] = finalVal
	}
	return row
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
