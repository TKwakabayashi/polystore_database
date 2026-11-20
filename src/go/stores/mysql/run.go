// Package mysql is the MySQL database.
package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"polystore_database/src/go/stores/internal/sqlq"

	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func runShippedSQL(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	db, ok := reg.MySQL()
	if !ok {
		return nil, fmt.Errorf("executor: mysql not connected")
	}
	rows := make([][]any, len(s.Rows))
	for i, r := range s.Rows {
		out := make([]any, len(r))
		for j, v := range r {
			out[j] = stores.ShipValue(v, true)
		}
		rows[i] = out
	}
	doc, err := json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	t := q.TextBody()
	args := append([]any{}, t.Args...)
	args[t.InArg] = string(doc)
	res, err := db.QueryContext(ctx, t.Text, args...)
	if err != nil {
		return nil, fmt.Errorf("sql %q: %w", t.Text, err)
	}
	defer res.Close()
	return stores.Collect(func(emit func([]any) error) error { return sqlq.Scan(res, len(q.Columns), emit) })
}

func streamSQL(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("executor: mysql not connected")
	}
	t := q.TextBody()
	text, args := sqlq.ExpandIn(t.Text, t.Args, t.InArg, stores.StoredList(in))
	rows, err := db.QueryContext(ctx, text, args...)
	if err != nil {
		return fmt.Errorf("sql %q: %w", text, err)
	}
	defer rows.Close()
	return sqlq.Scan(rows, len(q.Columns), emit)
}
