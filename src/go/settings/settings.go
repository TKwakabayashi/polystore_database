// Package settings holds the default values of internal behavior, switched in source code (edit and rebuild),
// not with CLI flags. Only entry points and experiment harnesses read it; Plan and Exec (options.go) turn the
// defaults into the planner and executor options that the query engine receives as arguments.
package settings

import "polystore_database/src/go/query/exec"

// ===== Execution =====
var (
	Engine     = exec.ModelStream // iteration model of single query runs
	VectorSize = 1024             // batch width of vectorized, stream and vecstream
)

// MaterializeChunkSize is the wanted number of ids per store query (0: the store's default). The executor
// clamps it to min(this or the store default, the store's limit store.LimitsOf(k).MaxInList).
var MaterializeChunkSize = 0

// CostBasedPlanning chooses, when statistics exist (stats.json next to the mapping; `admin stats`), the start,
// split and edge order of pattern components by estimated cost (sum of intermediate rows). Without statistics or
// when false, rules decide (unique-key equality > equality and IN > number of predicates, then pattern order).
var CostBasedPlanning = true

// SplitAnchoredPatterns: when a connected pattern has two or more anchors (bound earlier, or unique-key
// equality), assign vertices to the nearest anchor, expand from each and NodeHashJoin on the shared vertices
// (search from both sides, meet in the middle). An all-Neo4j pattern stays one Cypher query.
var SplitAnchoredPatterns = true

// DeferPredicates moves fetches (GetProperty) and filters (Filter), except those on the start, after all
// pattern expansions (logical.DeferPredicates). Neo4j then runs the traversal alone in one query and the
// conditions on interior nodes, endpoints and edges apply to its result. The physical planner also splits less:
// a `WITH DISTINCT <ids>` is absorbed into the traversal even with several starts (duplicates across input
// batches are removed by the coordinator), and moved fetches are grouped per store (Neo4j first).
// false (default) applies them as soon as their variables are bound (fewer rows early). Results are the same.
var DeferPredicates = false

// DeferOffGraphPredicates moves fetches and filters on columns not in Neo4j after the pattern expansions
// (physical.Build; filters on the start stay). The Neo4j traversal then stays one query instead of splitting at
// non-Neo4j conditions; Neo4j conditions apply inside it. false applies them when bound (fewer rows, split traversal). Results are the same.
var DeferOffGraphPredicates = true

// AbsorbDistinctOverOneRow absorbs `WITH DISTINCT <ids>` into a Neo4j query whose input is one row by unique-key
// equality (physical.Build), so the traversal does not split at WITH DISTINCT even when the start is outside Neo4j.
var AbsorbDistinctOverOneRow = true

// SplitAggregation splits an aggregation into a partial aggregation per key owner id, the fetch of key values,
// and a final aggregation (logical.SplitAggregation; eager aggregation). Key values are fetched once per owner,
// and with ShipRows partial aggregations can be absorbed into the stores holding values. Not for DISTINCT or collect. Results are the same.
var SplitAggregation = false

// ShipRows absorbs partial aggregations into queries on the stores holding the values and passes the boundary
// as a table of id tuples with weights (ir.ShipAggregate). false keeps the plain IN on one id column.
var ShipRows = false

// ShipStores lists the stores that receive tables with ShipRows (store.Kind name -> on). Each receives differently:
// - MySQL: JSON_TABLE with a left outer join and GROUP BY (aggregated in the store)
// - Neo4j: MATCH on the uuids of UNWIND rows, weighted aggregation (in the store)
// - MongoDB: $documents -> $lookup -> $group. MongoDB 7 has no hash join with a passed table; $lookup re-reads
// the index per row (about 340 ms for 20k rows vs about 90 ms to fetch values with $in), so it is off by default
// - ClickHouse: no way to receive a table yet; the executor emulates it
// - Cassandra / kvs: no way to join a table; the executor emulates it (fetch values with IN on the keys, join
// to the table, aggregate). The store's work equals a plain value fetch
var ShipStores = map[string]bool{
	"relational": true,
	"graph":      true,
	"document":   false,
	"columnar":   true,
	"widecolumn": true,
	"leveldb":    true,
	"rocksdb":    true,
}

// ShipChunkRows is the number of table rows per shipped query (the executor splits by it).
var ShipChunkRows = 20000

// StoreFetchParallelism is the number of concurrent chunks when an operator splits an IN list for one store.
// 1 means serial.
var StoreFetchParallelism = 4

// WeightedRows makes Neo4j queries with variable-length expansions return one row per distinct output with its
// count (weight), curbing per-path duplicate rows (Q11, ...). Results are the same.
var WeightedRows = true

// PathPredicates is how conditions on the elements inside a path (all / none / any over nodes(p) or
// relationships(p)) are evaluated when their values are outside Neo4j. Results are the same:
// - "post": the traversal returns the paths, element values are fetched from their stores and the coordinator filters
// - "prefilter": the condition runs first over all entities of the element labels in their stores, and the ids that
// pass go to the Neo4j traversal as `x.uuid IN $set` (Neo4j prunes while expanding). Nodes only; not single
// - "stepwise": the variable-length expansion runs hop by hop from the coordinator, and partial paths whose new
// relationship or inner node fails are dropped before the next hop. Bounded ranges from 1; all / none only
// Shapes a strategy does not cover use "post".
var PathPredicates = "post"

// GraphFetchParallelism is the number of concurrent chunks to Neo4j. A Neo4j chunk may run variable-length
// traversals from 50k starts, and in parallel Neo4j ran out of memory (BI9 on ldbc_sf1 was OOM-killed with
// heap 4G / page cache 1G), so the default is serial.
var GraphFetchParallelism = 1

// KVBatchKeys is the number of ids a key-value store reads per batch: the keys of their properties go in one
// MultiGet on RocksDB (LevelDB has none and reads key by key). 0 reads key by key.
var KVBatchKeys = 1024

// KVFetchParallelism is the number of batches a key-value store reads at once inside one store query (a scan,
// or one chunk of ids; chunks themselves run StoreFetchParallelism at a time). 1 means serial.
var KVFetchParallelism = 4

// ===== Migration =====
// MigrationDeleteSource deletes the source after a successful migration (lab workflow).
var MigrationDeleteSource = true
