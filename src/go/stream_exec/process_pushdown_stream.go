package stream_executor

import (
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
	"strconv"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// streamStorePushdown は集約を単一ストアへ委譲して最終行を得る。
// 出力行のキーは RETURN 別名（＝ ReturnItem.Name）に一致させ、コーディネータ経路と揃える。
func streamStorePushdown(qp *QueryProcessor, o *plan.StorePushdown, out chan<- []Row) int {
	switch o.Store {
	case "graph":
		return runGraphPushdown(qp, o.Query, o.Params, out)
	case "relational":
		return runRelationalPushdown(qp, o, out)
	case "document":
		return runDocumentPushdown(qp, o, out)
	case "columnar":
		return runColumnarPushdown(qp, o, out)
	default:
		fmt.Printf("StorePushdown: unsupported store %q\n", o.Store)
		return 0
	}
}

// ===== graph: パラメータ埋め込み済み Cypher をそのまま実行 =====

// baseline(RunNeo4j) と同じ session.Run(原クエリ, params) に揃えて公平に比較する。
func runGraphPushdown(qp *QueryProcessor, query string, params map[string]string, out chan<- []Row) int {
	if qp.neoDriver == nil {
		fmt.Println("StorePushdown[graph]: neo4j driver is nil")
		return 0
	}
	session := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(qp.ctx)

	res, err := session.Run(qp.ctx, query, typeParams(params))
	if err != nil {
		fmt.Printf("StorePushdown[graph] run error: %v\n", err)
		return 0
	}
	var rows []Row
	for res.Next(qp.ctx) {
		rec := res.Record()
		row := make(Row, len(rec.Keys))
		for i, k := range rec.Keys {
			row[k] = rec.Values[i]
		}
		rows = append(rows, row)
	}
	if err := res.Err(); err != nil {
		fmt.Printf("StorePushdown[graph] iterate error: %v\n", err)
	}
	if len(rows) > 0 {
		out <- rows
	}
	return len(rows)
}

// ===== relational (MySQL): SELECT ... aggexprs ... GROUP BY ... ORDER BY ... LIMIT =====

func runRelationalPushdown(qp *QueryProcessor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.sqlDb == nil {
		return 0
	}
	sql, args := buildRelationalSQL(o)
	rows, err := qp.sqlDb.QueryContext(qp.ctx, sql, args...)
	if err != nil {
		fmt.Printf("StorePushdown[relational] error: %v\n  SQL: %s\n", err, sql)
		return 0
	}
	defer rows.Close()
	cols, _ := rows.Columns()

	var result []Row
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(Row, len(cols))
		for i, c := range cols {
			row[c] = coerceScalar(vals[i])
		}
		result = append(result, row)
	}
	if len(result) > 0 {
		out <- result
	}
	return len(result)
}

func buildRelationalSQL(o *plan.StorePushdown) (string, []interface{}) {
	var selects, groupCols []string
	for _, it := range o.Items {
		alias := sqlIdent(it.Name)
		if it.IsAggregate {
			selects = append(selects, sqlAggExpr(*it.Agg)+" AS "+alias)
		} else {
			col := itemProp(it)
			selects = append(selects, sqlIdent(col)+" AS "+alias)
			groupCols = append(groupCols, sqlIdent(col))
		}
	}

	var where []string
	var args []interface{}
	for _, c := range o.Filters {
		if c == nil {
			continue
		}
		where = append(where, sqlIdent(c.Property)+" "+sqlOp(c.Type)+" ?")
		v, _ := codec.ConvertToNativeType(c.Value, c.DataType)
		args = append(args, v)
	}

	q := "SELECT " + strings.Join(selects, ", ") + " FROM " + sqlIdent(o.Table)
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	if len(groupCols) > 0 {
		q += " GROUP BY " + strings.Join(groupCols, ", ")
	}
	if len(o.OrderItems) > 0 {
		var ords []string
		for _, oi := range o.OrderItems {
			dir := "ASC"
			if oi.Direction == plan.OrderDesc {
				dir = "DESC"
			}
			ords = append(ords, sqlIdent(oi.Key)+" "+dir)
		}
		q += " ORDER BY " + strings.Join(ords, ", ")
	}
	if o.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", o.Limit)
	}
	return q, args
}

