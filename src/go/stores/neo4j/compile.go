// Package neo4j is the Neo4j database: Cypher compilation, execution and capabilities.
package neo4j

import (
	"fmt"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/internal/lower"
	"regexp"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
)

type cyKind int

const (
	cyIdent cyKind = iota // node or relationship variable
	cyValue
	cySeg  // variable-length segment (expr is the path variable)
	cyPath // path column (nodes/rels expressions in pathExprs)
	cyUUID // stored value of a shipped id column (not a node variable; returned as the uuid)
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
	visible []string // columns of the current row (for WITH DISTINCT and WITH *)
	paths   map[string][2]string
	params  map[string]any
	seq     int
	fresh   int
	own     map[string]lower.Owner
	types   stores.TypeFn
	relVars map[string]bool

	// optional is the clause being written for the inside of an OPTIONAL MATCH (MATCH parts and WHERE go here).
	inOptional bool
	optional   *cyClause

	// Shipped boundary (ir.ShipAggregate): ship holds the table's columns, shipOptional makes the key match an
	// OPTIONAL MATCH (a value fetch is a left outer join, so rows without the node stay), weight is the expression the first aggregation counts.
	ship         *lower.ShipTable
	shipOptional bool
	weight       string
	// unwindIn receives even a single input column with UNWIND. Neo4j 5.26 drops a later DISTINCT when a
	// variable-length expansion starts from `p.uuid IN $in` on a unique index (it prunes duplicates per start only).
	unwindIn bool
}

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func lowerCypher(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	b := &cyBuilder{cols: map[string]cyCol{}, paths: map[string][2]string{}, params: map[string]any{},
		own: lower.Owners(q.Ops), types: types, relVars: map[string]bool{}, unwindIn: q.Replace}
	for a, o := range b.own {
		if o.Kind == ir.RelObj {
			b.relVars[a] = true
		}
	}
	if q.ShipMode != ir.ShipNone {
		if q.ShipMode != ir.ShipAggregate {
			return nil, lower.Unsupported("shipped %s query on graph", q.ShipMode)
		}
		arg := lower.ShipLeaf(q.Ops)
		if arg == nil {
			return nil, lower.Unsupported("shipped graph query without an argument")
		}
		tab, err := lower.NewShipTable(q, arg)
		if err != nil {
			return nil, err
		}
		b.ship, b.shipOptional = tab, !lower.HasTraversal(q.Ops)
	}
	if err := b.node(q.Ops); err != nil {
		return nil, err
	}

	// RETURN (a preceding WITH * that only sorts or limits folds into it)
	var mods string
	if n := len(b.clauses); n > 0 {
		last := b.clauses[n-1]
		if last.head == "WITH *" && len(last.where) == 0 {
			mods = last.mods
			b.clauses = b.clauses[:n-1]
		}
	}
	rets := make([]string, len(q.Outputs))
	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		e, err := b.output(col)
		if err != nil {
			return nil, err
		}
		rets[i] = e + " AS " + lower.ColName(i)
		switch b.cols[col].kind {
		case cyIdent, cyUUID:
			kinds[i] = stores.KindIdent
		case cySeg, cyPath:
			kinds[i] = stores.KindList
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
		// Collapse equal output rows and return their count as the weight (Cypher groups by the non-aggregate RETURN items).
		if mods != "" {
			return nil, lower.Unsupported("weighted query with ORDER BY / LIMIT")
		}
		rets = append(rets, "count(*) AS "+lower.ColName(len(rets)))
		kinds = append(kinds, stores.KindValue)
		columns = append(append([]string{}, q.Outputs...), ir.WeightCol)
	}
	ret := "RETURN " + strings.Join(rets, ", ")
	if mods != "" {
		ret += " " + mods
	}
	parts = append(parts, ret)

	out := &stores.Query{Store: kind.Neo4j, Columns: columns, ColKinds: kinds, InKeys: q.Keys}
	body := &stores.TextBody{Text: strings.Join(parts, " "), InArg: -1}
	out.Body = body
	if b.ship != nil {
		out.ShipCols, out.ShipNames, out.ShipMode = b.ship.Cols, b.ship.ColNames(), q.ShipMode
	}
	if len(b.params) > 0 {
		body.Params = b.params
	}
	return out, nil
}

