package logical

import (
	"fmt"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/settings"
)

// translate.go は段階 B: クエリ AST を論理計画へ翻訳する（docs/operator_plan.md §4.2〜§4.5）。
// ストアは一切考えない。述語は変数が揃った直後に置く（述語の前倒しを組み立て時に済ませる）。
// パス全体の条件は、できるだけ普通の述語と区間条件へ分解する（§4.4）。

// VarKind は束縛変数の種類。
type VarKind int

const (
	VarNode VarKind = iota
	VarRel
	VarPath
	VarValue
	VarSeg // 可変長区間（内部用。ユーザーからは参照できない）
)

// varInfo はスコープ内の変数の情報。
type varInfo struct {
	kind   VarKind
	labels []string // ノードのラベル / 関係のタイプ
	elem   *varInfo // 値列が束縛のリスト（collect(n)）なら要素の情報。UNWIND で要素を束縛に戻す
}

// pathDef はパス変数の構成（パターンの左から右の順）。
type pathDef struct {
	elems []PathElem
	nodes []string // 束縛ノード（端点を含む）
	rels  []string // 固定長の関係
	segs  []string // 可変長区間
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ast.ErrUnsupported, fmt.Sprintf(format, args...))
}

// Translate はクエリ AST を論理計画へ翻訳する。翻訳できない形は ast.ErrUnsupported を包んだエラー。
func Translate(q *ast.Query) (Node, error) {
	b := &builder{
		q:     q,
		scope: map[string]varInfo{},
		avail: map[string]bool{},
		paths: map[string]pathDef{},
		bound: map[string]bool{},
	}
	for i, c := range q.Clauses {
		b.clauseIdx = i
		var err error
		switch c := c.(type) {
		case *ast.MatchClause:
			err = b.match(c)
		case *ast.ProjectionClause:
			err = b.project(c)
		case *ast.UnwindClause:
			err = b.unwind(c)
		}
		if err != nil {
			return nil, err
		}
	}
	return b.plan, nil
}

type builder struct {
	q         *ast.Query
	clauseIdx int
	plan      Node
	scope     map[string]varInfo
	avail     map[string]bool // 行に既にある列（値列・パス列）
	paths     map[string]pathDef
	bound     map[string]bool // 束縛済みの変数（ノード・関係・区間・値）
}

// ================================
// 式の変換
// ================================

// convCtx は量化子の反復変数 → 対象パス（と部分）。
type convCtx struct {
	quant map[string]quantSrc
}

type quantSrc struct {
	path string
	part SegPart
}

// conv は AST の式を論理式へ変換する。必要な GetProperty / GetListProperty / ProjectPath は
// 呼び出し側が needs で求めて置く。
func (b *builder) conv(e ast.Expr, ctx convCtx) (Expr, error) {
	switch x := e.(type) {
	case *ast.Lit:
		return &Lit{Val: x.Val}, nil
	case *ast.VarRef:
		if _, ok := ctx.quant[x.Name]; ok {
			return nil, unsupported("bare iteration variable %s", x.Name)
		}
		vi, ok := b.scope[x.Name]
		if !ok {
			return nil, unsupported("unknown variable %s", x.Name)
		}
		if vi.kind == VarSeg {
			return nil, unsupported("variable-length relationship variable %s as a value", x.Name)
		}
		return &Column{Key: x.Name}, nil
	case *ast.PropRef:
		if _, ok := ctx.quant[x.Var]; ok {
			return &ElemProp{Var: x.Var, Prop: x.Prop}, nil
		}
		vi, ok := b.scope[x.Var]
		if !ok {
			return nil, unsupported("unknown variable %s", x.Var)
		}
		if vi.kind != VarNode && vi.kind != VarRel {
			return nil, unsupported("property access on non-entity %s", x.Var)
		}
		return &Column{Key: x.Var + "." + x.Prop}, nil
	case *ast.Bin:
		l, err := b.conv(x.L, ctx)
		if err != nil {
			return nil, err
		}
		r, err := b.conv(x.R, ctx)
		if err != nil {
			return nil, err
		}
		return &Bin{Op: x.Op, L: l, R: r}, nil
	case *ast.Un:
		v, err := b.conv(x.X, ctx)
		if err != nil {
			return nil, err
		}
		return &Un{Op: x.Op, X: v}, nil
	case *ast.Call:
		if isPathFn(x.Fn) {
			p, err := b.pathArg(x)
			if err != nil {
				return nil, err
			}
			return &Call{Fn: x.Fn, Args: []Expr{&Column{Key: p}}}, nil
		}
		c := &Call{Fn: x.Fn}
		for _, a := range x.Args {
			v, err := b.conv(a, ctx)
			if err != nil {
				return nil, err
			}
			c.Args = append(c.Args, v)
		}
		return c, nil
	case *ast.Case:
		c := &Case{}
		var subj Expr
		if x.Subject != nil {
			v, err := b.conv(x.Subject, ctx)
			if err != nil {
				return nil, err
			}
			subj = v
		}
		for _, w := range x.Whens {
			wv, err := b.conv(w.When, ctx)
			if err != nil {
				return nil, err
			}
			if subj != nil { // 単純 CASE は WHEN subject = 値 として評価する（null とは一致しない）
				wv = &Bin{Op: "=", L: subj, R: wv}
			}
			tv, err := b.conv(w.Then, ctx)
			if err != nil {
				return nil, err
			}
			c.Whens = append(c.Whens, CaseWhen{When: wv, Then: tv})
		}
		if x.Else != nil {
			ev, err := b.conv(x.Else, ctx)
			if err != nil {
				return nil, err
			}
			c.Else = ev
		}
		return c, nil
	case *ast.List:
		l := &List{}
		for _, el := range x.Elems {
			v, err := b.conv(el, ctx)
			if err != nil {
				return nil, err
			}
			l.Elems = append(l.Elems, v)
		}
		return l, nil
	case *ast.Quant:
		src, err := b.quantSource(x)
		if err != nil {
			return nil, err
		}
		inner := convCtx{quant: map[string]quantSrc{}}
		for k, v := range ctx.quant {
			inner.quant[k] = v
		}
		inner.quant[x.Var] = src
		pred, err := b.conv(x.Pred, inner)
		if err != nil {
			return nil, err
		}
		fn := "nodes"
		if src.part == SegRels {
			fn = "relationships"
		}
		return &Quant{Kind: x.Kind, Var: x.Var,
			List: &Call{Fn: fn, Args: []Expr{&Column{Key: src.path}}}, Pred: pred}, nil
	case *ast.Agg:
		return nil, unsupported("aggregate outside projection")
	}
	return nil, unsupported("expression %T", e)
}

