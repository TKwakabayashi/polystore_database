package settings

// options.go turns the defaults into the planner and executor options the query engine takes as arguments.

import (
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/logical"
	"polystore_database/src/go/query/physical"
)

// Plan returns the planning options from the defaults. An unknown PathPredicates name panics (a source edit error).
func Plan() query.PlanOptions {
	ship := map[kind.Kind]bool{}
	for name, on := range ShipStores {
		if k, err := kind.ParseKind(name); err == nil && on {
			ship[k] = true
		}
	}
	paths, err := physical.ParsePathStrategy(PathPredicates)
	if err != nil {
		panic("settings.PathPredicates: " + err.Error())
	}
	return query.PlanOptions{
		CostBased: CostBasedPlanning,
		Logical: logical.Options{
			SplitAnchoredPatterns: SplitAnchoredPatterns,
			DeferPredicates:       DeferPredicates,
			SplitAggregation:      SplitAggregation,
		},
		Physical: physical.Options{
			DeferOffGraphPredicates:  DeferOffGraphPredicates,
			DeferPredicates:          DeferPredicates,
			AbsorbDistinctOverOneRow: AbsorbDistinctOverOneRow,
			ShipRows:                 ShipRows,
			ShipStores:               ship,
			WeightedRows:             WeightedRows,
			PathPredicates:           paths,
		},
	}
}

// Exec returns the execution options from settings for the given iteration model.
func Exec(model exec.Model) exec.ExecOptions {
	return exec.ExecOptions{
		Model:                 model,
		VectorSize:            VectorSize,
		MaterializeChunkSize:  MaterializeChunkSize,
		ShipChunkRows:         ShipChunkRows,
		StoreFetchParallelism: StoreFetchParallelism,
		GraphFetchParallelism: GraphFetchParallelism,
		KVBatchKeys:           KVBatchKeys,
		KVFetchParallelism:    KVFetchParallelism,
	}
}
