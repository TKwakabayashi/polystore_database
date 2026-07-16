package bulk_executor

import (
	"polystore_database/src/go/plan"
)

// bulkProjection は record を消費し、必要列を単一パスでクロスストア収集して
// wide row（キー: "alias.prop" と 束縛 uuid の "alias"）を生成する。
// record-stream と row-stream の橋渡し点。
func bulkProjection(qp *QueryProcessor, o *plan.Projection, in []Record) []Row {
	aliasToSlot := o.InputSlot.VarToSlot

	// --- Unit ごとに ID 収集（全行分） ---
	unitIDMap := make(map[string]map[string]struct{})
	for _, unit := range o.Units {
		unitIDMap[unit.Alias] = make(map[string]struct{})
	}
	for _, r := range in {
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

	// --- プロパティ一括フェッチ cache[alias][id][prop] = value ---
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
	rows := make([]Row, 0, len(in))
	for _, r := range in {
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
	return rows
}

// FetchPropertiesBulk は各ストアからプロパティを一括取得する共通入口。
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
