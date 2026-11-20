package physical

// build.go walks the logical plan from leaves to root, one operator at a time:
// - an operator assigned to a store joins the open query of that store when it can (merge, absorb);
// otherwise the open query is closed and a new one opens that receives the boundary ids with IN $ids;
// - a coordinator operator closes the open query and is stacked as a Coord, except that a leaf query
// (one without input) absorbs it when its store can run it.
// Values are fetched where the logical plan put GetProperty (just before first use), so deferred
// materialization follows from the logical positions and the merge rules. Finally finalize (finalize.go)
// decides the columns each query returns.

import (
	"fmt"
	"polystore_database/src/go/query/ir"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/logical"
)

// Build turns a logical plan into a physical plan. Shapes it cannot express wrap ast.ErrUnsupported.
func Build(root ir.Node, cat *Placement, opt Options) (Node, error) {
	var sets []*setSpec
	if opt.PathPredicates == PathPrefilter {
		pre := newPlanner(opt, cat)
		pre.index(root, nil)
		root, sets = pre.prefilter(root)
	}
	if opt.DeferOffGraphPredicates {
		root = logical.DeferPredicatesWhere(root, cat.offGraph)
	}
	if opt.DeferPredicates {
		root = logical.GroupFetches(root, cat.fetchGroup)
	}
	pl := newPlanner(opt, cat)
	pl.index(root, nil)
	st, err := pl.build(root)
	if err != nil {
		return nil, err
	}
	pl.close(st)
	pl.finalize(st.phys, nil, true)
	if len(sets) == 0 {
		return st.phys, nil
	}
	w := &WithSets{Input: st.phys}
	for _, s := range sets {
		sp, err := buildSet(s, cat, opt)
		if err != nil {
			return nil, err
		}
		w.Sets = append(w.Sets, sp)
	}
	return w, nil
}

func newPlanner(opt Options, cat *Placement) *planner {
	return &planner{opt: opt, cat: cat, parent: map[ir.Node]ir.Node{},
		nodeLabels: map[string][]string{}, relTypes: map[string][]string{},
		segTypes: map[string][]string{}, pathElems: map[string][]ir.PathElem{},
		owners: map[string]owner{}, vars: map[string]bool{}}
}

type planner struct {
	opt    Options
	cat    *Placement
	parent map[ir.Node]ir.Node

	// Variable types (to resolve list element labels).
	nodeLabels map[string][]string
	relTypes   map[string][]string
	segTypes   map[string][]string
	pathElems  map[string][]ir.PathElem
	owners     map[string]owner // alias -> owner of its properties (to type literals)
	vars       map[string]bool  // id columns bound by patterns (node and relationship variables)
	args       []*ArgRows       // the leaf of the OPTIONAL inside being built (picked up by buildOptional)
}

type owner struct {
	kind   ir.ObjKind
	labels []string
}

func (pl *planner) index(n, parent ir.Node) {
	if n == nil {
		return
	}
	pl.parent[n] = parent
	switch x := n.(type) {
	case *ir.EntityScan:
		pl.vars[x.Alias] = true
		pl.nodeLabels[x.Alias] = x.Labels
	case *ir.Expand:
		pl.vars[x.Rel], pl.vars[x.To] = true, true
		pl.relTypes[x.Rel] = x.Types
		if len(x.ToLabels) > 0 {
			pl.nodeLabels[x.To] = x.ToLabels
		}
	case *ir.ExpandInto:
		pl.vars[x.Rel] = true
		pl.relTypes[x.Rel] = x.Types
	case *ir.VarLengthExpand:
		pl.vars[x.To] = true
		pl.segTypes[x.Seg] = x.Types
		if len(x.ToLabels) > 0 {
			pl.nodeLabels[x.To] = x.ToLabels
		}
	case *ir.ProjectPath:
		pl.pathElems[x.Path] = x.Elems
	case *ir.Argument:
		for _, c := range x.Cols {
			pl.vars[c] = true
		}
		for k, ls := range x.Labels {
			if len(pl.nodeLabels[k]) == 0 {
				pl.nodeLabels[k] = ls
			}
		}
	case *ir.GetProperty:
		if x.Kind == ir.NodeObj && len(x.Labels) > 0 && len(pl.nodeLabels[x.Alias]) == 0 {
			pl.nodeLabels[x.Alias] = x.Labels
		}
		if _, ok := pl.owners[x.Alias]; !ok {
			pl.owners[x.Alias] = owner{kind: x.Kind, labels: x.Labels}
		}
	}
	for _, c := range n.Children() {
		pl.index(c, n)
	}
}

