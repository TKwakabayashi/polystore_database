// Package testopts holds the planner options that tests and goldens are pinned to.
// They equal the defaults in settings when the goldens were last regenerated; editing settings
// does not change test results.
package testopts

import (
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/logical"
	"polystore_database/src/go/query/physical"
)

// UniqueKeys are the unique-key property names the goldens assume.
var UniqueKeys = []string{"id"}

// Logical returns the logical planning options of the goldens.
func Logical() logical.Options {
	return logical.Options{
		SplitAnchoredPatterns: true,
		UniqueKeys:            UniqueKeys,
	}
}

// Physical returns the physical planning options of the goldens.
func Physical() physical.Options {
	return physical.Options{
		DeferOffGraphPredicates:  true,
		AbsorbDistinctOverOneRow: true,
		WeightedRows:             true,
		ShipStores: map[kind.Kind]bool{kind.MySQL: true, kind.Neo4j: true,
			kind.ClickHouse: true, kind.Cassandra: true, kind.LevelDB: true, kind.RocksDB: true},
		UniqueKeys: UniqueKeys,
	}
}

// AllShipStores lets every store receive a shipped table.
func AllShipStores() map[kind.Kind]bool {
	return map[kind.Kind]bool{kind.Neo4j: true, kind.MySQL: true, kind.Mongo: true,
		kind.ClickHouse: true, kind.Cassandra: true, kind.LevelDB: true, kind.RocksDB: true}
}
