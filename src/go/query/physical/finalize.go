package physical

// finalize.go is the requirement analysis: the columns each query returns, and segment output.

import (
	"polystore_database/src/go/query/ir"
	"sort"

	"polystore_database/src/go/base/codec"

	"polystore_database/src/go/base/kind"
)

func filterRefs(f *ir.Filter) []string { return ir.Refs(f) }

// complexRef reports whether cols include a non-scalar (list / json) property column. Predicates on such
// columns are evaluated only in Neo4j (stores holding JSON strings compare differently, and = on a MongoDB array means "contains").
func (pl *planner) complexRef(cols []string) bool {
	for _, c := range cols {
		if codec.IsComplexType(pl.exprType(&ir.Column{Key: c}, nil)) {
			return true
		}
	}
	return false
}

func availAll(o *openQ, cols []string) bool {
	for _, c := range cols {
		if !o.avail[c] {
			return false
		}
	}
	return true
}

func identAll(o *openQ, cols []string) bool {
	for _, c := range cols {
		if !o.idents[c] {
			return false
		}
	}
	return true
}

func anyKey(o *openQ, cols []string) bool {
	for _, c := range cols {
		for _, k := range o.keys {
			if c == k {
				return true
			}
		}
	}
	return false
}

// produce records the columns op adds to the query's rows.
func produce(o *openQ, op ir.Node) {
	switch x := op.(type) {
	case *ir.GetProperty:
		o.avail[x.Key()] = true
	case *ir.GetListProperty:
		o.avail[x.Key()] = true
	case *ir.Expand:
		o.avail[x.Rel], o.idents[x.Rel] = true, true
		o.avail[x.To], o.idents[x.To] = true, true
	case *ir.ExpandInto:
		o.avail[x.Rel], o.idents[x.Rel] = true, true
	case *ir.VarLengthExpand:
		o.avail[x.To], o.idents[x.To] = true, true
		o.avail[ir.SegColumn(x.Seg)] = true
	case *ir.ProjectPath:
		o.avail[x.Path] = true
	case *ir.Projection:
		resetTo(o, x.Items)
	case *ir.Aggregation:
		o.grouped = true
		items := append([]ir.ProjItem{}, x.Keys...)
		for _, a := range x.Aggs {
			items = append(items, ir.ProjItem{Name: a.Name})
		}
		resetTo(o, items)
	}
}

func resetTo(o *openQ, items []ir.ProjItem) {
	avail, idents := map[string]bool{}, map[string]bool{}
	for _, it := range items {
		avail[it.Name] = true
		if c, ok := it.Expr.(*ir.Column); ok && o.idents[c.Key] {
			idents[it.Name] = true
		}
	}
	o.avail, o.idents = avail, idents
}

