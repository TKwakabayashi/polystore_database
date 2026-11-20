package exec

// stepexpand.go: a variable-length expansion run hop by hop with the segment conditions checked at each hop
// (physical.StepExpand, PathStepwise). It means the same as VarLengthExpand in Neo4j: relationships are not
// repeated within a path, node conditions apply to inner nodes (a node is checked when a path is extended
// past it, not when a path ends at it), relationship conditions to every relationship.

import (
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"slices"
	"sync"
)

// stepMemo keeps the condition results per element id for one StepExpand during a run (shared by batches).
type stepMemo struct {
	mu    sync.Mutex
	nodes map[id.UUID]bool
	rels  map[id.UUID]bool
}

func (m *stepMemo) get(part ir.SegPart, u id.UUID) (pass, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if part == ir.SegRels {
		pass, ok = m.rels[u]
	} else {
		pass, ok = m.nodes[u]
	}
	return pass, ok
}

func (m *stepMemo) set(part ir.SegPart, u id.UUID, pass bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if part == ir.SegRels {
		m.rels[u] = pass
	} else {
		m.nodes[u] = pass
	}
}

func (e *Executor) memo(x *physical.StepExpand) *stepMemo {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.memos[x]
	if !ok {
		m = &stepMemo{nodes: map[id.UUID]bool{}, rels: map[id.UUID]bool{}}
		e.memos[x] = m
	}
	return m
}

// partialPath is a path from an input row's start: the nodes after the start (the last one included) and the relationships.
type partialPath struct {
	row   int
	last  id.UUID
	nodes []id.UUID
	rels  []id.UUID
}

// hopEdge is one relationship from a frontier node, with the hop row (the Neo4j values the conditions read).
type hopEdge struct {
	rel, to id.UUID
	vals    Row
}

func (e *Executor) stepExpand(x *physical.StepExpand, in []Row) ([]Row, error) {
	op := x.Op
	seg := ir.SegColumn(op.Seg)
	memo := e.memo(x)
	var nodePreds, relPreds []ir.Expr
	for _, p := range x.Preds {
		if p.Part == ir.SegRels {
			relPreds = append(relPreds, p.Pred)
		} else {
			nodePreds = append(nodePreds, p.Pred)
		}
	}

	var parts []partialPath
	for i, r := range in {
		if u, ok := r[op.From].(id.UUID); ok && !u.Empty() {
			parts = append(parts, partialPath{row: i, last: u})
		}
	}

	var out []Row
	type end struct {
		row int
		to  id.UUID
	}
	counts := map[end]int64{}
	var ends []end
	emit := func(p partialPath, h hopEdge) {
		if x.Weighted {
			k := end{p.row, h.to}
			if _, ok := counts[k]; !ok {
				ends = append(ends, k)
			}
			counts[k]++
			return
		}
		nr := in[p.row].Clone()
		nr[op.To] = h.to
		if x.EmitSeg {
			nr[seg] = &ListVal{Nodes: slices.Clone(p.nodes), Rels: append(slices.Clone(p.rels), h.rel)}
		}
		out = append(out, nr)
	}

	for depth := 1; depth <= op.Range.Max && len(parts) > 0; depth++ {
		adj, err := e.hop(x, parts)
		if err != nil {
			return nil, err
		}
		if len(relPreds) > 0 {
			vals := map[id.UUID]Row{}
			for _, hs := range adj {
				for _, h := range hs {
					vals[h.rel] = h.vals
				}
			}
			if err := e.check(x, memo, ir.SegRels, relPreds, vals); err != nil {
				return nil, err
			}
		}
		var next []partialPath
		nodeVals := map[id.UUID]Row{}
		for _, p := range parts {
			for _, h := range adj[p.last] {
				if len(relPreds) > 0 {
					if pass, _ := memo.get(ir.SegRels, h.rel); !pass {
						continue
					}
				}
				if slices.Contains(p.rels, h.rel) {
					continue
				}
				if depth >= op.Range.Min {
					emit(p, h)
				}
				if depth < op.Range.Max {
					next = append(next, partialPath{row: p.row, last: h.to,
						nodes: append(slices.Clone(p.nodes), h.to), rels: append(slices.Clone(p.rels), h.rel)})
					nodeVals[h.to] = h.vals
				}
			}
		}
		if len(nodePreds) > 0 && len(next) > 0 {
			if err := e.check(x, memo, ir.SegNodes, nodePreds, nodeVals); err != nil {
				return nil, err
			}
			kept := next[:0]
			for _, p := range next {
				if pass, _ := memo.get(ir.SegNodes, p.last); pass {
					kept = append(kept, p)
				}
			}
			next = kept
		}
		parts = next
	}

	if x.Weighted {
		for _, k := range ends {
			nr := in[k.row].Clone()
			nr[op.To] = k.to
			eval.SetWeight(nr, eval.WeightOf(in[k.row])*counts[k])
			out = append(out, nr)
		}
	}
	return out, nil
}

