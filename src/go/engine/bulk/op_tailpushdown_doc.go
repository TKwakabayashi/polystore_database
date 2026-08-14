package bulk

import (
	"fmt"
	"strconv"
	"time"

	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// runDocumentTail は MySQL 版 runRelationalTail の document(Mongo) 対応。
// tail（GROUP BY / 集約 / ORDER BY / LIMIT）を Mongo のネイティブ集約で丸ごと実行する。
//   - load: 中間 Record の staging エンティティ uuid を一時コレクションへ挿入（重複を保持＝multiplicity）。
//   - query: 一時コレクションから各永続コレクションへ $lookup（uuid 結合）→ $group → $sort → $limit。
//
// 出力列名は ReturnItem.Name（コーディネータ経路の行キーと一致）。
func runDocumentTail(qp *Processor, o *plan.TailPushdown, recs []Record) ([]Row, time.Duration, time.Duration, error) {
	if qp.mDb == nil {
		return nil, 0, 0, fmt.Errorf("TailPushdown[document]: mDb is nil")
	}
	if len(o.Entities) == 0 {
		return nil, 0, 0, fmt.Errorf("TailPushdown[document]: staging エンティティが空")
	}

	// alias → エンティティ index（temp 列 c{i} / lookup 別名 e{i}）。
	idxOf := make(map[string]int, len(o.Entities))
	slotIdx := make([]int, len(o.Entities))
	for i, e := range o.Entities {
		idxOf[e.Alias] = i
		if s, ok := o.InputSlot.VarToSlot[e.Alias]; ok {
			slotIdx[i] = s
		} else {
			slotIdx[i] = -1
		}
	}

	tmp := fmt.Sprintf("_tp_tail_%d", time.Now().UnixNano())
	coll := qp.mDb.Collection(tmp)

	// ---- load 区間: 中間 Record を一時コレクションへ挿入（重複保持） ----
	loadStart := time.Now()
	batch := make([]interface{}, 0, tailInsertBatch)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if _, err := coll.InsertMany(qp.ctx, batch); err != nil {
			return fmt.Errorf("TailPushdown[document]: insert 失敗: %w", err)
		}
		batch = batch[:0]
		return nil
	}
	for _, r := range recs {
		d := bson.M{}
		for i, si := range slotIdx {
			if si >= 0 && si < len(r.Slots) {
				d[fmt.Sprintf("c%d", i)] = r.Slots[si].String()
			} else {
				d[fmt.Sprintf("c%d", i)] = nil
			}
		}
		batch = append(batch, d)
		if len(batch) >= tailInsertBatch {
			if err := flush(); err != nil {
				_ = coll.Drop(qp.ctx)
				return nil, 0, 0, err
			}
		}
	}
	if err := flush(); err != nil {
		_ = coll.Drop(qp.ctx)
		return nil, 0, 0, err
	}
	loadDur := time.Since(loadStart)
	defer coll.Drop(qp.ctx)

	// ---- query 区間: $lookup + $group + $sort + $limit ----
	queryStart := time.Now()
	pipeline := buildDocTailPipeline(o, idxOf)
	cur, err := coll.Aggregate(qp.ctx, pipeline)
	if err != nil {
		return nil, loadDur, 0, fmt.Errorf("TailPushdown[document] aggregate error: %w", err)
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
		row := make(Row, len(o.Return))
		gi, ai := 0, 0
		for _, it := range o.Return {
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
	queryDur := time.Since(queryStart)
	return result, loadDur, queryDur, cur.Err()
}

// buildDocTailPipeline は $lookup（uuid 結合）→ $group → $sort → $limit のパイプラインを組む。
// _id.g{gi} は非集約 RETURN 項目（＝GROUP BY キー）を RETURN 順に、a{ai} は集約を RETURN 順に並べる
// （runDocumentTail の結果マッピングと一致させる）。
func buildDocTailPipeline(o *plan.TailPushdown, idxOf map[string]int) mongo.Pipeline {
	var pipeline mongo.Pipeline

	// staging エンティティごとに $lookup + $unwind（INNER JOIN 相当）。
	for i, e := range o.Entities {
		pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.M{
			"from":         e.Table,
			"localField":   fmt.Sprintf("c%d", i),
			"foreignField": uid.PropName,
			"as":           fmt.Sprintf("e%d", i),
		}}})
		pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: "$e" + strconv.Itoa(i)}})
	}

	// $group: _id は非集約 RETURN 項目、値は集約。
	idDoc := bson.D{}
	gi := 0
	for _, it := range o.Return {
		if it.IsAggregate {
			continue
		}
		i, ok := idxOf[it.Alias]
		if !ok {
			continue
		}
		prop := ""
		if len(it.Props) > 0 {
			prop = it.Props[0]
		}
		field := fmt.Sprintf("$c%d", i) // 束縛 uuid（prop 無し）
		if prop != "" {
			field = fmt.Sprintf("$e%d.%s", i, prop)
		}
		idDoc = append(idDoc, bson.E{Key: fmt.Sprintf("g%d", gi), Value: field})
		gi++
	}
	group := bson.D{{Key: "_id", Value: idDoc}}
	var distinctIdx []int
	ai := 0
	for _, it := range o.Return {
		if !it.IsAggregate || it.Agg == nil {
			continue
		}
		group = append(group, bson.E{Key: fmt.Sprintf("a%d", ai), Value: docTailAggExpr(*it.Agg, idxOf)})
		if it.Agg.Distinct {
			distinctIdx = append(distinctIdx, ai)
		}
		ai++
	}
	pipeline = append(pipeline, bson.D{{Key: "$group", Value: group}})

	// DISTINCT は $addToSet で集合化 → $size で個数化。
	if len(distinctIdx) > 0 {
		set := bson.D{}
		for _, i := range distinctIdx {
			f := fmt.Sprintf("a%d", i)
			set = append(set, bson.E{Key: f, Value: bson.M{"$size": "$" + f}})
		}
		pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: set}})
	}

	// $sort（出力別名で並べる → _id.g{gi} / a{ai} へ解決）。
	if len(o.OrderItems) > 0 {
		sortDoc := bson.D{}
		for _, oi := range o.OrderItems {
			dir := 1
			if oi.Direction == plan.OrderDesc {
				dir = -1
			}
			sortDoc = append(sortDoc, bson.E{Key: docTailSortField(o, oi), Value: dir})
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortDoc}})
	}

	if o.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(o.Limit)}})
	}
	return pipeline
}

