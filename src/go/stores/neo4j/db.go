package neo4j

import (
	"context"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// DB is Neo4j. It evaluates every predicate and operator, and runs Cypher text verbatim.
type DB struct{}

func (DB) Kind() kind.Kind         { return kind.Neo4j }
func (DB) CanExpress(ir.Expr) bool { return true }
func (DB) CanAbsorb(ir.Node) bool  { return true }
func (DB) Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerCypher(q, types)
}
func (DB) CompileShipped(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerShipCypher(q, types)
}
func (DB) Stream(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	return streamGraph(ctx, reg, q, in, emit)
}
func (DB) RunShipped(ctx context.Context, reg *conn.Registry, q *stores.Query, s stores.Shipment) ([][]any, error) {
	return runShippedGraph(ctx, reg, q, s)
}
