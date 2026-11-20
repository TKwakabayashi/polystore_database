package exec

// backend.go: what the executor needs from the stores (Backend and its optional extensions).

import (
	"context"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
)

// Key is one input row's id columns (stores.Key).
type Key = stores.Key

// Backend accesses the stores. Results are executor values (ids as id.UUID, segments as *ListVal, values
// normalized; dbs.Session decodes them) in the order of stores.Query.Columns.
type Backend interface {
	Run(ctx context.Context, q *stores.Query, in []Key) ([][]any, error)
	// Verbatim sends the original Cypher to Neo4j as is (for queries that cannot be planned).
	Verbatim(ctx context.Context, cypher string, params map[string]any) ([]map[string]any, error)
}

// Streamer is a backend that passes results row by row. The executor uses it when available and decodes rows
// as they arrive (no raw result is kept). The slice passed to emit is the receiver's. When emit returns an
// error or ctx is cancelled, close the cursor and return that error.
type Streamer interface {
	Stream(ctx context.Context, q *stores.Query, in []Key, emit func(vals []any) error) error
}

// Shipment is the table passed to a shipped query (stores.Shipment).
type Shipment = stores.Shipment

// Shipper is a backend that runs shipped queries (stores.Query.ShipCols set). The backend converts values to
// their stored form (the driver boundary).
type Shipper interface {
	RunShipped(ctx context.Context, q *stores.Query, s Shipment) ([][]any, error)
}

// Reporter is a backend that measures each call (stores.Report). The executor uses it when
// ExecOptions.Breakdown or StoreProfile is set.
type Reporter interface {
	StreamReport(ctx context.Context, q *stores.Query, in []Key, emit func(vals []any) error, profile bool) (stores.Report, error)
}

// Compiler is a backend that compiles store queries itself (tests evaluate plans on an in-memory graph).
// Other backends get dbs.Compile.
type Compiler interface {
	Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error)
}

type compileFn func(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error)
