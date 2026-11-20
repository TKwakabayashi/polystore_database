package cypher

// paths.go decomposes whole-path conditions and quantified path patterns into per-element segment conditions.

import (
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

func (b *builder) decomposePathPreds(pending []ast.Expr, segRange map[string]ir.HopRange, segPreds map[string][]ir.SegPred) ([]ast.Expr, error) {
	var out []ast.Expr
	for _, e := range pending {
		repl, handled, err := b.decomposeOne(e, segRange, segPreds)
		if err != nil {
			return nil, err
		}
		if handled {
			out = append(out, repl...)
		} else {
			out = append(out, e)
		}
	}
	return out, nil
}

func (b *builder) decomposeOne(e ast.Expr, segRange map[string]ir.HopRange, segPreds map[string][]ir.SegPred) ([]ast.Expr, bool, error) {
	// length(p) θ k
	if bin, ok := e.(*ast.Bin); ok {
		if c, ok := bin.L.(*ast.Call); ok && c.Fn == "length" {
			p, perr := b.pathArg(c)
			lit, lok := bin.R.(*ast.Lit)
			if perr != nil || !lok {
				return nil, false, nil
			}
			k, kok := lit.Val.(int64)
			if !kok {
				return nil, false, nil
			}
			pd := b.paths[p]
			fixed := len(pd.rels)
			switch len(pd.segs) {
			case 0: // the length is constant
				if cmpInt(int64(fixed), bin.Op, k) {
					return nil, true, nil
				}
				return []ast.Expr{&ast.Lit{Val: false}}, true, nil
			case 1:
				seg := pd.segs[0]
				r, ok := segRange[seg].Constrain(bin.Op, int(k)-fixed)
				if !ok {
					return nil, false, nil
				}
				segRange[seg] = r
				return nil, true, nil
			}
			return nil, false, nil
		}
	}
	// all / none(x IN nodes(p)|relationships(p) WHERE c(x))
	q, ok := e.(*ast.Quant)
	if !ok || (q.Kind != "all" && q.Kind != "none") {
		return nil, false, nil
	}
	src, err := b.quantSource(q)
	if err != nil {
		return nil, false, err
	}
	if !onlyRefersTo(q.Pred, q.Var) {
		return nil, false, nil
	}
	pred := q.Pred
	if q.Kind == "none" {
		pred = &ast.Un{Op: "NOT", X: pred}
	}
	pd := b.paths[src.path]
	var repl []ast.Expr
	elems := pd.nodes
	if src.part == ir.SegRels {
		elems = pd.rels
	}
	for _, v := range elems {
		repl = append(repl, substituteVar(pred, q.Var, v))
	}
	for _, seg := range pd.segs {
		lp, err := b.conv(pred, convCtx{quant: map[string]quantSrc{q.Var: src}})
		if err != nil {
			return nil, false, err
		}
		segPreds[seg] = append(segPreds[seg], ir.SegPred{Part: src.part, Var: q.Var, Pred: lp})
	}
	return repl, true, nil
}

func cmpInt(a int64, op string, k int64) bool {
	switch op {
	case "=":
		return a == k
	case "<>":
		return a != k
	case "<":
		return a < k
	case "<=":
		return a <= k
	case ">":
		return a > k
	case ">=":
		return a >= k
	}
	return false
}

// onlyRefersTo reports whether an expression references nothing but v (and literals); no nested quantifiers.
func onlyRefersTo(e ast.Expr, v string) bool {
	ok := true
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.VarRef:
			ok = false
		case *ast.PropRef:
			if x.Var != v {
				ok = false
			}
		case *ast.Bin:
			walk(x.L)
			walk(x.R)
		case *ast.Un:
			walk(x.X)
		case *ast.Call:
			for _, a := range x.Args {
				walk(a)
			}
		case *ast.List:
			for _, el := range x.Elems {
				walk(el)
			}
		case *ast.Case:
			for _, w := range x.Whens {
				walk(w.When)
				walk(w.Then)
			}
			if x.Else != nil {
				walk(x.Else)
			}
		case *ast.Quant, *ast.Agg:
			ok = false
		}
	}
	walk(e)
	return ok
}

// substituteVar copies an expression with from.prop replaced by to.prop.
func substituteVar(e ast.Expr, from, to string) ast.Expr {
	switch x := e.(type) {
	case *ast.PropRef:
		if x.Var == from {
			return &ast.PropRef{Var: to, Prop: x.Prop}
		}
		return x
	case *ast.Bin:
		return &ast.Bin{Op: x.Op, L: substituteVar(x.L, from, to), R: substituteVar(x.R, from, to)}
	case *ast.Un:
		return &ast.Un{Op: x.Op, X: substituteVar(x.X, from, to)}
	case *ast.Call:
		c := &ast.Call{Fn: x.Fn}
		for _, a := range x.Args {
			c.Args = append(c.Args, substituteVar(a, from, to))
		}
		return c
	case *ast.List:
		l := &ast.List{}
		for _, el := range x.Elems {
			l.Elems = append(l.Elems, substituteVar(el, from, to))
		}
		return l
	case *ast.Case:
		c := &ast.Case{}
		for _, w := range x.Whens {
			c.Whens = append(c.Whens, ast.CaseWhen{When: substituteVar(w.When, from, to), Then: substituteVar(w.Then, from, to)})
		}
		if x.Else != nil {
			c.Else = substituteVar(x.Else, from, to)
		}
		return c
	}
	return e
}

