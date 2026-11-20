// Package cassandra is the Cassandra database.
package cassandra

import (
	"context"
	"fmt"
	"reflect"

	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func streamCQL(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	sess, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("executor: widecolumn (cassandra) not connected")
	}
	t := q.TextBody()
	args := append([]any{}, t.Args...)
	if t.InArg >= 0 {
		args[t.InArg] = stores.StoredIDs(in)
	}
	iter := sess.Query(t.Text, args...).WithContext(ctx).Iter()
	// MapScan returns zero values for null, so it is not used: each column is read into a **T and null stays nil.
	// The **T holders are made once and their *T reset to nil per row (the values are new on every row).
	cols := iter.Columns()
	pos := make([]int, len(q.Columns))
	for i, name := range stores.ColNames(len(q.Columns)) {
		pos[i] = -1
		for j, c := range cols {
			if c.Name == name {
				pos[i] = j
			}
		}
	}
	dest := make([]any, len(cols))
	slots := make([]reflect.Value, len(cols)) // the *T that dest[j] points to
	for j, c := range cols {
		pp := reflect.New(reflect.TypeOf(c.TypeInfo.New()))
		dest[j], slots[j] = pp.Interface(), pp.Elem()
	}
	for {
		for _, s := range slots {
			s.Set(reflect.Zero(s.Type()))
		}
		if !iter.Scan(dest...) {
			break
		}
		vals := make([]any, len(q.Columns))
		for i, j := range pos {
			if j >= 0 && !slots[j].IsNil() {
				vals[i] = slots[j].Elem().Interface()
			}
		}
		if err := emit(vals); err != nil {
			_ = iter.Close()
			return err
		}
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("cql %q: %w", t.Text, err)
	}
	return nil
}
