package native

import (
	"fmt"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/store"
)

// mongo.go は document（MongoDB）向けの変換。1 コレクション（＝1 ラベル）への集約パイプラインにする。
//   $match（uuid IN 入力 ＋ 条件）→ [$group] → $project（行の列を v0, v1… に）→ [$sort] → [$skip/$limit]
//   → $project（出力を c0, c1… に）
// 遅延取得の GetProperty（Limit の後ろ）も同じ文書の項目なので、最初の $project で一緒に運んでおく。

type mongoBuilder struct {
	coll    string
	alias   string
	match   Doc
	fields  map[string]string // 列キー → 文書上の参照（"$prop" / "$_id.k0" / "$v3" …）
	visible []string
	stages  []Stage
	fresh   int
	grouped bool
	own     map[string]owner
	types   TypeFn
	keyed   bool
	carry   map[string]string // 最初の $project で運ぶ素の項目（列キー → "$prop"）
}

func lowerMongo(q *physical.StoreQuery, types TypeFn) (*Query, error) {
	ops := chain(q.Ops)
	if ops == nil {
		return nil, unsupported("document query with a join")
	}
	b := &mongoBuilder{fields: map[string]string{}, own: owners(q.Ops), types: types, carry: map[string]string{}}
	// 先に、どこかで取る素の項目を集める（遅延取得の分も最初の $project で運ぶため）。
	for _, op := range ops {
		if g, ok := op.(*logical.GetProperty); ok {
			b.carry[g.Key()] = "$" + g.Prop
		}
	}
	for _, op := range ops {
		if err := b.op(op); err != nil {
			return nil, err
		}
	}
	if b.coll == "" {
		return nil, unsupported("document query without a collection")
	}

	out := Doc{{Key: "_id", Val: 0}}
	for i, col := range q.Outputs {
		f, ok := b.fields[col]
		if !ok {
			return nil, unsupported("output column %s not produced", col)
		}
		out = append(out, Field{Key: colName(i), Val: f})
	}
	pipeline := append([]Stage{{Op: "$match", Body: b.match}}, b.stages...)
	pipeline = append(pipeline, Stage{Op: "$project", Body: out})
	return &Query{Store: store.Document, Columns: q.Outputs, InKeys: q.Keys, Collection: b.coll, Pipeline: pipeline, InArg: -1}, nil
}

func (b *mongoBuilder) setColl(alias string, labels []string) error {
	if len(labels) != 1 {
		return unsupported("collection for %s needs exactly one label (%v)", alias, labels)
	}
	if b.coll != "" && b.coll != labels[0] {
		return unsupported("query spans collections %s and %s", b.coll, labels[0])
	}
	b.coll, b.alias = labels[0], alias
	b.fields[alias] = "$uuid"
	b.visible = appendNew(b.visible, alias)
	return nil
}

func (b *mongoBuilder) freshField() string {
	f := fmt.Sprintf("v%d", b.fresh)
	b.fresh++
	return f
}

// project は現在の行の列を v0, v1… の項目へ移し替える $project を足す（素の項目も運ぶ）。
func (b *mongoBuilder) project(items map[string]string, order []string) {
	body := Doc{{Key: "_id", Val: 0}}
	fields := map[string]string{}
	for _, name := range order {
		f := b.freshField()
		body = append(body, Field{Key: f, Val: items[name]})
		fields[name] = "$" + f
	}
	if !b.grouped {
		body = append(body, Field{Key: "uuid", Val: 1})
		fields[b.alias] = "$uuid"
		for key, src := range b.carry {
			if _, ok := fields[key]; ok {
				continue
			}
			f := b.freshField()
			body = append(body, Field{Key: f, Val: src})
			b.carry[key] = "$" + f
		}
	}
	b.stages = append(b.stages, Stage{Op: "$project", Body: body})
	b.fields = fields
	b.visible = order
}

