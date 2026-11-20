// Package logical plans patterns (start, edge order, anchored splits) into a store-independent plan of ir operators and
// rewrites it (deferred fetches, predicate placement, aggregation split, pruning).
package logical

// Options are the logical planning switches. The zero value disables every optional rewrite.
type Options struct {
	// SplitAnchoredPatterns grows a pattern with two or more anchors from each anchor and joins them.
	SplitAnchoredPatterns bool
	// DeferPredicates moves non-anchor fetches and filters after the expansions.
	DeferPredicates bool
	// SplitAggregation splits an aggregation into partial aggregations per key owner and a final one.
	SplitAggregation bool
	// UniqueKeys are property names whose equality with a literal selects at most one entity.
	UniqueKeys []string
}

func (o Options) isUniqueKey(prop string) bool {
	for _, k := range o.UniqueKeys {
		if k == prop {
			return true
		}
	}
	return false
}
