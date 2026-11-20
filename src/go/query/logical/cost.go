package logical

import (
	"math"

	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/query/ir"
)

// cost.go estimates row counts from statistics (stats.Stats: label counts and relationship counts per
// type) for the cost-based orderer:
// - the start of a component (the vertex with the lowest estimated cost),
// - whether to split a component with several anchors (only when splitting is cheaper),
// - the edge order (at each step the edge that grows the rows least).
//
// The cost is the sum of the rows each step makes before its predicates. Value distributions are not
// kept, so selectivities are defaults (unique-key equality 1/count, other equality 1/10, ranges 1/3, after
// System R; equality is kept weak so low-cardinality columns such as type are not underestimated).
// Without statistics (b.st == nil) the rules decide.

const (
	eqSel    = 0.1
	rangeSel = 1.0 / 3
	unkSel   = 0.5
	// An unbounded variable-length expansion is estimated up to maxExtraHops beyond its minimum.
	maxExtraHops = 2
	// Estimated elements per UNWIND row.
	unwindFanout = 10
)

func (b *builder) labelsOf(g *growCtx, v string) []string {
	if vx, ok := g.vertices[v]; ok {
		return vx.labels
	}
	return b.labels[v]
}

// sel is the selectivity of one predicate term.
func (b *builder) sel(g *growCtx, e ir.Expr) float64 {
	switch x := e.(type) {
	case *ir.Bin:
		switch x.Op {
		case "=":
			if v, ok := b.uniqueEq(x); ok {
				return 1 / b.st.Count(b.labelsOf(g, v))
			}
			return eqSel
		case "<>":
			return 1 - eqSel
		case "<", ">", "<=", ">=":
			return rangeSel
		case "IN":
			if l, ok := x.R.(*ir.List); ok {
				return math.Min(1, float64(len(l.Elems))*eqSel)
			}
			return 10 * eqSel
		case "AND":
			return b.sel(g, x.L) * b.sel(g, x.R)
		case "OR":
			l, r := b.sel(g, x.L), b.sel(g, x.R)
			return l + r - l*r
		}
	case *ir.Un:
		switch x.Op {
		case "NOT":
			return 1 - b.sel(g, x.X)
		case "IS NULL":
			return eqSel
		case "IS NOT NULL":
			return 1 - eqSel
		}
	}
	return unkSel
}

// predFactor is the product of the selectivities of the predicates that become evaluable when added joins bound.
func (b *builder) predFactor(g *growCtx, pending []ir.PatternPred, bound map[string]bool, added ...string) float64 {
	f := 1.0
	for _, p := range pending {
		ready, touches := true, false
		for _, d := range p.Deps {
			isNew := false
			for _, a := range added {
				if d == a {
					isNew = true
				}
			}
			if isNew {
				touches = true
			} else if !bound[d] {
				ready = false
			}
		}
		if ready && touches {
			f *= b.sel(g, p.Pred)
		}
	}
	return f
}

func statsDir(d ir.Direction) stats.Dir {
	switch d {
	case ir.Out:
		return stats.Out
	case ir.In:
		return stats.In
	}
	return stats.Both
}

// edgeFactor is the row multiplier of following edge e: before predicates (work, the rows the expansion
// makes) and after the predicates evaluable at that point (out).
func (b *builder) edgeFactor(g *growCtx, e *patEdge, bound map[string]bool, pending []ir.PatternPred) (work, out float64) {
	from, to, dir := orient(e, bound)
	fl, tl := b.labelsOf(g, from), b.labelsOf(g, to)
	sd := statsDir(dir)
	added := []string{e.Var}
	if !bound[to] {
		added = append(added, to)
	}
	var f float64
	if e.Range == nil {
		f = b.st.Degree(fl, e.Types, sd, tl)
	} else {
		// Degree of one hop (the end-label share frac applies to the last hop only).
		step := b.st.Degree(fl, e.Types, sd, nil)
		frac := 1.0
		if step > 0 {
			frac = b.st.Degree(fl, e.Types, sd, tl) / step
		}
		r := *e.Range
		hi := r.Max
		if hi < 0 {
			hi = r.Min + maxExtraHops
		}
		for k := r.Min; k <= hi; k++ {
			if k == 0 {
				f++
			} else {
				f += math.Pow(step, float64(k)) * frac
			}
		}
	}
	work = f
	if bound[to] { // both ends bound: the chance of arriving at that vertex
		f /= b.st.Count(tl)
	}
	return work, f * b.predFactor(g, pending, bound, added...)
}