func (b *cyBuilder) add(head string) *cyClause {
	if b.inOptional {
		// The MATCH parts inside an OPTIONAL MATCH go into one clause (separate clauses would allow partial matches).
		pat, ok := strings.CutPrefix(head, "MATCH ")
		if !ok {
			b.optional = nil // non-pattern clauses (WITH, ...) are not absorbed here
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
	b.visible = lower.AppendNew(b.visible, name)
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

// nodePat returns (v:L); with several labels, (v) plus an OR in WHERE.
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

func (b *cyBuilder) node(n ir.Node) error {
	switch x := n.(type) {
	case *ir.Boundary:
		if b.inOptional {
			return nil // a leaf inside OPTIONAL: use the variable bound outside
		}
		return b.argument(x)
	case *ir.NodeHashJoin:
		if err := b.node(x.Left); err != nil {
			return err
		}
		return b.node(x.Right)
	case *ir.Optional:
		if err := b.node(x.Outer); err != nil {
			return err
		}
		if b.inOptional {
			return lower.Unsupported("nested OPTIONAL MATCH in one query")
		}
		b.inOptional, b.optional = true, nil
		err := b.node(x.Inner)
		if err == nil && b.optional == nil {
			err = lower.Unsupported("OPTIONAL MATCH without a pattern")
		}
		b.inOptional, b.optional = false, nil
		return err
	}
	if c := ir.Child(n); c != nil {
		if err := b.node(c); err != nil {
			return err
		}
	}

	switch x := n.(type) {
	case *ir.EntityScan:
		pat, cond := b.nodePat(x.Alias, x.Labels)
		b.add("MATCH " + pat)
		if cond != "" {
			b.where(cond)
		}
		b.bindIdent(x.Alias)

	case *ir.GetProperty:
		v, ok := b.cols[x.Alias]
		if !ok || v.kind != cyIdent {
			return lower.Unsupported("property of unbound %s", x.Alias)
		}
		b.cols[x.Key()] = cyCol{expr: v.expr + "." + x.Prop, kind: cyValue}
		b.visible = lower.AppendNew(b.visible, x.Key())

	case *ir.GetListProperty:
		list, err := b.listExpr(x.Source, x.Part)
		if err != nil {
			return err
		}
		b.cols[x.Key()] = cyCol{expr: fmt.Sprintf("[_x IN %s | _x.%s]", list, x.Prop), kind: cyValue}
		b.visible = lower.AppendNew(b.visible, x.Key())

	case *ir.Filter:
		e, err := b.expr(x.Pred)
		if err != nil {
			return err
		}
		b.where(e)

	case *ir.Expand:
		pat, cond := b.nodePat(x.To, x.ToLabels)
		b.add("MATCH (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody(x.Rel, x.Types, "")) + pat)
		if cond != "" {
			b.where(cond)
		}
		b.relVars[x.Rel] = true
		b.bindIdent(x.Rel)
		b.bindIdent(x.To)

	case *ir.ExpandInto:
		b.add("MATCH (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody(x.Rel, x.Types, "")) + "(" + b.cols[x.To].expr + ")")
		b.relVars[x.Rel] = true
		b.bindIdent(x.Rel)

	case *ir.VarLengthExpand:
		sp := "_sp_" + sanitize(x.Seg)
		pat, cond := b.nodePat(x.To, x.ToLabels)
		b.add("MATCH " + sp + " = (" + b.cols[x.From].expr + ")" + x.Dir.Pattern(relBody("", x.Types, x.Range.Cypher())) + pat)
		if cond != "" {
			b.where(cond)
		}
		b.bindIdent(x.To)
		seg := ir.SegColumn(x.Seg)
		b.cols[seg] = cyCol{expr: sp, kind: cySeg}
		b.visible = lower.AppendNew(b.visible, seg)
		for _, p := range x.SegPreds {
			list := "nodes(" + sp + ")[1..-1]"
			if p.Part == ir.SegRels {
				list = "relationships(" + sp + ")"
			}
			e, err := b.expr(p.Pred)
			if err != nil {
				return err
			}
			b.where(fmt.Sprintf("all(%s IN %s WHERE %s)", p.Var, list, e))
		}
		if len(x.InnerLabels) > 0 {
			b.where(fmt.Sprintf("all(_l IN nodes(%s)[1..-1] WHERE _l:%s)", sp, strings.Join(x.InnerLabels, ":")))
		}

	case *ir.ProjectPath:
		var nodes, rels []string
		for _, el := range x.Elems {
			c, ok := b.cols[elemKey(el)]
			if !ok {
				return lower.Unsupported("path element %s not bound", el.Var)
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
		b.visible = lower.AppendNew(b.visible, x.Path)

	case *ir.RelUniqueFilter:
		var conds []string
		for i := 0; i < len(x.Rels); i++ {
			for j := i + 1; j < len(x.Rels); j++ {
				conds = append(conds, b.cols[x.Rels[i]].expr+" <> "+b.cols[x.Rels[j]].expr)
			}
		}
		for _, s := range x.Segs {
			sp := b.cols[ir.SegColumn(s)].expr
			for _, r := range x.Rels {
				conds = append(conds, "NOT "+b.cols[r].expr+" IN relationships("+sp+")")
			}
		}
		for i := 0; i < len(x.Segs); i++ {
			for j := i + 1; j < len(x.Segs); j++ {
				a, c := b.cols[ir.SegColumn(x.Segs[i])].expr, b.cols[ir.SegColumn(x.Segs[j])].expr
				conds = append(conds, fmt.Sprintf("none(_x IN relationships(%s) WHERE _x IN relationships(%s))", a, c))
			}
		}
		for _, c := range conds {
			b.where(c)
		}

	case *ir.Projection:
		return b.with(x.Items, nil, false)

	case *ir.Aggregation:
		return b.with(x.Keys, x.Aggs, false)

	case *ir.Distinct:
		items := make([]ir.ProjItem, len(b.visible))
		for i, v := range b.visible {
			items[i] = ir.ProjItem{Name: v, Expr: &ir.Column{Key: v}}
		}
		return b.with(items, nil, true)

	case *ir.Sort:
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

	case *ir.Limit:
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

	case *ir.Boundary:
	default:
		return lower.Unsupported("graph operator %T", n)
	}
	return nil
}

func elemKey(el ir.PathElem) string {
	if el.Seg {
		return ir.SegColumn(el.Var)
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

// argument receives the boundary id columns: IN for one column, UNWIND rows for more. The uuid indexes are per
// label (per type for relationships), so the planner's label candidates are always added (no candidates: match without a label).
func (b *cyBuilder) argument(a *ir.Boundary) error {
	head := "MATCH "
	if b.shipOptional {
		head = "OPTIONAL MATCH "
	}
	match := func(k, props string) {
		ls := a.Labels[k]
		if b.relVars[k] {
			b.add(head + "()-[" + relBody(k, ls, "") + props + "]->()")
			return
		}
		pat, cond := b.nodePat(k, ls)
		b.add(head + strings.TrimSuffix(pat, ")") + props + ")")
		if cond != "" {
			b.where(cond)
		}
	}
	if b.ship != nil {
		// UNWIND the table rows: keys match on uuid, shipped columns and the weight are used as they are.
		b.add("UNWIND $" + stores.InParam + " AS _k")
		for _, k := range a.Cols {
			match(k, " {uuid: _k."+b.ship.Names[k]+"}")
			b.bindIdent(k)
		}
		for _, s := range a.Ship {
			b.cols[s] = cyCol{expr: "_k." + b.ship.Names[s], kind: cyUUID}
			b.visible = lower.AppendNew(b.visible, s)
		}
		b.weight = "_k." + b.ship.Names[ir.WeightCol]
		return nil
	}
	if len(a.Cols) == 1 && b.unwindIn {
		k := a.Cols[0]
		b.add("UNWIND $" + stores.InParam + " AS _k")
		match(k, " {uuid: _k}")
		b.bindIdent(k)
		return nil
	}
	if len(a.Cols) == 1 {
		k := a.Cols[0]
		match(k, "")
		b.where(k + ".uuid IN $" + stores.InParam)
		b.bindIdent(k)
		return nil
	}
	b.add("UNWIND $" + stores.InParam + " AS _k")
	for _, k := range a.Cols {
		match(k, " {uuid: _k."+k+"}")
		b.bindIdent(k)
	}
	return nil
}

// with turns Projection / Aggregation / Distinct into a WITH clause. Value columns get new variables vN;
// id columns are rebound to their name when it is an identifier.
func (b *cyBuilder) with(items []ir.ProjItem, aggs []ir.AggItem, distinct bool) error {
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
			return lower.Unsupported("carrying %s through WITH", it.Name)
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
		if b.weight != "" {
			// Count with the shipped rows' weight (count(*) = sum w, count(x) = sum of w where x is not null, sum(x) = sum w*x).
			e, err := weightedCypher(a, arg, b.weight)
			if err != nil {
				return err
			}
			parts = append(parts, e+" AS "+target)
			newCols[a.Name] = cyCol{expr: target, kind: cyValue}
			visible = append(visible, a.Name)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s(%s) AS %s", a.Fn, arg, target))
		newCols[a.Name] = cyCol{expr: target, kind: cyValue}
		visible = append(visible, a.Name)
	}
	head := "WITH "
	if distinct {
		head += "DISTINCT "
	}
	b.add(head + strings.Join(parts, ", "))
	b.weight = "" // the weight (an UNWIND row) does not survive the first WITH
	b.cols, b.visible, b.paths = newCols, visible, map[string][2]string{}
	return nil
}

// itemExpr returns a WITH item's expression (id columns pass as variables).
func (b *cyBuilder) itemExpr(e ir.Expr) (string, cyKind, error) {
	if c, ok := e.(*ir.Column); ok {
		col, ok := b.cols[c.Key]
		if !ok {
			return "", 0, lower.Unsupported("unknown column %s", c.Key)
		}
		return col.expr, col.kind, nil
	}
	s, err := b.expr(e)
	return s, cyValue, err
}

// output returns one RETURN column: the uuid for id columns, {nodes, rels} uuid lists for segments and paths.
func (b *cyBuilder) output(key string) (string, error) {
	c, ok := b.cols[key]
	if !ok {
		return "", lower.Unsupported("output column %s not produced", key)
	}
	switch c.kind {
	case cyIdent:
		return c.expr + ".uuid", nil
	case cyUUID:
		return c.expr, nil
	case cySeg:
		return fmt.Sprintf("{nodes: [_x IN nodes(%s)[1..-1] | _x.uuid], rels: [_x IN relationships(%s) | _x.uuid]}", c.expr, c.expr), nil
	case cyPath:
		p := b.paths[key]
		return fmt.Sprintf("{nodes: [_x IN %s | _x.uuid], rels: [_x IN %s | _x.uuid]}", p[0], p[1]), nil
	}
	return c.expr, nil
}

func (b *cyBuilder) listExpr(source string, part ir.SegPart) (string, error) {
	c, ok := b.cols[source]
	if !ok {
		return "", lower.Unsupported("list source %s not bound", source)
	}
	switch c.kind {
	case cySeg:
		if part == ir.SegRels {
			return "relationships(" + c.expr + ")", nil
		}
		return "nodes(" + c.expr + ")[1..-1]", nil
	case cyPath:
		p := b.paths[source]
		if part == ir.SegRels {
			return p[1], nil
		}
		return p[0], nil
	}
	return "", lower.Unsupported("list source %s is not a path", source)
}

// expr renders a logical expression as Cypher; literals become parameters typed like the column they are compared with.
func (b *cyBuilder) expr(e ir.Expr) (string, error) {
	return b.exprT(e, "")
}

func (b *cyBuilder) exprT(e ir.Expr, typ string) (string, error) {
	switch x := e.(type) {
	case *ir.Column:
		c, ok := b.cols[x.Key]
		if !ok {
			return "", lower.Unsupported("unknown column %s", x.Key)
		}
		if c.kind == cyPath || c.kind == cySeg {
			return "", lower.Unsupported("path column %s used as value", x.Key)
		}
		return c.expr, nil
	case *ir.ElemProp:
		return x.Var + "." + x.Prop, nil
	case *ir.InSet:
		v := x.Var
		if !x.Elem {
			c, ok := b.cols[x.Var]
			if !ok || c.kind != cyIdent {
				return "", lower.Unsupported("id set on unbound %s", x.Var)
			}
			v = c.expr
		}
		var ids []id.UUID
		if x.Bound != nil { // unbound when only rendering or checking the plan
			ids = x.Bound.IDs
		}
		return "(" + v + "." + id.PropName + " IN " + b.param(id.StoredSlice(ids)) + ")", nil
	case *ir.Lit:
		if b, ok := x.Val.(bool); ok {
			if b {
				return "true", nil
			}
			return "false", nil
		}
		if x.Val == nil {
			return "null", nil
		}
		return b.param(cypherValue(lower.Typed(x.Val, typ), typ)), nil
	case *ir.Bin:
		lt := lower.LitType(x.R, b.own, b.types)
		rt := lower.LitType(x.L, b.own, b.types)
		l, err := b.exprT(x.L, lt)
		if err != nil {
			return "", err
		}
		r, err := b.exprT(x.R, rt)
		if err != nil {
			return "", err
		}
		return "(" + l + " " + x.Op + " " + r + ")", nil
	case *ir.Un:
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
	case *ir.Call:
		if x.Fn == "length" || x.Fn == "nodes" || x.Fn == "relationships" {
			src, ok := x.Args[0].(*ir.Column)
			if !ok {
				return "", lower.Unsupported("%s argument", x.Fn)
			}
			part := ir.SegNodes
			if x.Fn != "nodes" {
				part = ir.SegRels
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
	case *ir.Case:
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
	case *ir.List:
		elems := make([]string, len(x.Elems))
		for i, el := range x.Elems {
			s, err := b.exprT(el, typ)
			if err != nil {
				return "", err
			}
			elems[i] = s
		}
		return "[" + strings.Join(elems, ", ") + "]", nil
	case *ir.Quant:
		call, ok := x.List.(*ir.Call)
		if !ok {
			return "", lower.Unsupported("quantifier list %s", x.List)
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
	return "", lower.Unsupported("cypher expression %T", e)
}

// cypherFn returns the Cypher spelling of a function name (the AST lowercases them).
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

// cypherValue types a parameter like its column. Dates go as Date (a time.Time becomes a DateTime, and
// comparing Date with DateTime is null in Cypher, so nothing would match).
func cypherValue(v any, typ string) any {
	if t, ok := v.(time.Time); ok && typ == "date" {
		return neo4j.DateOf(t)
	}
	return v
}

// lowerShipCypher compiles a shipped Neo4j query (UNWIND $in AS _k -> match keys on uuid -> traverse -> weighted aggregation).
func lowerShipCypher(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerCypher(q, types)
}

// weightedCypher is the Cypher form of weightedSQL (arg is the argument expression, w the weight).
func weightedCypher(a ir.AggItem, arg, w string) (string, error) {
	if a.Distinct {
		return "", lower.Unsupported("%s(DISTINCT) in a shipped aggregation", a.Fn)
	}
	if a.Arg == nil {
		if a.Fn != "count" {
			return "", lower.Unsupported("%s(*)", a.Fn)
		}
		return "sum(" + w + ")", nil
	}
	switch a.Fn {
	case "count":
		return "sum(CASE WHEN " + arg + " IS NULL THEN 0 ELSE " + w + " END)", nil
	case "sum":
		return "sum(" + w + " * " + arg + ")", nil
	case "min", "max":
		return a.Fn + "(" + arg + ")", nil
	}
	return "", lower.Unsupported("%s in a shipped aggregation", a.Fn)
}
