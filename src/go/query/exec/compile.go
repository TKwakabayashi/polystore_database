package exec

// compile.go: physical store queries and list fetches to native queries (cached per run).

import (
	"fmt"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/dbs"
)

// lower returns a StoreQuery's native query (compiled once per run).
func (e *Executor) lower(q *physical.StoreQuery) (*stores.Query, error) {
	e.mu.Lock()
	nq, ok := e.lowered[q]
	e.mu.Unlock()
	if ok {
		return nq, nil
	}
	cq := q
	if readsSets(q.Ops) {
		ops, err := e.bindNode(q.Ops)
		if err != nil {
			return nil, err
		}
		cp := *q
		cp.Ops = ops
		cq = &cp
	}
	nq, err := compileStoreQuery(cq, e.types, e.compile)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.lowered[q] = nq
	e.mu.Unlock()
	return nq, nil
}

func (e *Executor) lowerList(f *physical.ListFetch) ([]*stores.Query, error) {
	e.mu.Lock()
	qs, ok := e.lists[f]
	e.mu.Unlock()
	if ok {
		return qs, nil
	}
	qs, err := compileList(f, e.types, e.compile)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.lists[f] = qs
	e.mu.Unlock()
	return qs, nil
}

// CompileStoreQuery compiles a physical store query, resolving the hand-built plan fields first.
func CompileStoreQuery(q *physical.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	return compileStoreQuery(q, types, dbs.Compile)
}

func compileStoreQuery(q *physical.StoreQuery, types stores.TypeFn, compile compileFn) (*stores.Query, error) {
	if h, ok := q.Prelowered.(*stores.Query); ok {
		return h, nil
	}
	if q.HandScan {
		leaf := q.StoreQuery
		leaf.ShipMode = ir.ShipNone
		return compile(&leaf, types)
	}
	out, err := compile(&q.StoreQuery, types)
	if err != nil {
		return nil, err
	}
	if q.TextRewrite != nil {
		out.TextBody().Text = q.TextRewrite(out.TextBody().Text)
	}
	return out, nil
}

// CompileList compiles a list fetch into one value query per element label (columns c0=uuid, c1=value).
func CompileList(f *physical.ListFetch, types stores.TypeFn) ([]*stores.Query, error) {
	return compileList(f, types, dbs.Compile)
}

func compileList(f *physical.ListFetch, types stores.TypeFn, compile compileFn) ([]*stores.Query, error) {
	kind := ir.NodeObj
	if f.Part == ir.SegRels {
		kind = ir.RelObj
	}
	var out []*stores.Query
	for _, label := range f.Labels {
		g := &ir.GetProperty{Input: &ir.Boundary{Cols: []string{"_e"}, Labels: map[string][]string{"_e": {label}}}, Alias: "_e", Prop: f.Prop, Kind: kind, Labels: []string{label}}
		sq := &ir.StoreQuery{Store: f.Store, Ops: g, Keys: []string{"_e"}, Outputs: []string{"_e", g.Key()}}
		q, err := compile(sq, types)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: list fetch without element labels", ast.ErrUnsupported)
	}
	return out, nil
}
