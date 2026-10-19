package executor

import (
	"fmt"
	"sort"

	"polystore_database/src/go/id"
	"polystore_database/src/go/logical"
)

// coord.go はコーディネータの演算子（入力行だけで計算するもの）。

func runCoord(op logical.Node, inputs [][]Row) ([]Row, error) {
	in := func() []Row {
		if len(inputs) == 0 {
			return nil
		}
		return inputs[0]
	}
	switch x := op.(type) {
	case *logical.Filter:
		var out []Row
		for _, r := range in() {
			v, err := Eval(x.Pred, r)
			if err != nil {
				return nil, err
			}
			if truthy(v) {
				out = append(out, r)
			}
		}
		return out, nil

	case *logical.Projection:
		out := make([]Row, 0, len(in()))
		for _, r := range in() {
			nr := make(Row, len(x.Items)+1)
			for _, it := range x.Items {
				v, err := Eval(it.Expr, r)
				if err != nil {
					return nil, err
				}
				nr[it.Name] = v
			}
			setWeight(nr, weightOf(r))
			out = append(out, nr)
		}
		return out, nil

	case *logical.Aggregation:
		return aggregate(x, in())

	case *logical.Sort:
		rows := append([]Row{}, in()...)
		type keyed struct {
			row  Row
			keys []any
		}
		ks := make([]keyed, len(rows))
		for i, r := range rows {
			vals := make([]any, len(x.Keys))
			for j, k := range x.Keys {
				v, err := Eval(k.Expr, r)
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
				c := orderCompare(va, vb)
				// null は最大値として並ぶ（Neo4j と同じ: ASC で末尾、DESC で先頭）。
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

	case *logical.Limit:
		out, _ := (&limitState{skip: x.Skip, count: x.Count}).take(in())
		return out, nil

	case *logical.Distinct:
		seen := map[string]bool{}
		var out []Row
		for _, r := range in() {
			cols := x.Cols
			if len(cols) == 0 {
				cols = sortedKeys(r)
			}
			var vals []any
			for _, c := range cols {
				if c != weightCol { // 重みは行の数であって値ではない
					vals = append(vals, r[c])
				}
			}
			k := keyString(vals...)
			if !seen[k] {
				seen[k] = true
				nr := r.clone()
				setWeight(nr, 1)
				out = append(out, nr)
			}
		}
		return out, nil

	case *logical.NodeHashJoin:
		if len(inputs) != 2 {
			return nil, fmt.Errorf("executor: join needs two inputs")
		}
		return hashJoin(inputs[0], inputs[1], x.Keys), nil

	case *logical.RelUniqueFilter:
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
				if lv, _ := r[logical.SegColumn(s)].(*ListVal); lv != nil {
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

	case *logical.ProjectPath:
		out := make([]Row, 0, len(in()))
		for _, r := range in() {
			lv := &ListVal{}
			for i, el := range x.Elems {
				switch {
				case el.Seg:
					if s, _ := r[logical.SegColumn(el.Var)].(*ListVal); s != nil {
						lv.Nodes = append(lv.Nodes, s.Nodes...)
						lv.Rels = append(lv.Rels, s.Rels...)
					}
				case i%2 == 0: // パスはノード・関係が交互に並ぶ
					u, _ := r[el.Var].(id.UUID)
					lv.Nodes = append(lv.Nodes, u)
				default:
					u, _ := r[el.Var].(id.UUID)
					lv.Rels = append(lv.Rels, u)
				}
			}
			nr := r.clone()
			nr[x.Path] = lv
			out = append(out, nr)
		}
		return out, nil
	}
	return nil, fmt.Errorf("executor: coordinator operator %T", op)
}

func sortedKeys(r Row) []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// hashJoin は識別列 keys で 2 入力を結合する（keys が空なら直積）。
func hashJoin(left, right []Row, keys []string) []Row {
	build := map[string][]Row{}
	for _, r := range right {
		vals := make([]any, len(keys))
		for i, k := range keys {
			vals[i] = r[k]
		}
		k := keyString(vals...)
		build[k] = append(build[k], r)
	}
	var out []Row
	for _, l := range left {
		vals := make([]any, len(keys))
		for i, k := range keys {
			vals[i] = l[k]
		}
		for _, r := range build[keyString(vals...)] {
			nr := l.clone()
			for k, v := range r {
				nr[k] = v
			}
			setWeight(nr, weightOf(l)*weightOf(r))
			out = append(out, nr)
		}
	}
	return out
}

// ---- 集約 ----

type aggState struct {
	count int64
	sum   any
	min   any
	max   any
	n     int64 // avg の分母（null 以外の数）
	seen  map[string]bool
	list  []any
}

func aggregate(x *logical.Aggregation, rows []Row) ([]Row, error) {
	type group struct {
		keys  []any
		state []*aggState
	}
	groups := map[string]*group{}
	var order []string
	for _, r := range rows {
		keys := make([]any, len(x.Keys))
		for i, k := range x.Keys {
			v, err := Eval(k.Expr, r)
			if err != nil {
				return nil, err
			}
			keys[i] = v
		}
		gk := keyString(keys...)
		g, ok := groups[gk]
		if !ok {
			g = &group{keys: keys, state: make([]*aggState, len(x.Aggs))}
			for i := range g.state {
				g.state[i] = &aggState{seen: map[string]bool{}}
			}
			groups[gk] = g
			order = append(order, gk)
		}
		rw := weightOf(r) // この行が表す元の行の数
		for i, a := range x.Aggs {
			s := g.state[i]
			if a.Arg == nil { // count(*)
				s.count += rw
				continue
			}
			v, err := Eval(a.Arg, r)
			if err != nil {
				return nil, err
			}
			if v == nil {
				continue
			}
			w := rw
			if a.Distinct { // DISTINCT は値の種類を数えるので重みを使わない
				k := keyString(v)
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
					if term, err = arith("*", v, w); err != nil {
						return nil, err
					}
				}
				sum, err := arith("+", orZero(s.sum), term)
				if err != nil {
					return nil, err
				}
				s.sum = sum
			case "min":
				if s.min == nil || orderCompare(v, s.min) < 0 {
					s.min = v
				}
			case "max":
				if s.max == nil || orderCompare(v, s.max) > 0 {
					s.max = v
				}
			case "collect":
				for j := int64(0); j < w; j++ {
					s.list = append(s.list, v)
				}
			}
		}
	}
	// キー無しの集約は、入力が空でも 1 行返す（count(*) = 0）。
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
					f, _ := toFloat(s.sum)
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