func isPathFn(fn string) bool { return fn == "length" || fn == "nodes" || fn == "relationships" }

func (b *builder) pathArg(c *ast.Call) (string, error) {
	if len(c.Args) != 1 {
		return "", unsupported("%s arity", c.Fn)
	}
	v, ok := c.Args[0].(*ast.VarRef)
	if !ok || b.scope[v.Name].kind != VarPath {
		return "", unsupported("%s argument must be a path variable", c.Fn)
	}
	if _, bound := b.scope[v.Name]; !bound {
		return "", unsupported("unknown path %s", v.Name)
	}
	return v.Name, nil
}

// quantSource は量化子のリストが nodes(p) / relationships(p) であることを確かめる。
func (b *builder) quantSource(x *ast.Quant) (quantSrc, error) {
	c, ok := x.List.(*ast.Call)
	if !ok || (c.Fn != "nodes" && c.Fn != "relationships") {
		return quantSrc{}, unsupported("quantifier over %s", x.List)
	}
	p, err := b.pathArg(c)
	if err != nil {
		return quantSrc{}, err
	}
	part := SegNodes
	if c.Fn == "relationships" {
		part = SegRels
	}
	return quantSrc{path: p, part: part}, nil
}

// needs は式の評価に必要な列を集める。
type needs struct {
	props     []propNeed
	listProps []listNeed
	paths     []string
}

type propNeed struct{ v, prop string }
type listNeed struct {
	path string
	part SegPart
	prop string
}

func (b *builder) collectNeeds(e ast.Expr, ctx convCtx, n *needs) {
	switch x := e.(type) {
	case *ast.PropRef:
		if src, ok := ctx.quant[x.Var]; ok {
			n.listProps = append(n.listProps, listNeed{src.path, src.part, x.Prop})
			return
		}
		n.props = append(n.props, propNeed{x.Var, x.Prop})
	case *ast.VarRef:
		if b.scope[x.Name].kind == VarPath {
			n.paths = append(n.paths, x.Name)
		}
	case *ast.Bin:
		b.collectNeeds(x.L, ctx, n)
		b.collectNeeds(x.R, ctx, n)
	case *ast.Un:
		b.collectNeeds(x.X, ctx, n)
	case *ast.Call:
		for _, a := range x.Args {
			b.collectNeeds(a, ctx, n)
		}
	case *ast.Case:
		if x.Subject != nil {
			b.collectNeeds(x.Subject, ctx, n)
		}
		for _, w := range x.Whens {
			b.collectNeeds(w.When, ctx, n)
			b.collectNeeds(w.Then, ctx, n)
		}
		if x.Else != nil {
			b.collectNeeds(x.Else, ctx, n)
		}
	case *ast.List:
		for _, el := range x.Elems {
			b.collectNeeds(el, ctx, n)
		}
	case *ast.Quant:
		src, err := b.quantSource(x)
		if err != nil {
			return // conv が同じエラーを返す
		}
		n.paths = append(n.paths, src.path)
		inner := convCtx{quant: map[string]quantSrc{x.Var: src}}
		for k, v := range ctx.quant {
			inner.quant[k] = v
		}
		b.collectNeeds(x.Pred, inner, n)
	case *ast.Agg:
		if x.Arg != nil {
			b.collectNeeds(x.Arg, ctx, n)
		}
	}
}

// supply は needs の列を、まだ行に無いものだけ plan に足す。
func (b *builder) supply(n Node, nd needs) Node {
	for _, p := range nd.paths {
		if !b.avail[p] {
			n = &ProjectPath{Input: n, Path: p, Elems: b.paths[p].elems}
			b.avail[p] = true
		}
	}
	for _, p := range nd.props {
		key := p.v + "." + p.prop
		if b.avail[key] {
			continue
		}
		vi := b.scope[p.v]
		kind := NodeObj
		if vi.kind == VarRel {
			kind = RelObj
		}
		n = &GetProperty{Input: n, Alias: p.v, Prop: p.prop, Kind: kind, Labels: vi.labels}
		b.avail[key] = true
	}
	for _, lp := range nd.listProps {
		key := ListPropKey(lp.path, lp.part, lp.prop)
		if b.avail[key] {
			continue
		}
		n = &GetListProperty{Input: n, Source: lp.path, Part: lp.part, Prop: lp.prop}
		b.avail[key] = true
	}
	return n
}

