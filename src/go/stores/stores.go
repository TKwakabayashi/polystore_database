package stores

import (
	"context"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores/conn"
)

// DB is one database. Each stores/<db> package implements it; stores/dbs lists them.
type DB interface {
	Kind() kind.Kind
	// CanExpress reports whether the database evaluates predicate e natively.
	CanExpress(e ir.Expr) bool
	// CanAbsorb reports whether a leaf query on the database can run whole-row operator op.
	CanAbsorb(op ir.Node) bool
	// Compile turns a query with the id-list boundary (ShipMode none) into a native query.
	Compile(q *ir.StoreQuery, types TypeFn) (*Query, error)
	// Stream runs q for the input keys and passes each result row to emit. The row slice is the
	// receiver's. Rows hold stored values; the executor decodes them with the query's column kinds.
	Stream(ctx context.Context, reg *conn.Registry, q *Query, in []Key, emit func(vals []any) error) error
}

// Shipper is a database that joins a shipped table inside the query.
type Shipper interface {
	CompileShipped(q *ir.StoreQuery, types TypeFn) (*Query, error)
	RunShipped(ctx context.Context, reg *conn.Registry, q *Query, s Shipment) ([][]any, error)
}

// PlainCols reports whether every non-nil expression is a column reference.
func PlainCols(es ...ir.Expr) bool {
	for _, e := range es {
		if e == nil {
			continue
		}
		if _, ok := e.(*ir.Column); !ok {
			return false
		}
	}
	return true
}
