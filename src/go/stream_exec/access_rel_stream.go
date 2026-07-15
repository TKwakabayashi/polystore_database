package stream_executor

import (
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
	"strings"
)

func sqlOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "<>"
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	default:
		return "="
	}
}

func ScanRdbStream(qp *QueryProcessor,
	o *plan.EntityScan, output chan<- []Record) (int, error) {
	if qp.sqlDb == nil {
		return 0, nil
	}

	// --- DB特有: WHERE 構築（全条件 AND） ---
	var clauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("%s %s ?", cond.Property, sqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(clauses) > 0 {
		whereStr = "WHERE " + strings.Join(clauses, " AND ")
	}

	// --- graph と同じストリーミング骨格 ---
	const outputBatchSize = 500
	rowCount := 0
	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]
	currentBatch := make([]Record, 0, outputBatchSize)

	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM %s %s", label, whereStr)
		rows, err := qp.sqlDb.QueryContext(qp.ctx, query, args...)
		if err != nil {
			return rowCount, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil || id == "" {
				continue
			}
			newSlots := make([]string, newSlotCount)
			newSlots[aliasIdx] = id
			currentBatch = append(currentBatch, Record{Slots: newSlots})
			rowCount++
			if len(currentBatch) >= outputBatchSize {
				output <- currentBatch
				currentBatch = make([]Record, 0, outputBatchSize)
			}
		}
		rows.Close()
	}
	if len(currentBatch) > 0 {
		output <- currentBatch
	}
	return rowCount, nil
}

func FilterRdbStream(qp *QueryProcessor,
	o *plan.Filter, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// --- DB特有: 共通条件 ---
	var filterClauses []string
	var commonArgs []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		filterClauses = append(filterClauses, fmt.Sprintf("%s %s ?", cond.Property, sqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonArgs = append(commonArgs, val)
	}
	whereBase := "1=1"
	if len(filterClauses) > 0 {
		whereBase = strings.Join(filterClauses, " AND ")
	}

	return runBatches(
		qp.ctx, qp.exec, qp.sem, OpFilter, inputStream, outputStream,
		noResource, closeNoResource,
		func(_ struct{}, batch []Record) ([]Record, error) {
			idMap := make(map[string]struct{})
			for _, r := range batch {
				idMap[r.Slots[filterIdxIn]] = struct{}{}
			}
			uniqueIDs := make([]string, 0, len(idMap))
			for id := range idMap {
				uniqueIDs = append(uniqueIDs, id)
			}
			if len(uniqueIDs) == 0 {
				return nil, nil
			}

			// --- DB特有: valid 抽出 ---
			placeholders := strings.Repeat("?,", len(uniqueIDs)-1) + "?"
			validMap := make(map[string]struct{})
			for _, label := range o.Labels {
				query := fmt.Sprintf("SELECT uuid FROM %s WHERE %s AND uuid IN (%s)", label, whereBase, placeholders)
				args := make([]interface{}, 0, len(commonArgs)+len(uniqueIDs))
				args = append(args, commonArgs...)
				for _, id := range uniqueIDs {
					args = append(args, id)
				}
				rows, err := qp.sqlDb.QueryContext(qp.ctx, query, args...)
				if err != nil {
					return nil, err
				}
				for rows.Next() {
					var id string
					if err := rows.Scan(&id); err == nil {
						validMap[id] = struct{}{}
					}
				}
				rows.Close()
			}

			// --- graph と同じ列引き継ぎ ---
			out := make([]Record, 0, len(batch))
			for _, r := range batch {
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
		},
	)
}

func fetchRdbPropsStream(qp *QueryProcessor,
	ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.sqlDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
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
