package logical

// rewrite.go rewrites logical plans without looking at stores:
// - DeferOutputFetch moves GetProperty used only for output after Sort / Limit (deferred materialization):
// in `RETURN f.firstName ORDER BY f.id LIMIT 10` firstName is fetched for the 10 rows only.
// - Prune removes GetProperty / GetListProperty / ProjectPath whose columns are not used above.
// - DeferPredicates (Options.DeferPredicates) moves non-anchor fetches and filters after the expansions.
// - SplitAggregation (Options.SplitAggregation) splits an aggregation into partial aggregations per key owner and a final one.
//
// Literal types are fixed by the AST (int64 / float64 / string): parameters are substituted before parsing,
// so they need no typing here. Conversion to store types happens when compiling native queries.

import (
	"fmt"
	"polystore_database/src/go/query/ir"
	"slices"
)

// Rewrite applies the logical rewrites enabled in opt.
func Rewrite(n ir.Node, opt Options) ir.Node {
	if opt.DeferPredicates {
		n = DeferPredicates(n)
	}
	if opt.SplitAggregation {
		n = SplitAggregation(n)
	}
	n = DeferOutputFetch(n)
	return Prune(n)
}

// ================================
// Separating traversal from filtering
// ================================

// DeferPredicates moves, in each chain of unary operators, the GetProperty / GetListProperty / Filter above the
// first expansion to just above the last expansion (keeping their order). Those below the first expansion
// (filters on the start) stay. An expansion only adds columns to its input rows (an inner join), so swapping
// it with a filter on columns available below gives the same result. The relative order of expansions,
// segments, paths and relationship uniqueness (Expand / ExpandInto / VarLengthExpand / ProjectPath /
// RelUniqueFilter) is kept; other operators (Projection, Aggregation, joins, Optional, ...) end a chain.
func DeferPredicates(n ir.Node) ir.Node {
	return DeferPredicatesWhere(n, func(*ir.GetProperty) bool { return true })
}

// DeferPredicatesWhere applies the DeferPredicates rewrite only to GetProperty for which late is true and the
// Filters that read their values (others stay). The caller (physical) decides late from the stores.
// A nil late moves every GetProperty (DeferPredicates).
func DeferPredicatesWhere(n ir.Node, late func(*ir.GetProperty) bool) ir.Node {
	if n == nil {
		return nil
	}
	var chain []ir.Node // top first
	cur := n
	for isTraversal(cur) || isPredicate(cur) {
		chain = append(chain, cur)
		cur = ir.Child(cur)
	}
	bottom := deferPredicatesBelow(cur, late)
	if len(chain) == 0 {
		return bottom
	}
	top, lowest := -1, -1 // the topmost and bottommost expansions in the chain
	for i, c := range chain {
		if isTraversal(c) {
			if top < 0 {
				top = i
			}
			lowest = i
		}
	}
	if lowest < 0 {
		return relink(chain, bottom)
	}
	// Values are made below, so decide the moved operators from the bottom up.
	moved := map[ir.Node]bool{}
	cols := map[string]bool{}
	for i := lowest; i >= top; i-- {
		switch x := chain[i].(type) {
		case *ir.GetProperty:
			if late(x) {
				moved[x] = true
				cols[x.Key()] = true
			}
		case *ir.GetListProperty:
			if late(nil) {
				moved[x] = true
				cols[x.Key()] = true
			}
		case *ir.Filter:
			if late(nil) || refersAny(x, cols) {
				moved[x] = true
			}
		}
	}
	reordered := append([]ir.Node{}, chain[:top]...)
	var stay []ir.Node
	for _, c := range chain[top : lowest+1] {
		if moved[c] {
			reordered = append(reordered, c)
		} else {
			stay = append(stay, c)
		}
	}
	reordered = append(append(reordered, stay...), chain[lowest+1:]...)
	return relink(reordered, bottom)
}

func refersAny(n ir.Node, cols map[string]bool) bool {
	for _, c := range ir.Refs(n) {
		if cols[c] {
			return true
		}
	}
	return false
}

