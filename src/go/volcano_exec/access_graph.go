package volcano_executor

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// newReadSession は Neo4j の読み取りセッションを生成する。
func (p *Processor) newReadSession() neo4j.SessionWithContext {
	return p.neoDriver.NewSession(p.ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
		FetchSize:  neo4j.FetchAll,
	})
}

// cypherOp は比較演算子を Cypher 表記へ。
func cypherOp(t plan.ConditionType) (string, error) {
	switch t {
	case plan.CondEq:
		return "=", nil
	case plan.CondNeq:
		return "<>", nil
	case plan.CondGreater:
		return ">", nil
	case plan.CondLess:
		return "<", nil
	default:
		return "", fmt.Errorf("unknown operator")
	}
}

// sqlToCypherOp は graph filter 用の演算子表記（<> 等）。
func sqlToCypherOp(t plan.ConditionType) string { return sqlOp(t) }

// ---------- EntityScan ----------

func (p *Processor) scanGraphIDs(o *plan.EntityScan) ([]string, error) {
	var whereSections []string
	params := make(map[string]interface{})

	if len(o.Labels) > 0 {
		var lc []string
		for _, l := range o.Labels {
			lc = append(lc, fmt.Sprintf("n:%s", l))
		}
		whereSections = append(whereSections, "("+strings.Join(lc, " OR ")+")")
	}
	for i, cond := range o.Filter {
		op, err := cypherOp(cond.Type)
		if err != nil {
			return nil, err
		}
		pn := fmt.Sprintf("val%d", i)
		whereSections = append(whereSections, fmt.Sprintf("n.%s %s $%s", cond.Property, op, pn))
		params[pn], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	query := "MATCH (n)"
	if len(whereSections) > 0 {
		query += "\nWHERE " + strings.Join(whereSections, " AND ")
	}
	query += "\nRETURN n.uuid AS id"

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, query, params)
	if err != nil {
		return nil, err
	}
	var ids []string
	for res.Next(p.ctx) {
		if v, ok := res.Record().Get("id"); ok && v != nil {
			if s, ok := v.(string); ok {
				ids = append(ids, s)
			}
		}
	}
	return ids, res.Err()
}

// ---------- Filter ----------

func (p *Processor) filterGraphValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
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
		paramName := fmt.Sprintf("val%d", i)
		whereClauses = append(whereClauses, fmt.Sprintf("%s.%s %s $%s", targetVar, cond.Property, sqlToCypherOp(cond.Type), paramName))
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
	params["ids"] = ids

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, finalQuery, params)
	if err != nil {
		return nil, err
	}
	valid := make(map[string]struct{})
	for res.Next(p.ctx) {
		if v, ok := res.Record().Get("id"); ok && v != nil {
			valid[v.(string)] = struct{}{}
		}
	}
	return valid, res.Err()
}

// ---------- Projection fetch ----------

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
