package mongo

import (
	"bytes"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func runShippedMongo(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	db, ok := reg.Mongo()
	if !ok {
		return nil, fmt.Errorf("executor: mongo not connected")
	}
	docs := make(bson.A, len(s.Rows))
	for i, r := range s.Rows {
		d := bson.D{}
		for j, v := range r {
			d = append(d, bson.E{Key: q.ShipNames[j], Value: stores.ShipValue(v, false)})
		}
		docs[i] = d
	}
	pipeline := mongo.Pipeline{}
	for _, st := range q.PipelineBody().Stages {
		var body any = toBSON(st.Body, nil)
		if st.Op == "$documents" {
			body = docs
		}
		if st.Op == "$limit" {
			body = st.Body[0].Val
		}
		pipeline = append(pipeline, bson.D{{Key: st.Op, Value: body}})
	}
	cur, err := db.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("mongo aggregate: %w", err)
	}
	return stores.Collect(func(emit func([]any) error) error { return scanCursor(ctx, cur, len(q.Columns), emit) })
}

// toBSON converts a stores.Doc to bson.D, replacing the input uuid marker with ids.
func toBSON(v any, ids any) any {
	switch x := v.(type) {
	case stores.Doc:
		d := bson.D{}
		for _, f := range x {
			d = append(d, bson.E{Key: f.Key, Value: toBSON(f.Val, ids)})
		}
		return d
	case []any:
		out := make(bson.A, len(x))
		for i, e := range x {
			out[i] = toBSON(e, ids)
		}
		return out
	}
	if v == stores.InMarker {
		return ids
	}
	return v
}

func streamMongo(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("executor: mongo not connected")
	}
	p := q.PipelineBody()
	cur, err := db.Collection(p.Collection).Aggregate(ctx, inPipeline(p, in))
	if err != nil {
		return fmt.Errorf("mongo %s: %w", p.Collection, err)
	}
	return scanCursor(ctx, cur, len(q.Columns), emit)
}

// inPipeline is p's stages with the input uuids in the first $match.
func inPipeline(p *stores.PipelineBody, in []stores.Key) mongo.Pipeline {
	ids := stores.StoredIDs(in)
	pipeline := mongo.Pipeline{}
	for _, s := range p.Stages {
		body := toBSON(s.Body, ids)
		if s.Op == "$limit" || s.Op == "$skip" {
			body = s.Body[0].Val // {$limit: n}
		}
		pipeline = append(pipeline, bson.D{{Key: s.Op, Value: body}})
	}
	return pipeline
}

// scanCursor passes the aggregation's documents to emit as they arrive (columns c0, c1, ...) and closes the cursor.
func scanCursor(ctx context.Context, cur *mongo.Cursor, n int, emit func([]any) error) error {
	defer cur.Close(context.WithoutCancel(ctx))
	names := stores.ColNames(n)
	for cur.Next(ctx) {
		// Read only columns c0, c1, ... from the raw BSON (cur.Current) instead of decoding the whole document.
		vals := make([]any, n)
		for i, name := range names {
			rv, err := cur.Current.LookupErr(name)
			if err != nil {
				continue // a missing column is null (as when decoding into bson.M)
			}
			v, err := rawToGo(rv)
			if err != nil {
				return err
			}
			vals[i] = fromBSON(v)
		}
		if err := emit(vals); err != nil {
			return err
		}
	}
	return cur.Err()
}

// rawToGo converts a raw BSON value to the Go value decoding into bson.M gives (common types are read directly).
// Values are copied so they do not point into the cursor's buffer (reused by the next Next).
func rawToGo(rv bson.RawValue) (any, error) {
	switch rv.Type {
	case bsontype.Null, bsontype.Undefined:
		return nil, nil
	case bsontype.String:
		return rv.StringValue(), nil
	case bsontype.Int32:
		return rv.Int32(), nil
	case bsontype.Int64:
		return rv.Int64(), nil
	case bsontype.Double:
		return rv.Double(), nil
	case bsontype.Boolean:
		return rv.Boolean(), nil
	case bsontype.DateTime:
		return primitive.DateTime(rv.DateTime()), nil
	case bsontype.Binary:
		st, data := rv.Binary()
		return primitive.Binary{Subtype: st, Data: bytes.Clone(data)}, nil
	case bsontype.EmbeddedDocument:
		var m bson.M
		err := rv.Unmarshal(&m)
		return m, err
	}
	var x any
	err := rv.Unmarshal(&x)
	return x, err
}

func fromBSON(v any) any {
	switch x := v.(type) {
	case primitive.DateTime:
		return x.Time().UTC()
	case primitive.Binary: // uuid stored as binary (BinData)
		return x.Data
	case bson.A:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = fromBSON(e)
		}
		return out
	}
	return v
}
