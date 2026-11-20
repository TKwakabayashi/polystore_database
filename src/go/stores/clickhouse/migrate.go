package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func chIdent(s string) string { return "`" + strings.ReplaceAll(s, "`", "``") + "`" }

func chStageTable(entity string) string { return entity + "__stage" }

// chUUID returns the stored uuid value for a String column.
func chUUID(u id.UUID) string {
	switch v := id.Stored(u).(type) {
	case []byte:
		return string(v)
	case string:
		return v
	}
	return ""
}

func clickHouseDB(reg *conn.Registry) (*sql.DB, error) {
	db, ok := reg.ClickHouse()
	if !ok {
		return nil, fmt.Errorf("columnar store not available")
	}
	return db, nil
}

func chCount(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var n uint64
	if err := db.QueryRowContext(ctx, "SELECT count() FROM "+chIdent(table)).Scan(&n); err != nil {
		return 0, err
	}
	return int64(n), nil
}

// chColumns returns the table's columns (except uuid) in definition order.
func chColumns(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM system.columns WHERE database = currentDatabase() AND table = ? ORDER BY position", table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		if n != id.PropName {
			out = append(out, n)
		}
	}
	return out, rows.Err()
}

// prepareClickHouse prepares the table and columns and sets cfg.Stage when the table has rows (write to a temporary table).
func prepareClickHouse(ctx context.Context, cfg *stores.Target, reg *conn.Registry, typeMap map[string]string) error {
	db, err := clickHouseDB(reg)
	if err != nil {
		return err
	}
	create := func(table string) error {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s String) ENGINE = MergeTree ORDER BY %s",
			chIdent(table), chIdent(id.PropName), chIdent(id.PropName))); err != nil {
			return fmt.Errorf("create %s: %w", table, err)
		}
		for _, p := range cfg.Props {
			if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s",
				chIdent(table), chIdent(p), codec.MapToClickHouseType(typeMap[p]))); err != nil {
				return fmt.Errorf("alter %s.%s: %w", table, p, err)
			}
		}
		return nil
	}
	if err := create(cfg.Entity); err != nil {
		return err
	}
	n, err := chCount(ctx, db, cfg.Entity)
	if err != nil {
		return err
	}
	cfg.Stage = ""
	if n == 0 {
		return nil
	}
	stage := chStageTable(cfg.Entity)
	if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS "+chIdent(stage)); err != nil {
		return err
	}
	if err := create(stage); err != nil {
		return err
	}
	cfg.Stage = stage
	return nil
}

// upsertClickHouse writes the rows with one INSERT (into the table when empty, else into the temporary table).
func upsertClickHouse(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, err := clickHouseDB(reg)
	if err != nil {
		return err
	}
	table := cfg.Entity
	if cfg.Stage != "" {
		table = cfg.Stage
	}
	cols := []string{chIdent(id.PropName)}
	for _, p := range cfg.Props {
		cols = append(cols, chIdent(p))
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (%s)", chIdent(table), strings.Join(cols, ", ")))
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clickhouse prepare insert: %w", err)
	}
	for _, r := range rows {
		args := []any{chUUID(r.UUID)}
		for _, p := range cfg.Props {
			v, err := codec.PrepareForDB(r.Values[p], typeMap[p], "columnar")
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			args = append(args, v)
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("clickhouse append %s: %w", r.UUID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("clickhouse insert into %s: %w", table, err)
	}
	return nil
}

// finalizeClickHouse, when there is a temporary table, swaps in a table joined with it on uuid and drops it.
func finalizeClickHouse(ctx context.Context, cfg stores.Target, reg *conn.Registry) error {
	if cfg.Stage == "" {
		return nil
	}
	db, err := clickHouseDB(reg)
	if err != nil {
		return err
	}
	all, err := chColumns(ctx, db, cfg.Entity)
	if err != nil {
		return err
	}
	moved := map[string]bool{}
	for _, p := range cfg.Props {
		moved[p] = true
	}
	u := chIdent(id.PropName)
	sel := []string{fmt.Sprintf("if(s.%s IS NOT NULL, s.%s, t.%s) AS %s", u, u, u, u)}
	for _, c := range all {
		q := chIdent(c)
		if moved[c] {
			// moved properties take the temporary table's value for its rows (null included)
			sel = append(sel, fmt.Sprintf("if(s.%s IS NOT NULL, s.%s, t.%s) AS %s", u, q, q, q))
		} else {
			sel = append(sel, fmt.Sprintf("t.%s AS %s", q, q))
		}
	}
	tmp := cfg.Entity + "__merged"
	stmts := []string{
		"DROP TABLE IF EXISTS " + chIdent(tmp),
		fmt.Sprintf("CREATE TABLE %s AS %s", chIdent(tmp), chIdent(cfg.Entity)),
		fmt.Sprintf("INSERT INTO %s SELECT %s FROM %s AS t FULL OUTER JOIN %s AS s ON t.%s = s.%s SETTINGS join_use_nulls = 1",
			chIdent(tmp), strings.Join(sel, ", "), chIdent(cfg.Entity), chIdent(cfg.Stage), u, u),
		fmt.Sprintf("EXCHANGE TABLES %s AND %s", chIdent(cfg.Entity), chIdent(tmp)),
		"DROP TABLE " + chIdent(tmp),
		"DROP TABLE " + chIdent(cfg.Stage),
	}
	for _, st := range stmts {
		if _, err := db.ExecContext(ctx, st); err != nil {
			return fmt.Errorf("clickhouse merge %s: %q: %w", cfg.Entity, st, err)
		}
	}
	return nil
}

// fetchClickHouse streams every row of the table.
func fetchClickHouse(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row) error {
	db, err := clickHouseDB(reg)
	if err != nil {
		return err
	}
	cols := []string{chIdent(id.PropName)}
	for _, p := range cfg.Props {
		cols = append(cols, chIdent(p))
	}
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), chIdent(cfg.Entity)))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		u := id.FromAny(chDeref(vals[0]))
		if u.Empty() {
			continue
		}
		payload := make(map[string]interface{}, len(cfg.Props))
		for i, p := range cfg.Props {
			v, err := codec.ParseToNative(chDeref(vals[i+1]), typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [columnar] UUID %s, Prop %s: %w", u, p, err)
			}
			payload[p] = v
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: u, Values: payload}:
		}
	}
	return rows.Err()
}