// finalize passes "columns needed above" downwards and sets StoreQuery.Outputs and VarLengthExpand.EmitPath.
// all means the root: every output column is needed.
func (pl *planner) finalize(n Node, need map[string]bool, all bool) {
	switch x := n.(type) {
	case *StoreQuery:
		cols := ir.TopCols(x.Ops)
		var out []string
		if all || x.ShipMode == ir.ShipAggregate {
			// A shipped aggregation is not joined to input rows: it returns the aggregation's columns (keys and partial aggregates), not keys.
			out = cols
		} else if x.Replace {
			for _, c := range cols {
				if need[c] {
					out = append(out, c)
				}
			}
			sort.Strings(out)
		} else {
			out = append(out, x.Keys...)
			var rest []string
			for _, c := range cols {
				if need[c] && !contains(x.Keys, c) {
					rest = append(rest, c)
				}
			}
			sort.Strings(rest)
			out = append(out, rest...)
		}
		x.Outputs = out
		markEmit(x.Ops, out)
		x.Weighted = pl.opt.WeightedRows && weighable(x)
		if x.Input != nil {
			below := map[string]bool{}
			for k := range need {
				if !contains(cols, k) {
					below[k] = true
				}
			}
			for _, k := range x.Keys {
				below[k] = true
			}
			for _, k := range x.Ship {
				below[k] = true
			}
			if x.ShipMode == ir.ShipAggregate || x.Replace {
				// The aggregation consumes the input, so below only the columns it reads (keys, shipped columns) and the weight are needed.
				below = map[string]bool{}
				for _, k := range append(append([]string{}, x.Keys...), x.Ship...) {
					below[k] = true
				}
			}
			pl.finalize(x.Input, below, all && x.ShipMode != ir.ShipAggregate && !x.Replace)
		}
	case *OptionalApply:
		// Remember the columns the inner side makes (null for rows without a match); ask the outer side for its
		// columns and Keys, the inner side for the columns used above and Keys.
		x.InnerCols = innerCols(x.Inner)
		outerNeed := map[string]bool{}
		innerNeed := map[string]bool{}
		for k := range need {
			if contains(x.InnerCols, k) {
				innerNeed[k] = true
			} else {
				outerNeed[k] = true
			}
		}
		for _, k := range x.Keys {
			outerNeed[k] = true
			innerNeed[k] = true
		}
		pl.finalize(x.Outer, outerNeed, all)
		pl.finalize(x.Inner, innerNeed, false)
	case *ListFetch:
		below := map[string]bool{x.Source: true}
		for k := range need {
			if k != x.Key() {
				below[k] = true
			}
		}
		pl.finalize(x.Input, below, all)
	case *StepExpand:
		// Its hop query has fixed outputs (not finalized). The segment is made only when needed above; otherwise
		// paths to the same end collapse into weighted rows.
		seg := ir.SegColumn(x.Op.Seg)
		x.EmitSeg = all || need[seg]
		x.Weighted = pl.opt.WeightedRows && !x.EmitSeg
		below := map[string]bool{x.Op.From: true}
		for k := range need {
			if k != x.Op.To && k != seg {
				below[k] = true
			}
		}
		pl.finalize(x.Input, below, all)
	case *Coord:
		below := map[string]bool{}
		belowAll := all
		switch x.Op.(type) {
		case *ir.Projection, *ir.Aggregation:
			belowAll = false
		default:
			for k := range need {
				below[k] = true
			}
		}
		for _, k := range ir.Refs(x.Op) {
			below[k] = true
		}
		for _, in := range x.Inputs {
			pl.finalize(in, below, belowAll)
		}
	}
}

// innerCols returns the columns an OptionalApply's inner side makes (StoreQuery top columns and Coord outputs, minus ArgRows columns).
func innerCols(n Node) []string {
	var out []string
	add := func(ks ...string) {
		for _, k := range ks {
			if !contains(out, k) {
				out = append(out, k)
			}
		}
	}
	var walk func(Node)
	walk = func(n Node) {
		switch x := n.(type) {
		case *StoreQuery:
			for _, c := range ir.TopCols(x.Ops) {
				if !contains(x.Keys, c) {
					add(c)
				}
			}
		case *ListFetch:
			add(x.Key())
		case *StepExpand:
			add(x.Op.To, ir.SegColumn(x.Op.Seg))
		case *Coord:
			if pp, ok := x.Op.(*ir.ProjectPath); ok {
				add(pp.Path)
			}
		case *OptionalApply:
			add(x.New...)
			add(innerCols(x.Inner)...)
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

// weighable reports whether a StoreQuery's results can be collapsed into weighted rows: Neo4j, a
// variable-length expansion (per-path duplicates), no segment or path returned, and no absorbed operator that reorders or rebuilds rows.
func weighable(q *StoreQuery) bool {
	if q.Store != kind.Neo4j || q.ShipMode != ir.ShipNone {
		return false
	}
	varLen, ok := false, true
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *ir.VarLengthExpand:
			varLen = true
			if contains(q.Outputs, ir.SegColumn(x.Seg)) {
				ok = false
			}
		case *ir.ProjectPath:
			if contains(q.Outputs, x.Path) {
				ok = false
			}
		case *ir.Projection, *ir.Aggregation, *ir.Sort, *ir.Limit, *ir.Distinct:
			ok = false
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(q.Ops)
	return varLen && ok
}

// markEmit sets EmitPath on VarLengthExpand whose segment column is returned.
func markEmit(n ir.Node, out []string) {
	if n == nil {
		return
	}
	if v, ok := n.(*ir.VarLengthExpand); ok && contains(out, ir.SegColumn(v.Seg)) {
		v.EmitPath = true
	}
	for _, c := range n.Children() {
		markEmit(c, out)
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
