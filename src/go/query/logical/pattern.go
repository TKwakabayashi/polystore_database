package logical

import (
	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/query/ir"
)

// pattern.go plans the patterns the front end leaves as ir.Pattern. Each connected component is either grown
// from one start vertex along its edges or split at its anchors, grown from each and joined; the orderer
// (order.go) decides which. Predicates are placed right after their variables are bound.

// PlanPatterns replaces every ir.Pattern in n with scans, expansions, joins and filters. With statistics st the
// start, split and edge order are chosen by estimated cost; without them (nil), by rules.
func PlanPatterns(n ir.Node, st *stats.Stats, opt Options) ir.Node {
	var ord orderer = ruleOrderer{}
	if st != nil {
		ord = costOrderer{}
	}
	b := &builder{order: ord, opt: opt, st: st, est: 1}
	return b.walk(n)
}

type builder struct {
	opt   Options
	order orderer
	st    *stats.Stats // statistics (nil: rules decide)
	est   float64      // estimated rows of the current plan (meaningful only with statistics)

	// The pattern being planned.
	plan   ir.Node
	avail  map[string]bool     // columns its fetches added
	labels map[string][]string // labels of the variables in scope
}

// walk plans the patterns bottom-up, keeping the row estimate the cost-based orderer starts from.
func (b *builder) walk(n ir.Node) ir.Node {
	switch x := n.(type) {
	case nil:
		return nil
	case *ir.Pattern:
		return b.pattern(x, b.walk(x.Input))
	case *ir.Optional:
		x.Outer = b.walk(x.Outer)
		outer := b.est
		b.est = 1 // the inner side is built per outer row
		x.Inner = b.walk(x.Inner)
		b.est = outer // a left outer join keeps every outer row (growth is not estimated)
		return x
	case *ir.NodeHashJoin:
		x.Left, x.Right = b.walk(x.Left), b.walk(x.Right)
		return x
	}
	if c := ir.Child(n); c != nil {
		ir.SetChild(n, b.walk(c))
	}
	switch x := n.(type) {
	case *ir.Unwind:
		b.est *= unwindFanout
	case *ir.Aggregation:
		if len(x.Keys) == 0 {
			b.est = 1
		}
	case *ir.Limit:
		if x.Count >= 0 && float64(x.Count) < b.est {
			b.est = float64(x.Count)
		}
	}
	return n
}

type patVertex struct {
	name   string
	labels []string
	order  int
}

type patEdge struct {
	*ir.PatternEdge
	order int
}

// growCtx is one pattern's graph (shared by components and split sides).
type growCtx struct {
	vertices map[string]*patVertex
	edges    []*patEdge
}

// pattern plans p on input (the planned rows so far).
func (b *builder) pattern(p *ir.Pattern, input ir.Node) ir.Node {
	b.plan, b.avail, b.labels = input, map[string]bool{}, p.Labels
	g := &growCtx{vertices: map[string]*patVertex{}}
	var vorder []string
	for i, v := range p.Vertices {
		g.vertices[v.Var] = &patVertex{name: v.Var, labels: v.Labels, order: i}
		vorder = append(vorder, v.Var)
	}
	for i := range p.Edges {
		g.edges = append(g.edges, &patEdge{PatternEdge: &p.Edges[i], order: i})
	}
	pending := append([]ir.PatternPred(nil), p.Preds...)
	bound := map[string]bool{}
	for _, v := range p.Bound {
		bound[v] = true
	}

	// 1. Each connected component (with two or more anchors, split, expand from both sides and join)
	for _, comp := range components(vorder, g.edges) {
		inComp := map[string]bool{}
		for _, v := range comp {
			inComp[v] = true
		}
		groups := b.anchorGroups(comp, bound, pending)
		split, start := b.order.component(b, g, comp, groups, inComp, bound, pending, b.opt.SplitAnchoredPatterns && len(groups) >= 2)
		if split {
			pending = b.buildSplit(g, comp, groups, inComp, bound, pending)
			continue
		}
		if start == "" {
			start = b.chooseStart(comp, bound, pending)
		}
		fresh := !bound[start]
		var sub ir.Node
		compBound := map[string]bool{}
		if fresh {
			// A new component is built independently of earlier bindings and joined by a cross product at the end.
			sub = &ir.EntityScan{Alias: start, Labels: g.vertices[start].labels}
		} else {
			sub = b.plan
			for k := range bound {
				compBound[k] = true
			}
		}
		compBound[start] = true
		sub, pending = b.grow(g, sub, compBound, pending, func(e *patEdge) bool { return inComp[e.Left] })
		if fresh && b.plan != nil {
			b.plan = &ir.NodeHashJoin{Left: b.plan, Right: sub}
		} else {
			b.plan = sub
		}
		for k := range compBound {
			bound[k] = true
		}
	}

	// 2. Relationship uniqueness (when two or more relationships may share types)
	if len(p.UniqueRels)+len(p.UniqueSegs) >= 2 {
		b.plan = &ir.RelUniqueFilter{Input: b.plan, Rels: p.UniqueRels, Segs: p.UniqueSegs}
	}

	// 3. The rest (predicates across components or needing path columns); the front end checked that none is left over.
	b.plan, _ = b.placeReady(b.plan, bound, pending)
	return b.plan
}