// state is the result of a subtree: the closed physical tree and the open query that can take more operators.
type state struct {
	phys Node
	open *openQ
}

type openQ struct {
	store   kind.Kind
	ops     ir.Node
	input   Node
	keys    []string
	leaf    bool
	avail   map[string]bool // columns in the query's rows
	idents  map[string]bool // the id columns among them
	single  string          // the only alias a non-Neo4j query handles
	grouped bool            // an aggregation is absorbed (later Filters act as HAVING)
	ship    []string        // input id columns shipped with the keys
	mode    ir.ShipMode     // the shipped boundary's mode (ir.ShipNone: plain IN)

	distinctOK bool // a WITH DISTINCT Projection was just absorbed (the following Distinct can be absorbed too)

	replace bool     // do not join results back to input rows (StoreQuery.Replace)
	redup   bool     // duplicates of an absorbed Distinct survive across input batches; the coordinator removes them on close
	dedup   []string // with redup, the columns that identify a row (the Distinct columns plus later expanded relationships and nodes)
}

func newOpen(k kind.Kind, ops ir.Node, input Node, keys []string, single string) *openQ {
	q := &openQ{store: k, ops: ops, input: input, keys: keys, leaf: input == nil && len(keys) == 0,
		avail: map[string]bool{}, idents: map[string]bool{}, single: single}
	for _, key := range keys {
		q.avail[key] = true
		q.idents[key] = true
	}
	return q
}

func (pl *planner) close(st *state) {
	if st.open == nil {
		return
	}
	o := st.open
	st.phys = &StoreQuery{StoreQuery: ir.StoreQuery{Store: o.store, Ops: o.ops, Keys: o.keys, Ship: o.ship, ShipMode: o.mode, Replace: o.replace}, Input: o.input}
	st.open = nil
	if o.redup {
		st.phys = &Coord{Op: &ir.Distinct{Input: &ir.Boundary{}, Cols: o.dedup}, Inputs: []Node{st.phys}}
	}
}

// boundary closes st and opens a new query that receives keys.
func (pl *planner) boundary(st *state, k kind.Kind, op ir.Node, keys []string, single string) {
	pl.close(st)
	// The first Neo4j query inside an OPTIONAL receives every id column of its leaf (ArgRows), so later
	// ExpandInto can use other outer variables (Neo4j receives several keys with UNWIND).
	if arg, ok := st.phys.(*ArgRows); ok && k == kind.Neo4j {
		for _, c := range arg.Cols {
			if !contains(keys, c) {
				keys = append(keys, c)
			}
		}
	}
	st.open = newOpen(k, ir.WithInput(op, &ir.Boundary{Cols: keys, Labels: pl.argLabels(keys)}), st.phys, keys, single)
	st.phys = nil
	produce(st.open, op)
}

// argLabels returns label (node) or type (relationship) candidates of the boundary columns; unknown columns are left out.
func (pl *planner) argLabels(keys []string) map[string][]string {
	m := map[string][]string{}
	for _, k := range keys {
		if ts, ok := pl.relTypes[k]; ok {
			if len(ts) > 0 {
				m[k] = ts
			}
		} else if ls := pl.nodeLabels[k]; len(ls) > 0 {
			m[k] = ls
		}
	}
	return m
}

func (pl *planner) extend(st *state, op ir.Node) {
	st.open.ops = ir.WithInput(op, st.open.ops)
	produce(st.open, op)
	if st.open.redup {
		switch x := op.(type) {
		case *ir.Expand:
			st.open.dedup = append(st.open.dedup, x.Rel, x.To)
		case *ir.ExpandInto:
			st.open.dedup = append(st.open.dedup, x.Rel)
		}
	}
}

func (pl *planner) coord(st *state, op ir.Node) {
	pl.close(st)
	st.phys = &Coord{Op: ir.WithInput(op, &ir.Boundary{}), Inputs: []Node{st.phys}}
}