func sqlAggExpr(a plan.AggregateItem) string {
	fn := strings.ToUpper(a.Func.String())
	if a.Func == plan.AggCount && a.Prop == "" {
		return "COUNT(*)"
	}
	arg := sqlIdent(a.Prop)
	if a.Distinct {
		return fn + "(DISTINCT " + arg + ")"
	}
	return fn + "(" + arg + ")"
}

func sqlIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

// ===== document (Mongo): $match → $group → ($addFields for DISTINCT) → $sort → $limit =====

func runDocumentPushdown(qp *QueryProcessor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.mDb == nil {
		return 0
	}
	pipeline := buildMongoPipeline(o)
	cur, err := qp.mDb.Collection(o.Table).Aggregate(qp.ctx, pipeline)
	if err != nil {
		fmt.Printf("StorePushdown[document] error: %v\n", err)
		return 0
	}
	defer cur.Close(qp.ctx)

	var result []Row
	for cur.Next(qp.ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			continue
		}
		var idDoc bson.M
		if raw, ok := doc["_id"].(bson.M); ok {
			idDoc = raw
		}
		row := make(Row, len(o.Items))
		gi, ai := 0, 0
		for _, it := range o.Items {
			if it.IsAggregate {
				row[it.Name] = doc[fmt.Sprintf("a%d", ai)]
				ai++
			} else {
				if idDoc != nil {
					row[it.Name] = idDoc[fmt.Sprintf("g%d", gi)]
				}
				gi++
			}
		}
		result = append(result, row)
	}
	if len(result) > 0 {
		out <- result
	}
	return len(result)
}

func buildMongoPipeline(o *plan.StorePushdown) mongo.Pipeline {
	var pipeline mongo.Pipeline

	if len(o.Filters) > 0 {
		match := bson.D{}
		for _, c := range o.Filters {
			if c == nil {
				continue
			}
			v, _ := codec.ConvertToNativeType(c.Value, c.DataType)
			match = append(match, bson.E{Key: c.Property, Value: bson.M{mongoOp(c.Type): v}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	var idVal interface{}
	if len(o.GroupKeys) > 0 {
		idDoc := bson.D{}
		for i, gk := range o.GroupKeys {
			idDoc = append(idDoc, bson.E{Key: fmt.Sprintf("g%d", i), Value: "$" + gk.Prop})
		}
		idVal = idDoc
	}
	group := bson.D{{Key: "_id", Value: idVal}}
	var distinctIdx []int
	for i, a := range o.Aggs {
		group = append(group, bson.E{Key: fmt.Sprintf("a%d", i), Value: mongoAggExpr(a)})
		if a.Distinct {
			distinctIdx = append(distinctIdx, i)
		}
	}
	pipeline = append(pipeline, bson.D{{Key: "$group", Value: group}})

	if len(distinctIdx) > 0 {
		set := bson.D{}
		for _, i := range distinctIdx {
			f := fmt.Sprintf("a%d", i)
			set = append(set, bson.E{Key: f, Value: bson.M{"$size": "$" + f}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: set}})
	}

	if len(o.OrderItems) > 0 {
		sort := bson.D{}
		for _, oi := range o.OrderItems {
			dir := 1
			if oi.Direction == plan.OrderDesc {
				dir = -1
			}
			sort = append(sort, bson.E{Key: mongoSortField(o, oi), Value: dir})
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sort}})
	}

	if o.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(o.Limit)}})
	}
	return pipeline
}

