package native

import (
	"fmt"
	"regexp"
	"strings"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/store"
)

// cypher.go は graph（Neo4j）向けの変換。論理部分木を葉から順に MATCH / WITH の節へ落とす。
//   - EntityScan / Expand / ExpandInto / VarLengthExpand → MATCH 節（ラベルが複数なら WHERE の OR）
//   - GetProperty / GetListProperty → 節を作らず、列 → Cypher 式の対応だけを記録（RETURN か述語で使う）
//   - Filter → 直前の節の WHERE
//   - Projection / Aggregation / Distinct → WITH 節。Sort / Limit → WITH * の ORDER BY / SKIP / LIMIT
//   - RelUniqueFilter → WHERE の不等式（MATCH 節を分けるので Neo4j の一意性に頼らない）
//   - 可変長区間はパス変数 _sp_<seg> で持ち、中間ノードは nodes(_sp)[1..-1]
// 境界の識別列は Argument が `UNWIND $in` / `WHERE k.uuid IN $in` で受け取る。

type cyKind int

const (
	cyIdent cyKind = iota // ノード/関係の変数
	cyValue
	cySeg  // 可変長区間（expr はパス変数）
	cyPath // パス列（nodes/rels の式を pathExprs に持つ）
)

type cyCol struct {
	expr string
	kind cyKind
}

type cyClause struct {
	head  string
	where []string
	mods  string // ORDER BY / SKIP / LIMIT
}

type cyBuilder struct {
	clauses []*cyClause
	cols    map[string]cyCol
	visible []string // 現在の行の列（WITH DISTINCT・WITH * 用）
	paths   map[string][2]string
	params  map[string]any
	seq     int
	fresh   int
	own     map[string]owner
	types   TypeFn
	relVars map[string]bool

	// optional は OPTIONAL MATCH の内側を書いている間の節（MATCH 片をここへ , で連ね、WHERE もここへ付ける）。
	inOptional bool
	optional   *cyClause
}

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func lowerCypher(q *physical.StoreQuery, types TypeFn) (*Query, error) {
	b := &cyBuilder{cols: map[string]cyCol{}, paths: map[string][2]string{}, params: map[string]any{},
		own: owners(q.Ops), types: types, relVars: map[string]bool{}}
	for a, o := range b.own {
		if o.kind == logical.RelObj {
			b.relVars[a] = true
		}
	}
	if err := b.node(q.Ops); err != nil {
		return nil, err
	}

	// RETURN（直前が WITH * の並べ替え・件数制限だけなら RETURN に畳む）
	var mods string
	if n := len(b.clauses); n > 0 {
		last := b.clauses[n-1]
		if last.head == "WITH *" && len(last.where) == 0 {
			mods = last.mods
			b.clauses = b.clauses[:n-1]
		}
	}
	rets := make([]string, len(q.Outputs))
	kinds := make([]ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		e, err := b.output(col)
		if err != nil {
			return nil, err
		}
		rets[i] = e + " AS " + colName(i)
		switch b.cols[col].kind {
		case cyIdent:
			kinds[i] = KindIdent
		case cySeg, cyPath:
			kinds[i] = KindList
		}
	}
	var parts []string
	for _, c := range b.clauses {
		s := c.head
		if c.mods != "" {
			s += " " + c.mods
		}
		if len(c.where) > 0 {
			s += " WHERE " + strings.Join(c.where, " AND ")
		}
		parts = append(parts, s)
	}
	columns := q.Outputs
	if q.Weighted {
		// 同じ出力の組をまとめ、件数を重みとして返す（Cypher は集約以外の RETURN 項目でグループ化する）。
		if mods != "" {
			return nil, unsupported("weighted query with ORDER BY / LIMIT")
		}
		rets = append(rets, "count(*) AS "+colName(len(rets)))
		kinds = append(kinds, KindValue)
		columns = append(append([]string{}, q.Outputs...), physical.WeightCol)
	}
	ret := "RETURN " + strings.Join(rets, ", ")
	if mods != "" {
		ret += " " + mods
	}
	parts = append(parts, ret)

	out := &Query{Store: store.Graph, Columns: columns, ColKinds: kinds, InKeys: q.Keys, Text: strings.Join(parts, " "), InArg: -1}
	if len(b.params) > 0 {
		out.Params = b.params
	}
	return out, nil
}