func (pl *planner) build(n ir.Node) (*state, error) {
	switch x := n.(type) {
	case *ir.EntityScan:
		k, single := pl.scanStore(x)
		o := newOpen(k, x, nil, nil, single)
		o.avail[x.Alias], o.idents[x.Alias] = true, true
		return &state{open: o}, nil

	case *ir.Argument:
		arg := &ArgRows{Cols: x.Cols}
		pl.args = append(pl.args, arg)
		return &state{phys: arg}, nil

	case *ir.Optional:
		return pl.buildOptional(x)

	case *ir.NodeHashJoin:
		l, err := pl.build(x.Left)
		if err != nil {
			return nil, err
		}
		r, err := pl.build(x.Right)
		if err != nil {
			return nil, err
		}
		if l.open != nil && r.open != nil && l.open.leaf && r.open.leaf &&
			l.open.store == kind.Neo4j && r.open.store == kind.Neo4j &&
			graphPattern(l.open.ops) && graphPattern(r.open.ops) {
			// When both sides are Neo4j patterns only, write one MATCH (a side that absorbed a projection or
			// aggregation cannot be part of a MATCH, so the coordinator takes the product).
			o := l.open
			o.ops = &ir.NodeHashJoin{Left: l.open.ops, Right: r.open.ops, Keys: x.Keys}
			for k := range r.open.avail {
				o.avail[k] = true
			}
			for k := range r.open.idents {
				o.idents[k] = true
			}
			return &state{open: o}, nil
		}
		pl.close(l)
		pl.close(r)
		return &state{phys: &Coord{
			Op:     &ir.NodeHashJoin{Left: &ir.Boundary{}, Right: &ir.Boundary{}, Keys: x.Keys},
			Inputs: []Node{l.phys, r.phys},
		}}, nil
	}

	c := ir.Child(n)
	if c == nil {
		return nil, unsupported("leaf operator %T", n)
	}
	st, err := pl.build(c)
	if err != nil {
		return nil, err
	}
	if err := pl.add(st, n); err != nil {
		return nil, err
	}
	return st, nil
}

// buildOptional builds an OPTIONAL MATCH: one Cypher OPTIONAL MATCH when outer and inner are Neo4j only,
// otherwise an OptionalApply (the outer rows go to the inner leaf ArgRows).
func (pl *planner) buildOptional(x *ir.Optional) (*state, error) {
	outer, err := pl.build(x.Outer)
	if err != nil {
		return nil, err
	}
	mark := len(pl.args)
	inner, err := pl.build(x.Inner)
	if err != nil {
		return nil, err
	}
	if len(pl.args) != mark+1 {
		return nil, unsupported("OPTIONAL MATCH inner plan must have one argument leaf")
	}
	arg := pl.args[mark]
	pl.args = pl.args[:mark]

	// Absorbed into Neo4j: the outer side is an open Neo4j leaf query and the inner side one Neo4j query on
	// ArgRows (all inner id columns are in the outer query).
	if o, in := outer.open, inner.open; o != nil && in != nil && o.leaf && o.store == kind.Neo4j &&
		in.store == kind.Neo4j && in.input == Node(arg) && inner.phys == nil && identAll(o, x.Keys) && graphPattern(in.ops) {
		o.ops = &ir.Optional{Outer: o.ops, Inner: in.ops, Keys: x.Keys, New: x.New}
		for k := range in.avail {
			o.avail[k] = true
		}
		for k := range in.idents {
			o.idents[k] = true
		}
		return &state{open: o}, nil
	}
	pl.close(outer)
	pl.close(inner)
	return &state{phys: &OptionalApply{Outer: outer.phys, Inner: inner.phys, Arg: arg, Keys: x.Keys, New: x.New}}, nil
}

// graphPattern reports whether the subtree is only patterns and conditions one OPTIONAL MATCH clause can hold.
func graphPattern(n ir.Node) bool {
	ok := true
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil || !ok {
			return
		}
		switch n.(type) {
		case *ir.Boundary, *ir.EntityScan, *ir.Expand, *ir.ExpandInto, *ir.VarLengthExpand,
			*ir.Filter, *ir.GetProperty, *ir.RelUniqueFilter, *ir.NodeHashJoin:
		default:
			ok = false
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return ok
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ast.ErrUnsupported, fmt.Sprintf(format, args...))
}