// deps は式が依存する束縛（ノード・関係・区間）。パス変数は構成要素に展開する。
func (b *builder) deps(e ast.Expr) []string {
	seen := map[string]bool{}
	var out []string
	add := func(v string) {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	var walk func(ast.Expr, map[string]bool)
	walk = func(e ast.Expr, qv map[string]bool) {
		switch x := e.(type) {
		case *ast.VarRef:
			if qv[x.Name] {
				return
			}
			if pd, ok := b.paths[x.Name]; ok {
				for _, el := range pd.elems {
					add(el.Var)
				}
				return
			}
			add(x.Name)
		case *ast.PropRef:
			if !qv[x.Var] {
				add(x.Var)
			}
		case *ast.Bin:
			walk(x.L, qv)
			walk(x.R, qv)
		case *ast.Un:
			walk(x.X, qv)
		case *ast.Call:
			for _, a := range x.Args {
				walk(a, qv)
			}
		case *ast.Case:
			if x.Subject != nil {
				walk(x.Subject, qv)
			}
			for _, w := range x.Whens {
				walk(w.When, qv)
				walk(w.Then, qv)
			}
			if x.Else != nil {
				walk(x.Else, qv)
			}
		case *ast.List:
			for _, el := range x.Elems {
				walk(el, qv)
			}
		case *ast.Quant:
			walk(x.List, qv)
			inner := map[string]bool{x.Var: true}
			for k := range qv {
				inner[k] = true
			}
			walk(x.Pred, inner)
		case *ast.Agg:
			if x.Arg != nil {
				walk(x.Arg, qv)
			}
		}
	}
	walk(e, map[string]bool{})
	return out
}

// ================================
// MATCH（§4.3）
// ================================

type patVertex struct {
	name   string
	labels []string
	order  int
}

type patEdge struct {
	rel     ast.RelPat
	left    string
	right   string
	order   int
	pathVar string
}

func (b *builder) match(m *ast.MatchClause) error {
	if m.Optional {
		return b.optionalMatch(m)
	}
	// 1. パターングラフ
	vertices := map[string]*patVertex{}
	var vorder []string
	var edges []*patEdge
	addVertex := func(n ast.NodePat) error {
		if vi, ok := b.scope[n.Var]; ok && vi.kind != VarNode {
			return unsupported("variable %s rebound as node", n.Var)
		}
		v, ok := vertices[n.Var]
		if !ok {
			v = &patVertex{name: n.Var, order: len(vorder)}
			if vi, bound := b.scope[n.Var]; bound {
				v.labels = vi.labels
			}
			vertices[n.Var] = v
			vorder = append(vorder, n.Var)
		}
		for _, l := range n.Labels {
			if !contains(v.labels, l) {
				if _, bound := b.scope[n.Var]; bound {
					return unsupported("new label on bound variable %s", n.Var)
				}
				v.labels = append(v.labels, l)
			}
		}
		return nil
	}

	var pending []ast.Expr
	if m.Where != nil {
		pending = splitAnd(m.Where)
	}
	for _, p := range m.Patterns {
		var pd pathDef
		for i, n := range p.Nodes {
			if err := addVertex(n); err != nil {
				return err
			}
			for _, pe := range n.Props {
				pending = append(pending, &ast.Bin{Op: "=", L: &ast.PropRef{Var: n.Var, Prop: pe.Key}, R: pe.Value})
			}
			pd.elems = append(pd.elems, PathElem{Var: n.Var})
			pd.nodes = append(pd.nodes, n.Var)
			if i < len(p.Rels) {
				r := p.Rels[i]
				if _, bound := b.scope[r.Var]; bound {
					return unsupported("relationship variable %s already bound", r.Var)
				}
				edges = append(edges, &patEdge{rel: r, left: n.Var, right: p.Nodes[i+1].Var, order: len(edges), pathVar: p.PathVar})
				if r.Range != nil {
					if len(r.Props) > 0 {
						return unsupported("properties on variable-length relationship %s", r.Var)
					}
					pd.elems = append(pd.elems, PathElem{Var: r.Var, Seg: true})
					pd.segs = append(pd.segs, r.Var)
				} else {
					pd.elems = append(pd.elems, PathElem{Var: r.Var})
					pd.rels = append(pd.rels, r.Var)
					for _, pe := range r.Props {
						pending = append(pending, &ast.Bin{Op: "=", L: &ast.PropRef{Var: r.Var, Prop: pe.Key}, R: pe.Value})
					}
				}
			}
		}
		if p.PathVar != "" {
			if _, bound := b.scope[p.PathVar]; bound {
				return unsupported("path variable %s already bound", p.PathVar)
			}
			b.paths[p.PathVar] = pd
		}
	}

	// 変数をスコープへ登録（述語の変換・deps 判定の前に種類が分かっている必要がある）。
	for _, v := range vorder {
		b.scope[v] = varInfo{kind: VarNode, labels: vertices[v].labels}
	}
	for _, e := range edges {
		kind := VarRel
		if e.rel.Range != nil {
			kind = VarSeg
		}
		b.scope[e.rel.Var] = varInfo{kind: kind, labels: e.rel.Types}
	}
	for _, p := range m.Patterns {
		if p.PathVar != "" {
			b.scope[p.PathVar] = varInfo{kind: VarPath}
		}
	}
	// ただし「束縛済み」の判定は bound 集合で行う（scope は種類の参照用）。
	prevBound := b.boundSet()

	// 2. パス全体の条件の分解（§4.4）
	segRange := map[string]HopRange{}
	segPreds := map[string][]SegPred{}
	for _, e := range edges {
		if e.rel.Range != nil {
			segRange[e.rel.Var] = hopRange(e.rel.Range)
		}
	}
	var err error
	pending, err = b.decomposePathPreds(pending, segRange, segPreds)
	if err != nil {
		return err
	}

	// 3. 連結成分ごとに積む（起点が 2 つ以上あれば分割して両側から展開して合流する）
	g := &growCtx{vertices: vertices, edges: edges, segRange: segRange, segPreds: segPreds}
	comps := components(vorder, edges)
	bound := prevBound
	for _, comp := range comps {
		inComp := map[string]bool{}
		for _, v := range comp {
			inComp[v] = true
		}
		if groups := b.anchorGroups(comp, vertices, bound, pending); settings.SplitAnchoredPatterns && len(groups) >= 2 {
			pending, err = b.buildSplit(g, comp, groups, inComp, bound, pending)
			if err != nil {
				return err
			}
			continue
		}
		start := b.chooseStart(comp, vertices, bound, pending)
		fresh := !bound[start]
		var sub Node
		compBound := map[string]bool{}
		if fresh {
			// 新しい成分は前の束縛と独立に組み、最後に直積で合流する。
			sub = &EntityScan{Alias: start, Labels: vertices[start].labels}
		} else {
			sub = b.plan
			for k := range bound {
				compBound[k] = true
			}
		}
		compBound[start] = true
		sub, pending, err = b.grow(g, sub, compBound, pending, func(e *patEdge) bool { return inComp[e.left] })
		if err != nil {
			return err
		}
		if fresh && b.plan != nil {
			b.plan = &NodeHashJoin{Left: b.plan, Right: sub}
		} else {
			b.plan = sub
		}
		for k := range compBound {
			bound[k] = true
		}
	}
	b.bound = bound

	// 4. 関係一意性（タイプが重なり得る関係が 2 つ以上あるとき）
	if rels, segs := uniqueCandidates(edges); len(rels)+len(segs) >= 2 {
		b.plan = &RelUniqueFilter{Input: b.plan, Rels: rels, Segs: segs}
	}

	// 5. 残り（成分をまたぐ述語・パス列が必要な述語）
	b.plan, pending, err = b.placeReady(b.plan, bound, pending)
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return unsupported("unplaceable predicate %s", pending[0])
	}
	return nil
}

