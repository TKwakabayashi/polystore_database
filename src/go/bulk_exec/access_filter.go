package bulk_executor

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// filterValid は対象ストアで uniqueIDs のうち条件を満たす uuid 集合を返す（volcano_exec と同一）。
func (p *Processor) filterValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	switch o.DataStore {
	case "graph", "", "unknown":
		return p.filterGraphValid(o, ids)
	case "document":
		return p.filterDocValid(o, ids)
	case "kvs":
		return p.filterKvsValid(o, ids)
	case "relational":
		return p.filterRdbValid(o, ids)
	case "columnar":
		return p.filterColValid(o, ids)
	default:
		return nil, fmt.Errorf("未知のフィルタ対象ストア: %s", o.DataStore)
	}
}

// ---------- graph (Neo4j) ----------

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

// ---------- document (MongoDB) ----------

func (p *Processor) filterDocValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	valid := make(map[string]struct{})
	if p.mDb == nil {
		return valid, nil
	}
	var commonConditions []bson.E
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonConditions = append(commonConditions, bson.E{Key: cond.Property, Value: bson.M{mongoOp(cond.Type): val}})
	}
	query := append(bson.D{{Key: "uuid", Value: bson.M{"$in": ids}}}, commonConditions...)

	for _, label := range o.Labels {
		p.countRoundTrip()
		cur, err := p.mDb.Collection(label).Find(p.ctx, query, options.Find().SetProjection(bson.M{"uuid": 1}))
		if err != nil {
			return nil, err
		}
		for cur.Next(p.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err == nil {
				valid[rec.UUID] = struct{}{}
			}
		}
		cur.Close(p.ctx)
	}
	return valid, nil
}

// ---------- kvs (LevelDB) ----------

func (p *Processor) filterKvsValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	valid := make(map[string]struct{})
	for _, id := range ids {
		for _, label := range o.Labels {
			if p.matchConditionsKVS(label, id, o.Filter) {
				valid[id] = struct{}{}
				break
			}
		}
	}
	return valid, nil
}

// ---------- relational (MySQL) ----------

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

// ---------- columnar (Cassandra) ----------

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
		commonClauses = append(commonClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, cqlOp(cond.Type)))
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

// remap は InputSlot 基準の行を OutputSlot 基準へ引き継ぐ。
func remap(in []string, inSlot, outSlot plan.SlotTable) []string {
	out := make([]string, len(outSlot.VarToSlot))
	for alias, outIdx := range outSlot.VarToSlot {
		if inIdx, ok := inSlot.VarToSlot[alias]; ok && inIdx < len(in) {
			out[outIdx] = in[inIdx]
		}
	}
	return out
}

// sqlToCypherOp は graph filter 用の演算子表記（<> 等）。
func sqlToCypherOp(t plan.ConditionType) string { return sqlOp(t) }
