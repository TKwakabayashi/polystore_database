package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	mysqldrv "github.com/go-sql-driver/mysql"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func fetchRdb(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	query := fmt.Sprintf("SELECT uuid, %s FROM %s", strings.Join(cfg.Props, ", "), cfg.Entity)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		payload := make(map[string]interface{}, len(cfg.Props))
		var uuid id.UUID
		for i, colName := range cols {
			val := columns[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}

			if colName == "uuid" {
				uuid = id.FromAny(val)
				continue
			}
			nativeVal, err := codec.ParseToNative(val, typeMap[colName])
			if err != nil {
				return fmt.Errorf("parse error [relational] UUID %s, Prop %s: %w", uuid.String(), colName, err)
			}
			payload[colName] = nativeVal
		}
		if uuid.Empty() {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: uuid, Values: payload}:
		}
	}
	return nil
}

func upsertRdb(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	// the schema was prepared once by prepareDestSchema (no DDL here)

	// take locks in the same order across workers to make deadlocks rarer
	sort.Slice(rows, func(i, j int) bool { return id.Less(rows[i].UUID, rows[j].UUID) })

	numProps := len(cfg.Props)
	placeholderGroup := "(" + strings.Repeat("?,", numProps) + "?)"
	batchPlaceholders := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*(numProps+1))

	for i, r := range rows {
		batchPlaceholders[i] = placeholderGroup
		args = append(args, id.Stored(r.UUID))
		for _, p := range cfg.Props {
			finalVal, err := codec.PrepareForDB(r.Values[p], typeMap[p], "relational")
			if err != nil {
				return err
			}
			args = append(args, finalVal)
		}
	}

	updateClauses := make([]string, numProps)
	quotedProps := make([]string, numProps)
	for i, p := range cfg.Props {
		quotedProps[i] = fmt.Sprintf("`%s`", p)
		updateClauses[i] = fmt.Sprintf("`%s`=VALUES(`%s`)", p, p)
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (uuid, %s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		cfg.Entity, strings.Join(quotedProps, ", "), strings.Join(batchPlaceholders, ", "), strings.Join(updateClauses, ", "),
	)

	// retry deadlocks (1213) and lock waits (1205)
	const maxAttempts = 5
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			lastErr = err
			if isMySQLRetryable(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(attempt+1) * 20 * time.Millisecond):
				}
				continue
			}
			return fmt.Errorf("mysql bulk upsert failed: %w", err)
		}
		return nil
	}
	return fmt.Errorf("mysql bulk upsert failed after %d attempts: %w", maxAttempts, lastErr)
}

func isMySQLRetryable(err error) bool {
	var me *mysqldrv.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1213 || me.Number == 1205
	}
	return false
}

// rdbNullConditions builds "`col` IS NULL" for every column but uuid. Called once when the delete phase starts and
// reused by deleteRdb for each batch.
func rdbNullConditions(ctx context.Context, reg *conn.Registry, entity string) ([]string, error) {
	db, ok := reg.MySQL()
	if !ok {
		return nil, fmt.Errorf("relational store not available")
	}
	colRows, err := db.QueryContext(ctx,
		"SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		entity)
	if err != nil {
		return nil, fmt.Errorf("fetch columns for %s: %w", entity, err)
	}
	defer colRows.Close()

	var nullConditions []string
	for colRows.Next() {
		var c string
		if err := colRows.Scan(&c); err != nil {
			return nil, err
		}
		if c != "uuid" {
			nullConditions = append(nullConditions, fmt.Sprintf("`%s` IS NULL", c))
		}
	}
	return nullConditions, colRows.Err()
}

func deleteRdb(ctx context.Context, cfg stores.Target, reg *conn.Registry, nullConditions []string, rows []stores.Row) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	uuids := make([]interface{}, len(rows))
	placeholders := make([]string, len(rows))
	for i, r := range rows {
		uuids[i] = id.Stored(r.UUID)
		placeholders[i] = "?"
	}
	inList := strings.Join(placeholders, ", ")

	// 1. set the given properties to NULL
	setParts := make([]string, len(cfg.Props))
	for i, p := range cfg.Props {
		setParts[i] = fmt.Sprintf("`%s` = NULL", p)
	}
	updateQuery := fmt.Sprintf("UPDATE `%s` SET %s WHERE uuid IN (%s)",
		cfg.Entity, strings.Join(setParts, ", "), strings.Join(placeholders, ", "))

	if _, err := db.ExecContext(ctx, updateQuery, uuids...); err != nil {
		return fmt.Errorf("relational property nullify failed: %w", err)
	}

	// 2. Delete rows whose columns other than uuid are all NULL (rows with unmigrated columns stay). Without columns
	// other than uuid no condition can be built, so nothing is deleted.
	if len(nullConditions) == 0 {
		return nil
	}

	cleanupQuery := fmt.Sprintf("DELETE FROM `%s` WHERE uuid IN (%s) AND %s",
		cfg.Entity, inList, strings.Join(nullConditions, " AND "))
	if _, err := db.ExecContext(ctx, cleanupQuery, uuids...); err != nil {
		return fmt.Errorf("relational cleanup delete failed: %w", err)
	}

	return nil
}