// growCtx は 1 つの MATCH のパターン情報（成分・分割側の展開で共有する）。
type growCtx struct {
	vertices map[string]*patVertex
	edges    []*patEdge
	segRange map[string]HopRange
	segPreds map[string][]SegPred
}

// grow は sub を、束縛済みの頂点から allowed の辺を出現順にたどって伸ばし、揃った述語を置く。
// 両端が束縛済みの可変長は、仮の頂点へ展開してから端点の一致で絞る（Filter(tmp = to)）。
func (b *builder) grow(g *growCtx, sub Node, compBound map[string]bool, pending []ast.Expr, allowed func(*patEdge) bool) (Node, []ast.Expr, error) {
	var err error
	sub, pending, err = b.placeReady(sub, compBound, pending)
	if err != nil {
		return nil, nil, err
	}
	done := map[int]bool{}
	for {
		e := nextEdge(g.edges, allowed, compBound, done)
		if e == nil {
			break
		}
		done[e.order] = true
		from, to, dir := orient(e, compBound)
		switch {
		case compBound[to] && e.rel.Range != nil:
			tmp := "_vl_" + e.rel.Var
			b.scope[tmp] = varInfo{kind: VarNode, labels: g.vertices[to].labels}
			sub = &VarLengthExpand{Input: sub, From: from, Seg: e.rel.Var, To: tmp, Types: e.rel.Types,
				Dir: dir, ToLabels: g.vertices[to].labels, Range: g.segRange[e.rel.Var], SegPreds: g.segPreds[e.rel.Var]}
			sub = &Filter{Input: sub, Pred: &Bin{Op: "=", L: &Column{Key: tmp}, R: &Column{Key: to}}}
			compBound[e.rel.Var] = true
		case compBound[to]:
			sub = &ExpandInto{Input: sub, From: from, Rel: e.rel.Var, To: to, Types: e.rel.Types, Dir: dir}
			compBound[e.rel.Var] = true
		case e.rel.Range != nil:
			sub = &VarLengthExpand{Input: sub, From: from, Seg: e.rel.Var, To: to, Types: e.rel.Types,
				Dir: dir, ToLabels: g.vertices[to].labels, Range: g.segRange[e.rel.Var], SegPreds: g.segPreds[e.rel.Var]}
			compBound[e.rel.Var] = true
			compBound[to] = true
		default:
			sub = &Expand{Input: sub, From: from, Rel: e.rel.Var, To: to, Types: e.rel.Types,
				Dir: dir, ToLabels: g.vertices[to].labels}
			compBound[e.rel.Var] = true
			compBound[to] = true
		}
		sub, pending, err = b.placeReady(sub, compBound, pending)
		if err != nil {
			return nil, nil, err
		}
	}
	return sub, pending, nil
}

// anchorGroups は成分の起点の組: 前の節で束縛済みの頂点（あれば 1 組にまとめて先頭）と、一意キーの等値で
// 絞れる新しい頂点（1 つずつ、出現順）。2 組以上なら分割できる。
func (b *builder) anchorGroups(comp []string, vertices map[string]*patVertex, bound map[string]bool, pending []ast.Expr) [][]string {
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

// uniqueAnchor は v だけで評価できる一意キーの等値述語（{id: $x}）があるか。
func (b *builder) uniqueAnchor(v string, pending []ast.Expr) bool {
	for _, e := range pending {
		d := b.deps(e)
		if len(d) != 1 || d[0] != v {
			continue
		}
		if bin, ok := e.(*ast.Bin); ok && bin.Op == "=" {
			if pr, ok := bin.L.(*ast.PropRef); ok && isUniqueKey(pr.Prop) {
				if _, lit := bin.R.(*ast.Lit); lit {
					return true
				}
			}
		}
	}
	return false
}

// buildSplit は起点が複数ある成分を分割して組む。頂点を最寄りの起点の組へ（多始点の幅優先・同距離なら先の組）
// 振り分け、辺は両端の組の大きい方の側が展開する。各側を起点から伸ばし、両側で束縛される頂点をキーに
// NodeHashJoin で合流する（束縛済みの組の側は前の計画の上に伸ばす）。
func (b *builder) buildSplit(g *growCtx, comp []string, groups [][]string, inComp map[string]bool, bound map[string]bool, pending []ast.Expr) ([]ast.Expr, error) {
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
			if !inComp[e.left] {
				continue
			}
			var w string
			switch v {
			case e.left:
				w = e.right
			case e.right:
				w = e.left
			default:
				continue
			}
			if _, ok := region[w]; !ok {
				region[w] = region[v]
				queue = append(queue, w)
			}
		}
	}
	// 組をまたぐ辺は後の組の側が展開する（同距離の頂点は先の組に入るので、後の組がそこまで伸びて中央で合流する）。
	side := func(e *patEdge) int {
		if region[e.left] > region[e.right] {
			return region[e.left]
		}
		return region[e.right]
	}

	var acc Node
	accBound := map[string]bool{}
	hasBound := bound[groups[0][0]]
	for gi, grp := range groups {
		var sub Node
		sb := map[string]bool{}
		if gi == 0 && hasBound {
			sub = b.plan
			for k := range bound {
				sb[k] = true
			}
		} else {
			sub = &EntityScan{Alias: grp[0], Labels: g.vertices[grp[0]].labels}
			sb[grp[0]] = true
		}
		gi := gi
		var err error
		sub, pending, err = b.grow(g, sub, sb, pending, func(e *patEdge) bool { return inComp[e.left] && side(e) == gi })
		if err != nil {
			return nil, err
		}
		if acc == nil {
			acc = sub
		} else {
			var keys []string
			for _, v := range comp {
				if accBound[v] && sb[v] {
					keys = append(keys, v)
				}
			}
			acc = &NodeHashJoin{Left: acc, Right: sub, Keys: keys}
		}
		for k := range sb {
			accBound[k] = true
		}
	}
	if !hasBound && b.plan != nil {
		acc = &NodeHashJoin{Left: b.plan, Right: acc}
	}
	b.plan = acc
	for k := range accBound {
		bound[k] = true
	}
	var err error
	b.plan, pending, err = b.placeReady(b.plan, bound, pending)
	return pending, err
}

