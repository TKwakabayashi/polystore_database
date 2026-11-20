package sqlq

import (
	"fmt"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/internal/lower"
	"strings"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
)

func LowerShip(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	arg, single, o, ops, err := lower.ShipParts(q)
	if err != nil {
		return nil, err
	}
	tab, err := lower.NewShipTable(q, arg)
	if err != nil {
		return nil, err
	}
	d := MySQL
	tcol := func(col string) string { return "t." + d.ident(tab.Names[col]) }
	uuidOf := func(col string) string {
		if id.Binary() {
			return "UNHEX(" + tcol(col) + ")" // binary uuids travel as hex (id.StoredText)
		}
		return tcol(col)
	}
	b := &sqlBuilder{d: d, cols: map[string]string{}, own: lower.Owners(q.Ops), types: types, inArg: 0, limit: -1}
	b.args = append(b.args, stores.InMarker) // the JSON_TABLE in FROM is the first placeholder
	idents := map[string]bool{}
	for _, c := range tab.Cols {
		if tab.IsIdent(c) {
			b.cols[c] = uuidOf(c)
			idents[c] = true
		} else {
			b.cols[c] = tcol(c)
		}
	}
	w := b.cols[ir.WeightCol]
	on := b.cols[single] // keep the join's match expression before the aggregation replaces b.cols

	var sel []string
	for _, op := range ops {
		switch x := op.(type) {
		case *ir.GetProperty:
			if x.Alias != single {
				return nil, lower.Unsupported("property of %s in a shipped query on %s", x.Alias, single)
			}
			b.cols[x.Key()] = "d." + d.ident(x.Prop)
		case *ir.Filter:
			if b.grouped {
				return nil, lower.Unsupported("filter after a shipped aggregation")
			}
			e, err := b.expr(x.Pred, "")
			if err != nil {
				return nil, err
			}
			b.where = append(b.where, e)
		case *ir.Aggregation:
			if q.ShipMode != ir.ShipAggregate {
				return nil, lower.Unsupported("aggregation in a %s shipped query", q.ShipMode)
			}
			cols := map[string]string{}
			for _, it := range x.Keys {
				e, err := b.expr(it.Expr, "")
				if err != nil {
					return nil, err
				}
				cols[it.Name] = e
				b.group = append(b.group, e)
				if c, ok := it.Expr.(*ir.Column); ok && idents[c.Key] {
					idents[it.Name] = true
				}
			}
			for _, a := range x.Aggs {
				e, err := weightedSQL(b, a, w)
				if err != nil {
					return nil, err
				}
				cols[a.Name] = e
			}
			b.cols, b.grouped = cols, true
		default:
			return nil, lower.Unsupported("%T in a shipped query", op)
		}
	}
	switch q.ShipMode {
	case ir.ShipAggregate:
		if !b.grouped {
			return nil, lower.Unsupported("shipped aggregation without an aggregation")
		}
	case ir.ShipFilter, ir.ShipProject:
		idents[ir.RidCol] = false
	}

	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		e, ok := b.cols[col]
		if !ok {
			return nil, lower.Unsupported("output column %s not produced", col)
		}
		sel = append(sel, e+" AS "+lower.ColName(i))
		if idents[col] {
			kinds[i] = stores.KindIdent
		}
	}
	defs := make([]string, len(tab.Cols))
	for i, c := range tab.Cols {
		typ := "BIGINT"
		if tab.IsIdent(c) {
			typ = fmt.Sprintf("VARCHAR(%d) COLLATE utf8mb4_bin", shipTextLen())
		}
		defs[i] = fmt.Sprintf("%s %s PATH '$[%d]'", d.ident(tab.Names[c]), typ, i)
	}
	text := "SELECT " + strings.Join(sel, ", ") +
		" FROM JSON_TABLE(CAST(? AS JSON), '$[*]' COLUMNS(" + strings.Join(defs, ", ") + ")) AS t" +
		" LEFT JOIN " + d.ident(o.Labels[0]) + " AS d ON d." + d.ident("uuid") + " = " + on
	if len(b.where) > 0 {
		text += " WHERE " + strings.Join(b.where, " AND ")
	}
	if len(b.group) > 0 {
		text += " GROUP BY " + strings.Join(b.group, ", ")
	}
	return &stores.Query{Store: kind.MySQL, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys,
		Body: &stores.TextBody{Text: text, Args: b.args, InArg: 0}, ShipCols: tab.Cols, ShipNames: tab.ColNames(), ShipMode: q.ShipMode}, nil
}

// weightedSQL aggregates rows with weight w: count(*) = sum w, count(x) = sum of w where x is not null, sum(x) = sum w*x.
func weightedSQL(b *sqlBuilder, a ir.AggItem, w string) (string, error) {
	if a.Distinct {
		return "", lower.Unsupported("%s(DISTINCT) in a shipped aggregation", a.Fn)
	}
	if a.Arg == nil {
		if a.Fn != "count" {
			return "", lower.Unsupported("%s(*)", a.Fn)
		}
		return "SUM(" + w + ")", nil
	}
	x, err := b.expr(a.Arg, "")
	if err != nil {
		return "", err
	}
	switch a.Fn {
	case "count":
		return "SUM(CASE WHEN " + x + " IS NULL THEN 0 ELSE " + w + " END)", nil
	case "sum":
		return "SUM(" + w + " * " + x + ")", nil
	case "min", "max":
		return strings.ToUpper(a.Fn) + "(" + x + ")", nil
	}
	return "", lower.Unsupported("%s in a shipped aggregation", a.Fn)
}

// shipTextLen is the length of a shipped uuid (36 as text; 32 hex digits for 16 bytes).
func shipTextLen() int {
	if id.Binary() {
		return 2 * id.Size
	}
	return id.TextSize
}
