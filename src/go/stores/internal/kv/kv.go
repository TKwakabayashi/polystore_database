// Package kv compiles store queries to key-value fetch plans and runs them on LevelDB or RocksDB.
package kv

import (
	"bytes"
	"context"
	"fmt"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/stores/internal/lower"
	"slices"
	"strings"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"

	"golang.org/x/sync/errgroup"
)

func lowerKVS(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	ops := lower.Chain(q.Ops)
	if ops == nil {
		return nil, lower.Unsupported("kvs query with a join")
	}
	own := lower.Owners(q.Ops)
	k := &stores.KVSQuery{Types: map[string]string{}}
	for _, op := range ops {
		switch x := op.(type) {
		case *ir.Boundary:
			if len(x.Cols) != 1 {
				return nil, lower.Unsupported("kvs query keyed by %v", x.Cols)
			}
			o := own[x.Cols[0]]
			if len(o.Labels) != 1 {
				return nil, lower.Unsupported("kvs label of %s", x.Cols[0])
			}
			k.Label, k.Alias = o.Labels[0], x.Cols[0]
		case *ir.EntityScan:
			if len(x.Labels) != 1 {
				return nil, lower.Unsupported("kvs scan needs one label")
			}
			k.Label, k.Alias, k.Scan = x.Labels[0], x.Alias, true
		case *ir.GetProperty:
			if k.Label == "" {
				if len(x.Labels) != 1 {
					return nil, lower.Unsupported("kvs label of %s", x.Alias)
				}
				k.Label, k.Alias = x.Labels[0], x.Alias
			}
			if x.Alias != k.Alias {
				return nil, lower.Unsupported("property of %s in a kvs query on %s", x.Alias, k.Alias)
			}
			k.Props = lower.AppendNew(k.Props, x.Prop)
			k.Types[x.Prop] = types(x.Kind, x.Labels, x.Prop)
		case *ir.Filter:
			k.Filters = append(k.Filters, x.Pred)
			if k.Scan && k.PrimaryEq == nil {
				if eq := primaryEq(x.Pred, own, types); eq != nil {
					k.PrimaryEq = eq
				}
			}
		default:
			return nil, lower.Unsupported("%T in a kvs query", op)
		}
	}
	kinds := make([]stores.ColKind, len(q.Outputs))
	for i, col := range q.Outputs {
		if col == k.Alias {
			kinds[i] = stores.KindIdent
			continue
		}
		if !lower.Contains(k.Props, lower.PropOf(col)) {
			return nil, lower.Unsupported("kvs output %s not fetched", col)
		}
	}
	return &stores.Query{Store: q.Store, Columns: q.Outputs, ColKinds: kinds, InKeys: q.Keys, Body: k}, nil
}

// primaryEq extracts `a.p = literal` as a condition on the reverse index.
func primaryEq(e ir.Expr, own map[string]lower.Owner, types stores.TypeFn) *stores.KVSEq {
	bin, ok := e.(*ir.Bin)
	if !ok || bin.Op != "=" {
		return nil
	}
	c, ok := bin.L.(*ir.Column)
	if !ok {
		return nil
	}
	lit, ok := bin.R.(*ir.Lit)
	if !ok {
		return nil
	}
	return &stores.KVSEq{Prop: lower.PropOf(c.Key), Value: lower.Typed(lit.Val, lower.ColType(c.Key, own, types))}
}

func streamKVS(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	kv, ok := reg.KV(q.Store)
	if !ok {
		return fmt.Errorf("executor: kvs not connected")
	}
	k := q.KVPlan()
	var ids []id.UUID
	if k.Scan {
		var err error
		if ids, err = scanKVS(ctx, kv, k); err != nil {
			return err
		}
	} else {
		for _, key := range in {
			ids = append(ids, key[0])
		}
	}
	return fetchKVS(ctx, kv, q, k, ids, stores.KVFetchFrom(ctx), emit)
}

