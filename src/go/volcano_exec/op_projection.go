package volcano_executor

import (
	"time"

	"polystore_database/src/go/plan"
)

// runProjection は Projection のマテリアライズ実装。子から vectorWidth 幅のバッチを pull しつつ、
// バッチ単位で必要プロパティを一括取得し、wide row を生成する。
// wide row のキーは "alias.prop"（取得プロパティ）と "alias"（束縛 uuid）。
// バッチ幅は vectorWidth に従うため、Volcano ではプロパティ取得も tuple-at-a-time となる。
func (p *Processor) runProjection(o *plan.Projection, child Iterator) ([]Row, error) {
	step := p.newStep()
	aliasToSlot := o.InputSlot.VarToSlot

	var out []Row
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
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= batch.slotCount() {
					continue
				}
				if id := batch.get(i, slotIdx); id != "" {
					unitIDMap[unit.Alias][id] = struct{}{}
				}
			}
		}

		// B. プロパティ一括取得 cache[alias][id][prop] = value
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
				data := p.fetchPropertiesBulk(ids, &unit, &unit.Fetches[fi])
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

		// C. wide row 生成
		for i := 0; i < batch.n; i++ {
			row := make(Row)
			for alias, slotIdx := range aliasToSlot {
				if slotIdx < batch.slotCount() {
					row[alias] = batch.get(i, slotIdx) // 束縛 uuid
				}
			}
			for _, unit := range o.Units {
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= batch.slotCount() {
					continue
				}
				props := cache[unit.Alias][batch.get(i, slotIdx)]
				for _, f := range unit.Fetches {
					for _, pr := range f.Props {
						row[unit.Alias+"."+pr] = props[pr]
					}
				}
			}
			out = append(out, row)
		}
		p.recordOp(step, "Projection", time.Since(start), batch.n)
	}
	return out, nil
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
