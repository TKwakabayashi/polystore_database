package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func runShippedGraph(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	rows := make([]map[string]any, len(s.Rows))
	for i, r := range s.Rows {
		m := make(map[string]any, len(r))
		for j, v := range r {
			m[q.ShipNames[j]] = stores.ShipValue(v, false)
		}
		rows[i] = m
	}
	return runGraphParams(ctx, reg, q, rows)
}

func driver(reg *conn.Registry) (neo4j.DriverWithContext, error) {
	d, ok := reg.Neo4j()
	if !ok {
		return nil, fmt.Errorf("executor: neo4j not connected")
	}
	return d, nil
}

func streamGraph(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	d, err := driver(reg)
	if err != nil {
		return err
	}
	return streamCypher(ctx, d, q, graphParams(q, in), emit)
}

// graphParams is q's parameters with $in set to the input keys.
func graphParams(q *stores.Query, in []stores.Key) map[string]any {
	params := map[string]any{}
	for k, v := range q.TextBody().Params {
		params[k] = v
	}
	if len(q.InKeys) == 1 {
		params[stores.InParam] = stores.StoredIDs(in)
	} else if len(q.InKeys) > 1 {
		rows := make([]map[string]any, len(in))
		for i, k := range in {
			m := make(map[string]any, len(q.InKeys))
			for j, name := range q.InKeys {
				m[name] = id.Stored(k[j])
			}
			rows[i] = m
		}
		params[stores.InParam] = rows
	}
	return params
}

// runGraphParams runs Cypher with $in set to the boundary values as they are.
func runGraphParams(ctx context.Context, reg *conn.Registry, q *stores.Query, in any) ([][]any, error) {
	d, err := driver(reg)
	if err != nil {
		return nil, err
	}
	params := map[string]any{stores.InParam: in}
	for k, v := range q.TextBody().Params {
		params[k] = v
	}
	return stores.Collect(func(emit func([]any) error) error { return streamCypher(ctx, d, q, params, emit) })
}

// streamCypher runs Cypher and passes records to emit as they arrive; column positions are looked up once.
// When stopped early, closing the session discards the rest (closing runs even with a cancelled ctx).
func streamCypher(ctx context.Context, d neo4j.DriverWithContext, q *stores.Query, params map[string]any, emit func([]any) error) error {
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(context.WithoutCancel(ctx))
	text := q.TextBody().Text
	res, err := sess.Run(ctx, text, params)
	if err != nil {
		return fmt.Errorf("cypher %q: %w", text, err)
	}
	keys, err := res.Keys()
	if err != nil {
		return fmt.Errorf("cypher %q: %w", text, err)
	}
	pos := make([]int, len(q.Columns))
	for i, name := range stores.ColNames(len(q.Columns)) {
		pos[i] = -1
		for j, k := range keys {
			if k == name {
				pos[i] = j
			}
		}
	}
	for res.Next(ctx) {
		rec := res.Record()
		vals := make([]any, len(pos))
		for i, j := range pos {
			if j >= 0 && j < len(rec.Values) {
				vals[i] = rec.Values[j]
			}
		}
		if err := emit(vals); err != nil {
			return err
		}
	}
	return res.Err()
}

func Verbatim(ctx context.Context, reg *conn.Registry, cypher string, params map[string]any) ([]map[string]any, error) {
	d, err := driver(reg)
	if err != nil {
		return nil, err
	}
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(ctx)
	res, err := sess.Run(ctx, cypher, params)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	for res.Next(ctx) {
		rec := res.Record()
		row := make(map[string]any, len(rec.Keys))
		for i, k := range rec.Keys {
			row[k] = rec.Values[i]
		}
		rows = append(rows, row)
	}
	return rows, res.Err()
}
