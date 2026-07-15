package volcano_executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// filterIterator は Filter の pull 実装。子から 1 バッチ pull し、対象ストアで
// 妥当な uuid を判定して残った行のみを出力する。空になったら次のバッチを引き続き pull。
type filterIterator struct {
	p     *Processor
	o     *plan.Filter
	child Iterator
	step  int
}

func (f *filterIterator) Open(ctx context.Context) error  { return f.child.Open(ctx) }
func (f *filterIterator) Close(ctx context.Context) error { return f.child.Close(ctx) }

func (f *filterIterator) Next(ctx context.Context) (*Batch, error) {
	for {
		in, err := f.child.Next(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, nil
		}
		start := time.Now()
		out, err := f.process(in)
		if err != nil {
			return nil, err
		}
		f.p.recordOp(f.step, "Filter", time.Since(start), out.n)
		if out.n > 0 {
			return out, nil
		}
		// 全滅したバッチは EOF と紛れないよう、次を pull する。
	}
}

func (f *filterIterator) process(in *Batch) (*Batch, error) {
	filterIdxIn := f.o.InputSlot.VarToSlot[f.o.Alias]

	// バッチ内 uuid のユニーク化
	idSet := make(map[string]struct{}, in.n)
	for i := 0; i < in.n; i++ {
		idSet[in.get(i, filterIdxIn)] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idSet))
	for id := range idSet {
		uniqueIDs = append(uniqueIDs, id)
	}

	valid, err := f.p.filterValid(f.o, uniqueIDs)
	if err != nil {
		return nil, err
	}

	outSlots := len(f.o.OutputSlot.VarToSlot)
	out := newBatch(outSlots, in.n)
	for i := 0; i < in.n; i++ {
		if _, ok := valid[in.get(i, filterIdxIn)]; ok {
			out.appendRow(remap(in.row(i), f.o.InputSlot, f.o.OutputSlot))
		}
	}
	return out, nil
}

// filterValid は対象ストアで uniqueIDs のうち条件を満たす uuid 集合を返す。
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