// deferPredicatesBelow continues the rewrite below the operator that ended a chain.
func deferPredicatesBelow(n ir.Node, late func(*ir.GetProperty) bool) ir.Node {
	switch x := n.(type) {
	case nil:
		return nil
	case *ir.NodeHashJoin:
		x.Left = DeferPredicatesWhere(x.Left, late)
		x.Right = DeferPredicatesWhere(x.Right, late)
		return x
	case *ir.Optional:
		x.Outer = DeferPredicatesWhere(x.Outer, late)
		x.Inner = DeferPredicatesWhere(x.Inner, late)
		return x
	}
	if c := ir.Child(n); c != nil {
		ir.SetChild(n, DeferPredicatesWhere(c, late))
	}
	return n
}

// GroupFetches reorders each chain of GetProperty and Filter so fetches of the same group are adjacent:
// group "" first, the rest in order of first appearance, and each Filter right after the values it reads.
// Each step only adds values or drops rows, so the order does not change the result, and the physical planner can merge more into one query.
func GroupFetches(n ir.Node, group func(*ir.GetProperty) string) ir.Node {
	if n == nil {
		return nil
	}
	var chain []ir.Node // top first
	cur := n
	for {
		switch cur.(type) {
		case *ir.GetProperty, *ir.Filter:
			chain = append(chain, cur)
			cur = ir.Child(cur)
			continue
		}
		break
	}
	var bottom ir.Node
	switch x := cur.(type) {
	case nil:
	case *ir.NodeHashJoin:
		x.Left = GroupFetches(x.Left, group)
		x.Right = GroupFetches(x.Right, group)
		bottom = x
	case *ir.Optional:
		x.Outer = GroupFetches(x.Outer, group)
		x.Inner = GroupFetches(x.Inner, group)
		bottom = x
	default:
		if c := ir.Child(cur); c != nil {
			ir.SetChild(cur, GroupFetches(c, group))
		}
		bottom = cur
	}
	if len(chain) == 0 {
		return bottom
	}
	// Read from the bottom up to order the groups by first appearance.
	order := []string{""}
	members := map[string][]*ir.GetProperty{}
	var filters []*ir.Filter
	for i := len(chain) - 1; i >= 0; i-- {
		switch x := chain[i].(type) {
		case *ir.GetProperty:
			g := group(x)
			if !slices.Contains(order, g) {
				order = append(order, g)
			}
			members[g] = append(members[g], x)
		case *ir.Filter:
			filters = append(filters, x)
		}
	}
	var sorted []*ir.GetProperty
	for _, g := range order {
		sorted = append(sorted, members[g]...)
	}
	// Stack from the bottom. A Filter goes to the bottom when none of its values is in the chain, else right after the last of them.
	last := map[string]int{}
	for i, g := range sorted {
		last[g.Key()] = i
	}
	after := map[int][]*ir.Filter{}
	for _, f := range filters {
		at := -1
		for _, c := range ir.Refs(f) {
			if i, ok := last[c]; ok && i > at {
				at = i
			}
		}
		after[at] = append(after[at], f)
	}
	var up []ir.Node // bottom first
	for _, f := range after[-1] {
		up = append(up, f)
	}
	for i, g := range sorted {
		up = append(up, g)
		for _, f := range after[i] {
			up = append(up, f)
		}
	}
	slices.Reverse(up)
	return relink(up, bottom)
}

// relink stacks the top-first chain on bottom again and returns the top.
func relink(chain []ir.Node, bottom ir.Node) ir.Node {
	in := bottom
	for i := len(chain) - 1; i >= 0; i-- {
		ir.SetChild(chain[i], in)
		in = chain[i]
	}
	return in
}

func isTraversal(n ir.Node) bool {
	switch n.(type) {
	case *ir.Expand, *ir.ExpandInto, *ir.VarLengthExpand, *ir.ProjectPath, *ir.RelUniqueFilter:
		return true
	}
	return false
}

func isPredicate(n ir.Node) bool {
	switch n.(type) {
	case *ir.GetProperty, *ir.GetListProperty, *ir.Filter:
		return true
	}
	return false
}

// ================================
// Partial aggregation (eager aggregation)
// ================================

