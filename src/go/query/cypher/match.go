package cypher

// match.go translates MATCH, OPTIONAL MATCH and UNWIND (MATCH becomes an ir.Pattern).

import (
	"sort"

	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

func (b *builder) match(m *ast.MatchClause) error {
	if m.Optional {
		return b.optionalMatch(m)
	}
	// 1. Pattern graph
	pat := &ir.Pattern{Input: b.plan}
	vidx := map[string]int{}
	addVertex := func(n ast.NodePat) error {
		if vi, ok := b.scope[n.Var]; ok && vi.kind != VarNode {
			return unsupported("variable %s rebound as node", n.Var)
		}
		i, ok := vidx[n.Var]
		if !ok {
			v := ir.PatternVertex{Var: n.Var}
			if vi, bound := b.scope[n.Var]; bound {
				v.Labels = vi.labels
			}
			i = len(pat.Vertices)
			vidx[n.Var] = i
			pat.Vertices = append(pat.Vertices, v)
		}
		v := &pat.Vertices[i]
		for _, l := range n.Labels {
			if !contains(v.Labels, l) {
				if _, bound := b.scope[n.Var]; bound {
					return unsupported("new label on bound variable %s", n.Var)
				}
				v.Labels = append(v.Labels, l)
			}
		}
		return nil
	}

	var pending []ast.Expr
	if m.Where != nil {
		pending = splitAnd(m.Where)
	}
	segRange := map[string]ir.HopRange{}
	groupPreds := map[string][]ir.SegPred{}
	var edges []ir.PatternEdge
	for _, p := range m.Patterns {
		var pd pathDef
		for i, n := range p.Nodes {
			if err := addVertex(n); err != nil {
				return err
			}
			for _, pe := range n.Props {
				pending = append(pending, &ast.Bin{Op: "=", L: &ast.PropRef{Var: n.Var, Prop: pe.Key}, R: pe.Value})
			}
			if n.Where != nil {
				pending = append(pending, splitAnd(n.Where)...)
			}
			pd.elems = append(pd.elems, ir.PathElem{Var: n.Var})
			pd.nodes = append(pd.nodes, n.Var)
			if i < len(p.Rels) {
				r := p.Rels[i]
				if _, bound := b.scope[r.Var]; bound {
					return unsupported("relationship variable %s already bound", r.Var)
				}
				if r.Quantified != nil {
					right := p.Nodes[i+1].Var
					g, err := b.quantified(r.Quantified, r.Var, n.Var, right)
					if err != nil {
						return err
					}
					if err := addVertex(ast.NodePat{Var: n.Var, Labels: g.leftLabels}); err != nil {
						return err
					}
					if err := addVertex(ast.NodePat{Var: right, Labels: g.rightLabels}); err != nil {
						return err
					}
					edges = append(edges, g.edge)
					segRange[r.Var] = g.rng
					groupPreds[r.Var] = g.segPreds
					pending = append(pending, g.outer...)
					pd.elems = append(pd.elems, ir.PathElem{Var: r.Var, Seg: true})
					pd.segs = append(pd.segs, r.Var)
					continue
				}
				edges = append(edges, ir.PatternEdge{Var: r.Var, Types: r.Types, Left: n.Var, Right: p.Nodes[i+1].Var, Dir: direction(r.Dir)})
				if r.Range != nil {
					if len(r.Props) > 0 {
						return unsupported("properties on variable-length relationship %s", r.Var)
					}
					if r.Where != nil {
						return unsupported("WHERE on variable-length relationship %s", r.Var)
					}
					segRange[r.Var] = hopRange(r.Range)
					pd.elems = append(pd.elems, ir.PathElem{Var: r.Var, Seg: true})
					pd.segs = append(pd.segs, r.Var)
				} else {
					pd.elems = append(pd.elems, ir.PathElem{Var: r.Var})
					pd.rels = append(pd.rels, r.Var)
					for _, pe := range r.Props {
						pending = append(pending, &ast.Bin{Op: "=", L: &ast.PropRef{Var: r.Var, Prop: pe.Key}, R: pe.Value})
					}
					if r.Where != nil {
						pending = append(pending, splitAnd(r.Where)...)
					}
				}
			}
		}
		if p.PathVar != "" {
			if _, bound := b.scope[p.PathVar]; bound {
				return unsupported("path variable %s already bound", p.PathVar)
			}
			b.paths[p.PathVar] = pd
		}
	}

	// Register variables in scope (their kinds are needed before converting predicates and computing deps).
	for _, v := range pat.Vertices {
		b.scope[v.Var] = varInfo{kind: VarNode, labels: v.Labels}
	}
	for _, e := range edges {
		kind := VarRel
		if _, seg := segRange[e.Var]; seg {
			kind = VarSeg
		}
		b.scope[e.Var] = varInfo{kind: kind, labels: e.Types}
	}
	for _, p := range m.Patterns {
		if p.PathVar != "" {
			b.scope[p.PathVar] = varInfo{kind: VarPath}
		}
	}
	// Whether a variable is bound is decided by the bound set (scope only tells kinds).
	prevBound := b.boundSet()

	// 2. Decompose whole-path conditions
	segPreds := groupPreds
	var err error
	pending, err = b.decomposePathPreds(pending, segRange, segPreds)
	if err != nil {
		return err
	}
	for i := range edges {
		if r, ok := segRange[edges[i].Var]; ok {
			edges[i].Range = &r
			edges[i].SegPreds = segPreds[edges[i].Var]
		}
	}
	pat.Edges = edges
	pat.UniqueRels, pat.UniqueSegs = uniqueCandidates(edges)

	// 3. Predicates: every one must be placeable once the pattern is bound
	bound := b.boundSet()
	for _, v := range pat.Vertices {
		bound[v.Var] = true
	}
	for _, e := range edges {
		bound[e.Var] = true
	}
	added := map[string]bool{}
	var unplaced []ast.Expr
	for _, e := range pending {
		deps := b.deps(e)
		if !allBound(deps, bound) {
			unplaced = append(unplaced, e)
			continue
		}
		var nd needs
		b.collectNeeds(e, convCtx{}, &nd)
		fetches := b.fetches(nd)
		pred, err := b.conv(e, convCtx{})
		if err != nil {
			return err
		}
		pat.Preds = append(pat.Preds, ir.PatternPred{Pred: pred, Deps: deps, Fetches: fetches})
		for _, f := range fetches {
			added[ir.FetchKey(f)] = true
		}
	}
	if len(unplaced) > 0 {
		return unsupported("unplaceable predicate %s", unplaced[0])
	}
	for k := range added {
		b.avail[k] = true
	}

	for v := range prevBound {
		pat.Bound = append(pat.Bound, v)
	}
	sort.Strings(pat.Bound)
	pat.Labels = map[string][]string{}
	for v, vi := range b.scope {
		if vi.kind == VarNode || vi.kind == VarRel {
			pat.Labels[v] = vi.labels
		}
	}
	b.plan = pat
	b.bound = bound
	return nil
}

func allBound(vars []string, bound map[string]bool) bool {
	for _, v := range vars {
		if !bound[v] {
			return false
		}
	}
	return true
}

func direction(d ast.Direction) ir.Direction {
	switch d {
	case ast.DirLeft:
		return ir.In
	case ast.DirBoth:
		return ir.Both
	}
	return ir.Out
}

// optionalMatch turns OPTIONAL MATCH into a left outer apply (Optional). Pattern and WHERE are built with the
// MATCH rules on an Argument that receives the outer rows' bound id columns (Keys); WHERE is part of the match.
func (b *builder) optionalMatch(m *ast.MatchClause) error {
	if b.plan == nil {
		return unsupported("OPTIONAL MATCH as the first clause")
	}
	// Keys: the bound nodes and relationships the pattern and WHERE reference, in order.
	var keys []string
	seen := map[string]bool{}
	consider := func(v string) error {
		if seen[v] || !b.bound[v] {
			return nil
		}
		seen[v] = true
		switch b.scope[v].kind {
		case VarNode, VarRel:
			keys = append(keys, v)
			return nil
		}
		return unsupported("OPTIONAL MATCH referring to value %s", v)
	}
	var order []string // candidate new variables, in order
	for _, p := range m.Patterns {
		for i, n := range p.Nodes {
			if err := consider(n.Var); err != nil {
				return err
			}
			order = append(order, n.Var)
			if i < len(p.Rels) {
				if err := consider(p.Rels[i].Var); err != nil {
					return err
				}
				order = append(order, p.Rels[i].Var)
			}
		}
	}
	if m.Where != nil {
		for _, d := range b.deps(m.Where) {
			if err := consider(d); err != nil {
				return err
			}
		}
	}

	outer, outerBound, outerAvail := b.plan, b.bound, b.avail
	b.plan = &ir.Argument{Cols: keys}
	b.bound = map[string]bool{}
	for _, k := range keys {
		b.bound[k] = true
	}
	b.avail = map[string]bool{}
	if err := b.match(&ast.MatchClause{Patterns: m.Patterns, Where: m.Where}); err != nil {
		return err
	}
	inner, innerBound, innerAvail := b.plan, b.bound, b.avail

	var newVars []string
	added := map[string]bool{}
	for _, v := range order {
		if innerBound[v] && !seen[v] && !added[v] {
			added[v] = true
			newVars = append(newVars, v)
		}
	}
	b.plan = &ir.Optional{Outer: outer, Inner: inner, Keys: keys, New: newVars}
	b.bound, b.avail = outerBound, outerAvail
	for k := range innerBound {
		b.bound[k] = true
	}
	for k := range innerAvail {
		b.avail[k] = true
	}
	return nil
}

// unwind is UNWIND. Unwinding a list of bindings (a collect(n) value column) makes an id column of the original kind.
func (b *builder) unwind(u *ast.UnwindClause) error {
	if b.plan == nil {
		return unsupported("UNWIND as the first clause")
	}
	if _, ok := b.scope[u.Var]; ok {
		return unsupported("UNWIND rebinding %s", u.Var)
	}
	var nd needs
	b.collectNeeds(u.Expr, convCtx{}, &nd)
	b.plan = b.supply(b.plan, nd)
	e, err := b.conv(u.Expr, convCtx{})
	if err != nil {
		return err
	}
	b.plan = &ir.Unwind{Input: b.plan, List: e, Var: u.Var}
	info := varInfo{kind: VarValue}
	if v, ok := u.Expr.(*ast.VarRef); ok && b.scope[v.Name].elem != nil {
		info = *b.scope[v.Name].elem
	}
	b.scope[u.Var] = info
	b.bound[u.Var] = true
	if info.kind == VarValue {
		b.avail[u.Var] = true
	}
	return nil
}

// onlyOutputs reports whether an expression references only projected column names (no property, aggregate or quantifier).
func onlyOutputs(e ast.Expr, outNames map[string]bool) bool {
	ok, any := true, false
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.VarRef:
			any = true
			if !outNames[x.Name] {
				ok = false
			}
		case *ast.Lit:
		case *ast.Bin:
			walk(x.L)
			walk(x.R)
		case *ast.Un:
			walk(x.X)
		case *ast.Call:
			if isPathFn(x.Fn) {
				ok = false
			}
			for _, a := range x.Args {
				walk(a)
			}
		default:
			ok = false
		}
	}
	walk(e)
	return ok && any
}