// chDeref turns a Nullable value (*T) into T or nil.
func chDeref(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	}
	return v
}

func countClickHouse(ctx context.Context, cfg stores.Target, reg *conn.Registry) (int64, error) {
	db, err := clickHouseDB(reg)
	if err != nil {
		return 0, err
	}
	n, err := chCount(ctx, db, cfg.Entity)
	if err != nil {
		return 0, fmt.Errorf("clickhouse: count %s: %w", cfg.Entity, err)
	}
	return n, nil
}

// deleteClickHouse deletes as a source: drops the moved properties' columns and the rows left without values
// (empties the table when no column remains).
func deleteClickHouse(ctx context.Context, cfg stores.Target, reg *conn.Registry) error {
	db, err := clickHouseDB(reg)
	if err != nil {
		return err
	}
	for _, p := range cfg.Props {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s", chIdent(cfg.Entity), chIdent(p))); err != nil {
			return fmt.Errorf("clickhouse drop column %s.%s: %w", cfg.Entity, p, err)
		}
	}
	rest, err := chColumns(ctx, db, cfg.Entity)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		_, err = db.ExecContext(ctx, "TRUNCATE TABLE "+chIdent(cfg.Entity))
		return err
	}
	conds := make([]string, len(rest))
	for i, c := range rest {
		conds[i] = chIdent(c) + " IS NULL"
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s", chIdent(cfg.Entity), strings.Join(conds, " AND ")))
	return err
}

// Prepare creates the table and columns, or a staging table when the table already has rows (t.Stage).
func (DB) Prepare(ctx context.Context, reg *conn.Registry, t *stores.Target) error {
	return prepareClickHouse(ctx, t, reg, t.Types)
}

// ScanWorkers is 1: one streamed SELECT is faster.
func (DB) ScanWorkers() int { return 1 }
func (DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, _, _ int) error {
	return fetchClickHouse(ctx, t, reg, t.Types, out)
}
func (DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertClickHouse(ctx, t, reg, t.Types, rows)
}

// Finish moves the staging table in.
func (DB) Finish(ctx context.Context, reg *conn.Registry, t stores.Target) error {
	return finalizeClickHouse(ctx, t, reg)
}

// Delete does nothing: a move deletes with DeleteAll (per-row deletes are heavy on MergeTree).
func (DB) Delete(context.Context, *conn.Registry, stores.Target, []stores.Row) error { return nil }

// DeleteAll drops the target's columns.
func (DB) DeleteAll(ctx context.Context, reg *conn.Registry, t stores.Target) error {
	return deleteClickHouse(ctx, t, reg)
}
func (DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countClickHouse(ctx, t, reg)
}
