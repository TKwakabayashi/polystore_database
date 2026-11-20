// Package mongo is the MongoDB database: aggregation pipeline compilation, execution and capabilities.
package mongo

import (
	"fmt"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/internal/lower"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
)

type mongoBuilder struct {
	coll    string
	alias   string
	match   stores.Doc
	fields  map[string]string // column key -> reference in the document ("$prop" / "$_id.k0" / "$v3" ...)
	visible []string
	stages  []stores.Stage
	fresh   int
	grouped bool
	own     map[string]lower.Owner
	types   stores.TypeFn
	keyed   bool
	carry   map[string]string // raw fields carried by the first $project (column key -> "$prop")
}

func lowerMongo(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	ops := lower.Chain(q.Ops)
	if ops == nil {
		return nil, lower.Unsupported("document query with a join")
	}
	b := &mongoBuilder{fields: map[string]string{}, own: lower.Owners(q.Ops), types: types, carry: map[string]string{}}
	// First collect every raw field used anywhere (deferred fetches are carried by the first $project too).
	for _, op := range ops {
		if g, ok := op.(*ir.GetProperty); ok {
			b.carry[g.Key()] = "$" + g.Prop
		}
	}
	for _, op := range ops {
		if err := b.op(op); err != nil {
			return nil, err
		}
	}
	if b.coll == "" {
		return nil, lower.Unsupported("document query without a collection")
	}

	out := stores.Doc{{Key: "_id", Val: 0}}
	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		if col == b.alias {
			kinds[i] = stores.KindIdent
		}
		f, ok := b.fields[col]
		if !ok {
			return nil, lower.Unsupported("output column %s not produced", col)
		}
		out = append(out, stores.Field{Key: lower.ColName(i), Val: f})
	}
	pipeline := append([]stores.Stage{{Op: "$match", Body: b.match}}, b.stages...)
	pipeline = append(pipeline, stores.Stage{Op: "$project", Body: out})
	return &stores.Query{Store: kind.Mongo, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys,
		Body: &stores.PipelineBody{Collection: b.coll, Stages: pipeline}}, nil
}

func (b *mongoBuilder) setColl(alias string, labels []string) error {
	if len(labels) != 1 {
		return lower.Unsupported("collection for %s needs exactly one label (%v)", alias, labels)
	}
	if b.coll != "" && b.coll != labels[0] {
		return lower.Unsupported("query spans collections %s and %s", b.coll, labels[0])
	}
	b.coll, b.alias = labels[0], alias
	b.fields[alias] = "$uuid"
	b.visible = lower.AppendNew(b.visible, alias)
	return nil
}

func (b *mongoBuilder) freshField() string {
	f := fmt.Sprintf("v%d", b.fresh)
	b.fresh++
	return f
}

// project adds a $project moving the current row's columns to fields v0, v1, ... (raw fields are carried along).
func (b *mongoBuilder) project(items map[string]string, order []string) {
	body := stores.Doc{{Key: "_id", Val: 0}}
	fields := map[string]string{}
	for _, name := range order {
		f := b.freshField()
		body = append(body, stores.Field{Key: f, Val: items[name]})
		fields[name] = "$" + f
	}
	if !b.grouped {
		body = append(body, stores.Field{Key: "uuid", Val: 1})
		fields[b.alias] = "$uuid"
		for key, src := range b.carry {
			if _, ok := fields[key]; ok {
				continue
			}
			f := b.freshField()
			body = append(body, stores.Field{Key: f, Val: src})
			b.carry[key] = "$" + f
		}
	}
	b.stages = append(b.stages, stores.Stage{Op: "$project", Body: body})
	b.fields = fields
	b.visible = order
}

