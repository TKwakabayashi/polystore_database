package physical

import (
	"fmt"
	"sort"
	"strings"

	"polystore_database/src/go/codec"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// build.go は §5.6 の計画アルゴリズム。論理計画を葉から根へ 1 演算子ずつ見て、
//   - ストアに割り当てられた演算子は、同じストアの「開いている」クエリに足せれば足す（統合・吸収）。
//     足せなければ今のクエリを閉じ、境界の識別列を IN $ids で受け取る新しいクエリを開く。
//   - コーディネータの演算子は、開いているクエリを閉じてから Coord として積む。
//     ただし葉のクエリ（入力を持たない）で、そのストアが実行できるなら吸収する。
// 値は GetProperty の位置（＝論理計画が決めた「最初に使う直前」）で取得するので、遅延実体化は論理側の
// 位置と統合の規則だけで成り立つ。最後に要求解析（finalize）で各クエリが返す列を決める。

// Build は論理計画を物理計画にする。表現できない形は ast.ErrUnsupported を包んだエラー。
func Build(root logical.Node, cat *Catalog) (Node, error) {
	pl := &planner{cat: cat, parent: map[logical.Node]logical.Node{},
		nodeLabels: map[string][]string{}, relTypes: map[string][]string{},
		segTypes: map[string][]string{}, pathElems: map[string][]logical.PathElem{},
		owners: map[string]owner{}}
	pl.index(root, nil)
	st, err := pl.build(root)
	if err != nil {
		return nil, err
	}
	pl.close(st)
	finalize(st.phys, nil, true)
	return st.phys, nil
}

type planner struct {
	cat    *Catalog
	parent map[logical.Node]logical.Node

	// 変数の型情報（リスト要素のラベル解決に使う）。
	nodeLabels map[string][]string
	relTypes   map[string][]string
	segTypes   map[string][]string
	pathElems  map[string][]logical.PathElem
	owners     map[string]owner // alias → プロパティの持ち主（リテラルの型付け用）
	args       []*ArgRows       // 組み立て中の OPTIONAL の内側の葉（buildOptional が拾う）
}

type owner struct {
	kind   logical.ObjKind
	labels []string
}

func (pl *planner) index(n, parent logical.Node) {
	if n == nil {
		return
	}
	pl.parent[n] = parent
	switch x := n.(type) {
	case *logical.EntityScan:
		pl.nodeLabels[x.Alias] = x.Labels
	case *logical.Expand:
		pl.relTypes[x.Rel] = x.Types
		if len(x.ToLabels) > 0 {
			pl.nodeLabels[x.To] = x.ToLabels
		}
	case *logical.ExpandInto:
		pl.relTypes[x.Rel] = x.Types
	case *logical.VarLengthExpand:
		pl.segTypes[x.Seg] = x.Types
		if len(x.ToLabels) > 0 {
			pl.nodeLabels[x.To] = x.ToLabels
		}
	case *logical.ProjectPath:
		pl.pathElems[x.Path] = x.Elems
	case *logical.GetProperty:
		if x.Kind == logical.NodeObj && len(x.Labels) > 0 && len(pl.nodeLabels[x.Alias]) == 0 {
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

// state は部分木を処理した結果: 閉じた物理木と、まだ演算子を足せる開いたクエリ。
type state struct {
	phys Node
	open *openQ
}

type openQ struct {
	store   store.Kind
	ops     logical.Node
	input   Node
	keys    []string
	leaf    bool
	avail   map[string]bool // クエリの行にある列
	idents  map[string]bool // そのうち識別列
	single  string          // 非 graph のクエリが扱う唯一の alias
	grouped bool            // 集約を吸収済み（以降の Filter は HAVING 相当）
}

func newOpen(k store.Kind, ops logical.Node, input Node, keys []string, single string) *openQ {
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
	st.phys = &StoreQuery{Store: o.store, Ops: o.ops, Input: o.input, Keys: o.keys}
	st.open = nil
}

// boundary は st を閉じ、keys を受け取る新しいクエリを開く。
func (pl *planner) boundary(st *state, k store.Kind, op logical.Node, keys []string, single string) {
	pl.close(st)
	// OPTIONAL の内側の最初の graph クエリは、葉（ArgRows）の識別列をすべて受け取る
	// （後続の ExpandInto などが外側の別の変数を使えるように。graph は複数キーを UNWIND で受け取れる）。
	if arg, ok := st.phys.(*ArgRows); ok && k == store.Graph {
		for _, c := range arg.Cols {
			if !contains(keys, c) {
				keys = append(keys, c)
			}
		}
	}
	st.open = newOpen(k, logical.WithInput(op, &Argument{Cols: keys}), st.phys, keys, single)
	st.phys = nil
	produce(st.open, op)
}

func (pl *planner) extend(st *state, op logical.Node) {
	st.open.ops = logical.WithInput(op, st.open.ops)
	produce(st.open, op)
}

func (pl *planner) coord(st *state, op logical.Node) {
	pl.close(st)
	st.phys = &Coord{Op: logical.WithInput(op, &Argument{}), Inputs: []Node{st.phys}}
}

func (pl *planner) build(n logical.Node) (*state, error) {
	switch x := n.(type) {
	case *logical.EntityScan:
		k, single := pl.scanStore(x)
		o := newOpen(k, x, nil, nil, single)
		o.avail[x.Alias], o.idents[x.Alias] = true, true
		return &state{open: o}, nil

	case *logical.Argument:
		arg := &ArgRows{Cols: x.Cols}
		pl.args = append(pl.args, arg)
		return &state{phys: arg}, nil

	case *logical.Optional:
		return pl.buildOptional(x)

	case *logical.NodeHashJoin:
		l, err := pl.build(x.Left)
		if err != nil {
			return nil, err
		}
		r, err := pl.build(x.Right)
		if err != nil {
			return nil, err
		}
		if l.open != nil && r.open != nil && l.open.leaf && r.open.leaf &&
			l.open.store == store.Graph && r.open.store == store.Graph &&
			graphPattern(l.open.ops) && graphPattern(r.open.ops) {
			// 両側が graph のパターンだけで完結するなら 1 本の MATCH にまとめる
			// （射影・集約を吸収した側は MATCH の 1 部分に書けないので、コーディネータで直積にする）。
			o := l.open
			o.ops = &logical.NodeHashJoin{Left: l.open.ops, Right: r.open.ops, Keys: x.Keys}
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
			Op:     &logical.NodeHashJoin{Left: &Argument{}, Right: &Argument{}, Keys: x.Keys},
			Inputs: []Node{l.phys, r.phys},
		}}, nil
	}

	c := logical.Child(n)
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

// buildOptional は OPTIONAL MATCH を組む。外側と内側がどちらも graph だけで完結するなら、
// 1 本の Cypher（OPTIONAL MATCH）にまとめる。それ以外は OptionalApply（内側の葉 ArgRows に外側の行を渡す）。
func (pl *planner) buildOptional(x *logical.Optional) (*state, error) {
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

	// graph に吸収: 外側が開いた graph の葉クエリで、内側が ArgRows を入力とする 1 本の graph クエリだけ
	// （内側の識別列はすべて外側のクエリにある）。
	if o, in := outer.open, inner.open; o != nil && in != nil && o.leaf && o.store == store.Graph &&
		in.store == store.Graph && in.input == Node(arg) && inner.phys == nil && identAll(o, x.Keys) && graphPattern(in.ops) {
		o.ops = &logical.Optional{Outer: o.ops, Inner: in.ops, Keys: x.Keys, New: x.New}
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

// graphPattern は委譲する部分木が、OPTIONAL MATCH の 1 節に書けるパターンと条件だけか。
func graphPattern(n logical.Node) bool {
	ok := true
	var walk func(logical.Node)
	walk = func(n logical.Node) {
		if n == nil || !ok {
			return
		}
		switch n.(type) {
		case *Argument, *logical.EntityScan, *logical.Expand, *logical.ExpandInto, *logical.VarLengthExpand,
			*logical.Filter, *logical.GetProperty, *logical.RelUniqueFilter, *logical.NodeHashJoin:
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

// scanStore は EntityScan のストア。直上の GetProperty/Filter の連なりを見て、
//   - 最初の Filter が参照する列のストア、
//   - Filter が無ければ、連なりの最初の GetProperty のストア（必要な列がそこにあるので吸収できる）、
//   - どちらも無ければ graph。
func (pl *planner) scanStore(s *logical.EntityScan) (store.Kind, string) {
	for _, l := range s.Labels {
		if !pl.cat.IsEntity(l) {
			return store.Graph, "" // サブラベルの走査は graph でしかできない
		}
	}
	if len(s.Labels) == 0 {
		return store.Graph, ""
	}
	props := map[string]*logical.GetProperty{}
	var first *logical.GetProperty
	pick := func(g *logical.GetProperty) (store.Kind, string) {
		k := pl.cat.StoreOf(g.Alias, g.Kind, g.Labels, g.Prop)
		if k == store.Graph {
			return k, ""
		}
		return k, s.Alias
	}
walk:
	for p := pl.parent[s]; p != nil; p = pl.parent[p] {
		switch x := p.(type) {
		case *logical.GetProperty:
			if x.Alias != s.Alias {
				break walk
			}
			props[x.Key()] = x
			if first == nil {
				first = x
			}
		case *logical.Filter:
			for _, col := range logical.Columns(x.Pred) {
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
	return store.Graph, ""
}

func (pl *planner) add(st *state, op logical.Node) error {
	pl.typeLiterals(op)
	if g, ok := op.(*logical.GetProperty); ok && g.Kind == logical.NodeObj {
		g.Labels = pl.cat.Entities(g.Labels) // 表・コレクションは entity 単位（Post → Message）
	}
	o := st.open
	switch x := op.(type) {
	case *logical.GetProperty:
		k := pl.cat.StoreOf(x.Alias, x.Kind, x.Labels, x.Prop)
		if o != nil && o.store == k && o.idents[x.Alias] && (k == store.Graph || o.single == x.Alias) {
			pl.extend(st, op)
			return nil
		}
		single := x.Alias
		if k == store.Graph {
			single = ""
		}
		pl.boundary(st, k, op, []string{x.Alias}, single)
		return nil

	case *logical.Filter:
		having := o != nil && o.grouped && (o.store == store.Columnar || o.store == store.Kvs)
		if o != nil && !having && availAll(o, filterRefs(x)) && CanExpress(o.store, x.Pred) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *logical.Expand:
		return pl.addGraph(st, op, []string{x.From})
	case *logical.ExpandInto:
		return pl.addGraph(st, op, []string{x.From, x.To})
	case *logical.VarLengthExpand:
		return pl.addVarLength(st, x)

	case *logical.ProjectPath:
		if o != nil && o.store == store.Graph && availAll(o, logical.Refs(x)) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *logical.GetListProperty:
		labels := x.ElemLabels
		if len(labels) == 0 {
			labels = pl.listLabels(x.Source, x.Part)
		}
		k, err := pl.listStore(labels, x.Part, x.Prop)
		if err != nil {
			return err
		}
		if o != nil && o.store == store.Graph && k == store.Graph && o.avail[x.Source] {
			pl.extend(st, op)
			return nil
		}
		pl.close(st)
		if x.Part == logical.SegNodes {
			labels = pl.cat.Entities(labels)
		}
		st.phys = &ListFetch{Input: st.phys, Store: k, Source: x.Source, Part: x.Part, Prop: x.Prop, Labels: labels}
		return nil

	case *logical.RelUniqueFilter:
		if o != nil && o.store == store.Graph && availAll(o, logical.Refs(x)) && !anyKey(o, logical.Refs(x)) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil

	case *logical.Unwind:
		pl.coord(st, op)
		return nil

	case *logical.Projection, *logical.Aggregation, *logical.Sort, *logical.Limit, *logical.Distinct:
		if o != nil && o.leaf && CanAbsorb(o.store, op) && availAll(o, logical.Refs(op)) && !collectsIdent(o, op) {
			pl.extend(st, op)
			return nil
		}
		pl.coord(st, op)
		return nil
	}
	return unsupported("operator %T", op)
}

// collectsIdent は op が識別列を collect するか。識別子のリストはストアから返させず（要素の復元を持たない）、
// コーディネータで集める。
func collectsIdent(o *openQ, op logical.Node) bool {
	agg, ok := op.(*logical.Aggregation)
	if !ok {
		return false
	}
	for _, a := range agg.Aggs {
		if c, ok := a.Arg.(*logical.Column); ok && a.Fn == "collect" && o.idents[c.Key] {
			return true
		}
	}
	return false
}

// addGraph は graph の演算子（トラバーサル）を足す。開いた graph クエリが必要な識別列を持っていれば統合する。
func (pl *planner) addGraph(st *state, op logical.Node, need []string) error {
	o := st.open
	if o != nil && o.store == store.Graph && identAll(o, need) {
		pl.extend(st, op)
		return nil
	}
	pl.boundary(st, store.Graph, op, need, "")
	return nil
}

// addVarLength は可変長を足す。区間条件のうち graph 以外の列を参照するものは、区間の中身を返させて
// 要素の値を取得し、コーディネータで評価する（§5.5 案 a）。
func (pl *planner) addVarLength(st *state, x *logical.VarLengthExpand) error {
	var keep, moved []logical.SegPred
	for _, sp := range x.SegPreds {
		labels := pl.cat.ElemLabels(x.Types, sp.Part)
		inGraph := true
		for _, prop := range elemProps(sp.Pred, sp.Var) {
			k, err := pl.listStore(labels, sp.Part, prop)
			if err != nil {
				return err
			}
			if k != store.Graph {
				inGraph = false
			}
		}
		if inGraph {
			keep = append(keep, sp)
		} else {
			moved = append(moved, sp)
		}
	}
	y := *x
	y.SegPreds = keep
	if err := pl.addGraph(st, &y, []string{x.From}); err != nil {
		return err
	}
	seg := logical.SegColumn(x.Seg)
	for _, sp := range moved {
		labels := pl.cat.ElemLabels(x.Types, sp.Part)
		for _, prop := range elemProps(sp.Pred, sp.Var) {
			if err := pl.add(st, &logical.GetListProperty{Input: nil, Source: seg, Part: sp.Part, Prop: prop, ElemLabels: labels}); err != nil {
				return err
			}
		}
		fn := "nodes"
		if sp.Part == logical.SegRels {
			fn = "relationships"
		}
		pred := &logical.Quant{Kind: "all", Var: sp.Var,
			List: &logical.Call{Fn: fn, Args: []logical.Expr{&logical.Column{Key: seg}}}, Pred: sp.Pred}
		if err := pl.add(st, &logical.Filter{Pred: pred}); err != nil {
			return err
		}
	}
	return nil
}

// listStore はリスト要素のプロパティのストア。要素のラベルが複数あり得るときは、すべてのラベルで
// 同じストアでなければならない（ListFetch は 1 ストアへの取得）。違えば未対応。
func (pl *planner) listStore(labels []string, part logical.SegPart, prop string) (store.Kind, error) {
	kind := logical.NodeObj
	if part == logical.SegRels {
		kind = logical.RelObj
	}
	if len(labels) == 0 {
		return store.Graph, nil
	}
	first := pl.cat.StoreOf("", kind, labels[:1], prop)
	for _, l := range labels[1:] {
		if k := pl.cat.StoreOf("", kind, []string{l}, prop); k != first {
			return 0, unsupported("list elements of %v keep %s in different stores", labels, prop)
		}
	}
	return first, nil
}

// listLabels はリスト列（パス列・区間列）の要素のラベル候補。
func (pl *planner) listLabels(source string, part logical.SegPart) []string {
	var out []string
	add := func(ls []string) {
		for _, l := range ls {
			if !contains(out, l) {
				out = append(out, l)
			}
		}
	}
	addSeg := func(seg string) { add(pl.cat.ElemLabels(pl.segTypes[seg], part)) }
	if elems, ok := pl.pathElems[source]; ok {
		for _, e := range elems {
			switch {
			case e.Seg:
				addSeg(e.Var)
			case part == logical.SegNodes && pl.nodeLabels[e.Var] != nil:
				add(pl.nodeLabels[e.Var])
			case part == logical.SegRels && pl.relTypes[e.Var] != nil:
				add(pl.relTypes[e.Var])
			}
		}
		return out
	}
	for seg := range pl.segTypes {
		if logical.SegColumn(seg) == source {
			addSeg(seg)
		}
	}
	return out
}

func elemProps(e logical.Expr, v string) []string {
	var out []string
	seen := map[string]bool{}
	logical.Walk(e, func(x logical.Expr) {
		if ep, ok := x.(*logical.ElemProp); ok && ep.Var == v && !seen[ep.Prop] {
			seen[ep.Prop] = true
			out = append(out, ep.Prop)
		}
	})
	return out
}

func filterRefs(f *logical.Filter) []string { return logical.Refs(f) }

func availAll(o *openQ, cols []string) bool {
	for _, c := range cols {
		if !o.avail[c] {
			return false
		}
	}
	return true
}

func identAll(o *openQ, cols []string) bool {
	for _, c := range cols {
		if !o.idents[c] {
			return false
		}
	}
	return true
}

func anyKey(o *openQ, cols []string) bool {
	for _, c := range cols {
		for _, k := range o.keys {
			if c == k {
				return true
			}
		}
	}
	return false
}

// produce は op がクエリの行に足す列を記録する。
func produce(o *openQ, op logical.Node) {
	switch x := op.(type) {
	case *logical.GetProperty:
		o.avail[x.Key()] = true
	case *logical.GetListProperty:
		o.avail[x.Key()] = true
	case *logical.Expand:
		o.avail[x.Rel], o.idents[x.Rel] = true, true
		o.avail[x.To], o.idents[x.To] = true, true
	case *logical.ExpandInto:
		o.avail[x.Rel], o.idents[x.Rel] = true, true
	case *logical.VarLengthExpand:
		o.avail[x.To], o.idents[x.To] = true, true
		o.avail[logical.SegColumn(x.Seg)] = true
	case *logical.ProjectPath:
		o.avail[x.Path] = true
	case *logical.Projection:
		resetTo(o, x.Items)
	case *logical.Aggregation:
		o.grouped = true
		items := append([]logical.ProjItem{}, x.Keys...)
		for _, a := range x.Aggs {
			items = append(items, logical.ProjItem{Name: a.Name})
		}
		resetTo(o, items)
	}
}

func resetTo(o *openQ, items []logical.ProjItem) {
	avail, idents := map[string]bool{}, map[string]bool{}
	for _, it := range items {
		avail[it.Name] = true
		if c, ok := it.Expr.(*logical.Column); ok && o.idents[c.Key] {
			idents[it.Name] = true
		}
	}
	o.avail, o.idents = avail, idents
}

// ================================
// 要求解析（各クエリが返す列・区間列の出力）
// ================================

// finalize は上から下へ「上が必要とする列」を伝え、StoreQuery.Outputs と VarLengthExpand.EmitPath を決める。
// all は「根なので出力列すべてが必要」。
func finalize(n Node, need map[string]bool, all bool) {
	switch x := n.(type) {
	case *StoreQuery:
		cols := topCols(x.Ops)
		var out []string
		if all {
			out = cols
		} else {
			out = append(out, x.Keys...)
			var rest []string
			for _, c := range cols {
				if need[c] && !contains(x.Keys, c) {
					rest = append(rest, c)
				}
			}
			sort.Strings(rest)
			out = append(out, rest...)
		}
		x.Outputs = out
		markEmit(x.Ops, out)
		x.Weighted = settings.WeightedRows && weighable(x)
		if x.Input != nil {
			below := map[string]bool{}
			for k := range need {
				if !contains(cols, k) {
					below[k] = true
				}
			}
			for _, k := range x.Keys {
				below[k] = true
			}
			finalize(x.Input, below, all)
		}
	case *OptionalApply:
		// 内側が作る列（一致しない行で null にする列）を控え、外側には自分の列と Keys を、内側には
		// 上が使う列と Keys を求める。
		x.InnerCols = innerCols(x.Inner)
		outerNeed := map[string]bool{}
		innerNeed := map[string]bool{}
		for k := range need {
			if contains(x.InnerCols, k) {
				innerNeed[k] = true
			} else {
				outerNeed[k] = true
			}
		}
		for _, k := range x.Keys {
			outerNeed[k] = true
			innerNeed[k] = true
		}
		finalize(x.Outer, outerNeed, all)
		finalize(x.Inner, innerNeed, false)
	case *ListFetch:
		below := map[string]bool{x.Source: true}
		for k := range need {
			if k != x.Key() {
				below[k] = true
			}
		}
		finalize(x.Input, below, all)
	case *Coord:
		below := map[string]bool{}
		belowAll := all
		switch x.Op.(type) {
		case *logical.Projection, *logical.Aggregation:
			belowAll = false
		default:
			for k := range need {
				below[k] = true
			}
		}
		for _, k := range logical.Refs(x.Op) {
			below[k] = true
		}
		for _, in := range x.Inputs {
			finalize(in, below, belowAll)
		}
	}
}

// innerCols は OptionalApply の内側が作る列（StoreQuery の topCols と Coord の生成列の和、ArgRows の列を除く）。
func innerCols(n Node) []string {
	var out []string
	add := func(ks ...string) {
		for _, k := range ks {
			if !contains(out, k) {
				out = append(out, k)
			}
		}
	}
	var walk func(Node)
	walk = func(n Node) {
		switch x := n.(type) {
		case *StoreQuery:
			for _, c := range topCols(x.Ops) {
				if !contains(x.Keys, c) {
					add(c)
				}
			}
		case *ListFetch:
			add(x.Key())
		case *Coord:
			if pp, ok := x.Op.(*logical.ProjectPath); ok {
				add(pp.Path)
			}
		case *OptionalApply:
			add(x.New...)
			add(innerCols(x.Inner)...)
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

// topCols はクエリの最上位の行にある列（生成順。Projection/Aggregation で付け替わる）。
func topCols(n logical.Node) []string {
	var cols []string
	add := func(ks ...string) {
		for _, k := range ks {
			if !contains(cols, k) {
				cols = append(cols, k)
			}
		}
	}
	var walk func(logical.Node)
	walk = func(n logical.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *logical.NodeHashJoin:
			walk(x.Left)
			walk(x.Right)
			return
		case *logical.Optional:
			walk(x.Outer)
			walk(x.Inner)
			return
		case *Argument:
			add(x.Cols...)
			return
		case *logical.EntityScan:
			add(x.Alias)
			return
		}
		walk(logical.Child(n))
		switch x := n.(type) {
		case *logical.GetProperty:
			add(x.Key())
		case *logical.GetListProperty:
			add(x.Key())
		case *logical.Expand:
			add(x.Rel, x.To)
		case *logical.ExpandInto:
			add(x.Rel)
		case *logical.VarLengthExpand:
			add(x.To, logical.SegColumn(x.Seg))
		case *logical.ProjectPath:
			add(x.Path)
		case *logical.Projection:
			cols = nil
			for _, it := range x.Items {
				add(it.Name)
			}
		case *logical.Aggregation:
			cols = nil
			for _, it := range x.Keys {
				add(it.Name)
			}
			for _, a := range x.Aggs {
				add(a.Name)
			}
		}
	}
	walk(n)
	return cols
}

// weighable は StoreQuery の結果を重み付きの 1 行にまとめてよいか: graph で、可変長の展開を含み
// （経路ごとの重複が出る）、区間・パスを返さず、行を並べ替え・作り直す演算子を吸収していないもの。
func weighable(q *StoreQuery) bool {
	if q.Store != store.Graph {
		return false
	}
	varLen, ok := false, true
	var walk func(logical.Node)
	walk = func(n logical.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *logical.VarLengthExpand:
			varLen = true
			if contains(q.Outputs, logical.SegColumn(x.Seg)) {
				ok = false
			}
		case *logical.ProjectPath:
			if contains(q.Outputs, x.Path) {
				ok = false
			}
		case *logical.Projection, *logical.Aggregation, *logical.Sort, *logical.Limit, *logical.Distinct:
			ok = false
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(q.Ops)
	return varLen && ok
}

// markEmit は返す列に区間列が含まれる VarLengthExpand に EmitPath を立てる。
func markEmit(n logical.Node, out []string) {
	if n == nil {
		return
	}
	if v, ok := n.(*logical.VarLengthExpand); ok && contains(out, logical.SegColumn(v.Seg)) {
		v.EmitPath = true
	}
	for _, c := range n.Children() {
		markEmit(c, out)
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// ================================
// リテラルの型付け
// ================================

// typeLiterals は比較・IN の相手が型の分かる列（mapping の type）なら、リテラルをその型へ変換する。
// ストアに委譲する式もコーディネータで評価する式も、同じ型付き値で比べられるように計画の段階で行う。
func (pl *planner) typeLiterals(op logical.Node) {
	switch x := op.(type) {
	case *logical.Filter:
		x.Pred = pl.typeExpr(x.Pred, nil)
	case *logical.Projection:
		for i := range x.Items {
			x.Items[i].Expr = pl.typeExpr(x.Items[i].Expr, nil)
		}
	case *logical.VarLengthExpand:
		for i, sp := range x.SegPreds {
			el := map[string]owner{sp.Var: pl.segElemOwner(x.Types, sp.Part)}
			x.SegPreds[i].Pred = pl.typeExpr(sp.Pred, el)
		}
	}
}

func (pl *planner) segElemOwner(types []string, part logical.SegPart) owner {
	if part == logical.SegRels {
		return owner{kind: logical.RelObj, labels: types}
	}
	return owner{kind: logical.NodeObj, labels: pl.cat.ElemLabels(types, part)}
}

// exprType は列・要素参照の型名（不明なら空）。
func (pl *planner) exprType(e logical.Expr, elems map[string]owner) string {
	switch x := e.(type) {
	case *logical.Column:
		i := strings.LastIndex(x.Key, ".")
		if i < 0 {
			return ""
		}
		o, ok := pl.owners[x.Key[:i]]
		if !ok {
			return ""
		}
		return pl.cat.TypeOf(o.kind, o.labels, x.Key[i+1:])
	case *logical.ElemProp:
		o, ok := elems[x.Var]
		if !ok {
			return ""
		}
		return pl.cat.TypeOf(o.kind, o.labels, x.Prop)
	}
	return ""
}

func (pl *planner) typeExpr(e logical.Expr, elems map[string]owner) logical.Expr {
	conv := func(v logical.Expr, typ string) logical.Expr {
		switch l := v.(type) {
		case *logical.Lit:
			if typ == "" || l.Val == nil {
				return l
			}
			if t, err := codec.ConvertToNativeType(l.Val, typ); err == nil {
				return &logical.Lit{Val: t}
			}
		case *logical.List:
			out := &logical.List{}
			for _, el := range l.Elems {
				if lit, ok := el.(*logical.Lit); ok && typ != "" && lit.Val != nil {
					if t, err := codec.ConvertToNativeType(lit.Val, typ); err == nil {
						out.Elems = append(out.Elems, &logical.Lit{Val: t})
						continue
					}
				}
				out.Elems = append(out.Elems, el)
			}
			return out
		}
		return v
	}
	switch x := e.(type) {
	case *logical.Bin:
		switch x.Op {
		case "=", "<>", "<", ">", "<=", ">=", "IN":
			l, r := x.L, x.R
			if t := pl.exprType(x.L, elems); t != "" {
				r = conv(r, t)
			}
			if t := pl.exprType(x.R, elems); t != "" && x.Op != "IN" {
				l = conv(l, t)
			}
			return &logical.Bin{Op: x.Op, L: l, R: r}
		}
		return &logical.Bin{Op: x.Op, L: pl.typeExpr(x.L, elems), R: pl.typeExpr(x.R, elems)}
	case *logical.Un:
		return &logical.Un{Op: x.Op, X: pl.typeExpr(x.X, elems)}
	case *logical.Case:
		c := &logical.Case{}
		for _, w := range x.Whens {
			c.Whens = append(c.Whens, logical.CaseWhen{When: pl.typeExpr(w.When, elems), Then: w.Then})
		}
		c.Else = x.Else
		return c
	case *logical.Quant:
		inner := map[string]owner{}
		for k, v := range elems {
			inner[k] = v
		}
		if call, ok := x.List.(*logical.Call); ok && len(call.Args) == 1 {
			if src, ok := call.Args[0].(*logical.Column); ok {
				part := logical.SegNodes
				kind := logical.NodeObj
				if call.Fn == "relationships" {
					part, kind = logical.SegRels, logical.RelObj
				}
				inner[x.Var] = owner{kind: kind, labels: pl.listLabels(src.Key, part)}
			}
		}
		return &logical.Quant{Kind: x.Kind, Var: x.Var, List: x.List, Pred: pl.typeExpr(x.Pred, inner)}
	}
	return e
}
