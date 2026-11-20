package stores

import (
	"context"
	"time"

	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores/conn"
)

// Report is what one store call measured (L3): FirstRow is from sending to the first row, Receive the
// rest of the reception (the receiver's work excluded), Decode turning raw values into executor values.
// Profile is the store's own report (L4), set only when asked and the database is a Profiler.
type Report struct {
	FirstRow, Receive, Decode time.Duration
	Rows                      int
	Profile                   *result.StoreProfile
}

// Profiler is a database that reports how it runs a query (L4). It runs q once more for the report (the
// rows are not returned), so profiles are taken in a trial of their own.
type Profiler interface {
	Profile(ctx context.Context, reg *conn.Registry, q *Query, in []Key) (*result.StoreProfile, error)
}
