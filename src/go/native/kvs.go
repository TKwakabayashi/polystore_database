package native

import (
	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/store"
)

// kvs.go は kvs（LevelDB）向けの変換。KVS にクエリ言語は無いので、取得計画（どのキーを引き、
// どの条件をメモリで評価するか）を組む。等値条件が 1 つあれば逆引きインデックスで候補を絞る。

func lowerKVS(q *physical.StoreQuery, types TypeFn) (*Query, error) {
	ops := chain(q.Ops)
	if ops == nil {
		return nil, unsupported("kvs query with a join")
	}
	own := owners(q.Ops)
	k := &KVSQuery{Types: map[string]string{}}
	for _, op := range ops {
		switch x := op.(type) {
		case *physical.Argument:
			if len(x.Cols) != 1 {
				return nil, unsupported("kvs query keyed by %v", x.Cols)
			}
			o := own[x.Cols[0]]
			if len(o.labels) != 1 {
				return nil, unsupported("kvs label of %s", x.Cols[0])
			}
			k.Label, k.Alias = o.labels[0], x.Cols[0]
		case *logical.EntityScan:
			if len(x.Labels) != 1 {
				return nil, unsupported("kvs scan needs one label")
			}
			k.Label, k.Alias, k.Scan = x.Labels[0], x.Alias, true
		case *logical.GetProperty:
			if k.Label == "" {
				if len(x.Labels) != 1 {
					return nil, unsupported("kvs label of %s", x.Alias)
				}
				k.Label, k.Alias = x.Labels[0], x.Alias
			}
			if x.Alias != k.Alias {
				return nil, unsupported("property of %s in a kvs query on %s", x.Alias, k.Alias)
			}
			k.Props = appendNew(k.Props, x.Prop)
			k.Types[x.Prop] = types(x.Kind, x.Labels, x.Prop)
		case *logical.Filter:
			k.Filters = append(k.Filters, x.Pred)
			if k.Scan && k.PrimaryEq == nil {
				if eq := primaryEq(x.Pred, own, types); eq != nil {
					k.PrimaryEq = eq
				}
			}
		default:
			return nil, unsupported("%T in a kvs query", op)
		}
	}
	kinds := make([]ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		if col == k.Alias {
			kinds[i] = KindIdent
			continue
		}
		if !contains(k.Props, propOf(col)) {
			return nil, unsupported("kvs output %s not fetched", col)
		}
	}
	return &Query{Store: store.Kvs, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys, KVS: k, InArg: -1}, nil
}

// primaryEq は `a.p = リテラル` を逆引きインデックスの条件として取り出す。
func primaryEq(e logical.Expr, own map[string]owner, types TypeFn) *KVSEq {
	bin, ok := e.(*logical.Bin)
	if !ok || bin.Op != "=" {
		return nil
	}
	c, ok := bin.L.(*logical.Column)
	if !ok {
		return nil
	}
	lit, ok := bin.R.(*logical.Lit)
	if !ok {
		return nil
	}
	return &KVSEq{Prop: propOf(c.Key), Value: typed(lit.Val, colType(c.Key, own, types))}
}

func propOf(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			return key[i+1:]
		}
	}
	return key
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
