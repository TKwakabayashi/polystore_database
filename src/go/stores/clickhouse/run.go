// Package clickhouse is the ClickHouse database.
package clickhouse

import (
	"context"
	"fmt"
	"polystore_database/src/go/stores/internal/sqlq"
	"reflect"
	"strings"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// clickHouseIDs renders the first id column of the input as ClickHouse literals (the uuid column is a String;
// in binary storage it holds the 16 bytes, so the literal is unhex(...)).
func clickHouseIDs(in []stores.Key) string {
	parts := make([]string, len(in))
	for i, k := range in {
		t, _ := id.StoredText(k[0]).(string)
		if id.Binary() {
			parts[i] = "unhex('" + t + "')"
		} else {
			parts[i] = sqlq.ClickHouseString(t)
		}
	}
	return strings.Join(parts, ", ")
}

func streamClickHouse(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	db, ok := reg.ClickHouse()
	if !ok {
		return fmt.Errorf("executor: clickhouse not connected")
	}
	// Literals are embedded in the text; the only positional parameter is the input uuid list.
	text := q.TextBody().Text
	if q.TextBody().InArg >= 0 {
		if len(in) == 0 {
			return nil
		}
		i := strings.IndexByte(text, '?')
		if i < 0 {
			return fmt.Errorf("clickhouse: no placeholder for input in %q", text)
		}
		text = text[:i] + clickHouseIDs(in) + text[i+1:]
	}
	rows, err := db.QueryContext(ctx, text)
	if err != nil {
		return fmt.Errorf("clickhouse %q: %w", abbrev(text), err)
	}
	defer rows.Close()
	for rows.Next() {
		vals := make([]any, len(q.Columns))
		ptrs := make([]any, len(vals))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		for i, v := range vals {
			vals[i] = derefNullable(v)
		}
		if err := emit(vals); err != nil {
			return err
		}
	}
	return rows.Err()
}

// derefNullable turns a ClickHouse Nullable value (*T, nil for null) into T or nil.
func derefNullable(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return v
	}
	if rv.IsNil() {
		return nil
	}
	return rv.Elem().Interface()
}

// abbrev shortens a query text for error messages (IN lists hold thousands of uuids).
func abbrev(s string) string {
	if len(s) > 512 {
		return s[:512] + "…"
	}
	return s
}