func (b *mongoBuilder) op(n logical.Node) error {
	switch x := n.(type) {
	case *physical.Argument:
		if len(x.Cols) != 1 {
			return unsupported("non-graph query keyed by %v", x.Cols)
		}
		o, ok := b.own[x.Cols[0]]
		if !ok {
			return unsupported("unknown owner of %s", x.Cols[0])
		}
		b.keyed = true
		b.match = append(b.match, Field{Key: "uuid", Val: Doc{{Key: "$in", Val: InMarker}}})
		return b.setColl(x.Cols[0], o.labels)
	case *logical.EntityScan:
		return b.setColl(x.Alias, x.Labels)
	case *logical.GetProperty:
		if x.Alias != b.alias && b.coll != "" {
			return unsupported("property of %s in a query on %s", x.Alias, b.alias)
		}
		if b.grouped {
			return unsupported("property fetch after aggregation")
		}
		if b.coll == "" {
			if err := b.setColl(x.Alias, x.Labels); err != nil {
				return err
			}
		}
		if src, ok := b.carry[x.Key()]; ok && len(b.stages) > 0 {
			b.fields[x.Key()] = src // 遅延取得: 最初の $project で運んだ項目
		} else {
			b.fields[x.Key()] = "$" + x.Prop
		}
		b.visible = appendNew(b.visible, x.Key())
	case *logical.Filter:
		if len(b.stages) > 0 {
			return unsupported("filter after projection in a document query")
		}
		d, err := b.cond(x.Pred)
		if err != nil {
			return err
		}
		b.match = append(b.match, d...)
	case *logical.Projection:
		items := map[string]string{}
		var order []string
		for _, it := range x.Items {
			c, ok := it.Expr.(*logical.Column)
			if !ok {
				return unsupported("computed projection in a document query")
			}
			f, ok := b.fields[c.Key]
			if !ok {
				return unsupported("unknown column %s", c.Key)
			}
			items[it.Name] = f
			order = append(order, it.Name)
		}
		b.project(items, order)
	case *logical.Aggregation:
		id := Doc{}
		group := Doc{}
		items := map[string]string{}
		sizeOf := map[string]bool{} // count(DISTINCT) は集めた配列の大きさ
		var order []string
		for i, it := range x.Keys {
			c, ok := it.Expr.(*logical.Column)
			if !ok {
				return unsupported("computed group key in a document query")
			}
			k := fmt.Sprintf("k%d", i)
			id = append(id, Field{Key: k, Val: b.fields[c.Key]})
			items[it.Name] = "$_id." + k
			order = append(order, it.Name)
		}
		group = append(group, Field{Key: "_id", Val: id})
		for i, a := range x.Aggs {
			k := fmt.Sprintf("a%d", i)
			var acc Doc
			var src string
			if a.Arg != nil {
				c, ok := a.Arg.(*logical.Column)
				if !ok {
					return unsupported("computed aggregate argument in a document query")
				}
				src = b.fields[c.Key]
			}
			switch {
			case a.Fn == "count" && a.Arg == nil:
				acc = Doc{{Key: "$sum", Val: int64(1)}}
			case a.Fn == "count" && a.Distinct:
				acc = Doc{{Key: "$addToSet", Val: src}}
			case a.Fn == "count":
				// count(x) は null でない x の数。
				acc = Doc{{Key: "$sum", Val: Doc{{Key: "$cond", Val: []any{Doc{{Key: "$ne", Val: []any{src, nil}}}, int64(1), int64(0)}}}}}
			case a.Distinct:
				return unsupported("%s(DISTINCT) in a document query", a.Fn)
			default:
				acc = Doc{{Key: "$" + a.Fn, Val: src}}
			}
			group = append(group, Field{Key: k, Val: acc})
			items[a.Name] = "$" + k
			sizeOf[a.Name] = a.Fn == "count" && a.Distinct
			order = append(order, a.Name)
		}
		b.stages = append(b.stages, Stage{Op: "$group", Body: group})
		b.grouped = true
		// $group の出力を v0… へ（count DISTINCT は配列の大きさ）。
		body := Doc{{Key: "_id", Val: 0}}
		fields := map[string]string{}
		for _, name := range order {
			f := b.freshField()
			var val any = items[name]
			if sizeOf[name] {
				val = Doc{{Key: "$size", Val: items[name]}}
			}
			body = append(body, Field{Key: f, Val: val})
			fields[name] = "$" + f
		}
		b.stages = append(b.stages, Stage{Op: "$project", Body: body})
		b.fields, b.visible = fields, order
	case *logical.Sort:
		if len(b.stages) == 0 {
			b.project(b.currentItems())
		}
		body := Doc{}
		for _, k := range x.Keys {
			c, ok := k.Expr.(*logical.Column)
			if !ok {
				return unsupported("computed sort key in a document query")
			}
			f, ok := b.fields[c.Key]
			if !ok {
				return unsupported("unknown sort column %s", c.Key)
			}
			dir := int64(1)
			if k.Desc {
				dir = -1
			}
			body = append(body, Field{Key: f[1:], Val: dir})
		}
		b.stages = append(b.stages, Stage{Op: "$sort", Body: body})
	case *logical.Limit:
		if x.Skip > 0 {
			b.stages = append(b.stages, Stage{Op: "$skip", Body: Doc{{Key: "n", Val: int64(x.Skip)}}})
		}
		if x.Count >= 0 {
			b.stages = append(b.stages, Stage{Op: "$limit", Body: Doc{{Key: "n", Val: int64(x.Count)}}})
		}
	case *logical.Distinct:
		items, order := b.currentItems()
		id := Doc{}
		for i, name := range order {
			id = append(id, Field{Key: fmt.Sprintf("k%d", i), Val: items[name]})
		}
		b.stages = append(b.stages, Stage{Op: "$group", Body: Doc{{Key: "_id", Val: id}}})
		b.grouped = true
		body := Doc{{Key: "_id", Val: 0}}
		fields := map[string]string{}
		for i, name := range order {
			f := b.freshField()
			body = append(body, Field{Key: f, Val: fmt.Sprintf("$_id.k%d", i)})
			fields[name] = "$" + f
		}
		b.stages = append(b.stages, Stage{Op: "$project", Body: body})
		b.fields = fields
	default:
		return unsupported("%T in a document query", n)
	}
	return nil
}