// optionalMatch は OPTIONAL MATCH を左外部 apply（Optional）にする。パターンと WHERE は、外側の行が持つ
// 束縛済みの識別列（Keys）だけを受け取る Argument の上に、通常の MATCH と同じ規則で組む（WHERE は一致の条件）。
func (b *builder) optionalMatch(m *ast.MatchClause) error {
	if b.plan == nil {
		return unsupported("OPTIONAL MATCH as the first clause")
	}
	// Keys: パターンと WHERE が参照する、束縛済みのノード・関係（出現順）。
	var keys []string
	seen := map[string]bool{}
	consider := func(v string) error {
		if seen[v] || !b.bound[v] {
			return nil
		}
		seen[v] = true
		switch b.scope[v].kind {
		case VarNode, VarRel:
			keys = append(keys, v)
			return nil
		}
		return unsupported("OPTIONAL MATCH referring to value %s", v)
	}
	var order []string // 新しく束縛される変数の候補（出現順）
	for _, p := range m.Patterns {
		for i, n := range p.Nodes {
			if err := consider(n.Var); err != nil {
				return err
			}
			order = append(order, n.Var)
			if i < len(p.Rels) {
				if err := consider(p.Rels[i].Var); err != nil {
					return err
				}
				order = append(order, p.Rels[i].Var)
			}
		}
	}
	if m.Where != nil {
		for _, d := range b.deps(m.Where) {
			if err := consider(d); err != nil {
				return err
			}
		}
	}

	outer, outerBound, outerAvail := b.plan, b.bound, b.avail
	b.plan = &Argument{Cols: keys}
	b.bound = map[string]bool{}
	for _, k := range keys {
		b.bound[k] = true
	}
	b.avail = map[string]bool{}
	if err := b.match(&ast.MatchClause{Patterns: m.Patterns, Where: m.Where}); err != nil {
		return err
	}
	inner, innerBound, innerAvail := b.plan, b.bound, b.avail

	var newVars []string
	added := map[string]bool{}
	for _, v := range order {
		if innerBound[v] && !seen[v] && !added[v] {
			added[v] = true
			newVars = append(newVars, v)
		}
	}
	b.plan = &Optional{Outer: outer, Inner: inner, Keys: keys, New: newVars}
	b.bound, b.avail = outerBound, outerAvail
	for k := range innerBound {
		b.bound[k] = true
	}
	for k := range innerAvail {
		b.avail[k] = true
	}
	return nil
}

// unwind は UNWIND。束縛のリスト（collect(n) の値列）を展開した変数は、元の束縛と同じ種類の識別列になる。
func (b *builder) unwind(u *ast.UnwindClause) error {
	if b.plan == nil {
		return unsupported("UNWIND as the first clause")
	}
	if _, ok := b.scope[u.Var]; ok {
		return unsupported("UNWIND rebinding %s", u.Var)
	}
	var nd needs
	b.collectNeeds(u.Expr, convCtx{}, &nd)
	b.plan = b.supply(b.plan, nd)
	e, err := b.conv(u.Expr, convCtx{})
	if err != nil {
		return err
	}
	b.plan = &Unwind{Input: b.plan, List: e, Var: u.Var}
	info := varInfo{kind: VarValue}
	if v, ok := u.Expr.(*ast.VarRef); ok && b.scope[v.Name].elem != nil {
		info = *b.scope[v.Name].elem
	}
	b.scope[u.Var] = info
	b.bound[u.Var] = true
	if info.kind == VarValue {
		b.avail[u.Var] = true
	}
	return nil
}

// onlyOutputs は式が射影後の列名（outNames）だけを変数として参照するか（プロパティ参照・集約・量化子は不可）。
func onlyOutputs(e ast.Expr, outNames map[string]bool) bool {
	ok, any := true, false
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.VarRef:
			any = true
			if !outNames[x.Name] {
				ok = false
			}
		case *ast.Lit:
		case *ast.Bin:
			walk(x.L)
			walk(x.R)
		case *ast.Un:
			walk(x.X)
		case *ast.Call:
			if isPathFn(x.Fn) {
				ok = false
			}
			for _, a := range x.Args {
				walk(a)
			}
		default:
			ok = false
		}
	}
	walk(e)
	return ok && any
}

