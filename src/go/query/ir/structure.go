package ir

// Refs returns the columns an operator reads from its input rows (id columns included).
func Refs(n Node) []string {
	var out []string
	add := func(ks ...string) { out = append(out, ks...) }
	switch x := n.(type) {
	case *GetProperty:
		add(x.Alias)
	case *GetListProperty:
		add(x.Source)
	case *Expand:
		add(x.From)
	case *ExpandInto:
		add(x.From, x.To)
	case *VarLengthExpand:
		add(x.From)
	case *Filter:
		add(Columns(x.Pred)...)
		addQuantLists(x.Pred, &out)
	case *Projection:
		for _, it := range x.Items {
			add(Columns(it.Expr)...)
			addQuantLists(it.Expr, &out)
		}
	case *ProjectPath:
		for _, e := range x.Elems {
			if e.Seg {
				add(SegColumn(e.Var))
			} else {
				add(e.Var)
			}
		}
	case *Aggregation:
		for _, it := range x.Keys {
			add(Columns(it.Expr)...)
		}
		for _, a := range x.Aggs {
			if a.Arg != nil {
				add(Columns(a.Arg)...)
			}
		}
	case *Sort:
		for _, k := range x.Keys {
			add(Columns(k.Expr)...)
		}
	case *Distinct:
		add(x.Cols...)
	case *NodeHashJoin:
		add(x.Keys...)
	case *Optional:
		add(x.Keys...)
	case *Unwind:
		add(Columns(x.List)...)
	case *RelUniqueFilter:
		add(x.Rels...)
		for _, s := range x.Segs {
			add(SegColumn(s))
		}
	}
	return out
}

// addQuantLists adds the value list columns (GetListProperty outputs) that a quantifier's element references need.
func addQuantLists(e Expr, out *[]string) {
	Walk(e, func(x Expr) {
		q, ok := x.(*Quant)
		if !ok {
			return
		}
		call, ok := q.List.(*Call)
		if !ok || len(call.Args) != 1 {
			return
		}
		src, ok := call.Args[0].(*Column)
		if !ok {
			return
		}
		part := SegNodes
		if call.Fn == "relationships" {
			part = SegRels
		}
		Walk(q.Pred, func(y Expr) {
			if ep, ok := y.(*ElemProp); ok && ep.Var == q.Var {
				*out = append(*out, ListPropKey(src.Key, part, ep.Prop))
			}
		})
	})
}

// child / SetChild read and write a unary operator's input.
func child(n Node) Node {
	switch x := n.(type) {
	case *GetProperty:
		return x.Input
	case *GetListProperty:
		return x.Input
	case *Expand:
		return x.Input
	case *ExpandInto:
		return x.Input
	case *VarLengthExpand:
		return x.Input
	case *Filter:
		return x.Input
	case *Projection:
		return x.Input
	case *ProjectPath:
		return x.Input
	case *Aggregation:
		return x.Input
	case *Sort:
		return x.Input
	case *Limit:
		return x.Input
	case *Distinct:
		return x.Input
	case *RelUniqueFilter:
		return x.Input
	case *Unwind:
		return x.Input
	}
	return nil
}

// SetChild replaces the input of a unary operator (no-op for binary operators and leaves).
func SetChild(n, c Node) {
	switch x := n.(type) {
	case *GetProperty:
		x.Input = c
	case *GetListProperty:
		x.Input = c
	case *Expand:
		x.Input = c
	case *ExpandInto:
		x.Input = c
	case *VarLengthExpand:
		x.Input = c
	case *Filter:
		x.Input = c
	case *Projection:
		x.Input = c
	case *ProjectPath:
		x.Input = c
	case *Aggregation:
		x.Input = c
	case *Sort:
		x.Input = c
	case *Limit:
		x.Input = c
	case *Distinct:
		x.Input = c
	case *RelUniqueFilter:
		x.Input = c
	case *Unwind:
		x.Input = c
	}
}

// Child returns a unary operator's input (nil for binary operators and leaves).
func Child(n Node) Node { return child(n) }

// WithInput returns a shallow copy of unary operator n with its input replaced by c.
func WithInput(n Node, c Node) Node {
	var cp Node
	switch x := n.(type) {
	case *GetProperty:
		y := *x
		cp = &y
	case *GetListProperty:
		y := *x
		cp = &y
	case *Expand:
		y := *x
		cp = &y
	case *ExpandInto:
		y := *x
		cp = &y
	case *VarLengthExpand:
		y := *x
		cp = &y
	case *Filter:
		y := *x
		cp = &y
	case *Projection:
		y := *x
		cp = &y
	case *ProjectPath:
		y := *x
		cp = &y
	case *Aggregation:
		y := *x
		cp = &y
	case *Sort:
		y := *x
		cp = &y
	case *Limit:
		y := *x
		cp = &y
	case *Distinct:
		y := *x
		cp = &y
	case *RelUniqueFilter:
		y := *x
		cp = &y
	case *Unwind:
		y := *x
		cp = &y
	default:
		return n
	}
	SetChild(cp, c)
	return cp
}