func (b *cyBuilder) add(head string) *cyClause {
	if b.inOptional {
		// OPTIONAL MATCH の内側の MATCH 片は 1 つの節に連ねる（節を分けると部分一致を許してしまう）。
		pat, ok := strings.CutPrefix(head, "MATCH ")
		if !ok {
			b.optional = nil // パターン以外の節（WITH など）は吸収しない形なのでここには来ない
		} else if b.optional == nil {
			b.optional = &cyClause{head: "OPTIONAL MATCH " + pat}
			b.clauses = append(b.clauses, b.optional)
			return b.optional
		} else {
			b.optional.head += ", " + pat
			return b.optional
		}
	}
	c := &cyClause{head: head}
	b.clauses = append(b.clauses, c)
	return c
}

func (b *cyBuilder) where(cond string) {
	if b.inOptional && b.optional != nil {
		b.optional.where = append(b.optional.where, cond)
		return
	}
	if len(b.clauses) == 0 {
		b.add("WITH *")
	}
	last := b.clauses[len(b.clauses)-1]
	last.where = append(last.where, cond)
}

func (b *cyBuilder) bindIdent(name string) {
	b.cols[name] = cyCol{expr: name, kind: cyIdent}
	b.visible = appendNew(b.visible, name)
}

func appendNew(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	return append(xs, s)
}

func (b *cyBuilder) param(v any) string {
	name := fmt.Sprintf("p%d", b.seq)
	b.seq++
	b.params[name] = v
	return "$" + name
}

func (b *cyBuilder) freshVar() string {
	v := fmt.Sprintf("v%d", b.fresh)
	b.fresh++
	return v
}

func labelCond(v string, labels []string) string {
	ors := make([]string, len(labels))
	for i, l := range labels {
		ors[i] = v + ":" + l
	}
	return "(" + strings.Join(ors, " OR ") + ")"
}

// nodePat は (v:L)。ラベルが複数なら (v) にして WHERE で OR を付ける。
func (b *cyBuilder) nodePat(v string, labels []string) (string, string) {
	switch len(labels) {
	case 0:
		return "(" + v + ")", ""
	case 1:
		return "(" + v + ":" + labels[0] + ")", ""
	}
	return "(" + v + ")", labelCond(v, labels)
}

func relBody(v string, types []string, rng string) string {
	s := v
	if len(types) > 0 {
		s += ":" + strings.Join(types, "|")
	}
	return s + rng
}