// boundSet は現在束縛済みの変数（前の節までの結果）。
func (b *builder) boundSet() map[string]bool {
	out := map[string]bool{}
	for k := range b.bound {
		out[k] = true
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func splitAnd(e ast.Expr) []ast.Expr {
	if bin, ok := e.(*ast.Bin); ok && bin.Op == "AND" {
		return append(splitAnd(bin.L), splitAnd(bin.R)...)
	}
	return []ast.Expr{e}
}

func hopRange(h *ast.HopRange) HopRange {
	r := HopRange{Min: 1, Max: -1}
	if h.Min != nil {
		r.Min = *h.Min
	}
	if h.Max != nil {
		r.Max = *h.Max
	}
	return r
}

// components は無向の連結成分を、最初に現れた頂点の順に返す。
func components(vorder []string, edges []*patEdge) [][]string {
	parent := map[string]string{}
	var find func(string) string
	find = func(x string) string {
		if parent[x] == x {
			return x
		}
		parent[x] = find(parent[x])
		return parent[x]
	}
	for _, v := range vorder {
		parent[v] = v
	}
	for _, e := range edges {
		a, c := find(e.left), find(e.right)
		if a != c {
			parent[c] = a
		}
	}
	idx := map[string]int{}
	var comps [][]string
	for _, v := range vorder {
		r := find(v)
		i, ok := idx[r]
		if !ok {
			i = len(comps)
			idx[r] = i
			comps = append(comps, nil)
		}
		comps[i] = append(comps[i], v)
	}
	return comps
}

// chooseStart: ① 束縛済みの頂点 → ② その頂点だけで評価できる述語が最多 → ③ 出現順。
func (b *builder) chooseStart(comp []string, vertices map[string]*patVertex, bound map[string]bool, pending []ast.Expr) string {
	for _, v := range comp {
		if bound[v] {
			return v
		}
	}
	// 統計を持たないので、述語が絞れる強さで順位を付ける（強い順に比べ、同じなら次の項目、最後は出現順）:
	//   一意キーの等値（{id: $x}）> 等値・IN > その他の述語の数
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
		for _, e := range pending {
			d := b.deps(e)
			if len(d) != 1 || d[0] != v {
				continue
			}
			s.n++
			bin, ok := e.(*ast.Bin)
			if !ok || (bin.Op != "=" && bin.Op != "IN") {
				continue
			}
			s.eq++
			if pr, ok := bin.L.(*ast.PropRef); ok && bin.Op == "=" && isUniqueKey(pr.Prop) {
				if _, lit := bin.R.(*ast.Lit); lit {
					s.unique++
				}
			}
		}
		if better(s, bestS) {
			best, bestS = v, s
		}
	}
	return best
}

func isUniqueKey(prop string) bool {
	for _, k := range settings.UniqueKeyProps {
		if k == prop {
			return true
		}
	}
	return false
}

// nextEdge は allowed の未処理の辺のうち、片端以上が束縛済みで出現順が最初のもの。
func nextEdge(edges []*patEdge, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool) *patEdge {
	for _, e := range edges {
		if done[e.order] || !allowed(e) {
			continue
		}
		if bound[e.left] || bound[e.right] {
			return e
		}
	}
	return nil
}

// orient は束縛側を From にした向きを返す。両端が束縛済みならパターンどおり。
func orient(e *patEdge, bound map[string]bool) (from, to string, dir Direction) {
	d := Out
	switch e.rel.Dir {
	case ast.DirLeft:
		d = In
	case ast.DirBoth:
		d = Both
	}
	if bound[e.left] {
		return e.left, e.right, d
	}
	return e.right, e.left, d.Reverse()
}

// uniqueCandidates は関係一意性の対象（タイプが他と重なり得る関係・区間）。
func uniqueCandidates(edges []*patEdge) (rels, segs []string) {
	overlaps := func(a, c []string) bool {
		if len(a) == 0 || len(c) == 0 {
			return true
		}
		for _, x := range a {
			if contains(c, x) {
				return true
			}
		}
		return false
	}
	for i, e := range edges {
		hit := false
		for j, o := range edges {
			if i != j && overlaps(e.rel.Types, o.rel.Types) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		if e.rel.Range != nil {
			segs = append(segs, e.rel.Var)
		} else {
			rels = append(rels, e.rel.Var)
		}
	}
	return rels, segs
}

// placeReady は依存変数がすべて束縛済みになった述語を、必要な列を足した上で Filter として置く。
func (b *builder) placeReady(n Node, bound map[string]bool, pending []ast.Expr) (Node, []ast.Expr, error) {
	var rest []ast.Expr
	for _, e := range pending {
		ready := true
		for _, d := range b.deps(e) {
			if !bound[d] {
				ready = false
				break
			}
		}
		if !ready {
			rest = append(rest, e)
			continue
		}
		var nd needs
		b.collectNeeds(e, convCtx{}, &nd)
		n = b.supply(n, nd)
		pred, err := b.conv(e, convCtx{})
		if err != nil {
			return nil, nil, err
		}
		n = &Filter{Input: n, Pred: pred}
	}
	return n, rest, nil
}

// ================================
// パス全体の条件の分解（§4.4）
// ================================

func (b *builder) decomposePathPreds(pending []ast.Expr, segRange map[string]HopRange, segPreds map[string][]SegPred) ([]ast.Expr, error) {
	var out []ast.Expr
	for _, e := range pending {
		repl, handled, err := b.decomposeOne(e, segRange, segPreds)
		if err != nil {
			return nil, err
		}
		if handled {
			out = append(out, repl...)
		} else {
			out = append(out, e)
		}
	}
	return out, nil
}

func (b *builder) decomposeOne(e ast.Expr, segRange map[string]HopRange, segPreds map[string][]SegPred) ([]ast.Expr, bool, error) {
	// length(p) θ k
	if bin, ok := e.(*ast.Bin); ok {
		if c, ok := bin.L.(*ast.Call); ok && c.Fn == "length" {
			p, perr := b.pathArg(c)
			lit, lok := bin.R.(*ast.Lit)
			if perr != nil || !lok {
				return nil, false, nil
			}
			k, kok := lit.Val.(int64)
			if !kok {
				return nil, false, nil
			}
			pd := b.paths[p]
			fixed := len(pd.rels)
			switch len(pd.segs) {
			case 0: // 長さは定数
				if cmpInt(int64(fixed), bin.Op, k) {
					return nil, true, nil
				}
				return []ast.Expr{&ast.Lit{Val: false}}, true, nil
			case 1:
				seg := pd.segs[0]
				r, ok := segRange[seg].Constrain(bin.Op, int(k)-fixed)
				if !ok {
					return nil, false, nil
				}
				segRange[seg] = r
				return nil, true, nil
			}
			return nil, false, nil
		}
	}
	// all / none(x IN nodes(p)|relationships(p) WHERE c(x))
	q, ok := e.(*ast.Quant)
	if !ok || (q.Kind != "all" && q.Kind != "none") {
		return nil, false, nil
	}
	src, err := b.quantSource(q)
	if err != nil {
		return nil, false, err
	}
	if !onlyRefersTo(q.Pred, q.Var) {
		return nil, false, nil
	}
	pred := q.Pred
	if q.Kind == "none" {
		pred = &ast.Un{Op: "NOT", X: pred}
	}
	pd := b.paths[src.path]
	var repl []ast.Expr
	elems := pd.nodes
	if src.part == SegRels {
		elems = pd.rels
	}
	for _, v := range elems {
		repl = append(repl, substituteVar(pred, q.Var, v))
	}
	for _, seg := range pd.segs {
		lp, err := b.conv(pred, convCtx{quant: map[string]quantSrc{q.Var: src}})
		if err != nil {
			return nil, false, err
		}
		segPreds[seg] = append(segPreds[seg], SegPred{Part: src.part, Var: q.Var, Pred: lp})
	}
	return repl, true, nil
}

func cmpInt(a int64, op string, k int64) bool {
	switch op {
	case "=":
		return a == k
	case "<>":
		return a != k
	case "<":
		return a < k
	case "<=":
		return a <= k
	case ">":
		return a > k
	case ">=":
		return a >= k
	}
	return false
}

// onlyRefersTo は式が変数 v（とリテラル）以外を参照しないか。量化子の入れ子は不可。
func onlyRefersTo(e ast.Expr, v string) bool {
	ok := true
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.VarRef:
			ok = false
		case *ast.PropRef:
			if x.Var != v {
				ok = false
			}
		case *ast.Bin:
			walk(x.L)
			walk(x.R)
		case *ast.Un:
			walk(x.X)
		case *ast.Call:
			for _, a := range x.Args {
				walk(a)
			}
		case *ast.List:
			for _, el := range x.Elems {
				walk(el)
			}
		case *ast.Case:
			for _, w := range x.Whens {
				walk(w.When)
				walk(w.Then)
			}
			if x.Else != nil {
				walk(x.Else)
			}
		case *ast.Quant, *ast.Agg:
			ok = false
		}
	}
	walk(e)
	return ok
}

