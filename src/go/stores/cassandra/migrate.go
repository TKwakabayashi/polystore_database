package cassandra

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
	"golang.org/x/sync/errgroup"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func fetchCol(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row, shardIdx, shardTotal int) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("widecolumn store not available")
	}
	quotedProps := make([]string, len(cfg.Props))
	for i, p := range cfg.Props {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	query := fmt.Sprintf("SELECT \"uuid\", %s FROM \"%s\"", strings.Join(quotedProps, ", "), cfg.Entity)
	// bound the shard by the token range of partition key uuid (tokens are uniform)
	var args []interface{}
	if shardTotal > 1 {
		lo, hi, hasHi := id.ShardTokenBounds(shardIdx, shardTotal)
		query += " WHERE token(uuid) >= ?"
		args = append(args, lo)
		if hasHi {
			query += " AND token(uuid) < ?"
			args = append(args, hi)
		}
	}
	iter := session.Query(query, args...).WithContext(ctx).Iter()
	next := nullableRows(iter)
	for {
		row, ok := next()
		if !ok {
			break
		}

		uuid := id.FromAny(row["uuid"])
		if uuid.Empty() {
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Props))
		for _, p := range cfg.Props {
			var val interface{}
			if v, ok := row[p]; ok {
				val = v
			} else {
				for k, v := range row {
					if strings.EqualFold(k, p) {
						val = v
						break
					}
				}
			}
			// ParseToNative
			nativeVal, err := codec.ParseToNative(val, typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [widecolumn] UUID %s, Prop %s: %w", uuid, p, err)
			}
			payload[p] = nativeVal
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: uuid, Values: payload}:
		}
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("cql %q: %w", query, err)
	}
	return nil
}

// nullableRows reads iter's rows as column name -> value with null as nil (MapScan returns zero values
// for null, which a migration would write as "" or 0).
func nullableRows(iter *gocql.Iter) func() (map[string]any, bool) {
	cols := iter.Columns()
	dest := make([]any, len(cols))
	slots := make([]reflect.Value, len(cols))
	for j, c := range cols {
		pp := reflect.New(reflect.TypeOf(c.TypeInfo.New()))
		dest[j], slots[j] = pp.Interface(), pp.Elem()
	}
	return func() (map[string]any, bool) {
		for _, sl := range slots {
			sl.Set(reflect.Zero(sl.Type()))
		}
		if !iter.Scan(dest...) {
			return nil, false
		}
		row := make(map[string]any, len(cols))
		for j, c := range cols {
			if !slots[j].IsNil() {
				row[c.Name] = slots[j].Elem().Interface()
			} else {
				row[c.Name] = nil
			}
		}
		return row, true
	}
}