func (b *mongoBuilder) currentItems() (map[string]string, []string) {
	items := map[string]string{}
	for _, v := range b.visible {
		items[v] = b.fields[v]
	}
	return items, b.visible
}

// cond は述語を $match の条件（AND の 1 項ずつ）にする。
func (b *mongoBuilder) cond(e logical.Expr) (Doc, error) {
	switch x := e.(type) {
	case *logical.Bin:
		switch x.Op {
		case "AND":
			l, err := b.cond(x.L)
			if err != nil {
				return nil, err
			}
			r, err := b.cond(x.R)
			if err != nil {
				return nil, err
			}
			return Doc{{Key: "$and", Val: []any{l, r}}}, nil
		case "OR":
			l, err := b.cond(x.L)
			if err != nil {
				return nil, err
			}
			r, err := b.cond(x.R)
			if err != nil {
				return nil, err
			}
			return Doc{{Key: "$or", Val: []any{l, r}}}, nil
		case "IN":
			field, typ, err := b.fieldOf(x.L)
			if err != nil {
				return nil, err
			}
			list, ok := x.R.(*logical.List)
			if !ok {
				return nil, unsupported("IN without a list")
			}
			vals := make([]any, len(list.Elems))
			for i, el := range list.Elems {
				lit, ok := el.(*logical.Lit)
				if !ok {
					return nil, unsupported("non-literal in IN list")
				}
				vals[i] = typed(lit.Val, typ)
			}
			return Doc{{Key: field, Val: Doc{{Key: "$in", Val: vals}}}}, nil
		}
		ops := map[string]string{"=": "$eq", "<>": "$ne", "<": "$lt", ">": "$gt", "<=": "$lte", ">=": "$gte"}
		op, ok := ops[x.Op]
		if !ok {
			return nil, unsupported("operator %s in a document filter", x.Op)
		}
		field, typ, err := b.fieldOf(x.L)
		if err != nil {
			return nil, err
		}
		lit, ok := x.R.(*logical.Lit)
		if !ok {
			return nil, unsupported("non-literal comparison in a document filter")
		}
		return Doc{{Key: field, Val: Doc{{Key: op, Val: typed(lit.Val, typ)}}}}, nil
	case *logical.Un:
		switch x.Op {
		case "IS NULL", "IS NOT NULL":
			field, _, err := b.fieldOf(x.X)
			if err != nil {
				return nil, err
			}
			if x.Op == "IS NULL" {
				return Doc{{Key: field, Val: nil}}, nil
			}
			return Doc{{Key: field, Val: Doc{{Key: "$ne", Val: nil}}}}, nil
		case "NOT":
			inner, err := b.cond(x.X)
			if err != nil {
				return nil, err
			}
			return Doc{{Key: "$nor", Val: []any{inner}}}, nil
		}
	}
	return nil, unsupported("document filter %s", e)
}

func (b *mongoBuilder) fieldOf(e logical.Expr) (string, string, error) {
	c, ok := e.(*logical.Column)
	if !ok {
		return "", "", unsupported("document filter on %s", e)
	}
	f, ok := b.fields[c.Key]
	if !ok || len(f) < 2 || f[0] != '$' {
		return "", "", unsupported("unknown column %s", c.Key)
	}
	return f[1:], colType(c.Key, b.own, b.types), nil
}