// SplitAggregation splits an aggregation into a partial aggregation per key owner id, the GetProperty of the
// key values, and a final aggregation:
//
//	Aggregation[K | f(x) AS n]                     [Projection (s/c for avg)]
//	  ...GetProperty(a.p)...        =>             Aggregation[K | merge(_p) AS n]
//	                                                  GetProperty(a.p)...
//	                                                    Aggregation[a | partial(x) AS _p]
//	                                                      ...(input without the key GetProperty)...
//
// A property is determined by its owner's id (a functional dependency), so combining per-owner partial
// aggregates by K gives the same result. It applies when every function is count / sum / min / max / avg
// without DISTINCT and every key is an id column or the value of a GetProperty in the chain right below that
// nothing but the aggregation reads. Otherwise the aggregation stays.
func SplitAggregation(n ir.Node) ir.Node {
	s := &aggSplitter{}
	return s.rewrite(n)
}

type aggSplitter struct{ seq int }

func (s *aggSplitter) rewrite(n ir.Node) ir.Node {
	switch x := n.(type) {
	case nil:
		return nil
	case *ir.NodeHashJoin:
		x.Left = s.rewrite(x.Left)
		x.Right = s.rewrite(x.Right)
		return x
	case *ir.Optional:
		x.Outer = s.rewrite(x.Outer)
		x.Inner = s.rewrite(x.Inner)
		return x
	}
	if c := ir.Child(n); c != nil {
		ir.SetChild(n, s.rewrite(c))
	}
	if a, ok := n.(*ir.Aggregation); ok {
		if r, ok := s.split(a); ok {
			return r
		}
	}
	return n
}

// splitPartial says how each aggregate function splits into partial and final aggregates.
var splitPartial = map[string]string{"count": "sum", "sum": "sum", "min": "min", "max": "max", "avg": ""}

func (s *aggSplitter) split(a *ir.Aggregation) (ir.Node, bool) {
	for _, g := range a.Aggs {
		if _, ok := splitPartial[g.Fn]; !ok || g.Distinct {
			return nil, false
		}
	}
	// The chain of unary operators right below (where key GetProperty are looked for).
	var chain []ir.Node // top first
	for cur := a.Input; isTraversal(cur) || isPredicate(cur); cur = ir.Child(cur) {
		chain = append(chain, cur)
	}
	bottom := a.Input
	if len(chain) > 0 {
		bottom = ir.Child(chain[len(chain)-1])
	}
	byKey := map[string]*ir.GetProperty{}
	for _, c := range chain {
		if g, ok := c.(*ir.GetProperty); ok {
			byKey[g.Key()] = g
		}
	}

	// Key -> owner id column. Value keys come from GetProperty in the chain; id keys are bound below.
	var owners []string
	moved := map[*ir.GetProperty]bool{}
	addOwner := func(o string) {
		if !contains(owners, o) {
			owners = append(owners, o)
		}
	}
	for _, k := range a.Keys {
		c, ok := k.Expr.(*ir.Column)
		if !ok {
			return nil, false
		}
		if g, ok := byKey[c.Key]; ok {
			moved[g] = true
			addOwner(g.Alias)
			continue
		}
		if !binds(a.Input, c.Key) {
			return nil, false
		}
		addOwner(c.Key)
	}
	// The moved GetProperty values must not be used by the aggregate arguments or other operators in the chain.
	movedKeys := map[string]bool{}
	for g := range moved {
		movedKeys[g.Key()] = true
	}
	for _, g := range a.Aggs {
		if g.Arg != nil {
			for _, c := range ir.Columns(g.Arg) {
				if movedKeys[c] {
					return nil, false
				}
			}
		}
	}
	for _, c := range chain {
		if g, ok := c.(*ir.GetProperty); ok && moved[g] {
			continue
		}
		for _, r := range ir.Refs(c) {
			if movedKeys[r] {
				return nil, false
			}
		}
	}

	// Input of the partial aggregation (the chain without the moved GetProperty).
	var rest []ir.Node
	for _, c := range chain {
		if g, ok := c.(*ir.GetProperty); ok && moved[g] {
			continue
		}
		rest = append(rest, c)
	}
	input := relink(rest, bottom)

	s.seq++
	name := func(i int, part string) string { return fmt.Sprintf("_p%d_%d%s", s.seq, i, part) }
	partial := &ir.Aggregation{Input: input, Partial: true}
	for _, o := range owners {
		partial.Keys = append(partial.Keys, ir.ProjItem{Name: o, Expr: &ir.Column{Key: o}})
	}
	final := &ir.Aggregation{Keys: a.Keys}
	var avgs []ir.ProjItem
	for i, g := range a.Aggs {
		if g.Fn == "avg" {
			sum, cnt := name(i, "s"), name(i, "c")
			partial.Aggs = append(partial.Aggs,
				ir.AggItem{Name: sum, Fn: "sum", Arg: g.Arg},
				ir.AggItem{Name: cnt, Fn: "count", Arg: g.Arg})
			final.Aggs = append(final.Aggs,
				ir.AggItem{Name: sum, Fn: "sum", Arg: &ir.Column{Key: sum}},
				ir.AggItem{Name: cnt, Fn: "sum", Arg: &ir.Column{Key: cnt}})
			avgs = append(avgs, ir.ProjItem{Name: g.Name, Expr: avgExpr(sum, cnt)})
			continue
		}
		p := name(i, "")
		partial.Aggs = append(partial.Aggs, ir.AggItem{Name: p, Fn: g.Fn, Arg: g.Arg})
		final.Aggs = append(final.Aggs, ir.AggItem{Name: g.Name, Fn: splitPartial[g.Fn], Arg: &ir.Column{Key: p}})
	}

	// Key values are fetched after the partial aggregation (in their original order).
	upper := ir.Node(partial)
	for i := len(chain) - 1; i >= 0; i-- {
		if g, ok := chain[i].(*ir.GetProperty); ok && moved[g] {
			cp := *g
			cp.Input = upper
			upper = &cp
		}
	}
	final.Input = upper
	if len(avgs) == 0 {
		return final, true
	}
	// Turn avg back into s / c and output only the original columns (keys and aggregate names).
	items := make([]ir.ProjItem, 0, len(a.Keys)+len(a.Aggs))
	for _, k := range a.Keys {
		items = append(items, ir.ProjItem{Name: k.Name, Expr: &ir.Column{Key: k.Name}})
	}
	for _, g := range a.Aggs {
		if g.Fn == "avg" {
			items = append(items, avgs[0])
			avgs = avgs[1:]
			continue
		}
		items = append(items, ir.ProjItem{Name: g.Name, Expr: &ir.Column{Key: g.Name}})
	}
	return &ir.Projection{Input: final, Items: items}, true
}