func upsertCol(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("widecolumn store not available")
	}
	// the schema was prepared once by prepareDestSchema (no DDL here)

	eg, gctx := errgroup.WithContext(ctx)
	// bound the goroutines: one per row would exhaust Cassandra's connection pool
	eg.SetLimit(32)
	quotedProps := make([]string, len(cfg.Props))
	for i, p := range cfg.Props {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	insertQuery := fmt.Sprintf("INSERT INTO \"%s\" (uuid, %s) VALUES (?, %s)",
		cfg.Entity, strings.Join(quotedProps, ", "), strings.Repeat("?, ", len(cfg.Props)-1)+"?")

	for _, r := range rows {
		r := r
		eg.Go(func() error {
			args := []interface{}{id.Stored(r.UUID)}
			for _, p := range cfg.Props {
				finalVal, _ := codec.PrepareForDB(r.Values[p], typeMap[p], "widecolumn")
				args = append(args, finalVal)
			}
			return session.Query(insertQuery, args...).WithContext(gctx).Exec()
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("widecolumn (cassandra) upsert failed: %w", err)
	}

	return nil
}

func deleteCol(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("widecolumn store not available")
	}
	quotedProps := make([]string, len(cfg.Props))
	for i, p := range cfg.Props {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	// Delete the columns, then the row when no other column has a value (as MySQL and MongoDB do), so a later
	// migration from this table does not bring back rows of uuids that hold nothing.
	deleteQuery := fmt.Sprintf("DELETE %s FROM \"%s\" WHERE uuid = ?", strings.Join(quotedProps, ", "), cfg.Entity)
	rowQuery := fmt.Sprintf("SELECT * FROM \"%s\" WHERE uuid = ?", cfg.Entity)
	deleteRow := fmt.Sprintf("DELETE FROM \"%s\" WHERE uuid = ?", cfg.Entity)

	eg, gctx := errgroup.WithContext(ctx)
	// bound the goroutines so per-row goroutines do not exhaust connections
	eg.SetLimit(32)
	for _, r := range rows {
		r := r
		eg.Go(func() error {
			u := id.Stored(r.UUID)
			if err := session.Query(deleteQuery, u).WithContext(gctx).Exec(); err != nil {
				return err
			}
			iter := session.Query(rowQuery, u).WithContext(gctx).Iter()
			row, found := nullableRows(iter)()
			if err := iter.Close(); err != nil {
				return err
			}
			if !found {
				return nil
			}
			for c, v := range row {
				if c != "uuid" && v != nil {
					return nil
				}
			}
			return session.Query(deleteRow, u).WithContext(gctx).Exec()
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("widecolumn (cassandra) column delete failed: %w", err)
	}

	return nil
}

func countCol(ctx context.Context, cfg stores.Target, reg *conn.Registry) (int64, error) {
	sess, ok := reg.Cassandra()
	if !ok {
		return 0, fmt.Errorf("widecolumn store not available")
	}
	// SELECT COUNT(*) scans the whole table on the coordinator and times out on large tables, so page through uuids
	// and count on the client (the same automatic paging as fetchCol).
	iter := sess.Query(fmt.Sprintf("SELECT uuid FROM \"%s\"", cfg.Entity)).
		WithContext(ctx).PageSize(5000).Iter()
	var (
		uuid  string
		count int64
	)
	for iter.Scan(&uuid) {
		count++
	}
	if err := iter.Close(); err != nil {
		return 0, fmt.Errorf("cassandra: count %s: %w", cfg.Entity, err)
	}
	return count, nil
}

func uuidCQLType() string {
	if id.Binary() {
		return "blob"
	}
	return "text"
}

// alignCQLUUID recreates an empty table whose uuid column (the primary key) has another type (a primary key type
// cannot change). No table: no-op. A table with rows and another type is an error.
func alignCQLUUID(ctx context.Context, sess *gocql.Session, table string) error {
	iter := sess.Query(fmt.Sprintf(`SELECT uuid FROM "%s" LIMIT 1`, table)).WithContext(ctx).Iter()
	cols := iter.Columns()
	var raw []byte
	hasRow := iter.Scan(&raw)
	if err := iter.Close(); err != nil {
		// no table yet (the server says unconfigured table, gocql says does not exist from its metadata)
		if strings.Contains(err.Error(), "unconfigured table") || strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}
	if len(cols) == 0 || (cols[0].TypeInfo.Type() == gocql.TypeBlob) == id.Binary() {
		return nil
	}
	if hasRow {
		return fmt.Errorf("table %s has rows and its uuid column is not in the storage form; empty it with setup and migrate again", table)
	}
	return sess.Query(fmt.Sprintf(`DROP TABLE "%s"`, table)).WithContext(ctx).Exec()
}

// Prepare creates the table and the property columns, recreating an empty table whose uuid type differs.
func (DB) Prepare(ctx context.Context, reg *conn.Registry, t *stores.Target) error {
	sess, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("cassandra not connected")
	}
	if err := alignCQLUUID(ctx, sess, t.Entity); err != nil {
		return err
	}
	if err := sess.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (uuid %s PRIMARY KEY)`, t.Entity, uuidCQLType())).WithContext(ctx).Exec(); err != nil {
		return err
	}
	for _, p := range t.Props {
		if err := sess.Query(fmt.Sprintf(`ALTER TABLE "%s" ADD "%s" %s`, t.Entity, p, codec.MapToCassandraType(t.Types[p]))).WithContext(ctx).Exec(); err != nil &&
			!strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

// ScanWorkers is 4 (token range shards).
func (DB) ScanWorkers() int { return 4 }
func (DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, shard, shards int) error {
	return fetchCol(ctx, t, reg, t.Types, out, shard, shards)
}
func (DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertCol(ctx, t, reg, t.Types, rows)
}
func (DB) Finish(context.Context, *conn.Registry, stores.Target) error { return nil }
func (DB) Delete(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return deleteCol(ctx, t, reg, t.Types, rows)
}
func (DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countCol(ctx, t, reg)
}