// hop expands the distinct last nodes of the partial paths with one Neo4j query and returns the relationships per node.
func (e *Executor) hop(x *physical.StepExpand, parts []partialPath) (map[id.UUID][]hopEdge, error) {
	var frontier []Key
	seen := map[id.UUID]bool{}
	for _, p := range parts {
		if !seen[p.last] {
			seen[p.last] = true
			frontier = append(frontier, Key{p.last})
		}
	}
	// The hop query is not a plan node with an input (storeQuery would issue it as a leaf): issue the frontier ids directly.
	nq, err := e.lower(x.Hop)
	if err != nil {
		return nil, err
	}
	tuples, err := e.run(x, nq, frontier)
	if err != nil {
		return nil, err
	}
	rows := newTupleSet(nq.Columns, tuples).toRows()
	adj := map[id.UUID][]hopEdge{}
	type edge struct{ from, rel, to id.UUID }
	dup := map[edge]bool{}
	for _, r := range rows {
		from, _ := r[physical.StepFrom].(id.UUID)
		rel, _ := r[physical.StepRel].(id.UUID)
		to, _ := r[physical.StepTo].(id.UUID)
		if from.Empty() || rel.Empty() || to.Empty() {
			continue
		}
		k := edge{from, rel, to}
		if dup[k] {
			continue
		}
		dup[k] = true
		adj[from] = append(adj[from], hopEdge{rel: rel, to: to, vals: r})
	}
	return adj, nil
}

// check evaluates the conditions of one part on the elements not decided yet and records the results in memo.
// hopVals are the hop rows that reached each element (Neo4j values); values outside Neo4j are fetched.
func (e *Executor) check(x *physical.StepExpand, memo *stepMemo, part ir.SegPart, preds []ir.Expr, hopVals map[id.UUID]Row) error {
	var todo []id.UUID
	for u := range hopVals {
		if _, ok := memo.get(part, u); !ok {
			todo = append(todo, u)
		}
	}
	if len(todo) == 0 {
		return nil
	}
	var cols []string
	for _, p := range preds {
		for _, c := range ir.Columns(p) {
			if !slices.Contains(cols, c) {
				cols = append(cols, c)
			}
		}
	}
	vals := make(map[id.UUID]Row, len(todo))
	for _, u := range todo {
		r := make(Row, len(cols))
		for _, c := range cols {
			r[c] = hopVals[u][c] // nil when the hop does not read it
		}
		vals[u] = r
	}
	keys := make([]Key, len(todo))
	for i, u := range todo {
		keys[i] = Key{u}
	}
	for _, f := range x.Fetches {
		if f.Part != part {
			continue
		}
		got, err := e.fetchValues(x, f, keys)
		if err != nil {
			return err
		}
		col := f.Source + "." + f.Prop
		for u, v := range got {
			if r, ok := vals[u]; ok && v != nil {
				r[col] = v
			}
		}
	}
	for _, u := range todo {
		pass := true
		for _, p := range preds {
			v, err := eval.Eval(p, vals[u])
			if err != nil {
				return err
			}
			if !eval.Truthy(v) {
				pass = false
				break
			}
		}
		memo.set(part, u, pass)
	}
	return nil
}
