package volcano

import (
	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------- EntityScan ----------

func (p *Processor) scanDocIDs(o *plan.EntityScan) ([]string, error) {
	if p.mDb == nil {
		return nil, nil
	}
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

	var ids []string
	for _, label := range o.Labels {
		p.countRoundTrip()
		cur, err := p.mDb.Collection(label).Find(p.ctx, query)
		if err != nil {
			return nil, err
		}
		for cur.Next(p.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err != nil || rec.UUID == "" {
				continue
			}
			ids = append(ids, rec.UUID)
		}
		cur.Close(p.ctx)
	}
	return ids, nil
}

// ---------- Filter ----------

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
		commonConditions = append(commonConditions, bson.E{Key: cond.Property, Value: bson.M{core.MongoOp(cond.Type): val}})
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

// ---------- Projection fetch ----------

func (p *Processor) fetchDocProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.mDb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}
	for _, label := range unit.Labels {
		p.countRoundTrip()
		cur, err := p.mDb.Collection(label).Find(p.ctx, bson.M{"uuid": bson.M{"$in": ids}})
		if err != nil {
			continue
		}
		for cur.Next(p.ctx) {
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
			for _, prop := range fetch.Props {
				if val, ok := raw[prop]; ok && val != nil {
					result[id][prop], _ = codec.ConvertToNativeType(val, fetch.TypeMap[prop])
				}
			}
		}
		cur.Close(p.ctx)
	}
	return result
}