// quantGroup is a quantified path pattern as a variable-length edge.
type quantGroup struct {
	edge                    ir.PatternEdge
	rng                     ir.HopRange
	segPreds                []ir.SegPred
	outer                   []ast.Expr // conditions on the left and right nodes
	leftLabels, rightLabels []string
}

// quantified turns a one-relationship quantified path pattern ((x)-[r]->(y) WHERE c){m,n} between left and right
// into a variable-length edge seg. Iteration i binds x to the i-th node of the path, r to its i-th relationship and
// y to the next node, so conditions on x hold on left and the inner nodes, conditions on y on the inner nodes and
// right, and conditions on r on every relationship. Each condition must read one group variable only.
func (b *builder) quantified(q *ast.QuantifiedPath, seg, left, right string) (quantGroup, error) {
	var g quantGroup
	if len(q.Rels) != 1 {
		return g, unsupported("quantified path pattern with %d relationships", len(q.Rels))
	}
	x, r, y := q.Nodes[0], q.Rels[0], q.Nodes[1]
	vars := []string{x.Var, r.Var, y.Var}
	for i, v := range vars {
		if _, bound := b.scope[v]; bound || v == left || v == right || b.group[v] {
			return g, unsupported("group variable %s bound elsewhere", v)
		}
		for _, w := range vars[:i] {
			if v == w {
				return g, unsupported("group variable %s repeated in its pattern", v)
			}
		}
	}
	conds := map[string][]ast.Expr{}
	own := func(v string, e ast.Expr) error {
		for _, c := range splitAnd(e) {
			if !onlyRefersTo(c, v) {
				return unsupported("condition %s in a quantified path pattern reads more than %s", c, v)
			}
			conds[v] = append(conds[v], c)
		}
		return nil
	}
	props := func(v string, ps []ast.PropEq) {
		for _, pe := range ps {
			conds[v] = append(conds[v], &ast.Bin{Op: "=", L: &ast.PropRef{Var: v, Prop: pe.Key}, R: pe.Value})
		}
	}
	props(x.Var, x.Props)
	props(r.Var, r.Props)
	props(y.Var, y.Props)
	for _, el := range []struct {
		v string
		w ast.Expr
	}{{x.Var, x.Where}, {r.Var, r.Where}, {y.Var, y.Where}} {
		if el.w != nil {
			if err := own(el.v, el.w); err != nil {
				return g, err
			}
		}
	}
	if q.Where != nil {
		for _, c := range splitAnd(q.Where) {
			owner := ""
			for _, v := range vars {
				if onlyRefersTo(c, v) {
					owner = v
					break
				}
			}
			if owner == "" {
				return g, unsupported("condition %s in a quantified path pattern reads more than one group variable", c)
			}
			conds[owner] = append(conds[owner], c)
		}
	}

	g.rng = hopRange(&q.Range)
	nodeConds := len(conds[x.Var]) > 0 || len(conds[y.Var]) > 0 || len(x.Labels) > 0 || len(y.Labels) > 0
	if g.rng.Min == 0 && nodeConds {
		// Zero iterations put no condition on left (= right): not expressible as conditions on the ends.
		return g, unsupported("quantified path pattern with lower bound 0 and node conditions")
	}
	inner := append([]ast.Expr{}, conds[x.Var]...)
	for _, c := range conds[y.Var] {
		inner = append(inner, substituteVar(c, y.Var, x.Var))
	}
	if len(inner) > 0 {
		pred, err := b.conv(andAll(inner), convCtx{quant: map[string]quantSrc{x.Var: {part: ir.SegNodes}}})
		if err != nil {
			return g, err
		}
		g.segPreds = append(g.segPreds, ir.SegPred{Part: ir.SegNodes, Var: x.Var, Pred: pred})
	}
	if rc := conds[r.Var]; len(rc) > 0 {
		pred, err := b.conv(andAll(rc), convCtx{quant: map[string]quantSrc{r.Var: {part: ir.SegRels}}})
		if err != nil {
			return g, err
		}
		g.segPreds = append(g.segPreds, ir.SegPred{Part: ir.SegRels, Var: r.Var, Pred: pred})
	}
	for _, c := range conds[x.Var] {
		g.outer = append(g.outer, substituteVar(c, x.Var, left))
	}
	for _, c := range conds[y.Var] {
		g.outer = append(g.outer, substituteVar(c, y.Var, right))
	}
	g.leftLabels, g.rightLabels = x.Labels, y.Labels
	var innerLabels []string
	for _, l := range append(append([]string{}, x.Labels...), y.Labels...) {
		if !contains(innerLabels, l) {
			innerLabels = append(innerLabels, l)
		}
	}
	g.edge = ir.PatternEdge{Var: seg, Types: r.Types, Left: left, Right: right, Dir: direction(r.Dir), InnerLabels: innerLabels}
	for _, v := range vars {
		b.group[v] = true
	}
	return g, nil
}

func andAll(es []ast.Expr) ast.Expr {
	out := es[0]
	for _, e := range es[1:] {
		out = &ast.Bin{Op: "AND", L: out, R: e}
	}
	return out
}