// substituteVar は式中の from.prop を to.prop に置き換えた複製を返す。
func substituteVar(e ast.Expr, from, to string) ast.Expr {
	switch x := e.(type) {
	case *ast.PropRef:
		if x.Var == from {
			return &ast.PropRef{Var: to, Prop: x.Prop}
		}
		return x
	case *ast.Bin:
		return &ast.Bin{Op: x.Op, L: substituteVar(x.L, from, to), R: substituteVar(x.R, from, to)}
	case *ast.Un:
		return &ast.Un{Op: x.Op, X: substituteVar(x.X, from, to)}
	case *ast.Call:
		c := &ast.Call{Fn: x.Fn}
		for _, a := range x.Args {
			c.Args = append(c.Args, substituteVar(a, from, to))
		}
		return c
	case *ast.List:
		l := &ast.List{}
		for _, el := range x.Elems {
			l.Elems = append(l.Elems, substituteVar(el, from, to))
		}
		return l
	case *ast.Case:
		c := &ast.Case{}
		for _, w := range x.Whens {
			c.Whens = append(c.Whens, ast.CaseWhen{When: substituteVar(w.When, from, to), Then: substituteVar(w.Then, from, to)})
		}
		if x.Else != nil {
			c.Else = substituteVar(x.Else, from, to)
		}
		return c
	}
	return e
}

// ================================
// WITH / RETURN（§4.5）
// ================================

