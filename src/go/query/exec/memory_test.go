package exec

import (
	"context"
	"crypto/md5"
	"fmt"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"strings"
)

// memory_test.go is a test backend that evaluates a StoreQuery's logical subtree directly on a small graph.
// It does not separate stores (every column is here), so every placement must give the same result; that
// checks the planner and the executor (joins, fetches, ListFetch, coordinator operators).

type memNode struct {
	labels []string
	props  map[string]any
}

type memRel struct {
	typ      string
	src, dst id.UUID
	props    map[string]any
}

type memGraph struct {
	nodes map[id.UUID]*memNode
	rels  map[id.UUID]*memRel
	order []id.UUID // a deterministic node order
	rord  []id.UUID
}

func newMemGraph() *memGraph {
	return &memGraph{nodes: map[id.UUID]*memNode{}, rels: map[id.UUID]*memRel{}}
}

// tid is a deterministic test id (md5 of a kind letter and a name; nodes and relationships never collide).
func tid(k, name string) id.UUID { return id.Make(md5.Sum([]byte(k + name))) }

func (g *memGraph) node(uuid string, label string, props map[string]any) id.UUID {
	u := tid("N", uuid)
	for k, v := range props {
		props[k] = eval.Normalize(v)
	}
	g.nodes[u] = &memNode{labels: []string{label}, props: props}
	g.order = append(g.order, u)
	return u
}

func (g *memGraph) rel(uuid, typ string, src, dst id.UUID, props map[string]any) {
	u := tid("E", uuid)
	if props == nil {
		props = map[string]any{}
	}
	for k, v := range props {
		props[k] = eval.Normalize(v)
	}
	g.rels[u] = &memRel{typ: typ, src: src, dst: dst, props: props}
	g.rord = append(g.rord, u)
}

func hasAny(xs []string, ys []string) bool {
	if len(ys) == 0 {
		return true
	}
	for _, x := range xs {
		for _, y := range ys {
			if x == y {
				return true
			}
		}
	}
	return false
}

func typeOK(r *memRel, types []string) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if r.typ == t {
			return true
		}
	}
	return false
}

// step lists the (relationship, neighbor) pairs reachable from from in direction dir.
func (g *memGraph) step(from id.UUID, types []string, dir ir.Direction) [][2]id.UUID {
	var out [][2]id.UUID
	for _, ru := range g.rord {
		r := g.rels[ru]
		if !typeOK(r, types) {
			continue
		}
		if (dir == ir.Out || dir == ir.Both) && r.src == from {
			out = append(out, [2]id.UUID{ru, r.dst})
		}
		if (dir == ir.In || dir == ir.Both) && r.dst == from && !(dir == ir.Both && r.src == from) {
			out = append(out, [2]id.UUID{ru, r.src})
		}
	}
	return out
}

func (g *memGraph) prop(u id.UUID, p string) any {
	if n, ok := g.nodes[u]; ok {
		return n.props[p]
	}
	if r, ok := g.rels[u]; ok {
		return r.props[p]
	}
	return nil
}

// ---- Backend ----

type memBackend struct {
	g    *memGraph
	ship []Row // input of a shipped boundary (only while running a shipped query; Boundary returns these rows)
}

// memPlan is a "compiled" query: the logical subtree, evaluated on the graph when run.
type memPlan struct{ q *ir.StoreQuery }

func (p *memPlan) Render(b *strings.Builder) { fmt.Fprintf(b, "  memory %s\n", p.q.Ops.Describe()) }

// Compile keeps the logical subtree as the query body. Results come back in Outputs order (with the weight
// for a weighted query); a shipped query's table is Keys, Ship and the weight (the executor's order).
func (b *memBackend) Compile(q *ir.StoreQuery, _ stores.TypeFn) (*stores.Query, error) {
	cols := append([]string{}, q.Outputs...)
	if q.Weighted {
		cols = append(cols, ir.WeightCol)
	}
	out := &stores.Query{Store: q.Store, Columns: cols, InKeys: q.Keys, Body: &memPlan{q: q}}
	if q.ShipMode != ir.ShipNone {
		out.ShipMode = q.ShipMode
		out.ShipCols = append(append(append([]string{}, q.Keys...), q.Ship...), ir.WeightCol)
	}
	return out, nil
}

