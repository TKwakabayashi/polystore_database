package physical

// ship.go decides when an input is shipped to a store as a table (ShipMode) instead of as ids.

import (
	"polystore_database/src/go/query/ir"
	"strings"

	"polystore_database/src/go/base/kind"
)

// canShipAggregate reports whether partial aggregation agg can be absorbed into open non-Neo4j query o
// (a value fetch receiving one id column, o.single) as a shipped boundary (ir.ShipAggregate), and returns the
// input id columns to ship.
// - Partial aggregations only: the final aggregation combines per-chunk results, so inputs may be split.
// - Keys are o.single's id column or its values, or input id columns (which get shipped).
// - Arguments are o.single's id column or its values, or count(*); functions count / sum / min / max without DISTINCT.
func (pl *planner) canShipAggregate(o *openQ, agg *ir.Aggregation) ([]string, bool) {
	if !agg.Partial || o == nil || o.leaf || o.grouped || o.mode != ir.ShipNone || !pl.opt.canShip(o.store, ir.ShipAggregate) {
		return nil, false
	}
	if o.store == kind.Neo4j {
		return pl.canShipGraph(o, agg)
	}
	if o.single == "" || len(o.keys) != 1 || o.keys[0] != o.single {
		return nil, false
	}
	own := func(c string) bool { return c == o.single || (o.avail[c] && strings.HasPrefix(c, o.single+".")) }
	var ship []string
	for _, k := range agg.Keys {
		c, ok := k.Expr.(*ir.Column)
		if !ok {
			return nil, false
		}
		switch {
		case own(c.Key):
		case pl.vars[c.Key]:
			if !contains(ship, c.Key) {
				ship = append(ship, c.Key)
			}
		default:
			return nil, false
		}
	}
	for _, a := range agg.Aggs {
		switch a.Fn {
		case "count", "sum", "min", "max":
		default:
			return nil, false
		}
		if a.Distinct {
			return nil, false
		}
		if a.Arg == nil {
			continue
		}
		c, ok := a.Arg.(*ir.Column)
		if !ok || !own(c.Key) {
			return nil, false
		}
	}
	return ship, true
}

// canShipGraph reports whether a partial aggregation can be absorbed into a Neo4j query with input
// (boundary ids -> traversal and fetches). Keys are columns of the query's rows or input id columns
// (shipped); arguments are columns of the query's rows or count(*).
func (pl *planner) canShipGraph(o *openQ, agg *ir.Aggregation) ([]string, bool) {
	var ship []string
	for _, k := range agg.Keys {
		c, ok := k.Expr.(*ir.Column)
		if !ok {
			return nil, false
		}
		switch {
		case o.avail[c.Key]:
		case pl.vars[c.Key]:
			if !contains(ship, c.Key) {
				ship = append(ship, c.Key)
			}
		default:
			return nil, false
		}
	}
	for _, a := range agg.Aggs {
		switch a.Fn {
		case "count", "sum", "min", "max":
		default:
			return nil, false
		}
		if a.Distinct {
			return nil, false
		}
		if a.Arg == nil {
			continue
		}
		c, ok := a.Arg.(*ir.Column)
		if !ok || !o.avail[c.Key] {
			return nil, false
		}
	}
	return ship, true
}

// ship turns the open query into a shipped boundary (writes the shipped columns in the leaf ir.Boundary) and adds op.
func (pl *planner) ship(st *state, op ir.Node, mode ir.ShipMode, ship []string) {
	o := st.open
	for n := o.ops; n != nil; n = ir.Child(n) {
		if a, ok := n.(*ir.Boundary); ok {
			cp := *a
			cp.Ship = ship
			cp.Rid = mode == ir.ShipFilter || mode == ir.ShipProject
			o.ops = replaceLeaf(o.ops, &cp)
			break
		}
	}
	o.ship, o.mode = ship, mode
	for _, c := range ship {
		o.avail[c], o.idents[c] = true, true
	}
	pl.extend(st, op)
}

// replaceLeaf copies chain n with its leaf (an ir.Boundary) replaced by leaf.
func replaceLeaf(n ir.Node, leaf *ir.Boundary) ir.Node {
	if _, ok := n.(*ir.Boundary); ok {
		return leaf
	}
	return ir.WithInput(n, replaceLeaf(ir.Child(n), leaf))
}

// collectsIdent reports whether op collects id columns. Lists of ids are not returned by stores (they do
// not rebuild the elements); the coordinator collects them.
func collectsIdent(o *openQ, op ir.Node) bool {
	agg, ok := op.(*ir.Aggregation)
	if !ok {
		return false
	}
	for _, a := range agg.Aggs {
		if c, ok := a.Arg.(*ir.Column); ok && a.Fn == "collect" && o.idents[c.Key] {
			return true
		}
	}
	return false
}
