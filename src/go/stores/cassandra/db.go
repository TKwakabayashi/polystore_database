package cassandra

import (
	"context"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/internal/sqlq"
)

// DB is Cassandra. It cannot join a shipped table; the executor emulates shipped queries.
type DB struct{}

var (
	exprRules   = stores.ExprRules{}
	absorbRules = stores.AbsorbRules{Projection: true, Aggregation: true}
)

func (DB) Kind() kind.Kind           { return kind.Cassandra }
func (DB) CanExpress(e ir.Expr) bool { return exprRules.Expressible(e) }
func (DB) CanAbsorb(op ir.Node) bool { return absorbRules.Absorbs(op) }
func (DB) Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return sqlq.Lower(q, types, sqlq.CQL)
}
func (DB) Stream(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	return streamCQL(ctx, reg, q, in, emit)
}