func (b *memBackend) Run(_ context.Context, nq *stores.Query, in []Key) ([][]any, error) {
	q := nq.Body.(*memPlan).q
	rows, err := b.eval(q.Ops, q.Keys, in)
	if err != nil {
		return nil, err
	}
	var out [][]any
	index := map[string]int{}
	for _, r := range rows {
		vals := make([]any, len(q.Outputs))
		for i, c := range q.Outputs {
			vals[i] = r[c]
		}
		if q.Weighted { // collapse equal output rows into one with the count as weight (like Cypher's count(*))
			k := eval.KeyString(vals...)
			if i, ok := index[k]; ok {
				out[i][len(vals)] = out[i][len(vals)].(int64) + 1
				continue
			}
			index[k] = len(out)
			vals = append(vals, int64(1))
		}
		out = append(out, vals)
	}
	return out, nil
}

// RunShipped evaluates a shipped query: the table rows (Keys, Ship, weight) are the Boundary's rows for the
// delegated subtree (a partial aggregation). The weight stays in column _w, so the aggregation counts it.
func (b *memBackend) RunShipped(_ context.Context, nq *stores.Query, s Shipment) ([][]any, error) {
	q := nq.Body.(*memPlan).q
	sb := &memBackend{g: b.g}
	for _, vals := range s.Rows {
		r := Row{}
		for i, c := range nq.ShipCols {
			r[c] = vals[i] // null id columns are columns too (the group of a null owner)
		}
		sb.ship = append(sb.ship, r)
	}
	res, err := sb.eval(q.Ops, q.Keys, nil)
	if err != nil {
		return nil, err
	}
	out := make([][]any, 0, len(res))
	for _, r := range res {
		vals := make([]any, len(q.Outputs))
		for i, c := range q.Outputs {
			vals[i] = r[c]
		}
		out = append(out, vals)
	}
	return out, nil
}

func (b *memBackend) Verbatim(context.Context, string, map[string]any) ([]map[string]any, error) {
	return nil, fmt.Errorf("memBackend: verbatim not supported")
}