func mongoAggExpr(a plan.AggregateItem) interface{} {
	switch a.Func {
	case plan.AggCount:
		if a.Distinct {
			return bson.M{"$addToSet": "$" + a.Prop}
		}
		if a.Prop == "" {
			return bson.M{"$sum": 1}
		}
		return bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$ne": bson.A{"$" + a.Prop, nil}}, 1, 0}}}
	case plan.AggSum:
		return bson.M{"$sum": "$" + a.Prop}
	case plan.AggAvg:
		return bson.M{"$avg": "$" + a.Prop}
	case plan.AggMin:
		return bson.M{"$min": "$" + a.Prop}
	case plan.AggMax:
		return bson.M{"$max": "$" + a.Prop}
	}
	return nil
}

func mongoSortField(o *plan.StorePushdown, oi plan.OrderItem) string {
	for i, gk := range o.GroupKeys {
		if oi.Key == gk.OutName || oi.Key == gk.Alias+"."+gk.Prop {
			return fmt.Sprintf("_id.g%d", i)
		}
	}
	for i, a := range o.Aggs {
		if oi.Key == a.OutName {
			return fmt.Sprintf("a%d", i)
		}
	}
	return oi.Key
}

// ===== columnar (Cassandra): 全体集約のみ（GROUP BY / ORDER BY / DISTINCT はプランナで除外済み） =====

func runColumnarPushdown(qp *QueryProcessor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.cqlSes == nil {
		return 0
	}
	cql, args := buildColumnarCQL(o)
	iter := qp.cqlSes.Query(cql, args...).Iter()

	scanned := make(map[string]interface{})
	row := make(Row, len(o.Items))
	if iter.MapScan(scanned) {
		for i, it := range o.Items {
			row[it.Name] = scanned[fmt.Sprintf("c%d", i)]
		}
	}
	if err := iter.Close(); err != nil {
		fmt.Printf("StorePushdown[columnar] error: %v\n  CQL: %s\n", err, cql)
		return 0
	}
	if len(row) > 0 {
		out <- []Row{row}
	}
	return 1
}

func buildColumnarCQL(o *plan.StorePushdown) (string, []interface{}) {
	var selects []string
	for i, it := range o.Items {
		if !it.IsAggregate || it.Agg == nil {
			continue
		}
		selects = append(selects, cqlAggExpr(*it.Agg)+" AS c"+strconv.Itoa(i))
	}
	var where []string
	var args []interface{}
	for _, c := range o.Filters {
		if c == nil {
			continue
		}
		where = append(where, "\""+c.Property+"\" "+cqlOp(c.Type)+" ?")
		v, _ := codec.ConvertToNativeType(c.Value, c.DataType)
		args = append(args, v)
	}
	q := "SELECT " + strings.Join(selects, ", ") + " FROM \"" + o.Table + "\""
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ALLOW FILTERING"
	return q, args
}

func cqlAggExpr(a plan.AggregateItem) string {
	fn := strings.ToLower(a.Func.String())
	if a.Func == plan.AggCount && a.Prop == "" {
		return "count(*)"
	}
	return fn + "(\"" + a.Prop + "\")"
}

// ===== 共通ヘルパ =====

// typeParams は string パラメータを Neo4j 用の typed 値へ変換する
// （int → time.Time(RFC3339) → string の順。RunNeo4j の toValuedParams と同一規則）。
func typeParams(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		if n, err := strconv.Atoi(v); err == nil {
			out[k] = n
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = t
		} else {
			out[k] = v
		}
	}
	return out
}

func itemProp(it plan.ReturnItem) string {
	if len(it.Props) > 0 {
		return it.Props[0]
	}
	return ""
}

// coerceScalar は SQL ドライバ等が返す []byte を数値/文字列へ寄せる。
func coerceScalar(v interface{}) interface{} {
	if b, ok := v.([]byte); ok {
		s := string(b)
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return s
	}
	return v
}