// grow extends sub from bound vertices along allowed edges in the orderer's order and places predicates once ready.
// A variable-length edge between two bound vertices expands to a temporary vertex filtered by Filter(tmp = to).
func (b *builder) grow(g *growCtx, sub ir.Node, compBound map[string]bool, pending []ir.PatternPred, allowed func(*patEdge) bool) (ir.Node, []ir.PatternPred) {
	sub, pending = b.placeReady(sub, compBound, pending)
	done := map[int]bool{}
	for {
		e := b.order.next(b, g, allowed, compBound, done, pending)
		if e == nil {
			break
		}
		done[e.order] = true
		from, to, dir := orient(e, compBound)
		switch {
		case compBound[to] && e.Range != nil:
			tmp := "_vl_" + e.Var
			sub = &ir.VarLengthExpand{Input: sub, From: from, Seg: e.Var, To: tmp, Types: e.Types,
				Dir: dir, ToLabels: g.vertices[to].labels, Range: *e.Range, SegPreds: e.SegPreds, InnerLabels: e.InnerLabels}
			sub = &ir.Filter{Input: sub, Pred: &ir.Bin{Op: "=", L: &ir.Column{Key: tmp}, R: &ir.Column{Key: to}}}
			compBound[e.Var] = true
		case compBound[to]:
			sub = &ir.ExpandInto{Input: sub, From: from, Rel: e.Var, To: to, Types: e.Types, Dir: dir}
			compBound[e.Var] = true
		case e.Range != nil:
			sub = &ir.VarLengthExpand{Input: sub, From: from, Seg: e.Var, To: to, Types: e.Types,
				Dir: dir, ToLabels: g.vertices[to].labels, Range: *e.Range, SegPreds: e.SegPreds, InnerLabels: e.InnerLabels}
			compBound[e.Var] = true
			compBound[to] = true
		default:
			sub = &ir.Expand{Input: sub, From: from, Rel: e.Var, To: to, Types: e.Types,
				Dir: dir, ToLabels: g.vertices[to].labels}
			compBound[e.Var] = true
			compBound[to] = true
		}
		sub, pending = b.placeReady(sub, compBound, pending)
	}
	return sub, pending
}

// buildSplit builds a component with several anchors split: vertices go to the nearest anchor group
// (multi-source BFS; ties to the earlier group), each edge is expanded by its side, each side grows from its
// anchor, and the sides meet in a NodeHashJoin on the vertices both bind (a bound group grows on the previous plan).
func (b *builder) buildSplit(g *growCtx, comp []string, groups [][]string, inComp map[string]bool, bound map[string]bool, pending []ir.PatternPred) []ir.PatternPred {
	side := b.splitSide(g, groups, inComp)

	var acc ir.Node
	accBound := map[string]bool{}
	hasBound := bound[groups[0][0]]
	for gi, grp := range groups {
		var sub ir.Node
		sb := map[string]bool{}
		if gi == 0 && hasBound {
			sub = b.plan
			for k := range bound {
				sb[k] = true
			}
		} else {
			sub = &ir.EntityScan{Alias: grp[0], Labels: g.vertices[grp[0]].labels}
			sb[grp[0]] = true
		}
		gi := gi
		sub, pending = b.grow(g, sub, sb, pending, func(e *patEdge) bool { return inComp[e.Left] && side(e) == gi })
		if acc == nil {
			acc = sub
		} else {
			var keys []string
			for _, v := range comp {
				if accBound[v] && sb[v] {
					keys = append(keys, v)
				}
			}
			acc = &ir.NodeHashJoin{Left: acc, Right: sub, Keys: keys}
		}
		for k := range sb {
			accBound[k] = true
		}
	}
	if !hasBound && b.plan != nil {
		acc = &ir.NodeHashJoin{Left: b.plan, Right: acc}
	}
	b.plan = acc
	for k := range accBound {
		bound[k] = true
	}
	b.plan, pending = b.placeReady(b.plan, bound, pending)
	return pending
}

// placeReady places, as Filters, the predicates whose variables are all bound, after adding the columns they need.
func (b *builder) placeReady(n ir.Node, bound map[string]bool, pending []ir.PatternPred) (ir.Node, []ir.PatternPred) {
	var rest []ir.PatternPred
	for _, p := range pending {
		ready := true
		for _, d := range p.Deps {
			if !bound[d] {
				ready = false
				break
			}
		}
		if !ready {
			rest = append(rest, p)
			continue
		}
		for _, f := range p.Fetches {
			if k := ir.FetchKey(f); !b.avail[k] {
				n = ir.WithInput(f, n)
				b.avail[k] = true
			}
		}
		n = &ir.Filter{Input: n, Pred: p.Pred}
	}
	return n, rest
}

// components returns the undirected connected components, ordered by their first vertex.
func components(vorder []string, edges []*patEdge) [][]string {
	parent := map[string]string{}
	var find func(string) string
	find = func(x string) string {
		if parent[x] == x {
			return x
		}
		parent[x] = find(parent[x])
		return parent[x]
	}
	for _, v := range vorder {
		parent[v] = v
	}
	for _, e := range edges {
		a, c := find(e.Left), find(e.Right)
		if a != c {
			parent[c] = a
		}
	}
	idx := map[string]int{}
	var comps [][]string
	for _, v := range vorder {
		r := find(v)
		i, ok := idx[r]
		if !ok {
			i = len(comps)
			idx[r] = i
			comps = append(comps, nil)
		}
		comps[i] = append(comps[i], v)
	}
	return comps
}

// orient returns the direction with the bound side as From (the pattern's direction when both are bound).
func orient(e *patEdge, bound map[string]bool) (from, to string, dir ir.Direction) {
	if bound[e.Left] {
		return e.Left, e.Right, e.Dir
	}
	return e.Right, e.Left, e.Dir.Reverse()
}
