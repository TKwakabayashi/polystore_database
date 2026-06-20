package stream_executor

import (
	"fmt"
	"polystore_database/src/go/plan"
	"sort"
	"sync"
	"time"
)

func streamProjection(qp *QueryProcessor, o *plan.Projection, inputStream <-chan []Record) error {
	const batchSize = 500
	const workerCount = 1

	if qp == nil || qp.slotTable == nil {
		return fmt.Errorf("query processor or slot table is nil")
	}

	aliasToSlot := o.InputSlot.VarToSlot
	projectedRowChan := make(chan map[string]interface{}, 1000)

	var wg sync.WaitGroup

	// 1. Worker Pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range inputStream {
				// --- A. ID収集（このバッチに含まれる全UnitのIDをセット化） ---
				unitIDMap := make(map[string]map[string]struct{})
				for _, unit := range o.Units {
					unitIDMap[unit.Alias] = make(map[string]struct{})
				}

				for _, r := range batch {
					for _, unit := range o.Units {
						slotIdx := aliasToSlot[unit.Alias]
						id := r.Slots[slotIdx]
						if id != "" {
							unitIDMap[unit.Alias][id] = struct{}{}
						}
					}
				}
				// --- B. プロパティフェッチ（Unit単位のループを維持） ---
				// cache[alias][id][propName] = value
				cache := make(map[string]map[string]map[string]interface{})

				for _, unit := range o.Units {
					idSet := unitIDMap[unit.Alias]
					if len(idSet) == 0 {
						continue
					}

					// IDセットをスライスに変換（このバッチで必要なIDのみ）
					ids := make([]string, 0, len(idSet))
					for id := range idSet {
						ids = append(ids, id)
					}

					cache[unit.Alias] = make(map[string]map[string]interface{})

					for _, plan := range unit.Fetches {
						data := FetchPropertiesBulk(qp, ids, &unit, &plan)

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

				// --- C. 射影（既存の ProjectRow ロジックを適用） ---
				for _, r := range batch {
					projectedRowChan <- ProjectRow(r, o.Items, aliasToSlot, cache)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(projectedRowChan)
	}()

	// 3. 集約・ソート・リミット (以前のロジックを維持)
	var allResults []map[string]interface{}
	for row := range projectedRowChan {
		allResults = append(allResults, row)
	}

	qp.results = applySortAndLimit(o, allResults) // 既存のソート・リミットロジックを分離して呼び出し

	return nil
}

// FetchPropertiesBulk: 各種ストレージからプロパティを一括取得する共通の入り口
func FetchPropertiesBulk(qp *QueryProcessor, ids []string, unit *plan.ProjectionUnit, plan *plan.FetchPlan) map[string]map[string]interface{} {
	switch plan.Store {
	case "graph":
		return fetchGraphPropsStream(qp, ids, unit, plan)
	case "document":
		// return fetchDocPropsBulk(qp, ids, unit, plan)
		return nil
	default:
		return nil
	}
}

// ProjectRow: 1つの Record と キャッシュから、指定された ReturnItem に基づいて Map を生成する
func ProjectRow(r Record, items []plan.ReturnItem, aliasToSlot map[string]int, cache map[string]map[string]map[string]interface{}) map[string]interface{} {
	row := make(map[string]interface{})
	for _, item := range items {
		id := r.Slots[aliasToSlot[item.Alias]]
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
	// 5. ソート処理
	if o.OrderItems != nil && len(o.OrderItems) > 0 {
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

	// 6. リミット処理
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
