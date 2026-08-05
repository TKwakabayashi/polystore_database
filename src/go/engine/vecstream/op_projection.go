package vecstream

import (
	"time"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// runProjection は Projection のマテリアライズ実装。child（exchange 経由の並列パイプライン）を
// pull しつつ、バッチ単位で必要プロパティをクロスストア一括取得し wide row を生成する。
// child.Next はここ（単一 goroutine）からのみ呼ばれるが、child 内部の exchange が上流を並列駆動する。
//
// graph プロパティ取得用の Neo4j セッションは 1 本だけ生成して全バッチで使い回す（session 再利用）。
// wide row のキーは "alias.prop"（取得プロパティ）と "alias"（束縛 uuid）。
func (p *Processor) runProjection(o *plan.Projection, child Iterator, step int) ([]Row, error) {
	aliasToSlot := o.InputSlot.VarToSlot

	sess := p.newReadSession() // projection 全体で 1 本を使い回す
	defer p.closeSession(sess)

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
				if id := batch.get(i, slotIdx); !id.Empty() {
					unitIDMap[unit.Alias][id.String()] = struct{}{}
				}
			}
		}

		// B. プロパティ一括取得 cache[alias][id][prop] = value
		var queries int64
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
				data := p.fetchPropertiesBulk(sess, ids, &unit, &unit.Fetches[fi])
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
					row[alias] = batch.get(i, slotIdx).String() // 束縛 uuid
				}
			}
			for _, unit := range o.Units {
				slotIdx, ok := aliasToSlot[unit.Alias]
				if !ok || slotIdx >= batch.slotCount() {
					continue
				}
				props := cache[unit.Alias][batch.get(i, slotIdx).String()]
				for _, f := range unit.Fetches {
					for _, pr := range f.Props {
						row[unit.Alias+"."+pr] = props[pr]
					}
				}
			}
			out = append(out, row)
		}
		now := time.Now()
		p.recordOp(step, "Projection", now.Sub(start), batch.n)
		p.recordFlow(step, "Projection", 1, 0, int64(batch.n), int64(batch.n), queries, start, now)
	}
	return out, nil
}

// fetchPropertiesBulk は各種ストアからプロパティを一括取得する共通の入り口（実装は access_<store>.go）。
// sess は graph 取得でのみ使う（projection が使い回すセッション）。非 graph は共有ハンドルで sess を無視。
func (p *Processor) fetchPropertiesBulk(sess neo4j.SessionWithContext, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	switch fetch.Store {
	case store.Graph:
		return p.fetchGraphProps(sess, ids, unit, fetch)
	case store.Document:
		return p.fetchDocProps(ids, unit, fetch)
	case store.Kvs:
		return p.fetchKvsProps(ids, unit, fetch)
	case store.Columnar:
		return p.fetchColProps(ids, unit, fetch)
	case store.Relational:
		return p.fetchRdbProps(ids, unit, fetch)
	default:
		return nil
	}
}