// avgExpr is CASE WHEN c = 0 THEN null ELSE toFloat(s) / c END (avg of an empty group is null).
func avgExpr(sum, cnt string) ir.Expr {
	return &ir.Case{
		Whens: []ir.CaseWhen{{When: &ir.Bin{Op: "=", L: &ir.Column{Key: cnt}, R: &ir.Lit{Val: int64(0)}}, Then: &ir.Lit{Val: nil}}},
		Else:  &ir.Bin{Op: "/", L: &ir.Call{Fn: "tofloat", Args: []ir.Expr{&ir.Column{Key: sum}}}, R: &ir.Column{Key: cnt}},
	}
}

// binds reports whether subtree n binds id column v (to check that an id key is the owner).
func binds(n ir.Node, v string) bool {
	found := false
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil || found {
			return
		}
		switch x := n.(type) {
		case *ir.EntityScan:
			found = x.Alias == v
		case *ir.Expand:
			found = x.To == v || x.Rel == v
		case *ir.ExpandInto:
			found = x.Rel == v
		case *ir.VarLengthExpand:
			found = x.To == v
		case *ir.Argument:
			found = contains(x.Cols, v)
		case *ir.Projection, *ir.Aggregation:
			// An id column of the same name may not survive a projection or aggregation, so stop here.
			if !passesIdent(n, v) {
				return
			}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return found
}

// passesIdent reports whether a projection or aggregation outputs column v unchanged (Column(v) -> v).
func passesIdent(n ir.Node, v string) bool {
	var items []ir.ProjItem
	switch x := n.(type) {
	case *ir.Projection:
		items = x.Items
	case *ir.Aggregation:
		items = x.Keys
	}
	for _, it := range items {
		if c, ok := it.Expr.(*ir.Column); ok && it.Name == v && c.Key == v {
			return true
		}
	}
	return false
}

// ================================
// Deferred materialization
// ================================

// DeferOutputFetch handles "GetProperty* -> Projection -> [Sort] -> [Limit] -> [Projection]": GetProperty of
// columns the Projection only passes through move after the Limit. Not when a Distinct is in between (it
// would change the deduplicated columns).
func DeferOutputFetch(n ir.Node) ir.Node {
	if n == nil {
		return nil
	}
	// Look top down (apply once at the top of the Sort/Limit chain).
	if r, below, ok := deferAt(n); ok {
		if below != nil {
			ir.SetChild(below, DeferOutputFetch(ir.Child(below)))
		}
		return r
	}
	switch x := n.(type) {
	case *ir.NodeHashJoin:
		x.Left = DeferOutputFetch(x.Left)
		x.Right = DeferOutputFetch(x.Right)
		return x
	case *ir.Optional:
		x.Outer = DeferOutputFetch(x.Outer)
		x.Inner = DeferOutputFetch(x.Inner)
		return x
	}
	if c := ir.Child(n); c != nil {
		ir.SetChild(n, DeferOutputFetch(c))
	}
	return n
}

// deferAt applies deferred materialization to the shape topped by n. When applied it returns the new root and
// the lower Projection whose input the recursion continues from.
func deferAt(n ir.Node) (ir.Node, ir.Node, bool) {
	// With n as the top Sort/Limit/trim, find the Projection below it.
	var trim *ir.Projection
	top := n
	if p, ok := top.(*ir.Projection); ok && isPassThrough(p) {
		trim = p
		top = p.Input
	}
	var sortLimit []ir.Node // top first
	for {
		switch x := top.(type) {
		case *ir.Limit:
			sortLimit = append(sortLimit, x)
			top = x.Input
			continue
		case *ir.Sort:
			sortLimit = append(sortLimit, x)
			top = x.Input
			continue
		}
		break
	}
	proj, ok := top.(*ir.Projection)
	if !ok || len(sortLimit) == 0 {
		return n, nil, false
	}

	// The consecutive GetProperty right below the Projection that are used only as pass-through outputs.
	// Items a Sort reads (sort keys) do not count as pass-through.
	sortRefs := map[string]bool{}
	for _, sl := range sortLimit {
		if s, ok := sl.(*ir.Sort); ok {
			for _, k := range s.Keys {
				for _, c := range ir.Columns(k.Expr) {
					sortRefs[c] = true
				}
			}
		}
	}
	used := map[string]int{} // references to a column inside the Projection (other than pass-through)
	passItem := map[string]int{}
	for i, it := range proj.Items {
		if c, ok := it.Expr.(*ir.Column); ok && !sortRefs[it.Name] {
			if _, dup := passItem[c.Key]; !dup {
				passItem[c.Key] = i
				continue
			}
		}
		for _, k := range ir.Columns(it.Expr) {
			used[k]++
		}
	}

	var deferred []*ir.GetProperty
	below := proj.Input
	var keep []*ir.GetProperty
	for {
		g, ok := below.(*ir.GetProperty)
		if !ok {
			break
		}
		_, isPass := passItem[g.Key()]
		if isPass && used[g.Key()] == 0 {
			deferred = append(deferred, g)
		} else {
			keep = append(keep, g)
		}
		below = g.Input
	}
	if len(deferred) == 0 {
		return n, nil, false
	}

	// Rebuild the lower part (kept GetProperty in their original order).
	rebuilt := below
	for i := len(keep) - 1; i >= 0; i-- {
		g := *keep[i]
		g.Input = rebuilt
		rebuilt = &g
	}
	proj.Input = rebuilt

	// Remove the moved columns' items from the Projection and carry their id columns instead.
	outName := map[string]string{} // column key -> original item name
	order := make([]string, 0, len(proj.Items))
	for _, it := range proj.Items {
		order = append(order, it.Name)
	}
	isDeferred := map[string]bool{}
	for _, g := range deferred {
		isDeferred[g.Key()] = true
	}
	var items []ir.ProjItem
	names := map[string]bool{}
	for _, it := range proj.Items {
		if c, ok := it.Expr.(*ir.Column); ok && isDeferred[c.Key] && passItem[c.Key] == indexOf(proj.Items, it.Name) {
			outName[c.Key] = it.Name
			continue
		}
		items = append(items, it)
		names[it.Name] = true
	}
	for _, g := range deferred {
		if !names[g.Alias] {
			items = append(items, ir.ProjItem{Name: g.Alias, Expr: &ir.Column{Key: g.Alias}})
			names[g.Alias] = true
		}
	}
	proj.Items = items

	// Stack the GetProperty again above Sort/Limit and restore the original item order with a final Projection.
	upper := ir.Node(sortLimit[0])
	for i := len(deferred) - 1; i >= 0; i-- {
		g := *deferred[i]
		g.Input = upper
		upper = &g
	}
	final := make([]ir.ProjItem, 0, len(order))
	for _, name := range order {
		key := name
		for col, on := range outName {
			if on == name {
				key = col
			}
		}
		final = append(final, ir.ProjItem{Name: name, Expr: &ir.Column{Key: key}})
	}
	if trim != nil {
		// Keep the existing trim's item names (the original output names) and repoint only the moved columns.
		for i, it := range trim.Items {
			for col, on := range outName {
				if c, ok := it.Expr.(*ir.Column); ok && c.Key == on {
					trim.Items[i].Expr = &ir.Column{Key: col}
				}
			}
		}
		trim.Input = upper
		return trim, proj, true
	}
	return &ir.Projection{Input: upper, Items: final}, proj, true
}

func indexOf(items []ir.ProjItem, name string) int {
	for i, it := range items {
		if it.Name == name {
			return i
		}
	}
	return -1
}

// isPassThrough reports whether a Projection only passes columns through (Column(name) -> name).
func isPassThrough(p *ir.Projection) bool {
	for _, it := range p.Items {
		c, ok := it.Expr.(*ir.Column)
		if !ok || c.Key != it.Name {
			return false
		}
	}
	return true
}

// ================================
// Pruning
// ================================

// Prune removes GetProperty / GetListProperty / ProjectPath whose columns are not used above.
func Prune(n ir.Node) ir.Node {
	return prune(n, nil, true)
}

// prune takes need (all columns when nil and all is true) and returns the tree without unneeded operators.
func prune(n ir.Node, need map[string]bool, all bool) ir.Node {
	if n == nil {
		return nil
	}
	needs := func(k string) bool { return all || need[k] }
	switch x := n.(type) {
	case *ir.GetProperty:
		if !needs(x.Key()) {
			return prune(x.Input, need, all)
		}
	case *ir.GetListProperty:
		if !needs(x.Key()) {
			return prune(x.Input, need, all)
		}
	case *ir.ProjectPath:
		if !needs(x.Path) {
			return prune(x.Input, need, all)
		}
	}

	// Columns required from the child.
	below := map[string]bool{}
	belowAll := false
	switch x := n.(type) {
	case *ir.Projection:
		for _, it := range x.Items {
			for _, k := range ir.Columns(it.Expr) {
				below[k] = true
			}
		}
	case *ir.Aggregation:
		for _, it := range x.Keys {
			for _, k := range ir.Columns(it.Expr) {
				below[k] = true
			}
		}
		for _, a := range x.Aggs {
			if a.Arg != nil {
				for _, k := range ir.Columns(a.Arg) {
					below[k] = true
				}
			}
		}
	case *ir.Distinct:
		belowAll = all
		for k := range need {
			below[k] = true
		}
	default:
		belowAll = all
		for k := range need {
			below[k] = true
		}
		for _, k := range ir.Refs(n) {
			below[k] = true
		}
	}

	switch x := n.(type) {
	case *ir.NodeHashJoin:
		x.Left = prune(x.Left, below, belowAll)
		x.Right = prune(x.Right, below, belowAll)
	case *ir.Optional:
		x.Outer = prune(x.Outer, below, belowAll)
		x.Inner = prune(x.Inner, below, belowAll)
	default:
		if c := ir.Child(n); c != nil {
			ir.SetChild(n, prune(c, below, belowAll))
		}
	}
	return n
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
