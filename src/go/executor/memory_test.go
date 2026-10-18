package executor

import (
	"context"
	"fmt"

	"polystore_database/src/go/id"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/native"
	"polystore_database/src/go/physical"
)

// memory_test.go はテスト用のバックエンド: StoreQuery の論理部分木を、小さなグラフのデータに対して
// 直接評価する。ストアの区別をしない（どの列もここにある）ので、どの配置で計画しても結果は同じに
// なるはずで、それが計画と実行器（結合・Fetch・ListFetch・コーディネータ）の正しさの検証になる。

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
	order []id.UUID // ノードの決定的な順序
	rord  []id.UUID
}

func newMemGraph() *memGraph {
	return &memGraph{nodes: map[id.UUID]*memNode{}, rels: map[id.UUID]*memRel{}}
}

func (g *memGraph) node(uuid string, label string, props map[string]any) id.UUID {
	u := id.FromAny(uuid)
	for k, v := range props {
		props[k] = normalize(v)
	}
	g.nodes[u] = &memNode{labels: []string{label}, props: props}
	g.order = append(g.order, u)
	return u
}

func (g *memGraph) rel(uuid, typ string, src, dst id.UUID, props map[string]any) {
	u := id.FromAny(uuid)
	if props == nil {
		props = map[string]any{}
	}
	for k, v := range props {
		props[k] = normalize(v)
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

// step は from から dir の向きに辿れる (関係, 相手) の一覧。
func (g *memGraph) step(from id.UUID, types []string, dir logical.Direction) [][2]id.UUID {
	var out [][2]id.UUID
	for _, ru := range g.rord {
		r := g.rels[ru]
		if !typeOK(r, types) {
			continue
		}
		if (dir == logical.Out || dir == logical.Both) && r.src == from {
			out = append(out, [2]id.UUID{ru, r.dst})
		}
		if (dir == logical.In || dir == logical.Both) && r.dst == from && !(dir == logical.Both && r.src == from) {
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

type memBackend struct{ g *memGraph }

func (b *memBackend) Run(context.Context, *native.Query, []Key) ([][]any, error) {
	return nil, fmt.Errorf("memBackend: native queries are not used")
}

func (b *memBackend) Verbatim(context.Context, string, map[string]any) ([]map[string]any, error) {
	return nil, fmt.Errorf("memBackend: verbatim not supported")
}

func (b *memBackend) RunLogical(q *physical.StoreQuery, in []Key) ([]Row, error) {
	rows, err := b.eval(q.Ops, q.Keys, in)
	if err != nil {
		return nil, err
	}
	out := make([]Row, len(rows))
	for i, r := range rows {
		nr := Row{}
		for _, c := range q.Outputs {
			nr[c] = r[c]
		}
		out[i] = nr
	}
	return out, nil
}

func (b *memBackend) ListValues(f *physical.ListFetch, ids []id.UUID) (map[id.UUID]any, error) {
	out := map[id.UUID]any{}
	for _, u := range ids {
		out[u] = b.g.prop(u, f.Prop)
	}
	return out, nil
}

func (b *memBackend) eval(n logical.Node, keys []string, in []Key) ([]Row, error) {
	g := b.g
	switch x := n.(type) {
	case *physical.Argument:
		out := make([]Row, len(in))
		for i, k := range in {
			r := Row{}
			for j, c := range keys {
				r[c] = k[j]
			}
			out[i] = r
		}
		return out, nil
	case *logical.EntityScan:
		var out []Row
		for _, u := range g.order {
			if hasAny(g.nodes[u].labels, x.Labels) {
				out = append(out, Row{x.Alias: u})
			}
		}
		return out, nil
	case *logical.NodeHashJoin:
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

	rows, err := b.eval(logical.Child(n), keys, in)
	if err != nil {
		return nil, err
	}
	switch x := n.(type) {
	case *logical.GetProperty:
		for _, r := range rows {
			u, _ := r[x.Alias].(id.UUID)
			r[x.Key()] = g.prop(u, x.Prop)
		}
		return rows, nil
	case *logical.GetListProperty:
		for _, r := range rows {
			lv, _ := r[x.Source].(*ListVal)
			var es []id.UUID
			if lv != nil {
				es = lv.Nodes
				if x.Part == logical.SegRels {
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
	case *logical.Expand:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			for _, st := range g.step(from, x.Types, x.Dir) {
				if !hasAny(g.nodes[st[1]].labels, x.ToLabels) {
					continue
				}
				nr := r.clone()
				nr[x.Rel], nr[x.To] = st[0], st[1]
				out = append(out, nr)
			}
		}
		return out, nil
	case *logical.ExpandInto:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			to, _ := r[x.To].(id.UUID)
			for _, st := range g.step(from, x.Types, x.Dir) {
				if st[1] == to {
					nr := r.clone()
					nr[x.Rel] = st[0]
					out = append(out, nr)
				}
			}
		}
		return out, nil
	case *logical.VarLengthExpand:
		var out []Row
		for _, r := range rows {
			from, _ := r[x.From].(id.UUID)
			var walk func(cur id.UUID, nodes, rels []id.UUID, used map[id.UUID]bool) error
			walk = func(cur id.UUID, nodes, rels []id.UUID, used map[id.UUID]bool) error {
				hops := len(rels)
				if hops >= x.Range.Min && hasAny(g.nodes[cur].labels, x.ToLabels) {
					interior := nodes
					if len(interior) > 0 {
						interior = interior[:len(interior)-1] // 終点は中間ノードではない
					}
					ok, err := b.segOK(x.SegPreds, interior, rels)
					if err != nil {
						return err
					}
					if ok {
						nr := r.clone()
						nr[x.To] = cur
						nr[logical.SegColumn(x.Seg)] = &ListVal{Nodes: append([]id.UUID{}, interior...), Rels: append([]id.UUID{}, rels...)}
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
	case *logical.Filter, *logical.Projection, *logical.Aggregation, *logical.Sort, *logical.Limit,
		*logical.Distinct, *logical.RelUniqueFilter, *logical.ProjectPath:
		return runCoord(n, [][]Row{rows})
	}
	return nil, fmt.Errorf("memBackend: operator %T", n)
}

// segOK は区間の要素がすべて SegPreds を満たすか。
func (b *memBackend) segOK(preds []logical.SegPred, nodes, rels []id.UUID) (bool, error) {
	for _, p := range preds {
		es := nodes
		if p.Part == logical.SegRels {
			es = rels
		}
		for _, u := range es {
			r := Row{}
			logical.Walk(p.Pred, func(x logical.Expr) {
				if ep, ok := x.(*logical.ElemProp); ok && ep.Var == p.Var {
					r[logical.ListPropKey("_e", p.Part, ep.Prop)] = []any{b.g.prop(u, ep.Prop)}
				}
			})
			v, err := eval(p.Pred, r, elemCtx{p.Var: {source: "_e", part: p.Part, index: 0}})
			if err != nil {
				return false, err
			}
			if !truthy(v) {
				return false, nil
			}
		}
	}
	return true, nil
}
