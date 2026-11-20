package mysql

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/internal/sqlq"
)

// rootStats reads the top operator's "actual time=first..last rows=n" from EXPLAIN ANALYZE.
var rootStats = regexp.MustCompile(`actual time=[0-9.e+-]+?\.\.([0-9.e+-]+) rows=([0-9.e+-]+)`)

// Profile runs q under EXPLAIN ANALYZE: the iterator tree with actual times and rows.
func (DB) Profile(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key) (*result.StoreProfile, error) {
	db, ok := reg.MySQL()
	if !ok {
		return nil, fmt.Errorf("executor: mysql not connected")
	}
	t := q.TextBody()
	text, args := sqlq.ExpandIn(t.Text, t.Args, t.InArg, stores.StoredList(in))
	var out string
	if err := db.QueryRowContext(ctx, "EXPLAIN ANALYZE "+text, args...).Scan(&out); err != nil {
		return nil, fmt.Errorf("sql %q: %w", text, err)
	}
	p := &result.StoreProfile{Store: "mysql", Query: text, Raw: strings.Split(strings.TrimSpace(out), "\n")}
	p.Time, p.Rows = analyzeRoot(out)
	return p, nil
}

// analyzeRoot reads the top operator's last-row time and rows from EXPLAIN ANALYZE output (times may be
// written like 250e-6).
func analyzeRoot(out string) (time.Duration, int64) {
	m := rootStats.FindStringSubmatch(out)
	if m == nil {
		return 0, 0
	}
	var d time.Duration
	var rows int64
	if ms, err := strconv.ParseFloat(m[1], 64); err == nil {
		d = time.Duration(ms * float64(time.Millisecond))
	}
	if n, err := strconv.ParseFloat(m[2], 64); err == nil {
		rows = int64(n)
	}
	return d, rows
}
