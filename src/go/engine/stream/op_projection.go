package stream

import (
	"time"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// streamProjection は record ストリームを消費し、必要列を単一パスでクロスストア収集して
// wide row（キー: "alias.prop" と 束縛 uuid の "alias"）を row ストリームへ emit する。
// フロー計測は per-batch で RecordFlow（RecordOp は呼び出し元 spawnRowOp が担当）。
func streamProjection(qp *Processor, o *plan.Projection, step int, inputStream <-chan []Record, out chan<- []Row) int {
	if qp == nil {
		return 0
	}
	aliasToSlot := o.InputSlot.VarToSlot
	emitted := 0

	for batch := range inputStream {
		t0 := time.Now()
		var queries int64
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
				if id := r.Slots[slotIdx]; !id.Empty() {
					unitIDMap[unit.Alias][id.String()] = struct{}{}
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
				queries++
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
					row[alias] = r.Slots[slotIdx].String() // 束縛 uuid
				}
			}
			for _, unit := range o.Units {
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= len(r.Slots) {
					continue
				}
				props := cache[unit.Alias][r.Slots[slotIdx].String()]
				for _, f := range unit.Fetches {
					for _, p := range f.Props {
						row[unit.Alias+"."+p] = props[p]
					}
				}
			}
			rows = append(rows, row)
		}
		emitted += len(rows)
		qp.recordFlow(step, "Projection", 1, 0, int64(len(batch)), int64(len(rows)), queries, t0, time.Now())
		out <- rows
	}
	return emitted
}

// FetchPropertiesBulk は各ストアからプロパティを一括取得する共通入口（既存のまま）。
func FetchPropertiesBulk(qp *Processor, ids []string, unit *plan.ProjectionUnit, plan *plan.FetchPlan) map[string]map[string]interface{} {
	switch plan.Store {
	case store.Graph:
		return fetchGraphPropsStream(qp, ids, unit, plan)
	case store.Document:
		return fetchDocPropsStream(qp, ids, unit, plan)
	case store.Kvs:
		return fetchKvsPropsStream(qp, ids, unit, plan)
	case store.Columnar:
		return fetchColPropsStream(qp, ids, unit, plan)
	case store.Relational:
		return fetchRdbPropsStream(qp, ids, unit, plan)
	default:
		return nil
	}
}