func (b *memBackend) eval(n ir.Node, keys []string, in []Key) ([]Row, error) {
	g := b.g
	switch x := n.(type) {
	case *ir.Boundary:
		if b.ship != nil {
			out := make([]Row, len(b.ship))
			for i, r := range b.ship {
				out[i] = r.Clone()
			}
			return out, nil
		}
		out := make([]Row, len(in))
		for i, k := range in {
			r := Row{}
			for j, c := range keys {
				r[c] = k[j]
			}
			out[i] = r
		}
		return out, nil
	case *ir.EntityScan:
		var out []Row
		for _, u := range g.order {
			if hasAny(g.nodes[u].labels, x.Labels) {
				out = append(out, Row{x.Alias: u})
			}
		}
		return out, nil
	case *ir.Optional:
		// An OPTIONAL MATCH absorbed into one Neo4j query (the meaning of Cypher's OPTIONAL MATCH).
		outer, err := b.eval(x.Outer, keys, in)
		if err != nil {
			return nil, err
		}
		var argCols []string
		var findArg func(ir.Node)
		findArg = func(n ir.Node) {
			if a, ok := n.(*ir.Boundary); ok {
				argCols = a.Cols
			}
			for _, c := range n.Children() {
				findArg(c)
			}
		}
		findArg(x.Inner)
		var innerKeys []Key
		seen := map[string]bool{}
		for _, r := range outer {
			k := make(Key, len(argCols))
			vals := make([]any, len(argCols))
			ok := true
			for i, c := range argCols {
				u, isU := r[c].(id.UUID)
				if !isU {
					ok = false
					break
				}
				k[i], vals[i] = u, u
			}
			if ks := eval.KeyString(vals...); ok && !seen[ks] {
				seen[ks] = true
				innerKeys = append(innerKeys, k)
			}
		}
		inner, err := b.eval(x.Inner, argCols, innerKeys)
		if err != nil {
			return nil, err
		}
		pad := append([]string{}, x.New...)
		for _, r := range inner {
			for c := range r {
				if _, inOuter := outer0(outer)[c]; !inOuter && !contains(pad, c) {
					pad = append(pad, c)
				}
			}
		}
		return leftApply(outer, inner, &physical.OptionalApply{Keys: x.Keys, New: x.New, InnerCols: pad}), nil
	case *ir.NodeHashJoin:
		l, err := b.eval(x.Left, keys, in)
		if err != nil {
			return nil, err
		}
		r, err := b.eval(x.Right, keys, in)
		if err != nil {
			return nil, err
		}
		return hashJoin(l, r, x.Keys), nil
	}

	rows, err := b.eval(ir.Child(n), keys, in)
	if err != nil {
		return nil, err
	}
	switch x := n.(type) {
	case *ir.GetProperty:
		for _, r := range rows {
			u, _ := r[x.Alias].(id.UUID)
			r[x.Key()] = g.prop(u, x.Prop)
		}
		return rows, nil
	case *ir.GetListProperty:
		for _, r := range rows {
			lv, _ := r[x.Source].(*ListVal)
			var es []id.UUID
			if lv != nil {
				es = lv.Nodes
				if x.Part == ir.SegRels {
					es = lv.Rels
				}
			}
			vals := make([]any, len(es))
			for i, u := range es {
				vals[i] = g.prop(u, x.Prop)
			}
			r[x.Key()] = vals
		}
		return rows, nil
	case *ir.Expand:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			for _, st := range g.step(from, x.Types, x.Dir) {
				if !hasAny(g.nodes[st[1]].labels, x.ToLabels) {
					continue
				}
				nr := r.Clone()
				nr[x.Rel], nr[x.To] = st[0], st[1]
				out = append(out, nr)
			}
		}
		return out, nil
	case *ir.ExpandInto:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			to, _ := r[x.To].(id.UUID)
			for _, st := range g.step(from, x.Types, x.Dir) {
				if st[1] == to {
					nr := r.Clone()
					nr[x.Rel] = st[0]
					out = append(out, nr)
				}
			}
		}
		return out, nil
	case *ir.VarLengthExpand:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			var walk func(cur id.UUID, nodes, rels []id.UUID, used map[id.UUID]bool) error
			walk = func(cur id.UUID, nodes, rels []id.UUID, used map[id.UUID]bool) error {
				hops := len(rels)
				if hops >= x.Range.Min && hasAny(g.nodes[cur].labels, x.ToLabels) {
					interior := nodes
					if len(interior) > 0 {
						interior = interior[:len(interior)-1] // the end node is not an inner node
					}
					ok, err := b.segOK(x.SegPreds, interior, rels)
					if err != nil {
						return err
					}
					for _, u := range interior {
						for _, l := range x.InnerLabels {
							ok = ok && hasAny(g.nodes[u].labels, []string{l})
						}
					}
					if ok {
						nr := r.Clone()
						nr[x.To] = cur
						nr[ir.SegColumn(x.Seg)] = &ListVal{Nodes: append([]id.UUID{}, interior...), Rels: append([]id.UUID{}, rels...)}
						out = append(out, nr)
					}
				}
				if x.Range.Max >= 0 && hops >= x.Range.Max {
					return nil
				}
				for _, st := range g.step(cur, x.Types, x.Dir) {
					if used[st[0]] {
						continue
					}
					used[st[0]] = true
					if err := walk(st[1], append(append([]id.UUID{}, nodes...), st[1]), append(append([]id.UUID{}, rels...), st[0]), used); err != nil {
						return err
					}
					delete(used, st[0])
				}
				return nil
			}
			if err := walk(from, nil, nil, map[id.UUID]bool{}); err != nil {
				return nil, err
			}
		}
		return out, nil
	case *ir.Filter, *ir.Projection, *ir.Aggregation, *ir.Sort, *ir.Limit,
		*ir.Distinct, *ir.RelUniqueFilter, *ir.ProjectPath:
		return runCoord(n, [][]Row{rows})
	}
	return nil, fmt.Errorf("memBackend: operator %T", n)
}

// segOK reports whether every element of the segment meets SegPreds.
func (b *memBackend) segOK(preds []ir.SegPred, nodes, rels []id.UUID) (bool, error) {
	for _, p := range preds {
		es := nodes
		if p.Part == ir.SegRels {
			es = rels
		}
		for _, u := range es {
			// The element itself, for an id-set condition (ir.InSet with Elem).
			elem := &ListVal{Nodes: []id.UUID{u}}
			if p.Part == ir.SegRels {
				elem = &ListVal{Rels: []id.UUID{u}}
			}
			r := Row{"_e": elem}
			ir.Walk(p.Pred, func(x ir.Expr) {
				if ep, ok := x.(*ir.ElemProp); ok && ep.Var == p.Var {
					r[ir.ListPropKey("_e", p.Part, ep.Prop)] = []any{b.g.prop(u, ep.Prop)}
				}
			})
			v, err := eval.EvalElem(p.Pred, r, p.Var, "_e", p.Part, 0)
			if err != nil {
				return false, err
			}
			if !eval.Truthy(v) {
				return false, nil
			}
		}
	}
	return true, nil
}

func outer0(rows []Row) Row {
	if len(rows) == 0 {
		return Row{}
	}
	return rows[0]
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