func countRdb(ctx context.Context, cfg stores.Target, reg *conn.Registry) (int64, error) {
	sqlDB, ok := reg.MySQL()
	if !ok {
		return 0, fmt.Errorf("relational store not available")
	}
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", cfg.Entity)
	if err := sqlDB.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("mysql: count %s: %w", cfg.Entity, err)
	}
	return count, nil
}

// uuidSQLType / uuidCQLType are the uuid column types (following the storage form id.CurrentStorage()).
func uuidSQLType() string {
	if id.Binary() {
		return fmt.Sprintf("BINARY(%d)", id.Size)
	}
	return "VARCHAR(255) COLLATE " + codec.SQLTextCollation
}

// alignSQLCollation changes, in one ALTER TABLE, the string columns whose collation is not codec.SQLTextCollation
// (values stay; it fixes tables earlier migrations made with MySQL's default collation). Nothing to change: no-op.
func alignSQLCollation(ctx context.Context, db *sql.DB, table string) error {
	rows, err := db.QueryContext(ctx, `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLLATION_NAME IS NOT NULL AND COLLATION_NAME <> ?
		ORDER BY ORDINAL_POSITION`, table, codec.SQLTextCollation)
	if err != nil {
		return fmt.Errorf("collation of %s: %w", table, err)
	}
	var mods []string
	for rows.Next() {
		var name, typ, nullable string
		if err := rows.Scan(&name, &typ, &nullable); err != nil {
			rows.Close()
			return err
		}
		m := fmt.Sprintf("MODIFY `%s` %s CHARACTER SET utf8mb4 COLLATE %s", name, typ, codec.SQLTextCollation)
		if nullable == "NO" {
			m += " NOT NULL"
		}
		mods = append(mods, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(mods) == 0 {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE `%s` %s", table, strings.Join(mods, ", "))); err != nil {
		return fmt.Errorf("align collation of %s: %w", table, err)
	}
	return nil
}

// alignSQLUUID makes an existing table's uuid column type match the storage form. Only empty tables change (setup
// keeps definitions with TRUNCATE); a table with rows and another type is an error (forms are never mixed).
func alignSQLUUID(ctx context.Context, db *sql.DB, table string) error {
	var typ string
	err := db.QueryRowContext(ctx, "SELECT DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = 'uuid'", table).Scan(&typ)
	if err != nil {
		return fmt.Errorf("uuid column of %s: %w", table, err)
	}
	if (strings.EqualFold(typ, "binary")) == id.Binary() {
		return nil
	}
	var n int
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("table %s has %[3]d rows and its uuid column (%[2]s) is not in the storage form; empty it with setup and migrate again", table, typ, n)
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE `%s` MODIFY uuid %s", table, uuidSQLType()))
	return err
}

// Prepare creates the table and the property columns (string columns with codec.SQLTextCollation) and
// aligns an existing table's uuid type and collation.
func (DB) Prepare(ctx context.Context, reg *conn.Registry, t *stores.Target) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("mysql not connected")
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (uuid %s PRIMARY KEY)", t.Entity, uuidSQLType())); err != nil {
		return err
	}
	if err := alignSQLUUID(ctx, db, t.Entity); err != nil {
		return err
	}
	for _, p := range t.Props {
		alter := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", t.Entity, p, codec.MapToSQLType(t.Types[p]))
		if _, err := db.ExecContext(ctx, alter); err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return fmt.Errorf("alter %s.%s: %w", t.Entity, p, err)
		}
	}
	return alignSQLCollation(ctx, db, t.Entity)
}

// ScanWorkers is 1: the uuid column's collation order differs from byte order, so range shards might miss rows.
func (DB) ScanWorkers() int { return 1 }
func (DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, _, _ int) error {
	return fetchRdb(ctx, t, reg, t.Types, out)
}
func (DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertRdb(ctx, t, reg, t.Types, rows)
}
func (DB) Finish(context.Context, *conn.Registry, stores.Target) error { return nil }

// Delete clears the target's columns of rows and deletes rows whose columns are all null.
func (DB) Delete(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	conds, err := rdbNullConditions(ctx, reg, t.Entity)
	if err != nil {
		return err
	}
	return deleteRdb(ctx, t, reg, conds, rows)
}
func (DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countRdb(ctx, t, reg)
}
