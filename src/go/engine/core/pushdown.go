// Package core は3実行エンジン（stream / bulk / volcano）が共有する基盤。
//
// pushdown のネイティブクエリ生成（SQL / Mongo / CQL）は3エンジンで byte-identical
// だった重複を1本化したもの。エンジンの run*Pushdown からは
// BuildRelationalSQL / BuildMongoPipeline / BuildColumnarCQL / CoerceScalar を呼ぶ。
package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func BuildRelationalSQL(o *plan.StorePushdown) (string, []interface{}) {
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
		where = append(where, sqlIdent(c.Property)+" "+SQLOp(c.Type)+" ?")
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

func BuildMongoPipeline(o *plan.StorePushdown) mongo.Pipeline {
	var pipeline mongo.Pipeline

	if len(o.Filters) > 0 {
		match := bson.D{}
		for _, c := range o.Filters {
			if c == nil {
				continue
			}
			v, _ := codec.ConvertToNativeType(c.Value, c.DataType)
			match = append(match, bson.E{Key: c.Property, Value: bson.M{MongoOp(c.Type): v}})
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
		sortDoc := bson.D{}
		for _, oi := range o.OrderItems {
			dir := 1
			if oi.Direction == plan.OrderDesc {
				dir = -1
			}
			sortDoc = append(sortDoc, bson.E{Key: mongoSortField(o, oi), Value: dir})
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortDoc}})
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

func BuildColumnarCQL(o *plan.StorePushdown) (string, []interface{}) {
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
		where = append(where, "\""+c.Property+"\" "+CQLOp(c.Type)+" ?")
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

func TypeParams(params map[string]string) map[string]interface{} {
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

func CoerceScalar(v interface{}) interface{} {
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
