package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// Profile runs q under explain (executionStats): the plan with documents examined and returned.
func (DB) Profile(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key) (*result.StoreProfile, error) {
	db, ok := reg.Mongo()
	if !ok {
		return nil, fmt.Errorf("executor: mongo not connected")
	}
	p := q.PipelineBody()
	if p.OnDatabase {
		return nil, nil // a database aggregation ($documents) has no collection to explain
	}
	cmd := bson.D{
		{Key: "explain", Value: bson.D{{Key: "aggregate", Value: p.Collection}, {Key: "pipeline", Value: inPipeline(p, in)}, {Key: "cursor", Value: bson.D{}}}},
		{Key: "verbosity", Value: "executionStats"},
	}
	var raw bson.Raw
	if err := db.RunCommand(ctx, cmd).Decode(&raw); err != nil {
		return nil, fmt.Errorf("mongo %s explain: %w", p.Collection, err)
	}
	js, err := bson.MarshalExtJSON(raw, false, false)
	if err != nil {
		return nil, err
	}
	out := &result.StoreProfile{Store: "mongo", Query: p.Collection, Raw: json.RawMessage(js)}
	var doc map[string]any
	if json.Unmarshal(js, &doc) == nil {
		if st := findStats(doc); st != nil {
			out.Rows = num(st["nReturned"])
			out.DBHits = num(st["totalDocsExamined"]) + num(st["totalKeysExamined"])
			out.Time = time.Duration(num(st["executionTimeMillis"])) * time.Millisecond
		}
	}
	return out, nil
}

// findStats returns the first executionStats object in an explain result.
func findStats(v any) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		if st, ok := x["executionStats"].(map[string]any); ok {
			return st
		}
		for _, c := range x {
			if st := findStats(c); st != nil {
				return st
			}
		}
	case []any:
		for _, c := range x {
			if st := findStats(c); st != nil {
				return st
			}
		}
	}
	return nil
}

func num(v any) int64 {
	f, _ := v.(float64)
	return int64(f)
}
