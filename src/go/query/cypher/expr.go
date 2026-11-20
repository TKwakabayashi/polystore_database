package cypher

// expr.go translates Cypher expressions to ir expressions.

import (
	"fmt"

	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

// convCtx maps a quantifier's iteration variable to its path (and part).
type convCtx struct {
	quant map[string]quantSrc
}

type quantSrc struct {
	path string
	part ir.SegPart
}

// unknown is the error for a variable not in scope: a group variable outside its quantified pattern (a list
// there) is unsupported, any other name a semantic error.
func (b *builder) unknown(v string) error {
	if b.group[v] {
		return unsupported("group variable %s outside its quantified path pattern", v)
	}
	return fmt.Errorf("%w: unknown variable %s", ast.ErrSemantic, v)
}

// conv converts an AST expression into a logical expression. The caller places the GetProperty /
// GetListProperty / ProjectPath it needs (found with needs).
func (b *builder) conv(e ast.Expr, ctx convCtx) (ir.Expr, error) {
	switch x := e.(type) {
	case *ast.Lit:
		return &ir.Lit{Val: x.Val}, nil
	case *ast.VarRef:
		if _, ok := ctx.quant[x.Name]; ok {
			return nil, unsupported("bare iteration variable %s", x.Name)
		}
		vi, ok := b.scope[x.Name]
		if !ok {
			return nil, b.unknown(x.Name)
		}
		if vi.kind == VarSeg {
			return nil, unsupported("variable-length relationship variable %s as a value", x.Name)
		}
		return &ir.Column{Key: x.Name}, nil
	case *ast.PropRef:
		if _, ok := ctx.quant[x.Var]; ok {
			return &ir.ElemProp{Var: x.Var, Prop: x.Prop}, nil
		}
		vi, ok := b.scope[x.Var]
		if !ok {
			return nil, b.unknown(x.Var)
		}
		if vi.kind != VarNode && vi.kind != VarRel {
			return nil, unsupported("property access on non-entity %s", x.Var)
		}
		return &ir.Column{Key: x.Var + "." + x.Prop}, nil
	case *ast.Bin:
		l, err := b.conv(x.L, ctx)
		if err != nil {
			return nil, err
		}
		r, err := b.conv(x.R, ctx)
		if err != nil {
			return nil, err
		}
		return &ir.Bin{Op: x.Op, L: l, R: r}, nil
	case *ast.Un:
		v, err := b.conv(x.X, ctx)
		if err != nil {
			return nil, err
		}
		return &ir.Un{Op: x.Op, X: v}, nil
	case *ast.Call:
		if isPathFn(x.Fn) {
			p, err := b.pathArg(x)
			if err != nil {
				return nil, err
			}
			return &ir.Call{Fn: x.Fn, Args: []ir.Expr{&ir.Column{Key: p}}}, nil
		}
		c := &ir.Call{Fn: x.Fn}
		for _, a := range x.Args {
			v, err := b.conv(a, ctx)
			if err != nil {
				return nil, err
			}
			c.Args = append(c.Args, v)
		}
		return c, nil
	case *ast.Case:
		c := &ir.Case{}
		var subj ir.Expr
		if x.Subject != nil {
			v, err := b.conv(x.Subject, ctx)
			if err != nil {
				return nil, err
			}
			subj = v
		}
		for _, w := range x.Whens {
			wv, err := b.conv(w.When, ctx)
			if err != nil {
				return nil, err
			}
			if subj != nil { // a simple CASE is WHEN subject = value (never matches null)
				wv = &ir.Bin{Op: "=", L: subj, R: wv}
			}
			tv, err := b.conv(w.Then, ctx)
			if err != nil {
				return nil, err
			}
			c.Whens = append(c.Whens, ir.CaseWhen{When: wv, Then: tv})
		}
		if x.Else != nil {
			ev, err := b.conv(x.Else, ctx)
			if err != nil {
				return nil, err
			}
			c.Else = ev
		}
		return c, nil
	case *ast.List:
		l := &ir.List{}
		for _, el := range x.Elems {
			v, err := b.conv(el, ctx)
			if err != nil {
				return nil, err
			}
			l.Elems = append(l.Elems, v)
		}
		return l, nil
	case *ast.Quant:
		src, err := b.quantSource(x)
		if err != nil {
			return nil, err
		}
		inner := convCtx{quant: map[string]quantSrc{}}
		for k, v := range ctx.quant {
			inner.quant[k] = v
		}
		inner.quant[x.Var] = src
		pred, err := b.conv(x.Pred, inner)
		if err != nil {
			return nil, err
		}
		fn := "nodes"
		if src.part == ir.SegRels {
			fn = "relationships"
		}
		return &ir.Quant{Kind: x.Kind, Var: x.Var,
			List: &ir.Call{Fn: fn, Args: []ir.Expr{&ir.Column{Key: src.path}}}, Pred: pred}, nil
	case *ast.Agg:
		return nil, unsupported("aggregate outside projection")
	}
	return nil, unsupported("expression %T", e)
}

func isPathFn(fn string) bool { return fn == "length" || fn == "nodes" || fn == "relationships" }

func (b *builder) pathArg(c *ast.Call) (string, error) {
	if len(c.Args) != 1 {
		return "", unsupported("%s arity", c.Fn)
	}
	v, ok := c.Args[0].(*ast.VarRef)
	if !ok || b.scope[v.Name].kind != VarPath {
		return "", unsupported("%s argument must be a path variable", c.Fn)
	}
	if _, bound := b.scope[v.Name]; !bound {
		return "", unsupported("unknown path %s", v.Name)
	}
	return v.Name, nil
}

// quantSource checks that a quantifier's list is nodes(p) / relationships(p).
func (b *builder) quantSource(x *ast.Quant) (quantSrc, error) {
	c, ok := x.List.(*ast.Call)
	if !ok || (c.Fn != "nodes" && c.Fn != "relationships") {
		return quantSrc{}, unsupported("quantifier over %s", x.List)
	}
	p, err := b.pathArg(c)
	if err != nil {
		return quantSrc{}, err
	}
	part := ir.SegNodes
	if c.Fn == "relationships" {
		part = ir.SegRels
	}
	return quantSrc{path: p, part: part}, nil
}

// needs collects the columns an expression needs.
type needs struct {
	props     []propNeed
	listProps []listNeed
	paths     []string
}

type propNeed struct{ v, prop string }
type listNeed struct {
	path string
	part ir.SegPart
	prop string
}

func (b *builder) collectNeeds(e ast.Expr, ctx convCtx, n *needs) {
	switch x := e.(type) {
	case *ast.PropRef:
		if src, ok := ctx.quant[x.Var]; ok {
			n.listProps = append(n.listProps, listNeed{src.path, src.part, x.Prop})
			return
		}
		n.props = append(n.props, propNeed{x.Var, x.Prop})
	case *ast.VarRef:
		if b.scope[x.Name].kind == VarPath {
			n.paths = append(n.paths, x.Name)
		}
	case *ast.Bin:
		b.collectNeeds(x.L, ctx, n)
		b.collectNeeds(x.R, ctx, n)
	case *ast.Un:
		b.collectNeeds(x.X, ctx, n)
	case *ast.Call:
		for _, a := range x.Args {
			b.collectNeeds(a, ctx, n)
		}
	case *ast.Case:
		if x.Subject != nil {
			b.collectNeeds(x.Subject, ctx, n)
		}
		for _, w := range x.Whens {
			b.collectNeeds(w.When, ctx, n)
			b.collectNeeds(w.Then, ctx, n)
		}
		if x.Else != nil {
			b.collectNeeds(x.Else, ctx, n)
		}
	case *ast.List:
		for _, el := range x.Elems {
			b.collectNeeds(el, ctx, n)
		}
	case *ast.Quant:
		src, err := b.quantSource(x)
		if err != nil {
			return // conv returns the same error
		}
		n.paths = append(n.paths, src.path)
		inner := convCtx{quant: map[string]quantSrc{x.Var: src}}
		for k, v := range ctx.quant {
			inner.quant[k] = v
		}
		b.collectNeeds(x.Pred, inner, n)
	case *ast.Agg:
		if x.Arg != nil {
			b.collectNeeds(x.Arg, ctx, n)
		}
	}
}

// fetches returns the operators (without input) that add the needed columns not in the row yet, in order and once each.
func (b *builder) fetches(nd needs) []ir.Node {
	var out []ir.Node
	seen := map[string]bool{}
	add := func(f ir.Node) {
		k := ir.FetchKey(f)
		if !b.avail[k] && !seen[k] {
			seen[k] = true
			out = append(out, f)
		}
	}
	for _, p := range nd.paths {
		add(&ir.ProjectPath{Path: p, Elems: b.paths[p].elems})
	}
	for _, p := range nd.props {
		vi := b.scope[p.v]
		kind := ir.NodeObj
		if vi.kind == VarRel {
			kind = ir.RelObj
		}
		add(&ir.GetProperty{Alias: p.v, Prop: p.prop, Kind: kind, Labels: vi.labels})
	}
	for _, lp := range nd.listProps {
		add(&ir.GetListProperty{Source: lp.path, Part: lp.part, Prop: lp.prop})
	}
	return out
}

// supply adds to plan the needed columns that are not in the row yet.
func (b *builder) supply(n ir.Node, nd needs) ir.Node {
	for _, f := range b.fetches(nd) {
		n = ir.WithInput(f, n)
		b.avail[ir.FetchKey(f)] = true
	}
	return n
}

// deps returns the bindings an expression depends on (nodes, relationships, segments); path variables expand to their parts.
func (b *builder) deps(e ast.Expr) []string {
	seen := map[string]bool{}
	var out []string
	add := func(v string) {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	var walk func(ast.Expr, map[string]bool)
	walk = func(e ast.Expr, qv map[string]bool) {
		switch x := e.(type) {
		case *ast.VarRef:
			if qv[x.Name] {
				return
			}
			if pd, ok := b.paths[x.Name]; ok {
				for _, el := range pd.elems {
					add(el.Var)
				}
				return
			}
			add(x.Name)
		case *ast.PropRef:
			if !qv[x.Var] {
				add(x.Var)
			}
		case *ast.Bin:
			walk(x.L, qv)
			walk(x.R, qv)
		case *ast.Un:
			walk(x.X, qv)
		case *ast.Call:
			for _, a := range x.Args {
				walk(a, qv)
			}
		case *ast.Case:
			if x.Subject != nil {
				walk(x.Subject, qv)
			}
			for _, w := range x.Whens {
				walk(w.When, qv)
				walk(w.Then, qv)
			}
			if x.Else != nil {
				walk(x.Else, qv)
			}
		case *ast.List:
			for _, el := range x.Elems {
				walk(el, qv)
			}
		case *ast.Quant:
			walk(x.List, qv)
			inner := map[string]bool{x.Var: true}
			for k := range qv {
				inner[k] = true
			}
			walk(x.Pred, inner)
		case *ast.Agg:
			if x.Arg != nil {
				walk(x.Arg, qv)
			}
		}
	}
	walk(e, map[string]bool{})
	return out
}