// docTailAggExpr は集約式を $group 用 bson へ翻訳する（core.mongoAggExpr の $lookup 版：$e{i}.prop 参照）。
func docTailAggExpr(a plan.AggregateItem, idxOf map[string]int) interface{} {
	i, ok := idxOf[a.Alias]
	switch a.Func {
	case plan.AggCount:
		if a.Prop == "" {
			if a.Distinct && ok {
				return bson.M{"$addToSet": fmt.Sprintf("$c%d", i)} // count(DISTINCT alias): uuid 集合
			}
			return bson.M{"$sum": 1} // count(*) / count(alias): 中間行数（INNER なので束縛あり）
		}
		if !ok {
			return bson.M{"$sum": 1}
		}
		f := fmt.Sprintf("$e%d.%s", i, a.Prop)
		if a.Distinct {
			return bson.M{"$addToSet": f}
		}
		return bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$ne": bson.A{f, nil}}, 1, 0}}}
	}
	if !ok {
		return bson.M{"$sum": 0}
	}
	f := fmt.Sprintf("$e%d.%s", i, a.Prop)
	switch a.Func {
	case plan.AggSum:
		return bson.M{"$sum": f}
	case plan.AggAvg:
		return bson.M{"$avg": f}
	case plan.AggMin:
		return bson.M{"$min": f}
	case plan.AggMax:
		return bson.M{"$max": f}
	}
	return nil
}

// docTailSortField は ORDER BY の出力別名（oi.Key）を $group 後のフィールド（_id.g{gi} / a{ai}）へ解決する。
func docTailSortField(o *plan.TailPushdown, oi plan.OrderItem) string {
	key := oi.Key
	if key == "" {
		key = oi.Alias + "." + oi.Prop
	}
	gi, ai := 0, 0
	for _, it := range o.Return {
		if it.IsAggregate {
			if it.Name == key {
				return fmt.Sprintf("a%d", ai)
			}
			ai++
		} else {
			if it.Name == key || it.Alias+"."+firstProp(it) == key {
				return fmt.Sprintf("_id.g%d", gi)
			}
			gi++
		}
	}
	return key
}

func firstProp(it plan.ReturnItem) string {
	if len(it.Props) > 0 {
		return it.Props[0]
	}
	return ""
}
