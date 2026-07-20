package stream

import (
	"polystore_database/src/go/plan"
)

// streamProjection は record ストリームを消費し、必要列を単一パスでクロスストア収集して
// wide row（キー: "alias.prop" と 束縛 uuid の "alias"）を row ストリームへ emit する。
func streamProjection(qp *QueryProcessor, o *plan.Projection, inputStream <-chan []Record, out chan<- []Row) int {
	if qp == nil {
		return 0
	}
	aliasToSlot := o.InputSlot.VarToSlot
	emitted := 0

	for batch := range inputStream {
		// --- Unit ごとに ID 収集 ---
		unitIDMap := make(map[string]map[string]struct{})
		for _, unit := range o.Units {
			unitIDMap[unit.Alias] = make(map[string]struct{})
		}
		for _, r := range batch {
			for _, unit := range o.Units {
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= len(r.Slots) {
					continue
				}
				if id := r.Slots[slotIdx]; id != "" {
					unitIDMap[unit.Alias][id] = struct{}{}
				}
			}
		}

		// --- プロパティ一括フェッチ ---
		cache := make(map[string]map[string]map[string]interface{})
		for ui := range o.Units {
			unit := o.Units[ui]
			idSet := unitIDMap[unit.Alias]
			if len(idSet) == 0 {
				continue
			}
			ids := make([]string, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			cache[unit.Alias] = make(map[string]map[string]interface{})
			for fi := range unit.Fetches {
				data := FetchPropertiesBulk(qp, ids, &unit, &unit.Fetches[fi])
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

		// --- wide row 生成 ---
		rows := make([]Row, 0, len(batch))
		for _, r := range batch {
			row := make(Row)
			for alias, slotIdx := range aliasToSlot {
				if slotIdx < len(r.Slots) {
					row[alias] = r.Slots[slotIdx] // 束縛 uuid
				}
			}
			for _, unit := range o.Units {
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= len(r.Slots) {
					continue
				}
				props := cache[unit.Alias][r.Slots[slotIdx]]
				for _, f := range unit.Fetches {
					for _, p := range f.Props {
						row[unit.Alias+"."+p] = props[p]
					}
				}
			}
			rows = append(rows, row)
		}
		emitted += len(rows)
		out <- rows
	}
	return emitted
}

// FetchPropertiesBulk は各ストアからプロパティを一括取得する共通入口（既存のまま）。
func FetchPropertiesBulk(qp *QueryProcessor, ids []string, unit *plan.ProjectionUnit, plan *plan.FetchPlan) map[string]map[string]interface{} {
	switch plan.Store {
	case "graph":
		return fetchGraphPropsStream(qp, ids, unit, plan)
	case "document":
		return fetchDocPropsStream(qp, ids, unit, plan)
	case "kvs":
		return fetchKvsPropsStream(qp, ids, unit, plan)
	case "columnar":
		return fetchColPropsStream(qp, ids, unit, plan)
	case "relational":
		return fetchRdbPropsStream(qp, ids, unit, plan)
	default:
		return nil
	}
}