func (b *cyBuilder) node(n logical.Node) error {
	switch x := n.(type) {
	case *physical.Argument:
		if b.inOptional {
			return nil // OPTIONAL の内側の葉: 外側で束縛済みの変数をそのまま使う
		}
		return b.argument(x.Cols)
	case *logical.NodeHashJoin:
		if err := b.node(x.Left); err != nil {
			return err
		}
		return b.node(x.Right)
	case *logical.Optional:
		if err := b.node(x.Outer); err != nil {
			return err
		}
		if b.inOptional {
			return unsupported("nested OPTIONAL MATCH in one query")
		}
		b.inOptional, b.optional = true, nil
		err := b.node(x.Inner)
		if err == nil && b.optional == nil {
			err = unsupported("OPTIONAL MATCH without a pattern")
		}
		b.inOptional, b.optional = false, nil
		return err
	}
	if c := logical.Child(n); c != nil {
		if err := b.node(c); err != nil {
			return err
		}
	}

	switch x := n.(type) {
	case *logical.EntityScan:
		pat, cond := b.nodePat(x.Alias, x.Labels)
		b.add("MATCH " + pat)
		if cond != "" {
			b.where(cond)
		}
		b.bindIdent(x.Alias)

	case *logical.GetProperty:
		v, ok := b.cols[x.Alias]
		if !ok || v.kind != cyIdent {
			return unsupported("property of unbound %s", x.Alias)
		}
		b.cols[x.Key()] = cyCol{expr: v.expr + "." + x.Prop, kind: cyValue}
		b.visible = appendNew(b.visible, x.Key())

	case *logical.GetListProperty:
		list, err := b.listExpr(x.Source, x.Part)
		if err != nil {
			return err
		}
		b.cols[x.Key()] = cyCol{expr: fmt.Sprintf("[_x IN %s | _x.%s]", list, x.Prop), kind: cyValue}
		b.visible = appendNew(b.visible, x.Key())

	case *logical.Filter:
		e, err := b.expr(x.Pred)
		if err != nil {
			return err
		}
		b.where(e)

	case *logical.Expand:
		pat, cond := b.nodePat(x.To, x.ToLabels)
		b.add("MATCH (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody(x.Rel, x.Types, "")) + pat)
		if cond != "" {
			b.where(cond)
		}
		b.relVars[x.Rel] = true
		b.bindIdent(x.Rel)
		b.bindIdent(x.To)

	case *logical.ExpandInto:
		b.add("MATCH (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody(x.Rel, x.Types, "")) + "(" + b.cols[x.To].expr + ")")
		b.relVars[x.Rel] = true
		b.bindIdent(x.Rel)

	case *logical.VarLengthExpand:
		sp := "_sp_" + sanitize(x.Seg)
		pat, cond := b.nodePat(x.To, x.ToLabels)
		b.add("MATCH " + sp + " = (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody("", x.Types, x.Range.Cypher())) + pat)
		if cond != "" {
			b.where(cond)
		}
		b.bindIdent(x.To)
		seg := logical.SegColumn(x.Seg)
		b.cols[seg] = cyCol{expr: sp, kind: cySeg}
		b.visible = appendNew(b.visible, seg)
		for _, p := range x.SegPreds {
			list := "nodes(" + sp + ")[1..-1]"
			if p.Part == logical.SegRels {
				list = "relationships(" + sp + ")"
			}
			e, err := b.expr(p.Pred)
			if err != nil {
				return err
			}
			b.where(fmt.Sprintf("all(%s IN %s WHERE %s)", p.Var, list, e))
		}

	case *logical.ProjectPath:
		var nodes, rels []string
		for _, el := range x.Elems {
			c, ok := b.cols[elemKey(el)]
			if !ok {
				return unsupported("path element %s not bound", el.Var)
			}
			switch {
			case el.Seg:
				nodes = append(nodes, "nodes("+c.expr+")[1..-1]")
				rels = append(rels, "relationships("+c.expr+")")
			case b.relVars[el.Var]:
				rels = append(rels, "["+c.expr+"]")
			default:
				nodes = append(nodes, "["+c.expr+"]")
			}
		}
		b.paths[x.Path] = [2]string{joinList(nodes), joinList(rels)}
		b.cols[x.Path] = cyCol{expr: x.Path, kind: cyPath}
		b.visible = appendNew(b.visible, x.Path)

	case *logical.RelUniqueFilter:
		var conds []string
		for i := 0; i < len(x.Rels); i++ {
			for j := i + 1; j < len(x.Rels); j++ {
				conds = append(conds, b.cols[x.Rels[i]].expr+" <> "+b.cols[x.Rels[j]].expr)
			}
		}
		for _, s := range x.Segs {
			sp := b.cols[logical.SegColumn(s)].expr
			for _, r := range x.Rels {
				conds = append(conds, "NOT "+b.cols[r].expr+" IN relationships("+sp+")")
			}
		}
		for i := 0; i < len(x.Segs); i++ {
			for j := i + 1; j < len(x.Segs); j++ {
				a, c := b.cols[logical.SegColumn(x.Segs[i])].expr, b.cols[logical.SegColumn(x.Segs[j])].expr
				conds = append(conds, fmt.Sprintf("none(_x IN relationships(%s) WHERE _x IN relationships(%s))", a, c))
			}
		}
		for _, c := range conds {
			b.where(c)
		}

	case *logical.Projection:
		return b.with(x.Items, nil, false)

	case *logical.Aggregation:
		return b.with(x.Keys, x.Aggs, false)

	case *logical.Distinct:
		items := make([]logical.ProjItem, len(b.visible))
		for i, v := range b.visible {
			items[i] = logical.ProjItem{Name: v, Expr: &logical.Column{Key: v}}
		}
		return b.with(items, nil, true)

	case *logical.Sort:
		keys := make([]string, len(x.Keys))
		for i, k := range x.Keys {
			e, err := b.expr(k.Expr)
			if err != nil {
				return err
			}
			if k.Desc {
				e += " DESC"
			}
			keys[i] = e
		}
		c := b.add("WITH *")
		c.mods = "ORDER BY " + strings.Join(keys, ", ")

	case *logical.Limit:
		var mods []string
		if x.Skip > 0 {
			mods = append(mods, fmt.Sprintf("SKIP %d", x.Skip))
		}
		if x.Count >= 0 {
			mods = append(mods, fmt.Sprintf("LIMIT %d", x.Count))
		}
		last := b.clauses[len(b.clauses)-1]
		if last.head == "WITH *" && len(last.where) == 0 && strings.HasPrefix(last.mods, "ORDER BY") && !strings.Contains(last.mods, "LIMIT") {
			last.mods += " " + strings.Join(mods, " ")
		} else {
			c := b.add("WITH *")
			c.mods = strings.Join(mods, " ")
		}

	case *physical.Argument:
	default:
		return unsupported("graph operator %T", n)
	}
	return nil
}

func elemKey(el logical.PathElem) string {
	if el.Seg {
		return logical.SegColumn(el.Var)
	}
	return el.Var
}

func joinList(xs []string) string {
	if len(xs) == 0 {
		return "[]"
	}
	return strings.Join(xs, " + ")
}

func sanitize(s string) string { return strings.TrimLeft(s, "_") }

// argument は境界の識別列を受け取る。1 列なら IN、2 列以上なら UNWIND した行で受け取る。
func (b *cyBuilder) argument(cols []string) error {
	pat := func(k, props string) string {
		if b.relVars[k] {
			return "MATCH ()-[" + k + props + "]->()"
		}
		return "MATCH (" + k + ":Entity" + props + ")"
	}
	if len(cols) == 1 {
		b.add(pat(cols[0], ""))
		b.where(cols[0] + ".uuid IN $" + InParam)
		b.bindIdent(cols[0])
		return nil
	}
	b.add("UNWIND $" + InParam + " AS _k")
	for _, k := range cols {
		b.add(pat(k, " {uuid: _k."+k+"}"))
		b.bindIdent(k)
	}
	return nil
}

// with は Projection / Aggregation / Distinct を WITH 節にする。値列は新しい変数 vN に、
// 識別列は（名前が識別子なら）その名前に束縛し直す。
func (b *cyBuilder) with(items []logical.ProjItem, aggs []logical.AggItem, distinct bool) error {
	newCols := map[string]cyCol{}
	var visible, parts []string
	for _, it := range items {
		e, kind, err := b.itemExpr(it.Expr)
		if err != nil {
			return err
		}
		target := b.freshVar()
		if kind == cyIdent && identRe.MatchString(it.Name) {
			target = it.Name
		}
		if kind == cySeg || kind == cyPath {
			return unsupported("carrying %s through WITH", it.Name)
		}
		parts = append(parts, e+" AS "+target)
		newCols[it.Name] = cyCol{expr: target, kind: kind}
		visible = append(visible, it.Name)
	}
	for _, a := range aggs {
		arg := "*"
		if a.Arg != nil {
			e, err := b.expr(a.Arg)
			if err != nil {
				return err
			}
			arg = e
			if a.Distinct {
				arg = "DISTINCT " + arg
			}
		}
		target := b.freshVar()
		parts = append(parts, fmt.Sprintf("%s(%s) AS %s", a.Fn, arg, target))
		newCols[a.Name] = cyCol{expr: target, kind: cyValue}
		visible = append(visible, a.Name)
	}
	head := "WITH "
	if distinct {
		head += "DISTINCT "
	}
	b.add(head + strings.Join(parts, ", "))
	b.cols, b.visible, b.paths = newCols, visible, map[string][2]string{}
	return nil
}

// itemExpr は WITH の項目の式（識別列はそのまま変数で渡す）。
func (b *cyBuilder) itemExpr(e logical.Expr) (string, cyKind, error) {
	if c, ok := e.(*logical.Column); ok {
		col, ok := b.cols[c.Key]
		if !ok {
			return "", 0, unsupported("unknown column %s", c.Key)
		}
		return col.expr, col.kind, nil
	}
	s, err := b.expr(e)
	return s, cyValue, err
}

// output は RETURN の 1 列。識別列は uuid、区間・パスは {nodes, rels} の uuid リスト。
func (b *cyBuilder) output(key string) (string, error) {
	c, ok := b.cols[key]
	if !ok {
		return "", unsupported("output column %s not produced", key)
	}
	switch c.kind {
	case cyIdent:
		return c.expr + ".uuid", nil
	case cySeg:
		return fmt.Sprintf("{nodes: [_x IN nodes(%s)[1..-1] | _x.uuid], rels: [_x IN relationships(%s) | _x.uuid]}", c.expr, c.expr), nil
	case cyPath:
		p := b.paths[key]
		return fmt.Sprintf("{nodes: [_x IN %s | _x.uuid], rels: [_x IN %s | _x.uuid]}", p[0], p[1]), nil
	}
	return c.expr, nil
}

func (b *cyBuilder) listExpr(source string, part logical.SegPart) (string, error) {
	c, ok := b.cols[source]
	if !ok {
		return "", unsupported("list source %s not bound", source)
	}
	switch c.kind {
	case cySeg:
		if part == logical.SegRels {
			return "relationships(" + c.expr + ")", nil
		}
		return "nodes(" + c.expr + ")[1..-1]", nil
	case cyPath:
		p := b.paths[source]
		if part == logical.SegRels {
			return p[1], nil
		}
		return p[0], nil
	}
	return "", unsupported("list source %s is not a path", source)
}

// expr は論理式を Cypher の式にする。リテラルは比較相手の列の型に合わせてパラメータ化する。
func (b *cyBuilder) expr(e logical.Expr) (string, error) {
	return b.exprT(e, "")
}

func (b *cyBuilder) exprT(e logical.Expr, typ string) (string, error) {
	switch x := e.(type) {
	case *logical.Column:
		c, ok := b.cols[x.Key]
		if !ok {
			return "", unsupported("unknown column %s", x.Key)
		}
		if c.kind == cyPath || c.kind == cySeg {
			return "", unsupported("path column %s used as value", x.Key)
		}
		return c.expr, nil
	case *logical.ElemProp:
		return x.Var + "." + x.Prop, nil
	case *logical.Lit:
		if b, ok := x.Val.(bool); ok {
			if b {
				return "true", nil
			}
			return "false", nil
		}
		if x.Val == nil {
			return "null", nil
		}
		return b.param(typed(x.Val, typ)), nil
	case *logical.Bin:
		lt := litType(x.R, b.own, b.types)
		rt := litType(x.L, b.own, b.types)
		l, err := b.exprT(x.L, lt)
		if err != nil {
			return "", err
		}
		r, err := b.exprT(x.R, rt)
		if err != nil {
			return "", err
		}
		return "(" + l + " " + x.Op + " " + r + ")", nil
	case *logical.Un:
		v, err := b.exprT(x.X, typ)
		if err != nil {
			return "", err
		}
		switch x.Op {
		case "IS NULL", "IS NOT NULL":
			return "(" + v + " " + x.Op + ")", nil
		case "NOT":
			return "(NOT " + v + ")", nil
		}
		return "(-" + v + ")", nil
	case *logical.Call:
		if x.Fn == "length" || x.Fn == "nodes" || x.Fn == "relationships" {
			src, ok := x.Args[0].(*logical.Column)
			if !ok {
				return "", unsupported("%s argument", x.Fn)
			}
			part := logical.SegNodes
			if x.Fn != "nodes" {
				part = logical.SegRels
			}
			list, err := b.listExpr(src.Key, part)
			if err != nil {
				return "", err
			}
			if x.Fn == "length" {
				return "size(" + list + ")", nil
			}
			return list, nil
		}
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			s, err := b.exprT(a, "")
			if err != nil {
				return "", err
			}
			args[i] = s
		}
		return cypherFn(x.Fn) + "(" + strings.Join(args, ", ") + ")", nil
	case *logical.Case:
		var sb strings.Builder
		sb.WriteString("CASE")
		for _, w := range x.Whens {
			wv, err := b.exprT(w.When, "")
			if err != nil {
				return "", err
			}
			tv, err := b.exprT(w.Then, "")
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&sb, " WHEN %s THEN %s", wv, tv)
		}
		if x.Else != nil {
			ev, err := b.exprT(x.Else, "")
			if err != nil {
				return "", err
			}
			sb.WriteString(" ELSE " + ev)
		}
		sb.WriteString(" END")
		return sb.String(), nil
	case *logical.List:
		elems := make([]string, len(x.Elems))
		for i, el := range x.Elems {
			s, err := b.exprT(el, typ)
			if err != nil {
				return "", err
			}
			elems[i] = s
		}
		return "[" + strings.Join(elems, ", ") + "]", nil
	case *logical.Quant:
		call, ok := x.List.(*logical.Call)
		if !ok {
			return "", unsupported("quantifier list %s", x.List)
		}
		list, err := b.exprT(call, "")
		if err != nil {
			return "", err
		}
		pred, err := b.exprT(x.Pred, "")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s(%s IN %s WHERE %s)", x.Kind, x.Var, list, pred), nil
	}
	return "", unsupported("cypher expression %T", e)
}

// cypherFn は関数名を Cypher の綴りにする（AST は小文字にそろえている）。
func cypherFn(fn string) string {
	switch fn {
	case "tointeger":
		return "toInteger"
	case "tofloat":
		return "toFloat"
	case "tostring":
		return "toString"
	case "tolower":
		return "toLower"
	case "toupper":
		return "toUpper"
	}
	return fn
}
