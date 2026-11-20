package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// Profile runs q with a query id and reads its row from system.query_log (rows and bytes read, duration).
func (DB) Profile(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key) (*result.StoreProfile, error) {
	db, ok := reg.ClickHouse()
	if !ok {
		return nil, fmt.Errorf("executor: clickhouse not connected")
	}
	qid := id.New().String()
	err := streamClickHouse(ch.Context(ctx, ch.WithQueryID(qid)), reg, q, in, func([]any) error { return nil })
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, "SYSTEM FLUSH LOGS"); err != nil {
		return nil, fmt.Errorf("clickhouse flush logs: %w", err)
	}
	var readRows, readBytes, resultRows, memory uint64
	var durMS uint64
	var text string
	err = db.QueryRowContext(ctx, `SELECT read_rows, read_bytes, result_rows, memory_usage, query_duration_ms, query
		FROM system.query_log WHERE query_id = ? AND type = 'QueryFinish' LIMIT 1`, qid).
		Scan(&readRows, &readBytes, &resultRows, &memory, &durMS, &text)
	if err != nil {
		return nil, fmt.Errorf("clickhouse query_log: %w", err)
	}
	return &result.StoreProfile{Store: "clickhouse", Query: strings.TrimSpace(text), Rows: int64(resultRows),
		DBHits: int64(readRows), Time: time.Duration(durMS) * time.Millisecond,
		Raw: map[string]any{"read_rows": readRows, "read_bytes": readBytes, "result_rows": resultRows,
			"memory_usage": memory, "query_duration_ms": durMS}}, nil
}
