package volcano_executor

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
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
	case plan.CondGreaterEq:
		return ">="
	case plan.CondLessEq:
		return "<="
	default:
		return "="
	}
}

// ---------- EntityScan ----------

func (p *Processor) scanRdbIDs(o *plan.EntityScan) ([]string, error) {
	if p.sqlDb == nil {
		return nil, nil
	}
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

	var ids []string
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM %s %s", label, whereStr)
		p.countRoundTrip()
		rows, err := p.sqlDb.QueryContext(p.ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil || id == "" {
				continue
			}
			ids = append(ids, id)
		}
		rows.Close()
	}
	return ids, nil
}

// ---------- Filter ----------

func (p *Processor) filterRdbValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	valid := make(map[string]struct{})
	if p.sqlDb == nil || len(ids) == 0 {
		return valid, nil
	}
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
	placeholders := strings.Repeat("?,", len(ids)-1) + "?"

	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM %s WHERE %s AND uuid IN (%s)", label, whereBase, placeholders)
		args := make([]interface{}, 0, len(commonArgs)+len(ids))
		args = append(args, commonArgs...)
		for _, id := range ids {
			args = append(args, id)
		}
		p.countRoundTrip()
		rows, err := p.sqlDb.QueryContext(p.ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				valid[id] = struct{}{}
			}
		}
		rows.Close()
	}
	return valid, nil
}

// ---------- Projection fetch ----------

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
