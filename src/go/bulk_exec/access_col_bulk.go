package bulk_executor

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
)

func cqlOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "!="
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	default:
		return "="
	}
}

func ScanColBulk(qp *QueryProcessor, o *plan.EntityScan) ([]Record, error) {
	if qp.cqlSes == nil {
		return nil, nil
	}

	// --- DB特有: WHERE 構築 ---
	var whereClauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, cqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(whereClauses) > 0 {
		whereStr = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	out := make([]Record, 0)
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" %s ALLOW FILTERING", label, whereStr)
		iter := qp.cqlSes.Query(query, args...).Iter()
		var id string
		for iter.Scan(&id) {
			if id == "" {
				continue
			}
			newSlots := make([]string, newSlotCount)
			newSlots[aliasIdx] = id
			out = append(out, Record{Slots: newSlots})
		}
		if err := iter.Close(); err != nil {
			return out, err
		}
	}
	return out, nil
}

func FilterColBulk(qp *QueryProcessor, o *plan.Filter, in []Record) ([]Record, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// --- DB特有: 共通条件 ---
	var commonClauses []string
	var commonArgs []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		commonClauses = append(commonClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, cqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonArgs = append(commonArgs, val)
	}
	whereClause := strings.Join(append([]string{"uuid IN ?"}, commonClauses...), " AND ")

	idMap := make(map[string]struct{})
	for _, r := range in {
		idMap[r.Slots[filterIdxIn]] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idMap))
	for id := range idMap {
		uniqueIDs = append(uniqueIDs, id)
	}

	// --- DB特有: valid 抽出 ---
	queryArgs := append([]interface{}{uniqueIDs}, commonArgs...)
	validMap := make(map[string]struct{})
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" WHERE %s ALLOW FILTERING", label, whereClause)
		iter := qp.cqlSes.Query(query, queryArgs...).Iter()
		var id string
		for iter.Scan(&id) {
			validMap[id] = struct{}{}
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	out := make([]Record, 0, len(in))
	for _, r := range in {
		if _, ok := validMap[r.Slots[filterIdxIn]]; ok {
			newRec := Record{Slots: make([]string, newSlotCount)}
			for alias, outIdx := range o.OutputSlot.VarToSlot {
				if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
					newRec.Slots[outIdx] = r.Slots[inIdx]
				}
			}
			out = append(out, newRec)
		}
	}
	return out, nil
}

func fetchColPropsBulk(qp *QueryProcessor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.cqlSes == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}

	const batchSize = 500
	quoted := make([]string, len(fetch.Props))
	for i, p := range fetch.Props {
		quoted[i] = fmt.Sprintf("\"%s\"", p)
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
			iter := qp.cqlSes.Query(query, batch).Iter()
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
				for _, p := range fetch.Props {
					if val, ok := row[p]; ok && val != nil {
						result[id][p], _ = codec.ConvertToNativeType(val, fetch.TypeMap[p])
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
