package stream

import (
	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ScanDocStream(qp *Processor,
	o *plan.EntityScan, output chan<- []Record) (int, error) {
	// --- DB特有: bson クエリ構築 ---
	query := bson.D{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		if cond.DataType == "int" || cond.DataType == "integer" {
			if v32, ok := val.(int32); ok {
				val = int64(v32)
			}
		}
		query = append(query, bson.E{Key: cond.Property, Value: bson.M{core.MongoOp(cond.Type): val}})
	}

	// --- 以下 ScanGraphStream と同じストリーミング骨格 ---
	const outputBatchSize = 500
	rowCount := 0
	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]
	currentBatch := make([]Record, 0, outputBatchSize)

	for _, label := range o.Labels {
		cur, err := qp.mDb.Collection(label).Find(qp.ctx, query)
		if err != nil {
			return rowCount, err
		}
		for cur.Next(qp.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err != nil || rec.UUID == "" {
				continue
			}
			newSlots := make([]uid.UUID, newSlotCount)
			newSlots[aliasIdx] = uid.UUID(rec.UUID)
			currentBatch = append(currentBatch, Record{Slots: newSlots})
			rowCount++
			if len(currentBatch) >= outputBatchSize {
				output <- currentBatch
				currentBatch = make([]Record, 0, outputBatchSize)
			}
		}
		cur.Close(qp.ctx)
	}
	if len(currentBatch) > 0 {
		output <- currentBatch
	}
	return rowCount, nil
}

func FilterDocStream(qp *Processor,
	o *plan.Filter, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// --- DB特有: 共通条件 (uuid 以外) ---
	var commonConditions []bson.E
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonConditions = append(commonConditions, bson.E{Key: cond.Property, Value: bson.M{core.MongoOp(cond.Type): val}})
	}

	return runBatches(
		qp.ctx, qp.exec, qp.sem, OpFilter, inputStream, outputStream,
		noResource, closeNoResource,
		func(_ struct{}, batch []Record) ([]Record, error) {
			// id ユニーク化（graph と同じ）
			idMap := make(map[uid.UUID]struct{})
			for _, r := range batch {
				idMap[r.Slots[filterIdxIn]] = struct{}{}
			}
			uniqueIDs := make([]string, 0, len(idMap))
			for id := range idMap {
				uniqueIDs = append(uniqueIDs, id.String())
			}

			// --- DB特有: valid 抽出 ---
			query := append(bson.D{{Key: "uuid", Value: bson.M{"$in": uniqueIDs}}}, commonConditions...)
			validMap := make(map[uid.UUID]struct{})
			for _, label := range o.Labels {
				cur, err := qp.mDb.Collection(label).Find(qp.ctx, query, options.Find().SetProjection(bson.M{uid.PropName: 1}))
				if err != nil {
					return nil, err
				}
				for cur.Next(qp.ctx) {
					var rec struct {
						UUID string `bson:"uuid"`
					}
					if err := cur.Decode(&rec); err == nil {
						validMap[uid.UUID(rec.UUID)] = struct{}{}
					}
				}
				cur.Close(qp.ctx)
			}

			// 列引き継ぎ（graph と同じ inline）
			out := make([]Record, 0, len(batch))
			for _, r := range batch {
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
		},
	)
}

// fetchGraphPropsStream と同じ構造。クエリと行パースだけ Mongo 用。
func fetchDocPropsStream(qp *Processor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.mDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	for _, label := range unit.Labels {
		cur, err := qp.mDb.Collection(label).Find(qp.ctx, bson.M{"uuid": bson.M{"$in": ids}})
		if err != nil {
			continue
		}
		for cur.Next(qp.ctx) {
			var raw bson.M
			if err := cur.Decode(&raw); err != nil {
				continue
			}
			id, ok := raw["uuid"].(string)
			if !ok {
				continue
			}
			if _, exists := result[id]; !exists {
				result[id] = make(map[string]interface{})
			}
			for _, p := range fetch.Props {
				if val, ok := raw[p]; ok && val != nil {
					result[id][p], _ = codec.ConvertToNativeType(val, fetch.TypeMap[p])
				}
			}
		}
		cur.Close(qp.ctx)
	}
	return result
}