// fetchKVS reads the properties of ids in batches of f.BatchKeys (one MultiGet each when the store has one),
// f.Parallelism batches at a time, and emits the rows that pass the filters in the order of ids.
func fetchKVS(ctx context.Context, kv conn.KV, q *stores.Query, k *stores.KVSQuery, ids []id.UUID, f stores.KVFetch, emit func([]any) error) error {
	size, multi := f.BatchKeys, f.BatchKeys > 0
	if size < 1 {
		size = 1024 // key by key; batches only bound cancellation checks and parallel work
	}
	var batches [][]id.UUID
	for i := 0; i < len(ids); i += size {
		batches = append(batches, ids[i:min(i+size, len(ids))])
	}
	par := max(f.Parallelism, 1)
	for w := 0; w < len(batches); w += par {
		win := batches[w:min(w+par, len(batches))]
		rows := make([][][]any, len(win))
		if len(win) == 1 {
			r, err := fetchBatch(ctx, kv, q, k, win[0], multi)
			if err != nil {
				return err
			}
			rows[0] = r
		} else {
			g, gctx := errgroup.WithContext(ctx)
			for i, b := range win {
				g.Go(func() error {
					r, err := fetchBatch(gctx, kv, q, k, b, multi)
					rows[i] = r
					return err
				})
			}
			if err := g.Wait(); err != nil {
				return err
			}
		}
		for _, rs := range rows {
			for _, vals := range rs {
				if err := emit(vals); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// fetchBatch reads the properties of ids (with one GetMany when multi, otherwise key by key) and returns the
// output rows that pass the filters.
func fetchBatch(ctx context.Context, kv conn.KV, q *stores.Query, k *stores.KVSQuery, ids []id.UUID, multi bool) ([][]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Keys go in ascending property order, so that ascending ids make the whole batch sorted (RocksDB then skips its sort).
	np := len(k.Props)
	order := make([]int, np)
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int { return strings.Compare(k.Props[a], k.Props[b]) })
	slot := make([]int, np) // slot[pi]: position of property pi among an entity's keys
	for pos, pi := range order {
		slot[pi] = pos
	}
	keys := make([][]byte, 0, len(ids)*np)
	for _, u := range ids {
		for _, pi := range order {
			keys = append(keys, codec.BuildEntityKey(k.Label, u, k.Props[pi]))
		}
	}
	var raws [][]byte
	var found []bool
	if multi {
		sorted := slices.IsSortedFunc(keys, bytes.Compare)
		var err error
		if raws, found, err = conn.GetMany(kv, keys, sorted); err != nil {
			return nil, err
		}
	} else {
		raws, found = make([][]byte, len(keys)), make([]bool, len(keys))
		for i, key := range keys {
			v, ok, err := kv.Get(key)
			if err != nil {
				return nil, err
			}
			raws[i], found[i] = v, ok
		}
	}
	var out [][]any
	for j, u := range ids {
		r := eval.Row{k.Alias: u}
		for pi, p := range k.Props {
			raw, ok := raws[j*np+slot[pi]], found[j*np+slot[pi]]
			if !ok {
				r[k.Alias+"."+p] = nil
				continue
			}
			// A present key with an empty value is the empty string (DecodeValue turns empty into nil, so it is handled first).
			// Values are decoded to their declared type before the filters run (so datetimes are not compared as strings).
			var v any
			if len(raw) == 0 && isStringType(k.Types[p]) {
				v = ""
			} else {
				v = codec.DecodeValue(raw, k.Types[p])
			}
			r[k.Alias+"."+p] = eval.Decode(v, stores.KindValue, k.Types[p])
		}
		keep := true
		for _, f := range k.Filters {
			v, err := eval.Eval(f, r)
			if err != nil {
				return nil, err
			}
			if !eval.Truthy(v) {
				keep = false
				break
			}
		}
		if !keep {
			continue
		}
		vals := make([]any, len(q.Columns))
		for i, c := range q.Columns {
			if c == k.Alias {
				vals[i] = u.String()
			} else {
				vals[i] = r[c]
			}
		}
		out = append(out, vals)
	}
	return out, nil
}

func isStringType(t string) bool {
	switch strings.ToLower(t) {
	case "", "string", "text":
		return true
	}
	return false
}

// scanKVS collects the uuids of a label (a prefix scan of the reverse index when there is an equality).
func scanKVS(ctx context.Context, kv conn.KV, k *stores.KVSQuery) ([]id.UUID, error) {
	var prefix []byte
	index := false
	if eq := k.PrimaryEq; eq != nil {
		var vb []byte
		switch v := eq.Value.(type) {
		case int64:
			vb = codec.EncodeInt(v)
		case string:
			vb = []byte(v)
		}
		if vb != nil {
			index = true
			prefix = []byte("index" + codec.Sep + k.Label + codec.Sep + eq.Prop + codec.Sep)
			prefix = append(prefix, vb...)
			prefix = append(prefix, []byte(codec.Sep)...)
		}
	}
	if !index {
		prefix = []byte(k.Label + codec.Sep)
	}
	seen := map[id.UUID]bool{}
	var out []id.UUID
	n := 0
	err := conn.ScanPrefix(kv, prefix, func(key, _ []byte) error {
		if n++; n%4096 == 0 && ctx.Err() != nil {
			return ctx.Err()
		}
		var u id.UUID
		if index {
			u = codec.IndexKeyUUID(key)
		} else if _, ku, _, ok := codec.ParseEntityKey(key); ok {
			u = ku
		}
		if !u.Empty() && !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
		return nil
	})
	return out, err
}

// DB is a key-value database (LevelDB or RocksDB behind conn.KV). It cannot join a shipped table.
type DB struct {
	K kind.Kind
}

var exprRules = stores.ExprRules{IsNull: true}

func (d DB) Kind() kind.Kind         { return d.K }
func (DB) CanExpress(e ir.Expr) bool { return exprRules.Expressible(e) }
func (DB) CanAbsorb(ir.Node) bool    { return false }
func (DB) Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return lowerKVS(q, types)
}
func (DB) Stream(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key, emit func([]any) error) error {
	return streamKVS(ctx, reg, q, in, emit)
}