func (b *builder) project(c *ast.ProjectionClause) error {
	if b.plan == nil {
		return unsupported("projection without MATCH")
	}

	hasAgg := false
	for _, it := range c.Items {
		if _, ok := it.Expr.(*ast.Agg); ok {
			hasAgg = true
		}
	}
	// 集約・DISTINCT で行がまとまる場合、ORDER BY は射影後の列か、射影された束縛のプロパティだけを参照できる。
	collapsing := hasAgg || c.Distinct
	projIdent := map[string]varInfo{} // 自分の名前のまま射影される束縛（ORDER BY がそのプロパティを使える）
	for _, it := range c.Items {
		if v, ok := it.Expr.(*ast.VarRef); ok && v.Name == it.Name {
			if vi := b.scope[v.Name]; vi.kind == VarNode || vi.kind == VarRel {
				projIdent[it.Name] = vi
			}
		}
	}

	// 1. 必要な列を足す（行がまとまる場合の ORDER BY 用プロパティは、まとめた後で取る）
	var nd needs
	for _, it := range c.Items {
		b.collectNeeds(it.Expr, convCtx{}, &nd)
	}
	if !collapsing {
		for _, o := range c.OrderBy {
			b.collectNeeds(o.Expr, convCtx{}, &nd)
		}
	}
	b.plan = b.supply(b.plan, nd)

	// 2. 項目の変換
	itemNames := map[string]string{} // 式の文字列 → 項目名（ORDER BY の解決用）
	outNames := map[string]bool{}
	for _, it := range c.Items {
		if outNames[it.Name] {
			return unsupported("duplicate column name %s", it.Name)
		}
		outNames[it.Name] = true
		itemNames[it.Expr.String()] = it.Name
	}

	var sortKeys []SortKey
	var hidden []ProjItem
	var postProps []propNeed
	resolveOrder := func(o ast.SortItem) (Expr, error) {
		if name, ok := itemNames[o.Expr.String()]; ok {
			return &Column{Key: name}, nil
		}
		if v, ok := o.Expr.(*ast.VarRef); ok && outNames[v.Name] {
			return &Column{Key: v.Name}, nil
		}
		if onlyOutputs(o.Expr, outNames) { // 射影後の列だけの式（toInteger(personId) など）
			saved := b.scope
			b.scope = map[string]varInfo{}
			for n := range outNames {
				b.scope[n] = varInfo{kind: VarValue}
			}
			e, err := b.conv(o.Expr, convCtx{})
			b.scope = saved
			return e, err
		}
		if collapsing {
			for _, d := range b.deps(o.Expr) {
				if _, ok := projIdent[d]; !ok {
					return nil, unsupported("ORDER BY %s not in projection", o.Expr)
				}
			}
			var ond needs
			b.collectNeeds(o.Expr, convCtx{}, &ond)
			if len(ond.paths) > 0 || len(ond.listProps) > 0 {
				return nil, unsupported("ORDER BY %s not in projection", o.Expr)
			}
			postProps = append(postProps, ond.props...)
			return b.conv(o.Expr, convCtx{})
		}
		e, err := b.conv(o.Expr, convCtx{})
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("_sort%d", len(hidden))
		hidden = append(hidden, ProjItem{Name: name, Expr: e})
		return &Column{Key: name}, nil
	}

	trim := func() []ProjItem {
		items := make([]ProjItem, 0, len(c.Items))
		for _, it := range c.Items {
			items = append(items, ProjItem{Name: it.Name, Expr: &Column{Key: it.Name}})
		}
		return items
	}

	var finalItems []ProjItem
	if hasAgg {
		agg := &Aggregation{Input: b.plan}
		for _, it := range c.Items {
			if a, ok := it.Expr.(*ast.Agg); ok {
				ai := AggItem{Name: it.Name, Fn: a.Fn, Distinct: a.Distinct}
				if a.Arg != nil {
					v, err := b.conv(a.Arg, convCtx{})
					if err != nil {
						return err
					}
					ai.Arg = v
				}
				agg.Aggs = append(agg.Aggs, ai)
				continue
			}
			v, err := b.conv(it.Expr, convCtx{})
			if err != nil {
				return err
			}
			agg.Keys = append(agg.Keys, ProjItem{Name: it.Name, Expr: v})
		}
		b.plan = agg
		// 集約の出力は keys → aggs の順。項目順と違えば最後に並べ替える。
		natural := make([]string, 0, len(c.Items))
		for _, k := range agg.Keys {
			natural = append(natural, k.Name)
		}
		for _, a := range agg.Aggs {
			natural = append(natural, a.Name)
		}
		for i, it := range c.Items {
			if natural[i] != it.Name {
				finalItems = trim()
				break
			}
		}
	} else {
		proj := &Projection{Input: b.plan}
		for _, it := range c.Items {
			v, err := b.conv(it.Expr, convCtx{})
			if err != nil {
				return err
			}
			proj.Items = append(proj.Items, ProjItem{Name: it.Name, Expr: v})
		}
		b.plan = proj
	}

	for _, o := range c.OrderBy {
		e, err := resolveOrder(o)
		if err != nil {
			return err
		}
		sortKeys = append(sortKeys, SortKey{Expr: e, Desc: o.Desc})
	}
	if len(hidden) > 0 {
		proj := b.plan.(*Projection)
		proj.Items = append(proj.Items, hidden...)
		finalItems = trim()
	}

	// 3. 後処理
	if c.Distinct {
		b.plan = &Distinct{Input: b.plan}
	}
	seenPost := map[string]bool{}
	for _, p := range postProps {
		key := p.v + "." + p.prop
		if seenPost[key] {
			continue
		}
		seenPost[key] = true
		vi := projIdent[p.v]
		kind := NodeObj
		if vi.kind == VarRel {
			kind = RelObj
		}
		b.plan = &GetProperty{Input: b.plan, Alias: p.v, Prop: p.prop, Kind: kind, Labels: vi.labels}
		finalItems = trim()
	}
	if len(sortKeys) > 0 {
		b.plan = &Sort{Input: b.plan, Keys: sortKeys}
	}
	if c.Skip != nil || c.Limit != nil {
		l := &Limit{Input: b.plan, Count: -1}
		if c.Skip != nil {
			l.Skip = *c.Skip
		}
		if c.Limit != nil {
			l.Count = *c.Limit
		}
		b.plan = l
	}
	if len(finalItems) > 0 {
		b.plan = &Projection{Input: b.plan, Items: finalItems}
	}

	// 4. スコープの付け替え（WITH の本質）
	newScope := map[string]varInfo{}
	newBound := map[string]bool{}
	newAvail := map[string]bool{}
	newPaths := map[string]pathDef{}
	for _, it := range c.Items {
		info := varInfo{kind: VarValue}
		switch x := it.Expr.(type) {
		case *ast.VarRef:
			vi := b.scope[x.Name]
			if vi.kind == VarNode || vi.kind == VarRel {
				newScope[it.Name] = vi
				newBound[it.Name] = true
				continue
			}
			if vi.kind == VarPath {
				return unsupported("carrying path %s across WITH", x.Name)
			}
			info = vi // 値列（束縛のリストなら要素の情報も）をそのまま引き継ぐ
		case *ast.Agg:
			if v, ok := x.Arg.(*ast.VarRef); ok && x.Fn == "collect" {
				if vi := b.scope[v.Name]; vi.kind == VarNode || vi.kind == VarRel {
					el := vi
					info.elem = &el
				}
			}
		}
		newScope[it.Name] = info
		newBound[it.Name] = true
		newAvail[it.Name] = true
	}
	b.scope, b.bound, b.avail, b.paths = newScope, newBound, newAvail, newPaths

	// 5. WITH の後ろの WHERE
	if c.Where != nil {
		var err error
		var rest []ast.Expr
		b.plan, rest, err = b.placeReady(b.plan, b.bound, splitAnd(c.Where))
		if err != nil {
			return err
		}
		if len(rest) > 0 {
			return unsupported("unplaceable WITH predicate %s", rest[0])
		}
	}
	return nil
}