// pickEdge picks, among unprocessed allowed edges with a bound end, the one with the smallest row multiplier (ties in pattern order).
func (b *builder) pickEdge(g *growCtx, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool, pending []ir.PatternPred) *patEdge {
	var best *patEdge
	bestF := 0.0
	for _, e := range g.edges {
		if done[e.order] || !allowed(e) || !(bound[e.Left] || bound[e.Right]) {
			continue
		}
		if _, f := b.edgeFactor(g, e, bound, pending); best == nil || f < bestF {
			best, bestF = e, f
		}
	}
	return best
}

// simulate follows the allowed edges from bound0 in pickEdge order and returns the estimated cost (sum of work) and the final rows.
func (b *builder) simulate(g *growCtx, bound0 map[string]bool, allowed func(*patEdge) bool, pending []ir.PatternPred, rows float64) (cost, out float64) {
	bound := map[string]bool{}
	for k := range bound0 {
		bound[k] = true
	}
	done := map[int]bool{}
	for {
		e := b.pickEdge(g, allowed, bound, done, pending)
		if e == nil {
			return cost, rows
		}
		done[e.order] = true
		work, out := b.edgeFactor(g, e, bound, pending)
		cost += rows * work
		rows *= out
		bound[e.Var], bound[e.Left], bound[e.Right] = true, true, true
	}
}

// startRows is the first row count when starting from v (count x selectivity of the predicates on v alone).
func (b *builder) startRows(g *growCtx, v string, pending []ir.PatternPred) float64 {
	return b.st.Count(b.labelsOf(g, v)) * b.predFactor(g, pending, map[string]bool{}, v)
}

// compPlan is the estimated way to build a component.
type compPlan struct {
	split bool
	start string  // the start when not split
	rows  float64 // estimated rows after the component (including previous rows for a bound component)
}

// planComp decides how to build a component from statistics: it compares the best shape grown from one start
// with the split shape (when there are two or more anchor groups).
func (b *builder) planComp(g *growCtx, comp []string, groups [][]string, inComp map[string]bool, bound map[string]bool, pending []ir.PatternPred, splitOK bool) compPlan {
	all := func(e *patEdge) bool { return inComp[e.Left] }
	var single compPlan
	singleCost := math.Inf(1)
	for _, v := range comp {
		if bound[v] { // start from a bound vertex if there is one (grow on top of the previous plan)
			c, r := b.simulate(g, bound, all, pending, b.est)
			single, singleCost = compPlan{start: v, rows: r}, c
			break
		}
	}
	if math.IsInf(singleCost, 1) {
		for _, v := range comp {
			r0 := b.startRows(g, v, pending)
			c, r := b.simulate(g, map[string]bool{v: true}, all, pending, r0)
			if c+r0 < singleCost {
				single, singleCost = compPlan{start: v, rows: r}, c+r0
			}
		}
	}
	if !splitOK || len(groups) < 2 {
		return single
	}
	side := b.splitSide(g, groups, inComp)
	hasBound := bound[groups[0][0]]
	var splitCost, accRows float64
	accBound := map[string]bool{}
	for gi, grp := range groups {
		sb := map[string]bool{}
		var r0 float64
		if gi == 0 && hasBound {
			for k := range bound {
				sb[k] = true
			}
			r0 = b.est
		} else {
			sb[grp[0]] = true
			r0 = b.startRows(g, grp[0], pending)
			splitCost += r0
		}
		gi := gi
		c, r := b.simulate(g, sb, func(e *patEdge) bool { return all(e) && side(e) == gi }, pending, r0)
		splitCost += c
		sideBound := map[string]bool{}
		for k := range sb {
			sideBound[k] = true
		}
		for _, e := range g.edges {
			if all(e) && side(e) == gi {
				sideBound[e.Left], sideBound[e.Right] = true, true
			}
		}
		if gi == 0 {
			accRows = r
		} else {
			j := accRows * r
			for _, v := range comp {
				if accBound[v] && sideBound[v] {
					j /= b.st.Count(b.labelsOf(g, v))
				}
			}
			accRows = j
			splitCost += j
		}
		for k := range sideBound {
			accBound[k] = true
		}
	}
	if splitCost < singleCost {
		return compPlan{split: true, rows: accRows}
	}
	return single
}
