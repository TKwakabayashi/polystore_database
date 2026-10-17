package native

import (
	"fmt"
	"strings"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
)

// sql.go は relational（MySQL）と columnar（Cassandra CQL）向けの変換。非 graph の StoreQuery は
// 1 つの alias（＝1 テーブル）だけを扱うので、単一テーブルの SELECT にまとまる。
//   SELECT <列> FROM <ラベル> WHERE uuid IN (<入力>) AND <条件> [GROUP BY] [ORDER BY] [LIMIT]

type dialect int

const (
	dialectMySQL dialect = iota
	dialectCQL
)

func (d dialect) ident(s string) string {
	if d == dialectCQL {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

type sqlBuilder struct {
	d        dialect
	table    string
	alias    string
	cols     map[string]string // 列キー → SQL 式
	visible  []string
	where    []string
	args     []any
	inArg    int
	group    []string
	grouped  bool
	order    []string
	limit    int
	offset   int
	distinct bool
	own      map[string]owner
	types    TypeFn
}

func lowerSQL(q *physical.StoreQuery, types TypeFn, d dialect) (*Query, error) {
	ops := chain(q.Ops)
	if ops == nil {
		return nil, unsupported("%s query with a join", q.Store)
	}
	b := &sqlBuilder{d: d, cols: map[string]string{}, own: owners(q.Ops), types: types, inArg: -1, limit: -1}
	for _, op := range ops {
		if err := b.op(op); err != nil {
			return nil, err
		}
	}
	if b.table == "" {
		return nil, unsupported("%s query without a table", q.Store)
	}

	sel := make([]string, len(q.Outputs))
	for i, col := range q.Outputs {
		e, ok := b.cols[col]
		if !ok {
			return nil, unsupported("output column %s not produced", col)
		}
		sel[i] = e + " AS " + colName(i)
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
	if len(b.order) > 0 {
		text += " ORDER BY " + strings.Join(b.order, ", ")
	}
	if b.limit >= 0 {
		text += fmt.Sprintf(" LIMIT %d", b.limit)
	}
	if b.offset > 0 {
		text += fmt.Sprintf(" OFFSET %d", b.offset)
	}
	if d == dialectCQL && (len(b.where) > 1 || (len(b.where) == 1 && b.inArg < 0)) {
		text += " ALLOW FILTERING"
	}
	return &Query{Store: q.Store, Columns: q.Outputs, InKeys: q.Keys, Text: text, Args: b.args, InArg: b.inArg}, nil
}

func (b *sqlBuilder) setTable(alias string, labels []string) error {
	if len(labels) != 1 {
		return unsupported("table for %s needs exactly one label (%v)", alias, labels)
	}
	if b.table != "" && b.table != labels[0] {
		return unsupported("query spans tables %s and %s", b.table, labels[0])
	}
	b.table, b.alias = labels[0], alias
	b.cols[alias] = b.d.ident("uuid")
	b.visible = appendNew(b.visible, alias)
	return nil
}

func (b *sqlBuilder) op(n logical.Node) error {
	switch x := n.(type) {
	case *physical.Argument:
		if len(x.Cols) != 1 {
			return unsupported("non-graph query keyed by %v", x.Cols)
		}
		o, ok := b.own[x.Cols[0]]
		if !ok {
			return unsupported("unknown owner of %s", x.Cols[0])
		}
		if err := b.setTable(x.Cols[0], o.labels); err != nil {
			return err
		}
		b.inArg = len(b.args)
		b.args = append(b.args, InMarker)
		if b.d == dialectCQL {
			b.where = append(b.where, b.d.ident("uuid")+" IN ?")
		} else {
			b.where = append(b.where, b.d.ident("uuid")+" IN (?)")
		}
	case *logical.EntityScan:
		return b.setTable(x.Alias, x.Labels)
	case *logical.GetProperty:
		if x.Alias != b.alias {
			return unsupported("property of %s in a query on %s", x.Alias, b.alias)
		}
		if b.grouped {
			return unsupported("property fetch after aggregation")
		}
		if b.table == "" {
			if err := b.setTable(x.Alias, x.Labels); err != nil {
				return err
			}
		}
		b.cols[x.Key()] = b.d.ident(x.Prop)
		b.visible = appendNew(b.visible, x.Key())
	case *logical.Filter:
		e, err := b.expr(x.Pred, "")
		if err != nil {
			return err
		}
		b.where = append(b.where, e)
	case *logical.Projection:
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
		// 単一テーブルなので、後の GetProperty（遅延取得）は同じ行の列として足せる。alias は残す。
		cols[b.alias] = b.d.ident("uuid")
		b.cols, b.visible = cols, vis
	case *logical.Aggregation:
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
	case *logical.Sort:
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
			b.order = append(b.order, e)
		}
	case *logical.Limit:
		b.limit, b.offset = x.Count, x.Skip
	case *logical.Distinct:
		b.distinct = true
	default:
		return unsupported("%T in a single-table query", n)
	}
	return nil
}

func (b *sqlBuilder) arg(v any) string {
	b.args = append(b.args, v)
	return "?"
}

func (b *sqlBuilder) expr(e logical.Expr, typ string) (string, error) {
	switch x := e.(type) {
	case *logical.Column:
		c, ok := b.cols[x.Key]
		if !ok {
			return "", unsupported("unknown column %s", x.Key)
		}
		return c, nil
	case *logical.Lit:
		return b.arg(typed(x.Val, typ)), nil
	case *logical.Bin:
		if x.Op == "IN" {
			l, err := b.expr(x.L, "")
			if err != nil {
				return "", err
			}
			list, ok := x.R.(*logical.List)
			if !ok {
				return "", unsupported("IN without a list")
			}
			t := litType(x.L, b.own, b.types)
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
		l, err := b.expr(x.L, litType(x.R, b.own, b.types))
		if err != nil {
			return "", err
		}
		r, err := b.expr(x.R, litType(x.L, b.own, b.types))
		if err != nil {
			return "", err
		}
		return "(" + l + " " + x.Op + " " + r + ")", nil
	case *logical.Un:
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
	return "", unsupported("SQL expression %T", e)
}
