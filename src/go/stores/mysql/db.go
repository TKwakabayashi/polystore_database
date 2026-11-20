package mysql

import (
	"context"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/internal/sqlq"
)

// DB is MySQL.
type DB struct{}

var (
	exprRules   = stores.ExprRules{Or: true, Arith: true, IsNull: true, Not: true, OtherUnary: true}
	absorbRules = stores.AbsorbRules{Projection: true, Aggregation: true, GroupedAggregation: true, DistinctAggregate: true, SortLimitDistinct: true}
)

func (DB) Kind() kind.Kind           { return kind.MySQL }
func (DB) CanExpress(e ir.Expr) bool { return exprRules.Expressible(e) }
func (DB) CanAbsorb(op ir.Node) bool { return absorbRules.Absorbs(op) }
func (DB) Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return sqlq.Lower(q, types, sqlq.MySQL)
}
func (DB) CompileShipped(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return sqlq.LowerShip(q, types)
}
func (DB) Stream(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	return streamSQL(ctx, reg, q, in, emit)
}
func (DB) RunShipped(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	return runShippedSQL(ctx, reg, q, s)
}
