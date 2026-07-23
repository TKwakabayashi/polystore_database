package bulk

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (qp *Processor) newReadSession() neo4j.SessionWithContext {
	return qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, FetchSize: neo4j.FetchAll})
}

func (qp *Processor) closeSession(s neo4j.SessionWithContext) { _ = s.Close(qp.ctx) }

func ScanGraphBulk(qp *Processor, o *plan.EntityScan) ([]Record, error) {
	var whereSections []string
	params := make(map[string]interface{})

	// 1. ラベル条件の構築 (n:L1 OR n:L2 ...)
	if len(o.Labels) > 0 {
		var labelConditions []string
		for _, l := range o.Labels {
			labelConditions = append(labelConditions, fmt.Sprintf("n:%s", l))
		}
		whereSections = append(whereSections, "("+strings.Join(labelConditions, " OR ")+")")
	}

	// 2. プロパティフィルタ条件の構築 (n.prop = $val ...)
	for i, cond := range o.Filter {
		var operator string
		switch cond.Type {
		case plan.CondEq:
			operator = "="
		case plan.CondNeq:
			operator = "<>"
		case plan.CondGreater:
			operator = ">"
		case plan.CondLess:
			operator = "<"
		default:
			return nil, fmt.Errorf("unknown operator")
		}
		paramName := fmt.Sprintf("val%d", i)
		whereSections = append(whereSections, fmt.Sprintf("n.%s %s $%s", cond.Property, operator, paramName))
		params[paramName], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	// 3. クエリの組み立て
	query := "MATCH (n)"
	if len(whereSections) > 0 {
		query += "\nWHERE " + strings.Join(whereSections, " AND ")
	}
	query += "\nRETURN n.uuid AS id"

	sess := qp.newReadSession()
	defer qp.closeSession(sess)

	res, err := sess.Run(qp.ctx, query, params)
	if err != nil {
		return nil, err
	}

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	out := make([]Record, 0)
	for res.Next(qp.ctx) {
		if idVal, ok := res.Record().Get("id"); ok && idVal != nil {
			newSlots := make([]uid.UUID, newSlotCount)
			newSlots[aliasIdx] = uid.FromAny(idVal)
			out = append(out, Record{Slots: newSlots})
		}
	}
	return out, res.Err()
}

func bulkFilterGraph(qp *Processor, o *plan.Filter, in []Record) ([]Record, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	var targetVar, matchPattern string
	if o.ObjType == plan.Relationship {
		targetVar = "r"
		matchPattern = "()-[r]->()"
	} else {
		targetVar = "n"
		matchPattern = "(n:Entity)"
	}

	var labelConditions []string
	for _, l := range o.Labels {
		labelConditions = append(labelConditions, fmt.Sprintf("%s:%s", targetVar, l))
	}
	labelFilter := strings.Join(labelConditions, " OR ")

	var whereClauses []string
	params := make(map[string]interface{})
	for i, cond := range o.Filter {
		operator := "="
		switch cond.Type {
		case plan.CondEq:
			operator = "="
		case plan.CondNeq:
			operator = "<>"
		case plan.CondGreater:
			operator = ">"
		case plan.CondLess:
			operator = "<"
		}
		paramName := fmt.Sprintf("val%d", i)
		whereClauses = append(whereClauses, fmt.Sprintf("%s.%s %s $%s", targetVar, cond.Property, operator, paramName))
		params[paramName], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	finalQuery := fmt.Sprintf(`
        MATCH %s
        WHERE (%s)
          AND %s.uuid IN $ids
          AND %s
        RETURN %s.uuid AS id`,
		matchPattern, labelFilter, targetVar, strings.Join(whereClauses, " AND "), targetVar,
	)

	// 全入力から uuid をユニーク化して 1 回で問い合わせる
	idMap := make(map[uid.UUID]struct{})
	for _, r := range in {
		idMap[r.Slots[filterIdxIn]] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idMap))
	for id := range idMap {
		uniqueIDs = append(uniqueIDs, id.String())
	}
	params["ids"] = uniqueIDs

	sess := qp.newReadSession()
	defer qp.closeSession(sess)

	res, err := sess.Run(qp.ctx, finalQuery, params)
	if err != nil {
		return nil, err
	}
	validMap := make(map[uid.UUID]struct{})
	for res.Next(qp.ctx) {
		if id, ok := res.Record().Get("id"); ok && id != nil {
			validMap[uid.FromAny(id)] = struct{}{}
		}
	}
	if err := res.Err(); err != nil {
		return nil, err
	}

	// 列引き継ぎ（InputSlot → OutputSlot）
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

func fetchGraphPropsBulk(qp *Processor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if len(ids) == 0 || len(fetch.Props) == 0 {
		return result
	}
	var targetVar, matchPattern string
	if unit.ObjType == plan.Relationship {
		targetVar = "r"
		matchPattern = fmt.Sprintf("()-[%s]->()", targetVar)
	} else if unit.ObjType == plan.Entity {
		targetVar = "n"
		matchPattern = fmt.Sprintf("(%s:Entity)", targetVar)
	} else {
		panic("unknown ObjectType passed")
	}

	var labelFilter string
	if len(unit.Labels) > 0 {
		var labelConditions []string
		for _, l := range unit.Labels {
			labelConditions = append(labelConditions, fmt.Sprintf("%s:%s", targetVar, l))
		}
		labelFilter = fmt.Sprintf("WHERE (%s) AND ", strings.Join(labelConditions, " OR "))
	} else {
		labelFilter = "WHERE "
	}

	var propReturns []string
	for _, p := range fetch.Props {
		propReturns = append(propReturns, fmt.Sprintf("%s.%s AS %s", targetVar, p, p))
	}

	query := fmt.Sprintf(`
        MATCH %s
        %s %s.uuid IN $ids
        RETURN %s.uuid AS uuid, %s`,
		matchPattern, labelFilter, targetVar, targetVar, strings.Join(propReturns, ", "))

	sess := qp.newReadSession()
	defer qp.closeSession(sess)

	res, err := sess.Run(qp.ctx, query, map[string]interface{}{"ids": ids})
	if err != nil {
		return result
	}
	for res.Next(qp.ctx) {
		rec := res.Record()
		idVal, _ := rec.Get("uuid")
		id, ok := idVal.(string)
		if !ok {
			continue
		}
		propsMap := make(map[string]interface{})
		for _, p := range fetch.Props {
			if val, ok := rec.Get(p); ok && val != nil {
				propsMap[p], _ = codec.ConvertToNativeType(val, fetch.TypeMap[p])
			}
		}
		result[id] = propsMap
	}
	return result
}