// boundSet returns the variables bound so far (results of earlier clauses).
func (b *builder) boundSet() map[string]bool {
	out := map[string]bool{}
	for k := range b.bound {
		out[k] = true
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func splitAnd(e ast.Expr) []ast.Expr {
	if bin, ok := e.(*ast.Bin); ok && bin.Op == "AND" {
		return append(splitAnd(bin.L), splitAnd(bin.R)...)
	}
	return []ast.Expr{e}
}

func hopRange(h *ast.HopRange) ir.HopRange {
	r := ir.HopRange{Min: 1, Max: -1}
	if h.Min != nil {
		r.Min = *h.Min
	}
	if h.Max != nil {
		r.Max = *h.Max
	}
	return r
}

// uniqueCandidates returns the relationships and segments subject to uniqueness (types that may overlap others).
func uniqueCandidates(edges []ir.PatternEdge) (rels, segs []string) {
	overlaps := func(a, c []string) bool {
		if len(a) == 0 || len(c) == 0 {
			return true
		}
		for _, x := range a {
			if contains(c, x) {
				return true
			}
		}
		return false
	}
	for i, e := range edges {
		hit := false
		for j, o := range edges {
			if i != j && overlaps(e.Types, o.Types) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		if e.Range != nil {
			segs = append(segs, e.Var)
		} else {
			rels = append(rels, e.Var)
		}
	}
	return rels, segs
}

// placeReady places, as Filters, the predicates whose variables are all bound, after adding the columns they need.
func (b *builder) placeReady(n ir.Node, bound map[string]bool, pending []ast.Expr) (ir.Node, []ast.Expr, error) {
	var rest []ast.Expr
	for _, e := range pending {
		ready := true
		for _, d := range b.deps(e) {
			if !bound[d] {
				ready = false
				break
			}
		}
		if !ready {
			rest = append(rest, e)
			continue
		}
		var nd needs
		b.collectNeeds(e, convCtx{}, &nd)
		n = b.supply(n, nd)
		pred, err := b.conv(e, convCtx{})
		if err != nil {
			return nil, nil, err
		}
		n = &ir.Filter{Input: n, Pred: pred}
	}
	return n, rest, nil
}
