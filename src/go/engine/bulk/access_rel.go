package bulk

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

func ScanRdbBulk(qp *Processor, o *plan.EntityScan) ([]Record, error) {
	if qp.sqlDb == nil {
		return nil, nil
	}

	// --- DB特有: WHERE 構築（全条件 AND） ---
	var clauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("%s %s ?", cond.Property, core.SQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(clauses) > 0 {
		whereStr = "WHERE " + strings.Join(clauses, " AND ")
	}

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	out := make([]Record, 0)
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM %s %s", label, whereStr)
		rows, err := qp.sqlDb.QueryContext(qp.ctx, query, args...)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil || id == "" {
				continue
			}
			newSlots := make([]uid.UUID, newSlotCount)
			newSlots[aliasIdx] = uid.UUID(id)
			out = append(out, Record{Slots: newSlots})
		}
		rows.Close()
	}
	return out, nil
}

func FilterRdbBulk(qp *Processor, o *plan.Filter, in []Record) ([]Record, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// --- DB特有: 共通条件 ---
	var filterClauses []string
	var commonArgs []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		filterClauses = append(filterClauses, fmt.Sprintf("%s %s ?", cond.Property, core.SQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonArgs = append(commonArgs, val)
	}
	whereBase := "1=1"
	if len(filterClauses) > 0 {
		whereBase = strings.Join(filterClauses, " AND ")
	}

	idMap := make(map[uid.UUID]struct{})
	for _, r := range in {
		idMap[r.Slots[filterIdxIn]] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idMap))
	for id := range idMap {
		uniqueIDs = append(uniqueIDs, id.String())
	}
	if len(uniqueIDs) == 0 {
		return nil, nil
	}

	// --- DB特有: valid 抽出 ---
	// ID はストアの実効チャンクサイズで分割する。全件を 1 クエリに載せると MySQL の
	// プレースホルダ上限（65535）を超えてエラーになるため必須。
	validMap := make(map[uid.UUID]struct{})
	for _, label := range o.Labels {
		if err := core.ForEachIDChunk(store.Relational, uniqueIDs, func(chunk []string) error {
			placeholders := strings.Repeat("?,", len(chunk)-1) + "?"
			query := fmt.Sprintf("SELECT uuid FROM %s WHERE %s AND uuid IN (%s)", label, whereBase, placeholders)
			args := make([]interface{}, 0, len(commonArgs)+len(chunk))
			args = append(args, commonArgs...)
			for _, id := range chunk {
				args = append(args, id)
			}
			return scanValidRdbIDs(qp, query, args, validMap)
		}); err != nil {
			return nil, err
		}
	}
	out := make([]Record, 0, len(in))
	for _, r := range in {
		if _, ok := validMap[r.Slots[filterIdxIn]]; ok {
			newRec := Record{Slots: make([]uid.UUID, newSlotCount)}
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

func fetchRdbPropsBulk(qp *Processor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.sqlDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}

	batchSize := core.ChunkSize(store.Relational)
	propList := strings.Join(fetch.Props, ", ")

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]
		placeholders := strings.Repeat("?,", len(batch)-1) + "?"

		// --- DB特有: 複数ラベルを UNION ALL ---
		var selects []string
		var args []interface{}
		for _, table := range unit.Labels {
			selects = append(selects, fmt.Sprintf("SELECT uuid, %s FROM %s WHERE uuid IN (%s)", propList, table, placeholders))
			for _, id := range batch {
				args = append(args, id)
			}
		}
		finalQuery := strings.Join(selects, " UNION ALL ")

		rows, err := qp.sqlDb.QueryContext(qp.ctx, finalQuery, args...)
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

// scanValidRdbIDs は 1 チャンク分のクエリを実行し、有効な uuid を validMap へ積む。
func scanValidRdbIDs(qp *Processor, query string, args []interface{}, validMap map[uid.UUID]struct{}) error {
	rows, err := qp.sqlDb.QueryContext(qp.ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			validMap[uid.UUID(id)] = struct{}{}
		}
	}
	return rows.Err()
}
