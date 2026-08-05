package vecstream

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
)

// ---------- EntityScan ----------

func (p *Processor) scanColIDs(o *plan.EntityScan) ([]string, error) {
	if p.cqlSes == nil {
		return nil, nil
	}
	var whereClauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, core.CQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(whereClauses) > 0 {
		whereStr = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	var ids []string
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" %s ALLOW FILTERING", label, whereStr)
		p.countRoundTrip()
		iter := p.cqlSes.Query(query, args...).Iter()
		var id string
		for iter.Scan(&id) {
			if id != "" {
				ids = append(ids, id)
			}
		}
		if err := iter.Close(); err != nil {
			return ids, err
		}
	}
	return ids, nil
}

// ---------- Filter ----------

func (p *Processor) filterColValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	valid := make(map[string]struct{})
	if p.cqlSes == nil || len(ids) == 0 {
		return valid, nil
	}
	var commonClauses []string
	var commonArgs []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		commonClauses = append(commonClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, core.CQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonArgs = append(commonArgs, val)
	}
	whereClause := strings.Join(append([]string{"uuid IN ?"}, commonClauses...), " AND ")
	queryArgs := append([]interface{}{ids}, commonArgs...)

	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" WHERE %s ALLOW FILTERING", label, whereClause)
		p.countRoundTrip()
		iter := p.cqlSes.Query(query, queryArgs...).Iter()
		var id string
		for iter.Scan(&id) {
			valid[id] = struct{}{}
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}
	return valid, nil
}

// ---------- Projection fetch ----------

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
