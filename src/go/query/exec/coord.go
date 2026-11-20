package exec

import (
	"fmt"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"sort"

	"polystore_database/src/go/base/id"
)

// coord.go runs coordinator operators (computed from input rows only).

func runCoord(op ir.Node, inputs [][]Row) ([]Row, error) {
	in := func() []Row {
		if len(inputs) == 0 {
			return nil
		}
		return inputs[0]
	}
	switch x := op.(type) {
	case *ir.Filter:
		var out []Row
		for _, r := range in() {
			v, err := eval.Eval(x.Pred, r)
			if err != nil {
				return nil, err
			}
			if eval.Truthy(v) {
				out = append(out, r)
			}
		}
		return out, nil

	case *ir.Projection:
		out := make([]Row, 0, len(in()))
		for _, r := range in() {
			nr := make(Row, len(x.Items)+1)
			for _, it := range x.Items {
				v, err := eval.Eval(it.Expr, r)
				if err != nil {
					return nil, err
				}
				nr[it.Name] = v
			}
			eval.SetWeight(nr, eval.WeightOf(r))
			out = append(out, nr)
		}
		return out, nil

	case *ir.Aggregation:
		return aggregate(x, in())

	case *ir.Sort:
		rows := append([]Row{}, in()...)
		type keyed struct {
			row  Row
			keys []any
		}
		ks := make([]keyed, len(rows))
		for i, r := range rows {
			vals := make([]any, len(x.Keys))
			for j, k := range x.Keys {
				v, err := eval.Eval(k.Expr, r)
				if err != nil {
					return nil, err
				}
				vals[j] = v
			}
			ks[i] = keyed{row: r, keys: vals}
		}
		idx := make([]int, len(ks))
		for i := range idx {
			idx[i] = i
		}
		less := func(a, b int) bool {
			for j, k := range x.Keys {
				va, vb := ks[a].keys[j], ks[b].keys[j]
				c := eval.OrderCompare(va, vb)
				// null sorts as the largest value (as in Neo4j: last in ASC, first in DESC).
				if k.Desc {
					c = -c
				}
				if c != 0 {
					return c < 0
				}
			}
			return false
		}
		sort.SliceStable(idx, func(i, j int) bool { return less(idx[i], idx[j]) })
		out := make([]Row, len(rows))
		for i, j := range idx {
			out[i] = ks[j].row
		}
		return out, nil

	case *ir.Limit:
		out, _ := (&limitState{skip: x.Skip, count: x.Count}).take(in())
		return out, nil

	case *ir.Distinct:
		seen := map[string]struct{}{}
		var out []Row
		var buf []byte
		var prev []string // the previous row's column names (sorted) when no columns are given
		for _, r := range in() {
			cols := x.Cols
			if len(cols) == 0 {
				if !sameKeys(r, prev) { // reuse the sort when the column set equals the previous row's
					prev = sortedKeys(r)
				}
				cols = prev
			}
			buf = buf[:0]
			for _, c := range cols {
				if c != eval.WeightCol { // the weight counts rows; it is not a value
					buf = append(eval.AppendKey(buf, r[c]), 0)
				}
			}
			if _, dup := seen[string(buf)]; !dup {
				seen[string(buf)] = struct{}{}
				nr := r.Clone()
				eval.SetWeight(nr, 1)
				out = append(out, nr)
			}
		}
		return out, nil

	case *ir.NodeHashJoin:
		if len(inputs) != 2 {
			return nil, fmt.Errorf("executor: join needs two inputs")
		}
		return hashJoin(inputs[0], inputs[1], x.Keys), nil

	case *ir.RelUniqueFilter:
		var out []Row
		for _, r := range in() {
			seen := map[id.UUID]bool{}
			ok := true
			add := func(u id.UUID) {
				if u.Empty() {
					return
				}
				if seen[u] {
					ok = false
				}
				seen[u] = true
			}
			for _, rel := range x.Rels {
				u, _ := r[rel].(id.UUID)
				add(u)
			}
			for _, s := range x.Segs {
				if lv, _ := r[ir.SegColumn(s)].(*ListVal); lv != nil {
					for _, u := range lv.Rels {
						add(u)
					}
				}
			}
			if ok {
				out = append(out, r)
			}
		}
		return out, nil

	case *ir.Unwind:
		var out []Row
		for _, r := range in() {
			v, err := eval.Eval(x.List, r)
			if err != nil {
				return nil, err
			}
			list, ok := v.([]any)
			if !ok {
				if v == nil {
					continue // UNWIND null keeps no row
				}
				list = []any{v} // a non-list value is one element (as in Neo4j)
			}
			for _, el := range list {
				nr := r.Clone()
				nr[x.Var] = el
				out = append(out, nr)
			}
		}
		return out, nil

	case *ir.ProjectPath:
		out := make([]Row, 0, len(in()))
		for _, r := range in() {
			lv := &ListVal{}
			for i, el := range x.Elems {
				switch {
				case el.Seg:
					if s, _ := r[ir.SegColumn(el.Var)].(*ListVal); s != nil {
						lv.Nodes = append(lv.Nodes, s.Nodes...)
						lv.Rels = append(lv.Rels, s.Rels...)
					}
				case i%2 == 0: // a path alternates nodes and relationships
					u, _ := r[el.Var].(id.UUID)
					lv.Nodes = append(lv.Nodes, u)
				default:
					u, _ := r[el.Var].(id.UUID)
					lv.Rels = append(lv.Rels, u)
				}
			}
			nr := r.Clone()
			nr[x.Path] = lv
			out = append(out, nr)
		}
		return out, nil
	}
	return nil, fmt.Errorf("executor: coordinator operator %T", op)
}