func (b *mongoBuilder) op(n ir.Node) error {
	switch x := n.(type) {
	case *ir.Boundary:
		if len(x.Cols) != 1 {
			return lower.Unsupported("non-graph query keyed by %v", x.Cols)
		}
		o, ok := b.own[x.Cols[0]]
		if !ok {
			return lower.Unsupported("unknown owner of %s", x.Cols[0])
		}
		b.keyed = true
		b.match = append(b.match, stores.Field{Key: "uuid", Val: stores.Doc{{Key: "$in", Val: stores.InMarker}}})
		return b.setColl(x.Cols[0], o.Labels)
	case *ir.EntityScan:
		return b.setColl(x.Alias, x.Labels)
	case *ir.GetProperty:
		if x.Alias != b.alias && b.coll != "" {
			return lower.Unsupported("property of %s in a query on %s", x.Alias, b.alias)
		}
		if b.grouped {
			return lower.Unsupported("property fetch after aggregation")
		}
		if b.coll == "" {
			if err := b.setColl(x.Alias, x.Labels); err != nil {
				return err
			}
		}
		if src, ok := b.carry[x.Key()]; ok && len(b.stages) > 0 {
			b.fields[x.Key()] = src // deferred fetch: a field carried by the first $project
		} else {
			b.fields[x.Key()] = "$" + x.Prop
		}
		b.visible = lower.AppendNew(b.visible, x.Key())
	case *ir.Filter:
		d, err := b.cond(x.Pred)
		if err != nil {
			return err
		}
		if len(b.stages) > 0 {
			b.stages = append(b.stages, stores.Stage{Op: "$match", Body: d}) // condition after a projection or aggregation
		} else {
			b.match = append(b.match, d...)
		}
	case *ir.Projection:
		items := map[string]string{}
		var order []string
		for _, it := range x.Items {
			c, ok := it.Expr.(*ir.Column)
			if !ok {
				return lower.Unsupported("computed projection in a document query")
			}
			f, ok := b.fields[c.Key]
			if !ok {
				return lower.Unsupported("unknown column %s", c.Key)
			}
			items[it.Name] = f
			order = append(order, it.Name)
		}
		b.project(items, order)
	case *ir.Aggregation:
		id := stores.Doc{}
		group := stores.Doc{}
		items := map[string]string{}
		sizeOf := map[string]bool{} // count(DISTINCT) is the size of the collected array
		var order []string
		for i, it := range x.Keys {
			c, ok := it.Expr.(*ir.Column)
			if !ok {
				return lower.Unsupported("computed group key in a document query")
			}
			k := fmt.Sprintf("k%d", i)
			id = append(id, stores.Field{Key: k, Val: b.fields[c.Key]})
			items[it.Name] = "$_id." + k
			order = append(order, it.Name)
		}
		group = append(group, stores.Field{Key: "_id", Val: id})
		for i, a := range x.Aggs {
			k := fmt.Sprintf("a%d", i)
			var acc stores.Doc
			var src string
			if a.Arg != nil {
				c, ok := a.Arg.(*ir.Column)
				if !ok {
					return lower.Unsupported("computed aggregate argument in a document query")
				}
				src = b.fields[c.Key]
			}
			switch {
			case a.Fn == "count" && a.Arg == nil:
				acc = stores.Doc{{Key: "$sum", Val: int64(1)}}
			case a.Fn == "count" && a.Distinct:
				acc = stores.Doc{{Key: "$addToSet", Val: src}}
			case a.Fn == "count":
				// count(x) counts non-null x.
				acc = stores.Doc{{Key: "$sum", Val: stores.Doc{{Key: "$cond", Val: []any{stores.Doc{{Key: "$ne", Val: []any{src, nil}}}, int64(1), int64(0)}}}}}
			case a.Distinct:
				return lower.Unsupported("%s(DISTINCT) in a document query", a.Fn)
			default:
				acc = stores.Doc{{Key: "$" + a.Fn, Val: src}}
			}
			group = append(group, stores.Field{Key: k, Val: acc})
			items[a.Name] = "$" + k
			sizeOf[a.Name] = a.Fn == "count" && a.Distinct
			order = append(order, a.Name)
		}
		b.stages = append(b.stages, stores.Stage{Op: "$group", Body: group})
		b.grouped = true
		// Move the $group output to v0, ... (count DISTINCT is the array size).
		body := stores.Doc{{Key: "_id", Val: 0}}
		fields := map[string]string{}
		for _, name := range order {
			f := b.freshField()
			var val any = items[name]
			if sizeOf[name] {
				val = stores.Doc{{Key: "$size", Val: items[name]}}
			}
			body = append(body, stores.Field{Key: f, Val: val})
			fields[name] = "$" + f
		}
		b.stages = append(b.stages, stores.Stage{Op: "$project", Body: body})
		b.fields, b.visible = fields, order
	case *ir.Sort:
		if len(b.stages) == 0 {
			b.project(b.currentItems())
		}
		body := stores.Doc{}
		for _, k := range x.Keys {
			c, ok := k.Expr.(*ir.Column)
			if !ok {
				return lower.Unsupported("computed sort key in a document query")
			}
			f, ok := b.fields[c.Key]
			if !ok {
				return lower.Unsupported("unknown sort column %s", c.Key)
			}
			dir := int64(1)
			if k.Desc {
				dir = -1
			}
			body = append(body, stores.Field{Key: f[1:], Val: dir})
		}
		b.stages = append(b.stages, stores.Stage{Op: "$sort", Body: body})
	case *ir.Limit:
		if x.Skip > 0 {
			b.stages = append(b.stages, stores.Stage{Op: "$skip", Body: stores.Doc{{Key: "n", Val: int64(x.Skip)}}})
		}
		if x.Count >= 0 {
			b.stages = append(b.stages, stores.Stage{Op: "$limit", Body: stores.Doc{{Key: "n", Val: int64(x.Count)}}})
		}
	case *ir.Distinct:
		items, order := b.currentItems()
		id := stores.Doc{}
		for i, name := range order {
			id = append(id, stores.Field{Key: fmt.Sprintf("k%d", i), Val: items[name]})
		}
		b.stages = append(b.stages, stores.Stage{Op: "$group", Body: stores.Doc{{Key: "_id", Val: id}}})
		b.grouped = true
		body := stores.Doc{{Key: "_id", Val: 0}}
		fields := map[string]string{}
		for i, name := range order {
			f := b.freshField()
			body = append(body, stores.Field{Key: f, Val: fmt.Sprintf("$_id.k%d", i)})
			fields[name] = "$" + f
		}
		b.stages = append(b.stages, stores.Stage{Op: "$project", Body: body})
		b.fields = fields
	default:
		return lower.Unsupported("%T in a document query", n)
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

// cond turns a predicate into $match conditions (one per AND term).
func (b *mongoBuilder) cond(e ir.Expr) (stores.Doc, error) {
	switch x := e.(type) {
	case *ir.Bin:
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
			return stores.Doc{{Key: "$and", Val: []any{l, r}}}, nil
		case "OR":
			l, err := b.cond(x.L)
			if err != nil {
				return nil, err
			}
			r, err := b.cond(x.R)
			if err != nil {
				return nil, err
			}
			return stores.Doc{{Key: "$or", Val: []any{l, r}}}, nil
		case "IN":
			field, typ, err := b.fieldOf(x.L)
			if err != nil {
				return nil, err
			}
			list, ok := x.R.(*ir.List)
			if !ok {
				return nil, lower.Unsupported("IN without a list")
			}
			vals := make([]any, len(list.Elems))
			for i, el := range list.Elems {
				lit, ok := el.(*ir.Lit)
				if !ok {
					return nil, lower.Unsupported("non-literal in IN list")
				}
				vals[i] = lower.Typed(lit.Val, typ)
			}
			return stores.Doc{{Key: field, Val: stores.Doc{{Key: "$in", Val: vals}}}}, nil
		}
		ops := map[string]string{"=": "$eq", "<>": "$ne", "<": "$lt", ">": "$gt", "<=": "$lte", ">=": "$gte"}
		op, ok := ops[x.Op]
		if !ok {
			return nil, lower.Unsupported("operator %s in a document filter", x.Op)
		}
		field, typ, err := b.fieldOf(x.L)
		if err != nil {
			return nil, err
		}
		lit, ok := x.R.(*ir.Lit)
		if !ok {
			return nil, lower.Unsupported("non-literal comparison in a document filter")
		}
		return stores.Doc{{Key: field, Val: stores.Doc{{Key: op, Val: lower.Typed(lit.Val, typ)}}}}, nil
	case *ir.Un:
		switch x.Op {
		case "IS NULL", "IS NOT NULL":
			field, _, err := b.fieldOf(x.X)
			if err != nil {
				return nil, err
			}
			if x.Op == "IS NULL" {
				return stores.Doc{{Key: field, Val: nil}}, nil
			}
			return stores.Doc{{Key: field, Val: stores.Doc{{Key: "$ne", Val: nil}}}}, nil
		case "NOT":
			inner, err := b.cond(x.X)
			if err != nil {
				return nil, err
			}
			return stores.Doc{{Key: "$nor", Val: []any{inner}}}, nil
		}
	}
	return nil, lower.Unsupported("document filter %s", e)
}

func (b *mongoBuilder) fieldOf(e ir.Expr) (string, string, error) {
	c, ok := e.(*ir.Column)
	if !ok {
		return "", "", lower.Unsupported("document filter on %s", e)
	}
	f, ok := b.fields[c.Key]
	if !ok || len(f) < 2 || f[0] != '$' {
		return "", "", lower.Unsupported("unknown column %s", c.Key)
	}
	return f[1:], lower.ColType(c.Key, b.own, b.types), nil
}

func lowerShipMongo(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	arg, single, o, ops, err := lower.ShipParts(q)
	if err != nil {
		return nil, err
	}
	tab, err := lower.NewShipTable(q, arg)
	if err != nil {
		return nil, err
	}
	b := &mongoBuilder{fields: map[string]string{}, own: lower.Owners(q.Ops), types: types, carry: map[string]string{}}
	idents := map[string]bool{}
	for _, c := range tab.Cols {
		b.fields[c] = "$" + tab.Names[c]
		idents[c] = tab.IsIdent(c)
	}
	w := b.fields[ir.WeightCol]
	var match stores.Doc
	var tail []stores.Stage
	grouped := false
	for _, op := range ops {
		switch x := op.(type) {
		case *ir.GetProperty:
			if x.Alias != single {
				return nil, lower.Unsupported("property of %s in a shipped query on %s", x.Alias, single)
			}
			b.fields[x.Key()] = "$d." + x.Prop
		case *ir.Filter:
			if grouped {
				return nil, lower.Unsupported("filter after a shipped aggregation")
			}
			c, err := b.cond(x.Pred)
			if err != nil {
				return nil, err
			}
			match = append(match, c...)
		case *ir.Aggregation:
			if q.ShipMode != ir.ShipAggregate {
				return nil, lower.Unsupported("aggregation in a %s shipped query", q.ShipMode)
			}
			idDoc := stores.Doc{}
			group := stores.Doc{}
			fields := map[string]string{}
			for i, it := range x.Keys {
				c, ok := it.Expr.(*ir.Column)
				if !ok {
					return nil, lower.Unsupported("computed group key in a shipped query")
				}
				k := fmt.Sprintf("g%d", i)
				idDoc = append(idDoc, stores.Field{Key: k, Val: b.fields[c.Key]})
				fields[it.Name] = "$_id." + k
				idents[it.Name] = idents[c.Key]
			}
			group = append(group, stores.Field{Key: "_id", Val: idDoc})
			for i, a := range x.Aggs {
				acc, err := weightedMongo(b, a, w)
				if err != nil {
					return nil, err
				}
				k := fmt.Sprintf("a%d", i)
				group = append(group, stores.Field{Key: k, Val: acc})
				fields[a.Name] = "$" + k
			}
			tail = append(tail, stores.Stage{Op: "$group", Body: group})
			b.fields, grouped = fields, true
		default:
			return nil, lower.Unsupported("%T in a shipped query", op)
		}
	}
	if q.ShipMode == ir.ShipAggregate && !grouped {
		return nil, lower.Unsupported("shipped aggregation without an aggregation")
	}

	out := stores.Doc{{Key: "_id", Val: 0}}
	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		f, ok := b.fields[col]
		if !ok {
			return nil, lower.Unsupported("output column %s not produced", col)
		}
		out = append(out, stores.Field{Key: lower.ColName(i), Val: f})
		if idents[col] {
			kinds[i] = stores.KindIdent
		}
	}
	pipeline := []stores.Stage{
		{Op: "$documents", Body: stores.Doc{{Key: "rows", Val: stores.InMarker}}},
		{Op: "$lookup", Body: stores.Doc{{Key: "from", Val: o.Labels[0]}, {Key: "localField", Val: tab.Names[single]},
			{Key: "foreignField", Val: "uuid"}, {Key: "as", Val: "d"}}},
		{Op: "$unwind", Body: stores.Doc{{Key: "path", Val: "$d"}, {Key: "preserveNullAndEmptyArrays", Val: true}}},
	}
	if len(match) > 0 {
		pipeline = append(pipeline, stores.Stage{Op: "$match", Body: match})
	}
	pipeline = append(pipeline, tail...)
	pipeline = append(pipeline, stores.Stage{Op: "$project", Body: out})
	return &stores.Query{Store: kind.Mongo, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys,
		Body: &stores.PipelineBody{Stages: pipeline, OnDatabase: true}, ShipCols: tab.Cols, ShipNames: tab.ColNames(), ShipMode: q.ShipMode}, nil
}

