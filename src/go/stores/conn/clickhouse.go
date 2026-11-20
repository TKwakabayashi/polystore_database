package conn

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/base/kind"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// clickhouse.go connects to ClickHouse (store.ClickHouse) through database/sql with the native protocol.

type ClickHouseStore struct{ db *sql.DB }

func clickHouseOptions(c ClickHouseConfig, database string) *clickhouse.Options {
	return &clickhouse.Options{
		Addr: c.Addr,
		Auth: clickhouse.Auth{Database: database, Username: c.User, Password: c.Password},
		Settings: clickhouse.Settings{
			// Boundary IN lists hold thousands of uuids, so raise the query text limit (256KiB by default).
			"max_query_size": 64 << 20,
		},
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Minute,
	}
}

func openClickHouse(ctx context.Context, c ClickHouseConfig) (*ClickHouseStore, error) {
	if err := EnsureClickHouseDatabase(ctx, c); err != nil {
		return nil, err
	}
	db := clickhouse.OpenDB(clickHouseOptions(c, c.Database))
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return &ClickHouseStore{db: db}, nil
}

// EnsureClickHouseDatabase creates c.Database if missing (connects through default first, since a missing DB refuses connections).
func EnsureClickHouseDatabase(ctx context.Context, c ClickHouseConfig) error {
	db := clickhouse.OpenDB(clickHouseOptions(c, "default"))
	defer db.Close()
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", c.Database)); err != nil {
		return fmt.Errorf("clickhouse create database %s: %w", c.Database, err)
	}
	return nil
}

func (s *ClickHouseStore) Kind() kind.Kind               { return kind.ClickHouse }
func (s *ClickHouseStore) DB() *sql.DB                   { return s.db }
func (s *ClickHouseStore) Close(_ context.Context) error { return s.db.Close() }

// OpenClickHouseDB connects to ClickHouse without a Registry (used by setup).
func OpenClickHouseDB(ctx context.Context, c ClickHouseConfig) (*sql.DB, error) {
	s, err := openClickHouse(ctx, c)
	if err != nil {
		return nil, err
	}
	return s.db, nil
}
