package physical

// absorb.go decides whether a leaf store query can take a distinct over rows it already returns.

import (
	"polystore_database/src/go/query/ir"

	"polystore_database/src/go/base/kind"
)

// absorbDistinct reports whether, with Options.DeferPredicates, a `WITH DISTINCT <ids>` can be absorbed as is
// into a Neo4j query that has input (called for the Projection and the Distinct after it). The input columns
// are projected away, so results are not joined back. Inputs may come in batches, so the coordinator removes
// duplicates across batches on close (not needed for a one-row input). Not inside an OPTIONAL, whose rows must match outer rows.
func (pl *planner) absorbDistinct(o *openQ, op ir.Node) bool {
	if !pl.opt.DeferPredicates || o == nil || o.store != kind.Neo4j || o.leaf || o.mode != ir.ShipNone || len(pl.args) > 0 {
		return false
	}
	switch x := op.(type) {
	case *ir.Distinct:
		return o.distinctOK && o.replace && len(x.Cols) == 0
	case *ir.Projection:
		if o.distinctOK {
			return false
		}
		if d, ok := pl.parent[x].(*ir.Distinct); !ok || len(d.Cols) > 0 {
			return false
		}
		var cols []string
		for _, it := range x.Items {
			c, ok := it.Expr.(*ir.Column)
			if !ok || c.Key != it.Name || !o.idents[c.Key] {
				return false
			}
			cols = append(cols, c.Key)
		}
		o.distinctOK, o.replace = true, true
		o.redup, o.dedup = !pl.oneRow(o), cols
		return true
	}
	return false
}

// redupSafe reports whether operators after an absorbed Distinct can join the same query and rows stay
// identifiable for the final deduplication: an expanded row is determined by the Distinct columns plus its
// relationships and nodes, and its values are functions of them. Variable-length expansions (a row per path)
// and operators that rebuild rows go to another query.
func redupSafe(op ir.Node) bool {
	switch op.(type) {
	case *ir.Distinct, *ir.Expand, *ir.ExpandInto, *ir.Filter, *ir.GetProperty, *ir.RelUniqueFilter:
		return true
	}
	return false
}

// distinctOverOneRow returns the operator to absorb when a `WITH DISTINCT <ids>` can go into a Neo4j query
// that has input (nil otherwise). Such queries are usually chunked, so a global distinct needs the
// coordinator and splits the traversal there. With at most one input row there is one chunk, so
// removing duplicates inside the query is the global distinct. The key column is kept to join results back
// (one key value, so duplicates are judged the same).
func (pl *planner) distinctOverOneRow(o *openQ, op ir.Node) ir.Node {
	if !pl.opt.AbsorbDistinctOverOneRow || o == nil || o.store != kind.Neo4j || o.leaf || o.mode != ir.ShipNone || !pl.oneRow(o) {
		return nil
	}
	switch x := op.(type) {
	case *ir.Distinct:
		if o.distinctOK && len(x.Cols) == 0 {
			return x
		}
	case *ir.Projection:
		if _, ok := pl.parent[x].(*ir.Distinct); !ok {
			return nil
		}
		items := append([]ir.ProjItem{}, x.Items...)
		for _, it := range x.Items {
			c, ok := it.Expr.(*ir.Column)
			if !ok || c.Key != it.Name || !o.idents[c.Key] {
				return nil
			}
		}
		for _, k := range o.keys {
			if indexOfItem(items, k) < 0 {
				items = append(items, ir.ProjItem{Name: k, Expr: &ir.Column{Key: k}})
			}
		}
		o.distinctOK = true
		return &ir.Projection{Input: x.Input, Items: items}
	}
	return nil
}

func indexOfItem(items []ir.ProjItem, name string) int {
	for i, it := range items {
		if it.Name == name {
			return i
		}
	}
	return -1
}

// oneRow reports whether o's input is known to be at most one row: the query receives the key of a scan
// filtered by unique-key equality. Uniqueness is taken from Options.UniqueKeys (duplicates against it make the result per input).
func (pl *planner) oneRow(o *openQ) bool {
	in, ok := o.input.(*StoreQuery)
	if !ok || in.Input != nil || len(o.keys) != 1 {
		return false
	}
	unique := false
	for n := in.Ops; n != nil; {
		switch x := n.(type) {
		case *ir.EntityScan:
			return unique && x.Alias == o.keys[0]
		case *ir.GetProperty:
			n = x.Input
		case *ir.Filter:
			if b, ok := x.Pred.(*ir.Bin); ok && b.Op == "=" {
				c, cok := b.L.(*ir.Column)
				_, lok := b.R.(*ir.Lit)
				if cok && lok {
					for _, k := range pl.opt.UniqueKeys {
						unique = unique || c.Key == o.keys[0]+"."+k
					}
				}
			}
			n = x.Input
		default:
			return false
		}
	}
	return false
}
