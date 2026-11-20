package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func fetchDoc(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row, shardIdx, shardTotal int) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	// BSON compares strings bytewise, so a uuid range bounds the shard (using the uuid index).
	filter := bson.M{}
	if lo, hi := id.ShardBounds(shardIdx, shardTotal); lo != nil || hi != nil {
		cond := bson.M{}
		if lo != nil {
			cond["$gte"] = id.StoredBound(lo)
		}
		if hi != nil {
			cond["$lt"] = id.StoredBound(hi)
		}
		filter["uuid"] = cond
	}
	cur, err := db.Collection(cfg.Entity).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		uuid := id.FromAny(bsonValue(doc["uuid"]))
		if uuid.Empty() {
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Props))
		for _, p := range cfg.Props {
			// ParseToNative
			nativeVal, err := codec.ParseToNative(doc[p], typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [document] UUID %s, Prop %s: %w", uuid, p, err)
			}
			payload[p] = nativeVal
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: uuid, Values: payload}:
		}
	}

	return nil
}

// bsonValue turns a bson binary (a uuid stored as bytes) into []byte.
func bsonValue(v interface{}) interface{} {
	if b, ok := v.(primitive.Binary); ok {
		return b.Data
	}
	return v
}

func upsertDoc(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	coll := db.Collection(cfg.Entity)
	models := make([]mongo.WriteModel, len(rows))
	for i, row := range rows {
		// rebuild the payload with MongoDB-ready types
		finalPayload := make(map[string]interface{})
		for p, v := range row.Values {
			finalVal, _ := codec.PrepareForDB(v, typeMap[p], "document")
			finalPayload[p] = finalVal
		}
		models[i] = mongo.NewUpdateOneModel().
			SetFilter(bson.M{"uuid": id.Stored(row.UUID)}).
			SetUpdate(bson.M{"$set": finalPayload}).
			SetUpsert(true)
	}
	if _, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false)); err != nil {
		return fmt.Errorf("mongo bulk write failed: %w", err)
	}

	return nil
}

func deleteDoc(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	coll := db.Collection(cfg.Entity)

	unsetMap := bson.M{}
	for _, p := range cfg.Props {
		unsetMap[p] = ""
	}

	var models []mongo.WriteModel
	for _, r := range rows {
		// delete the properties
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"uuid": id.Stored(r.UUID)}).
			SetUpdate(bson.M{"$unset": unsetMap}))

		// delete the whole document when only uuid and _id remain
		deleteFilter := bson.M{
			"uuid":  id.Stored(r.UUID),
			"$expr": bson.M{"$lte": bson.A{bson.M{"$size": bson.M{"$objectToArray": "$$ROOT"}}, 2}},
		}
		models = append(models, mongo.NewDeleteOneModel().SetFilter(deleteFilter))
	}
	// run in order
	if _, err := coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(true)); err != nil {
		return fmt.Errorf("mongodb bulk delete failed: %w", err)
	}

	return nil
}

func countDoc(ctx context.Context, cfg stores.Target, reg *conn.Registry) (int64, error) {
	db, ok := reg.Mongo()
	if !ok {
		return 0, fmt.Errorf("document store not available")
	}
	count, err := db.Collection(cfg.Entity).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("mongo: count %s: %w", cfg.Entity, err)
	}
	return count, nil
}

// Prepare creates the unique uuid index (without it every upsert scans the collection). Idempotent.
func (DB) Prepare(ctx context.Context, reg *conn.Registry, t *stores.Target) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("mongo not connected")
	}
	if _, err := db.Collection(t.Entity).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("create uuid index on %s: %w", t.Entity, err)
	}
	return nil
}

// ScanWorkers is 4 (uuid range shards).
func (DB) ScanWorkers() int { return 4 }
func (DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, shard, shards int) error {
	return fetchDoc(ctx, t, reg, t.Types, out, shard, shards)
}
func (DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertDoc(ctx, t, reg, t.Types, rows)
}
func (DB) Finish(context.Context, *conn.Registry, stores.Target) error { return nil }
func (DB) Delete(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return deleteDoc(ctx, t, reg, t.Types, rows)
}
func (DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countDoc(ctx, t, reg)
}
