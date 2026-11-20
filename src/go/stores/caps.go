package stores

import "polystore_database/src/go/query/ir"

// ExprRules says which predicate forms a database evaluates natively. Comparisons, IN and AND
// over columns, literals and lists are always expressible.
type ExprRules struct {
	Or, Arith, IsNull, Not, OtherUnary, Calls bool
}

// Expressible reports whether every part of e is allowed by r.
func (r ExprRules) Expressible(e ir.Expr) bool {
	ok := true
	ir.Walk(e, func(x ir.Expr) {
		switch v := x.(type) {
		case *ir.Column, *ir.Lit, *ir.List:
		case *ir.Bin:
			switch v.Op {
			case "=", "<>", "<", ">", "<=", ">=", "IN", "AND":
			case "OR":
				ok = ok && r.Or
			default:
				ok = ok && r.Arith
			}
		case *ir.Un:
			switch v.Op {
			case "IS NULL", "IS NOT NULL":
				ok = ok && r.IsNull
			case "NOT":
				ok = ok && r.Not
			default:
				ok = ok && r.OtherUnary
			}
		default: // Call / Case / Quant / ElemProp
			ok = ok && r.Calls
		}
	})
	return ok
}

// AbsorbRules says which whole-row operators a leaf query on the database can run. Operator
// arguments must be plain columns.
type AbsorbRules struct {
	Projection         bool
	Aggregation        bool
	GroupedAggregation bool // aggregation with grouping keys
	DistinctAggregate  bool // count(DISTINCT x) and the like
	SortLimitDistinct  bool
}

// Absorbs reports whether op is allowed by r.
func (r AbsorbRules) Absorbs(op ir.Node) bool {
	switch x := op.(type) {
	case *ir.Projection:
		if !r.Projection {
			return false
		}
		for _, it := range x.Items {
			if !PlainCols(it.Expr) {
				return false
			}
		}
		return true
	case *ir.Aggregation:
		if !r.Aggregation || len(x.Keys) > 0 && !r.GroupedAggregation {
			return false
		}
		for _, it := range x.Keys {
			if !PlainCols(it.Expr) {
				return false
			}
		}
		for _, a := range x.Aggs {
			if a.Fn == "collect" || a.Distinct && !r.DistinctAggregate || !PlainCols(a.Arg) {
				return false
			}
		}
		return true
	case *ir.Sort:
		if !r.SortLimitDistinct {
			return false
		}
		for _, sk := range x.Keys {
			if !PlainCols(sk.Expr) {
				return false
			}
		}
		return true
	case *ir.Limit, *ir.Distinct:
		return r.SortLimitDistinct
	}
	return false
}
