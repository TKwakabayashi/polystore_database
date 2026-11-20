package physical

// typing.go types literals by the property types of the columns they are compared with.

import (
	"polystore_database/src/go/query/ir"
	"strings"

	"polystore_database/src/go/base/codec"
)

// typeLiterals converts a literal to the type of the column it is compared with (the mapping's type), at
// planning time, so store-side and coordinator-side expressions compare the same typed values.
func (pl *planner) typeLiterals(op ir.Node) {
	switch x := op.(type) {
	case *ir.Filter:
		x.Pred = pl.typeExpr(x.Pred, nil)
	case *ir.Projection:
		for i := range x.Items {
			x.Items[i].Expr = pl.typeExpr(x.Items[i].Expr, nil)
		}
	case *ir.VarLengthExpand:
		for i, sp := range x.SegPreds {
			el := map[string]owner{sp.Var: pl.segElemOwner(x.Types, sp.Part)}
			x.SegPreds[i].Pred = pl.typeExpr(sp.Pred, el)
		}
	}
}

func (pl *planner) segElemOwner(types []string, part ir.SegPart) owner {
	if part == ir.SegRels {
		return owner{kind: ir.RelObj, labels: types}
	}
	return owner{kind: ir.NodeObj, labels: pl.cat.ElemLabels(types, part)}
}

// exprType returns the type name of a column or element reference ("" if unknown).
func (pl *planner) exprType(e ir.Expr, elems map[string]owner) string {
	switch x := e.(type) {
	case *ir.Column:
		i := strings.LastIndex(x.Key, ".")
		if i < 0 {
			return ""
		}
		o, ok := pl.owners[x.Key[:i]]
		if !ok {
			return ""
		}
		return pl.cat.TypeOf(o.kind, o.labels, x.Key[i+1:])
	case *ir.ElemProp:
		o, ok := elems[x.Var]
		if !ok {
			return ""
		}
		return pl.cat.TypeOf(o.kind, o.labels, x.Prop)
	}
	return ""
}

func (pl *planner) typeExpr(e ir.Expr, elems map[string]owner) ir.Expr {
	conv := func(v ir.Expr, typ string) ir.Expr {
		switch l := v.(type) {
		case *ir.Lit:
			if typ == "" || l.Val == nil {
				return l
			}
			if t, err := codec.ConvertToNativeType(l.Val, typ); err == nil {
				return &ir.Lit{Val: t}
			}
		case *ir.List:
			out := &ir.List{}
			for _, el := range l.Elems {
				if lit, ok := el.(*ir.Lit); ok && typ != "" && lit.Val != nil {
					if t, err := codec.ConvertToNativeType(lit.Val, typ); err == nil {
						out.Elems = append(out.Elems, &ir.Lit{Val: t})
						continue
					}
				}
				out.Elems = append(out.Elems, el)
			}
			return out
		}
		return v
	}
	switch x := e.(type) {
	case *ir.Bin:
		switch x.Op {
		case "=", "<>", "<", ">", "<=", ">=", "IN":
			l, r := x.L, x.R
			if t := pl.exprType(x.L, elems); t != "" {
				r = conv(r, t)
			}
			if t := pl.exprType(x.R, elems); t != "" && x.Op != "IN" {
				l = conv(l, t)
			}
			return &ir.Bin{Op: x.Op, L: l, R: r}
		}
		return &ir.Bin{Op: x.Op, L: pl.typeExpr(x.L, elems), R: pl.typeExpr(x.R, elems)}
	case *ir.Un:
		return &ir.Un{Op: x.Op, X: pl.typeExpr(x.X, elems)}
	case *ir.Case:
		c := &ir.Case{}
		for _, w := range x.Whens {
			c.Whens = append(c.Whens, ir.CaseWhen{When: pl.typeExpr(w.When, elems), Then: w.Then})
		}
		c.Else = x.Else
		return c
	case *ir.Quant:
		inner := map[string]owner{}
		for k, v := range elems {
			inner[k] = v
		}
		if call, ok := x.List.(*ir.Call); ok && len(call.Args) == 1 {
			if src, ok := call.Args[0].(*ir.Column); ok {
				part := ir.SegNodes
				kind := ir.NodeObj
				if call.Fn == "relationships" {
					part, kind = ir.SegRels, ir.RelObj
				}
				inner[x.Var] = owner{kind: kind, labels: pl.listLabels(src.Key, part)}
			}
		}
		return &ir.Quant{Kind: x.Kind, Var: x.Var, List: x.List, Pred: pl.typeExpr(x.Pred, inner)}
	}
	return e
}
