package mongo

import (
	"context"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// DB is MongoDB.
type DB struct{}

var (
	exprRules   = stores.ExprRules{Or: true, IsNull: true, Not: true}
	absorbRules = stores.AbsorbRules{Projection: true, Aggregation: true, GroupedAggregation: true, DistinctAggregate: true, SortLimitDistinct: true}
)

func (DB) Kind() kind.Kind           { return kind.Mongo }
func (DB) CanExpress(e ir.Expr) bool { return exprRules.Expressible(e) }
func (DB) CanAbsorb(op ir.Node) bool { return absorbRules.Absorbs(op) }
func (DB) Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerMongo(q, types)
}
func (DB) CompileShipped(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerShipMongo(q, types)
}
func (DB) Stream(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	return streamMongo(ctx, reg, q, in, emit)
}
func (DB) RunShipped(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	return runShippedMongo(ctx, reg, q, s)
}
