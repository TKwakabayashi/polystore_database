package volcano_executor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
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

// fetchPropertiesBulk は各種ストアからプロパティを一括取得する共通の入り口。
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

// ---------- graph (Neo4j) ----------

func (p *Processor) fetchGraphProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.neoDriver == nil || len(ids) == 0 || len(fetch.Props) == 0 {
		return result
	}
	var targetVar, matchPattern string
	if unit.ObjType == plan.Relationship {
		targetVar = "r"
		matchPattern = "()-[r]->()"
	} else {
		targetVar = "n"
		matchPattern = "(n:Entity)"
	}
	labelFilter := "WHERE "
	if len(unit.Labels) > 0 {
		var lc []string
		for _, l := range unit.Labels {
			lc = append(lc, fmt.Sprintf("%s:%s", targetVar, l))
		}
		labelFilter = fmt.Sprintf("WHERE (%s) AND ", strings.Join(lc, " OR "))
	}
	var propReturns []string
	for _, prop := range fetch.Props {
		propReturns = append(propReturns, fmt.Sprintf("%s.%s AS %s", targetVar, prop, prop))
	}
	query := fmt.Sprintf(`
        MATCH %s
        %s %s.uuid IN $ids
        RETURN %s.uuid AS uuid, %s`,
		matchPattern, labelFilter, targetVar, targetVar, strings.Join(propReturns, ", "))

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, query, map[string]interface{}{"ids": ids})
	if err != nil {
		return result
	}
	for res.Next(p.ctx) {
		rec := res.Record()
		idVal, _ := rec.Get("uuid")
		id, ok := idVal.(string)
		if !ok {
			continue
		}
		props := make(map[string]interface{})
		for _, prop := range fetch.Props {
			if val, ok := rec.Get(prop); ok && val != nil {
				props[prop], _ = codec.ConvertToNativeType(val, fetch.TypeMap[prop])
			}
		}
		result[id] = props
	}
	return result
}

// ---------- document (MongoDB) ----------

func (p *Processor) fetchDocProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.mDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	for _, label := range unit.Labels {
		p.countRoundTrip()
		cur, err := p.mDb.Collection(label).Find(p.ctx, bson.M{"uuid": bson.M{"$in": ids}})
		if err != nil {
			continue
		}
		for cur.Next(p.ctx) {
			var raw bson.M
			if err := cur.Decode(&raw); err != nil {
				continue
			}
			id, ok := raw["uuid"].(string)
			if !ok {
				continue
			}
			if _, exists := result[id]; !exists {
				result[id] = make(map[string]interface{})
			}
			for _, prop := range fetch.Props {
				if val, ok := raw[prop]; ok && val != nil {
					result[id][prop], _ = codec.ConvertToNativeType(val, fetch.TypeMap[prop])
				}
			}
		}
		cur.Close(p.ctx)
	}
	return result
}

// ---------- kvs (LevelDB) ----------

func (p *Processor) fetchKvsProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.ldb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	for _, uuid := range ids {
		if uuid == "" {
			continue
		}
		if _, exists := result[uuid]; !exists {
			result[uuid] = make(map[string]interface{})
		}
		for _, label := range unit.Labels {
			for _, propName := range fetch.Props {
				p.countRoundTrip()
				valByte, err := p.ldb.Get(codec.BuildEntityKey(label, uuid, propName), nil)
				if err != nil {
					if _, ok := result[uuid][propName]; !ok {
						result[uuid][propName] = nil
					}
					continue
				}
				finalVal, _ := codec.ConvertToNativeType(codec.DecodeValue(valByte, fetch.TypeMap[propName]), fetch.TypeMap[propName])
				if finalVal != nil {
					result[uuid][propName] = finalVal
				} else if _, ok := result[uuid][propName]; !ok {
					result[uuid][propName] = nil
				}
			}
		}
	}
	return result
}

// ---------- relational (MySQL) ----------

func (p *Processor) fetchRdbProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.sqlDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	const batchSize = 1000
	propList := strings.Join(fetch.Props, ", ")

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]
		placeholders := strings.Repeat("?,", len(batch)-1) + "?"

		var selects []string
		var args []interface{}
		for _, table := range unit.Labels {
			selects = append(selects, fmt.Sprintf("SELECT uuid, %s FROM %s WHERE uuid IN (%s)", propList, table, placeholders))
			for _, id := range batch {
				args = append(args, id)
			}
		}
		finalQuery := strings.Join(selects, " UNION ALL ")

		p.countRoundTrip()
		rows, err := p.sqlDb.QueryContext(p.ctx, finalQuery, args...)
		if err != nil {
			continue
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			columns := make([]interface{}, len(cols))
			pointers := make([]interface{}, len(cols))
			for j := range columns {
				pointers[j] = &columns[j]
			}
			if err := rows.Scan(pointers...); err != nil {
				continue
			}
			var id string
			if b, ok := columns[0].([]byte); ok {
				id = string(b)
			} else {
				id = fmt.Sprintf("%v", columns[0])
			}
			if _, exists := result[id]; !exists {
				result[id] = make(map[string]interface{})
			}
			for j, colName := range cols {
				if j == 0 {
					continue
				}
				val := columns[j]
				if b, ok := val.([]byte); ok {
					val = string(b)
				}
				result[id][colName], _ = codec.ConvertToNativeType(val, fetch.TypeMap[colName])
			}
		}
		rows.Close()
	}
	return result
}

// ---------- columnar (Cassandra) ----------

func (p *Processor) fetchColProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.cqlSes == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	const batchSize = 500
	quoted := make([]string, len(fetch.Props))
	for i, prop := range fetch.Props {
		quoted[i] = fmt.Sprintf("\"%s\"", prop)
	}
	propList := strings.Join(quoted, ", ")

	for _, table := range unit.Labels {
		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			batch := ids[i:end]
			query := fmt.Sprintf("SELECT uuid, %s FROM \"%s\" WHERE uuid IN ?", propList, table)
			p.countRoundTrip()
			iter := p.cqlSes.Query(query, batch).Iter()
			for {
				row := make(map[string]interface{})
				if !iter.MapScan(row) {
					break
				}
				idRaw, ok := row["uuid"]
				if !ok || idRaw == nil {
					continue
				}
				id := fmt.Sprintf("%v", idRaw)
				if _, exists := result[id]; !exists {
					result[id] = make(map[string]interface{})
				}
				for _, prop := range fetch.Props {
					if val, ok := row[prop]; ok && val != nil {
						result[id][prop], _ = codec.ConvertToNativeType(val, fetch.TypeMap[prop])
					}
				}
			}
			if err := iter.Close(); err != nil {
				continue
			}
		}
	}
	return result
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