// scanStore picks the store of an EntityScan from the GetProperty/Filter chain right above it:
// - the store of the columns the first Filter reads;
// - without a Filter, the store of the chain's first GetProperty (its column is there, so it can be absorbed);
// - otherwise Neo4j.
func (pl *planner) scanStore(s *ir.EntityScan) (kind.Kind, string) {
	for _, l := range s.Labels {
		if !pl.cat.IsEntity(l) {
			return kind.Neo4j, "" // only Neo4j can scan a sub-label
		}
	}
	if len(s.Labels) == 0 {
		return kind.Neo4j, ""
	}
	props := map[string]*ir.GetProperty{}
	var first *ir.GetProperty
	pick := func(g *ir.GetProperty) (kind.Kind, string) {
		k := pl.cat.StoreOf(g.Alias, g.Kind, g.Labels, g.Prop)
		if k == kind.Neo4j {
			return k, ""
		}
		return k, s.Alias
	}
walk:
	for p := pl.parent[s]; p != nil; p = pl.parent[p] {
		switch x := p.(type) {
		case *ir.GetProperty:
			if x.Alias != s.Alias {
				break walk
			}
			props[x.Key()] = x
			if first == nil {
				first = x
			}
		case *ir.Filter:
			for _, col := range ir.Columns(x.Pred) {
				if g, ok := props[col]; ok {
					return pick(g)
				}
			}
			break walk
		default:
			break walk
		}
	}
	if first != nil {
		return pick(first)
	}
	return kind.Neo4j, ""
}

func (pl *planner) add(st *state, op ir.Node) error {
	pl.typeLiterals(op)
	if g, ok := op.(*ir.GetProperty); ok && g.Kind == ir.NodeObj {
		g.Labels = pl.cat.Entities(g.Labels) // tables and collections are per entity (Post -> Message)
	}
	if st.open != nil && st.open.mode != ir.ShipNone {
		// A shipped aggregation query ends at the aggregation (stores that emulate it rely on the aggregation being on top).
		pl.close(st)
	}
	if st.open != nil && st.open.redup && !redupSafe(op) {
		pl.close(st)
	}
	o := st.open
	switch x := op.(type) {
	case *ir.GetProperty:
		k := pl.cat.StoreOf(x.Alias, x.Kind, x.Labels, x.Prop)
		if o != nil && o.store == k && o.idents[x.Alias] && (k == kind.Neo4j || o.single == x.Alias) {
			pl.extend(st, op)
			return nil
		}
		single := x.Alias
		if k == kind.Neo4j {
			single = ""
		}
		pl.boundary(st, k, op, []string{x.Alias}, single)
		return nil

	case *ir.Filter:
		having := o != nil && o.grouped && (o.store == kind.Cassandra || o.store.IsKV())
		if o != nil && !having && availAll(o, filterRefs(x)) && canExpress(o.store, x.Pred) &&
			(o.store == kind.Neo4j || !pl.complexRef(filterRefs(x))) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *ir.Expand:
		return pl.addGraph(st, op, []string{x.From})
	case *ir.ExpandInto:
		return pl.addGraph(st, op, []string{x.From, x.To})
	case *ir.VarLengthExpand:
		return pl.addVarLength(st, x)

	case *ir.ProjectPath:
		if o != nil && o.store == kind.Neo4j && availAll(o, ir.Refs(x)) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *ir.GetListProperty:
		labels := x.ElemLabels
		if len(labels) == 0 {
			labels = pl.listLabels(x.Source, x.Part)
		}
		k, err := pl.listStore(labels, x.Part, x.Prop)
		if err != nil {
			return err
		}
		if o != nil && o.store == kind.Neo4j && k == kind.Neo4j && o.avail[x.Source] {
			pl.extend(st, op)
			return nil
		}
		pl.close(st)
		if x.Part == ir.SegNodes {
			labels = pl.cat.Entities(labels)
		}
		st.phys = &ListFetch{Input: st.phys, Store: k, Source: x.Source, Part: x.Part, Prop: x.Prop, Labels: labels}
		return nil

	case *ir.RelUniqueFilter:
		if o != nil && o.store == kind.Neo4j && availAll(o, ir.Refs(x)) && !anyKey(o, ir.Refs(x)) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *ir.Unwind:
		pl.coord(st, op)
		return nil

	case *ir.Projection, *ir.Aggregation, *ir.Sort, *ir.Limit, *ir.Distinct:
		if o != nil && o.leaf && canAbsorb(o.store, op) && availAll(o, ir.Refs(op)) && !collectsIdent(o, op) {
			pl.extend(st, op)
			return nil
		}
		if pl.absorbDistinct(o, op) {
			pl.extend(st, op)
			return nil
		}
		if abs := pl.distinctOverOneRow(o, op); abs != nil {
			pl.extend(st, abs)
			return nil
		}
		if agg, ok := op.(*ir.Aggregation); ok && pl.opt.ShipRows {
			if ship, ok := pl.canShipAggregate(o, agg); ok {
				pl.ship(st, op, ir.ShipAggregate, ship)
				return nil
			}
		}
		pl.coord(st, op)
		return nil
	}
	return unsupported("operator %T", op)
}
