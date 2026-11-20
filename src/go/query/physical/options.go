package physical

import (
	"fmt"

	"polystore_database/src/go/base/kind"
)

// Options are the physical planning switches. The zero value disables every optional rewrite.
type Options struct {
	// DeferOffGraphPredicates moves predicates on values outside the graph after the traversal.
	DeferOffGraphPredicates bool
	// DeferPredicates groups deferred fetches by store (matches logical.Options.DeferPredicates).
	DeferPredicates bool
	// AbsorbDistinctOverOneRow lets a graph query absorb DISTINCT when its input is one row.
	AbsorbDistinctOverOneRow bool
	// ShipRows passes partial aggregations to the next store as a table instead of an id list.
	ShipRows bool
	// ShipStores are the stores that may receive a shipped table.
	ShipStores map[kind.Kind]bool
	// WeightedRows lets graph queries with variable-length expansions return one weighted row per distinct output.
	WeightedRows bool
	// UniqueKeys are property names whose equality with a literal selects at most one entity.
	UniqueKeys []string
	// PathPredicates is how conditions on the elements inside a path are evaluated when their values are not in Neo4j.
	PathPredicates PathStrategy
}

// PathStrategy is how conditions on path elements (all / none / any over nodes(p), segment conditions) whose
// values are outside Neo4j are evaluated. A shape a strategy does not cover falls back to PathPost.
type PathStrategy int

const (
	// PathPost returns the paths from Neo4j, fetches the element values from their stores and evaluates the
	// condition in the coordinator (after the whole traversal).
	PathPost PathStrategy = iota
	// PathPrefilter evaluates the condition over all entities of the element labels in their stores first
	// (set plans) and passes the ids that satisfy it to the Neo4j traversal (`x.uuid IN $set`). Node
	// conditions in all / none over a path or segment and in any over a path; not single, not relationships.
	PathPrefilter
	// PathStepwise runs a variable-length expansion hop by hop from the coordinator and drops partial paths
	// whose new relationship or inner node fails the segment condition before expanding further (StepExpand).
	// Bounded ranges with a minimum of at least 1; any / single fall back.
	PathStepwise
)

var pathStrategyNames = [...]string{"post", "prefilter", "stepwise"}

func (s PathStrategy) String() string {
	if int(s) < len(pathStrategyNames) {
		return pathStrategyNames[s]
	}
	return fmt.Sprintf("PathStrategy(%d)", int(s))
}

// ParsePathStrategy reads a strategy name (post / prefilter / stepwise).
func ParsePathStrategy(s string) (PathStrategy, error) {
	for i, n := range pathStrategyNames {
		if n == s {
			return PathStrategy(i), nil
		}
	}
	return 0, fmt.Errorf("unknown path predicate strategy %q (post / prefilter / stepwise)", s)
}
