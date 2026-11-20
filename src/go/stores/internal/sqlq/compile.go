// Package sqlq compiles store queries to SQL-like text (MySQL, ClickHouse, CQL) and runs them.
package sqlq

import (
	"fmt"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/internal/lower"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/query/ir"
)

type Dialect int

const (
	MySQL Dialect = iota
	CQL
	// ClickHouse gets literals embedded in the text (clickhouse-go rounds datetime parameters to seconds).
	// Its only positional parameter is the input uuid list (InArg), replaced by the executor.
	ClickHouse
)

func (d Dialect) ident(s string) string {
	if d == CQL {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

type sqlBuilder struct {
	d        Dialect
	table    string
	alias    string
	cols     map[string]string // column key -> SQL expression
	visible  []string
	where    []string
	args     []any
	inArg    int
	group    []string
	grouped  bool
	having   []string
	order    []string
	limit    int
	offset   int
	distinct bool
	own      map[string]lower.Owner
	types    stores.TypeFn
}

func Lower(q *ir.StoreQuery, types stores.TypeFn, d Dialect) (*stores.Query, error) {
	ops := lower.Chain(q.Ops)
	if ops == nil {
		return nil, lower.Unsupported("%s query with a join", q.Store)
	}
	b := &sqlBuilder{d: d, cols: map[string]string{}, own: lower.Owners(q.Ops), types: types, inArg: -1, limit: -1}
	for _, op := range ops {
		if err := b.op(op); err != nil {
			return nil, err
		}
	}
	if b.table == "" {
		return nil, lower.Unsupported("%s query without a table", q.Store)
	}

	sel := make([]string, len(q.Outputs))
	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		e, ok := b.cols[col]
		if !ok {
			return nil, lower.Unsupported("output column %s not produced", col)
		}
		sel[i] = e + " AS " + lower.ColName(i)
		if col == b.alias {
			kinds[i] = stores.KindIdent
		}
	}
	head := "SELECT "
	if b.distinct {
		head += "DISTINCT "
	}
	text := head + strings.Join(sel, ", ") + " FROM " + d.ident(b.table)
	if len(b.where) > 0 {
		text += " WHERE " + strings.Join(b.where, " AND ")
	}
	if b.grouped && len(b.group) > 0 {
		text += " GROUP BY " + strings.Join(b.group, ", ")
	}
	if len(b.having) > 0 {
		if d == CQL {
			return nil, lower.Unsupported("HAVING in CQL")
		}
		text += " HAVING " + strings.Join(b.having, " AND ")
	}
	if len(b.order) > 0 {
		text += " ORDER BY " + strings.Join(b.order, ", ")
	}
	if b.limit >= 0 {
		text += fmt.Sprintf(" LIMIT %d", b.limit)
	}
	if b.offset > 0 {
		text += fmt.Sprintf(" OFFSET %d", b.offset)
	}
	if d == CQL && (len(b.where) > 1 || (len(b.where) == 1 && b.inArg < 0)) {
		text += " ALLOW FILTERING"
	}
	return &stores.Query{Store: q.Store, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys,
		Body: &stores.TextBody{Text: text, Args: b.args, InArg: b.inArg}}, nil
}

func (b *sqlBuilder) setTable(alias string, labels []string) error {
	if len(labels) != 1 {
		return lower.Unsupported("table for %s needs exactly one label (%v)", alias, labels)
	}
	if b.table != "" && b.table != labels[0] {
		return lower.Unsupported("query spans tables %s and %s", b.table, labels[0])
	}
	b.table, b.alias = labels[0], alias
	b.cols[alias] = b.d.ident("uuid")
	b.visible = lower.AppendNew(b.visible, alias)
	return nil
}

func (b *sqlBuilder) op(n ir.Node) error {
	switch x := n.(type) {
	case *ir.Boundary:
		if len(x.Cols) != 1 {
			return lower.Unsupported("non-graph query keyed by %v", x.Cols)
		}
		o, ok := b.own[x.Cols[0]]
		if !ok {
			return lower.Unsupported("unknown owner of %s", x.Cols[0])
		}
		if err := b.setTable(x.Cols[0], o.Labels); err != nil {
			return err
		}
		b.inArg = len(b.args)
		b.args = append(b.args, stores.InMarker)
		if b.d == CQL {
			b.where = append(b.where, b.d.ident("uuid")+" IN ?")
		} else {
			b.where = append(b.where, b.d.ident("uuid")+" IN (?)")
		}
	case *ir.EntityScan:
		return b.setTable(x.Alias, x.Labels)
	case *ir.GetProperty:
		if x.Alias != b.alias {
			return lower.Unsupported("property of %s in a query on %s", x.Alias, b.alias)
		}
		if b.grouped {
			return lower.Unsupported("property fetch after aggregation")
		}
		if b.table == "" {
			if err := b.setTable(x.Alias, x.Labels); err != nil {
				return err
			}
		}
		b.cols[x.Key()] = b.d.ident(x.Prop)
		b.visible = lower.AppendNew(b.visible, x.Key())
	case *ir.Filter:
		e, err := b.expr(x.Pred, "")
		if err != nil {
			return err
		}
		if b.grouped {
			b.having = append(b.having, e) // condition after the aggregation
		} else {
			b.where = append(b.where, e)
		}
	case *ir.Projection:
		cols := map[string]string{}
		var vis []string
		for _, it := range x.Items {
			e, err := b.expr(it.Expr, "")
			if err != nil {
				return err
			}
			cols[it.Name] = e
			vis = append(vis, it.Name)
		}
		// One table: later GetProperty (deferred fetches) add columns of the same row. The alias stays.
		cols[b.alias] = b.d.ident("uuid")
		b.cols, b.visible = cols, vis
	case *ir.Aggregation:
		cols := map[string]string{}
		var vis []string
		for _, it := range x.Keys {
			e, err := b.expr(it.Expr, "")
			if err != nil {
				return err
			}
			cols[it.Name] = e
			b.group = append(b.group, e)
			vis = append(vis, it.Name)
		}
		for _, a := range x.Aggs {
			e := "count(*)"
			if a.Arg != nil {
				arg, err := b.expr(a.Arg, "")
				if err != nil {
					return err
				}
				if a.Distinct {
					arg = "DISTINCT " + arg
				}
				e = a.Fn + "(" + arg + ")"
			}
			cols[a.Name] = e
			vis = append(vis, a.Name)
		}
		b.cols, b.visible, b.grouped = cols, vis, true
	case *ir.Sort:
		for _, k := range x.Keys {
			e, err := b.expr(k.Expr, "")
			if err != nil {
				return err
			}
			if k.Desc {
				e += " DESC"
			} else {
				e += " ASC"
			}
			if b.d == ClickHouse { // Cypher sorts null as the largest value (ClickHouse defaults to NULLS LAST in both directions)
				if k.Desc {
					e += " NULLS FIRST"
				} else {
					e += " NULLS LAST"
				}
			}
			b.order = append(b.order, e)
		}
	case *ir.Limit:
		b.limit, b.offset = x.Count, x.Skip
	case *ir.Distinct:
		b.distinct = true
	default:
		return lower.Unsupported("%T in a single-table query", n)
	}
	return nil
}

func (b *sqlBuilder) arg(v any) string {
	b.args = append(b.args, v)
	return "?"
}

func (b *sqlBuilder) expr(e ir.Expr, typ string) (string, error) {
	switch x := e.(type) {
	case *ir.Column:
		c, ok := b.cols[x.Key]
		if !ok {
			return "", lower.Unsupported("unknown column %s", x.Key)
		}
		return c, nil
	case *ir.Lit:
		if b.d == ClickHouse {
			return clickHouseLiteral(lower.Typed(x.Val, typ), typ)
		}
		return b.arg(lower.Typed(x.Val, typ)), nil
	case *ir.Bin:
		if x.Op == "IN" {
			l, err := b.expr(x.L, "")
			if err != nil {
				return "", err
			}
			list, ok := x.R.(*ir.List)
			if !ok {
				return "", lower.Unsupported("IN without a list")
			}
			t := lower.LitType(x.L, b.own, b.types)
			ph := make([]string, len(list.Elems))
			for i, el := range list.Elems {
				s, err := b.expr(el, t)
				if err != nil {
					return "", err
				}
				ph[i] = s
			}
			return l + " IN (" + strings.Join(ph, ", ") + ")", nil
		}
		l, err := b.expr(x.L, lower.LitType(x.R, b.own, b.types))
		if err != nil {
			return "", err
		}
		r, err := b.expr(x.R, lower.LitType(x.L, b.own, b.types))
		if err != nil {
			return "", err
		}
		return "(" + l + " " + x.Op + " " + r + ")", nil
	case *ir.Un:
		v, err := b.expr(x.X, typ)
		if err != nil {
			return "", err
		}
		switch x.Op {
		case "IS NULL", "IS NOT NULL":
			return v + " " + x.Op, nil
		case "NOT":
			return "NOT (" + v + ")", nil
		}
		return "(-" + v + ")", nil
	}
	return "", lower.Unsupported("SQL expression %T", e)
}

// clickHouseLiteral renders a ClickHouse literal. Datetimes keep fractional seconds (compared with DateTime64(6)).
func clickHouseLiteral(v any, typ string) (string, error) {
	switch x := v.(type) {
	case nil:
		return "NULL", nil
	case bool:
		return strconv.FormatBool(x), nil
	case int:
		return strconv.Itoa(x), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case string:
		return ClickHouseString(x), nil
	case time.Time:
		if strings.EqualFold(typ, "date") {
			return "toDate32('" + x.UTC().Format("2006-01-02") + "')", nil
		}
		return "toDateTime64('" + x.UTC().Format("2006-01-02 15:04:05.000000") + "', 6, 'UTC')", nil
	}
	return "", lower.Unsupported("ClickHouse literal %T", v)
}

// ClickHouseString renders a ClickHouse string literal ('...').
func ClickHouseString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + r.Replace(s) + "'"
}
