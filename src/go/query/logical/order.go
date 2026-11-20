package logical

import (
	"strings"

	"polystore_database/src/go/query/ir"
)

// order.go decides, for each connected component of a pattern, whether to split the component at its
// anchors, which vertex to start from and which edge to expand next; pattern.go builds the operators
// for those decisions. With statistics the cost-based orderer (cost.go) decides; without them, rules.

// orderer decides the shape of a pattern component's plan.
type orderer interface {
	// component returns whether to split the component at its anchors and, if not, where to start
	// ("" lets the rules choose).
	component(b *builder, g *growCtx, comp []string, groups [][]string, inComp, bound map[string]bool, pending []ir.PatternPred, splitOK bool) (split bool, start string)
	// next returns the edge to expand next (nil when none is allowed and ready).
	next(b *builder, g *growCtx, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool, pending []ir.PatternPred) *patEdge
}

// ruleOrderer: anchors by binding and unique-key equality, edges in pattern order.
type ruleOrderer struct{}

func (ruleOrderer) component(b *builder, g *growCtx, comp []string, groups [][]string, inComp, bound map[string]bool, pending []ir.PatternPred, splitOK bool) (bool, string) {
	if splitOK {
		return true, ""
	}
	return false, b.chooseStart(comp, bound, pending)
}

func (ruleOrderer) next(b *builder, g *growCtx, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool, pending []ir.PatternPred) *patEdge {
	return nextEdge(g.edges, allowed, bound, done)
}

// costOrderer: the plan with the fewest estimated intermediate rows (cost.go).
type costOrderer struct{}

func (costOrderer) component(b *builder, g *growCtx, comp []string, groups [][]string, inComp, bound map[string]bool, pending []ir.PatternPred, splitOK bool) (bool, string) {
	cp := b.planComp(g, comp, groups, inComp, bound, pending, splitOK)
	if !bound[cp.start] && !cp.split || cp.split && !bound[groups[0][0]] {
		b.est *= cp.rows // a new component is a cross product with the rows so far
	} else {
		b.est = cp.rows
	}
	return cp.split, cp.start
}

func (costOrderer) next(b *builder, g *growCtx, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool, pending []ir.PatternPred) *patEdge {
	return b.pickEdge(g, allowed, bound, done, pending)
}

// anchorGroups returns a component's anchor groups: vertices bound by earlier clauses (one group, first) and
// new vertices filtered by unique-key equality (one each, in order). Two or more groups can be split.
func (b *builder) anchorGroups(comp []string, bound map[string]bool, pending []ir.PatternPred) [][]string {
	var groups [][]string
	var boundGroup []string
	for _, v := range comp {
		if bound[v] {
			boundGroup = append(boundGroup, v)
		}
	}
	if len(boundGroup) > 0 {
		groups = append(groups, boundGroup)
	}
	for _, v := range comp {
		if !bound[v] && b.uniqueAnchor(v, pending) {
			groups = append(groups, []string{v})
		}
	}
	return groups
}

// uniqueAnchor reports whether v has a unique-key equality predicate on v alone ({id: $x}).
func (b *builder) uniqueAnchor(v string, pending []ir.PatternPred) bool {
	for _, p := range pending {
		if len(p.Deps) != 1 || p.Deps[0] != v {
			continue
		}
		if _, ok := b.uniqueEq(p.Pred); ok {
			return true
		}
	}
	return false
}

// uniqueEq reports whether e is a unique-key property equal to a literal (v.id = 1) and returns the variable.
func (b *builder) uniqueEq(e ir.Expr) (string, bool) {
	bin, ok := e.(*ir.Bin)
	if !ok || bin.Op != "=" {
		return "", false
	}
	v, prop, ok := b.propOf(bin.L)
	if !ok || !b.opt.isUniqueKey(prop) {
		return "", false
	}
	if _, lit := bin.R.(*ir.Lit); !lit {
		return "", false
	}
	return v, true
}

// propOf splits a property column "v.prop" of a node or relationship variable v.
func (b *builder) propOf(e ir.Expr) (v, prop string, ok bool) {
	c, isCol := e.(*ir.Column)
	if !isCol {
		return "", "", false
	}
	i := strings.IndexByte(c.Key, '.')
	if i < 0 {
		return "", "", false
	}
	if _, entity := b.labels[c.Key[:i]]; !entity {
		return "", "", false
	}
	return c.Key[:i], c.Key[i+1:], true
}

// splitSide returns, for a split, the group that expands each edge: vertices go to the nearest anchor group
// (multi-source BFS; ties to the earlier group).
func (b *builder) splitSide(g *growCtx, groups [][]string, inComp map[string]bool) func(*patEdge) int {
	region := map[string]int{}
	var queue []string
	for gi, grp := range groups {
		for _, v := range grp {
			region[v] = gi
			queue = append(queue, v)
		}
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, e := range g.edges {
			if !inComp[e.Left] {
				continue
			}
			var w string
			switch v {
			case e.Left:
				w = e.Right
			case e.Right:
				w = e.Left
			default:
				continue
			}
			if _, ok := region[w]; !ok {
				region[w] = region[v]
				queue = append(queue, w)
			}
		}
	}
	// An edge between groups is expanded by the later group (equidistant vertices belong to the earlier group, so the later one grows to meet it in the middle).
	side := func(e *patEdge) int {
		if region[e.Left] > region[e.Right] {
			return region[e.Left]
		}
		return region[e.Right]
	}

	return side
}

// chooseStart: (1) a bound vertex, (2) the most predicates on that vertex alone, (3) pattern order.
func (b *builder) chooseStart(comp []string, bound map[string]bool, pending []ir.PatternPred) string {
	for _, v := range comp {
		if bound[v] {
			return v
		}
	}
	// Without statistics, rank by how strongly the predicates filter (compare in this order, ties go to the
	// next item, finally pattern order): unique-key equality ({id: $x}) > equality and IN > number of other predicates.
	type score struct{ unique, eq, n int }
	better := func(a, c score) bool {
		if a.unique != c.unique {
			return a.unique > c.unique
		}
		if a.eq != c.eq {
			return a.eq > c.eq
		}
		return a.n > c.n
	}
	best, bestS := comp[0], score{-1, -1, -1}
	for _, v := range comp {
		var s score
		for _, p := range pending {
			if len(p.Deps) != 1 || p.Deps[0] != v {
				continue
			}
			s.n++
			bin, ok := p.Pred.(*ir.Bin)
			if !ok || (bin.Op != "=" && bin.Op != "IN") {
				continue
			}
			s.eq++
			if _, ok := b.uniqueEq(bin); ok {
				s.unique++
			}
		}
		if better(s, bestS) {
			best, bestS = v, s
		}
	}
	return best
}

// nextEdge picks the first unprocessed allowed edge, in pattern order, that has a bound end.
func nextEdge(edges []*patEdge, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool) *patEdge {
	for _, e := range edges {
		if done[e.order] || !allowed(e) {
			continue
		}
		if bound[e.Left] || bound[e.Right] {
			return e
		}
	}
	return nil
}