// sameKeys reports whether row r's column names are exactly keys.
func sameKeys(r Row, keys []string) bool {
	if keys == nil || len(r) != len(keys) {
		return false
	}
	for _, k := range keys {
		if _, ok := r[k]; !ok {
			return false
		}
	}
	return true
}

func sortedKeys(r Row) []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// hashJoin joins two inputs on id columns keys (a cross product when keys is empty).
func hashJoin(left, right []Row, keys []string) []Row {
	// Keys are built with AppendKey into a reused buffer (lookups with m[string(buf)] do not copy).
	var buf []byte
	keyOf := func(r Row) []byte {
		buf = buf[:0]
		for _, k := range keys {
			buf = append(eval.AppendKey(buf, r[k]), 0)
		}
		return buf
	}
	build := make(map[string][]Row, len(right))
	for _, r := range right {
		k := keyOf(r)
		build[string(k)] = append(build[string(k)], r)
	}
	var out []Row
	for _, l := range left {
		for _, r := range build[string(keyOf(l))] {
			nr := make(Row, len(l)+len(r)) // room for the right side's columns without regrowing the map
			for k, v := range l {
				nr[k] = v
			}
			for k, v := range r {
				nr[k] = v
			}
			eval.SetWeight(nr, eval.WeightOf(l)*eval.WeightOf(r))
			out = append(out, nr)
		}
	}
	return out
}

// ---- aggregation ----

type aggState struct {
	count int64
	sum   any
	min   any
	max   any
	n     int64 // avg's denominator (non-null count)
	seen  map[string]bool
	list  []any
}

func aggregate(x *ir.Aggregation, rows []Row) ([]Row, error) {
	type group struct {
		keys  []any
		state []*aggState
	}
	groups := map[string]*group{}
	var order []string
	for _, r := range rows {
		keys := make([]any, len(x.Keys))
		for i, k := range x.Keys {
			v, err := eval.Eval(k.Expr, r)
			if err != nil {
				return nil, err
			}
			keys[i] = v
		}
		gk := eval.KeyString(keys...)
		g, ok := groups[gk]
		if !ok {
			g = &group{keys: keys, state: make([]*aggState, len(x.Aggs))}
			for i := range g.state {
				g.state[i] = &aggState{seen: map[string]bool{}}
			}
			groups[gk] = g
			order = append(order, gk)
		}
		rw := eval.WeightOf(r) // how many original rows this row stands for
		for i, a := range x.Aggs {
			s := g.state[i]
			if a.Arg == nil { // count(*)
				s.count += rw
				continue
			}
			v, err := eval.Eval(a.Arg, r)
			if err != nil {
				return nil, err
			}
			if v == nil {
				continue
			}
			w := rw
			if a.Distinct { // DISTINCT counts kinds of values, so the weight is not used
				k := eval.KeyString(v)
				if s.seen[k] {
					continue
				}
				s.seen[k] = true
				w = 1
			}
			s.count += w
			s.n += w
			switch a.Fn {
			case "sum", "avg":
				term := v
				if w != 1 {
					if term, err = eval.Arith("*", v, w); err != nil {
						return nil, err
					}
				}
				sum, err := eval.Arith("+", orZero(s.sum), term)
				if err != nil {
					return nil, err
				}
				s.sum = sum
			case "min":
				if s.min == nil || eval.OrderCompare(v, s.min) < 0 {
					s.min = v
				}
			case "max":
				if s.max == nil || eval.OrderCompare(v, s.max) > 0 {
					s.max = v
				}
			case "collect":
				for j := int64(0); j < w; j++ {
					s.list = append(s.list, v)
				}
			}
		}
	}
	// An aggregation without keys returns one row even for empty input (count(*) = 0).
	if len(x.Keys) == 0 && len(order) == 0 {
		g := &group{state: make([]*aggState, len(x.Aggs))}
		for i := range g.state {
			g.state[i] = &aggState{seen: map[string]bool{}}
		}
		groups[""] = g
		order = append(order, "")
	}
	out := make([]Row, 0, len(order))
	for _, gk := range order {
		g := groups[gk]
		nr := Row{}
		for i, k := range x.Keys {
			nr[k.Name] = g.keys[i]
		}
		for i, a := range x.Aggs {
			s := g.state[i]
			switch a.Fn {
			case "count":
				nr[a.Name] = s.count
			case "sum":
				if s.sum == nil {
					nr[a.Name] = int64(0)
				} else {
					nr[a.Name] = s.sum
				}
			case "avg":
				if s.n == 0 {
					nr[a.Name] = nil
				} else {
					f, _ := eval.ToFloat(s.sum)
					nr[a.Name] = f / float64(s.n)
				}
			case "min":
				nr[a.Name] = s.min
			case "max":
				nr[a.Name] = s.max
			case "collect":
				if s.list == nil {
					nr[a.Name] = []any{}
				} else {
					nr[a.Name] = s.list
				}
			default:
				return nil, fmt.Errorf("executor: aggregate %s", a.Fn)
			}
		}
		out = append(out, nr)
	}
	return out, nil
}

func orZero(v any) any {
	if v == nil {
		return int64(0)
	}
	return v
}
