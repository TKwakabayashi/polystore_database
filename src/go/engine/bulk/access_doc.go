package bulk

import (
	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ScanDocBulk(qp *Processor, o *plan.EntityScan) ([]Record, error) {
	if qp.mDb == nil {
		return nil, nil
	}
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

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	out := make([]Record, 0)
	for _, label := range o.Labels {
		cur, err := qp.mDb.Collection(label).Find(qp.ctx, query)
		if err != nil {
			return out, err
		}
		for cur.Next(qp.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err != nil || rec.UUID == "" {
				continue
			}
			newSlots := make([]string, newSlotCount)
			newSlots[aliasIdx] = rec.UUID
			out = append(out, Record{Slots: newSlots})
		}
		cur.Close(qp.ctx)
	}
	return out, nil
}

func FilterDocBulk(qp *Processor, o *plan.Filter, in []Record) ([]Record, error) {
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

	idMap := make(map[string]struct{})
	for _, r := range in {
		idMap[r.Slots[filterIdxIn]] = struct{}{}
	}
	uniqueIDs := make([]string, 0, len(idMap))
	for id := range idMap {
		uniqueIDs = append(uniqueIDs, id)
	}

	// --- DB特有: valid 抽出 ---
	query := append(bson.D{{Key: "uuid", Value: bson.M{"$in": uniqueIDs}}}, commonConditions...)
	validMap := make(map[string]struct{})
	for _, label := range o.Labels {
		cur, err := qp.mDb.Collection(label).Find(qp.ctx, query, options.Find().SetProjection(bson.M{"uuid": 1}))
		if err != nil {
			return nil, err
		}
		for cur.Next(qp.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err == nil {
				validMap[rec.UUID] = struct{}{}
			}
		}
		cur.Close(qp.ctx)
	}

	out := make([]Record, 0, len(in))
	for _, r := range in {
		if _, ok := validMap[r.Slots[filterIdxIn]]; ok {
			newRec := Record{Slots: make([]string, newSlotCount)}
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

func fetchDocPropsBulk(qp *Processor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
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