// weightedMongo is the $group form of weightedSQL.
func weightedMongo(b *mongoBuilder, a ir.AggItem, w string) (stores.Doc, error) {
	if a.Distinct {
		return nil, lower.Unsupported("%s(DISTINCT) in a shipped aggregation", a.Fn)
	}
	if a.Arg == nil {
		if a.Fn != "count" {
			return nil, lower.Unsupported("%s(*)", a.Fn)
		}
		return stores.Doc{{Key: "$sum", Val: w}}, nil
	}
	c, ok := a.Arg.(*ir.Column)
	if !ok {
		return nil, lower.Unsupported("computed aggregate argument in a shipped query")
	}
	x, ok := b.fields[c.Key]
	if !ok {
		return nil, lower.Unsupported("unknown column %s", c.Key)
	}
	switch a.Fn {
	case "count":
		// A missing field (no value in a left outer join) is not counted either.
		isNull := stores.Doc{{Key: "$eq", Val: []any{stores.Doc{{Key: "$ifNull", Val: []any{x, nil}}}, nil}}}
		return stores.Doc{{Key: "$sum", Val: stores.Doc{{Key: "$cond", Val: []any{isNull, int64(0), w}}}}}, nil
	case "sum":
		return stores.Doc{{Key: "$sum", Val: stores.Doc{{Key: "$multiply", Val: []any{w, x}}}}}, nil
	case "min", "max":
		return stores.Doc{{Key: "$" + a.Fn, Val: x}}, nil
	}
	return nil, lower.Unsupported("%s in a shipped aggregation", a.Fn)
}
