package exec

// ship.go: store queries that receive their input as a table (shipped, emulated or hand-built).

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/dbs"
	"slices"
)

// shipQuery issues a shipped query. ShipAggregate: the input rows become a table of (Keys, Ship) tuples with
// summed weights, sent ShipChunkRows rows at a time; the stores' partially aggregated rows are concatenated
// (not joined back; the final aggregation above combines groups that span chunks). Rows with null ids are
// sent too (count(*) counts rows without values; stores treat them as null in a left outer join).
func (e *Executor) shipQuery(q *physical.StoreQuery, in []Row) ([]Row, error) {
	if q.ShipMode != ir.ShipAggregate {
		return nil, fmt.Errorf("executor: ship mode %s is not supported yet", q.ShipMode)
	}
	cols := append(append([]string{}, q.Keys...), q.Ship...)
	index := map[string]int{}
	var rows [][]any
	for _, r := range in {
		vals := make([]any, len(cols)+1)
		for i, c := range cols {
			if u, ok := r[c].(id.UUID); ok && !u.Empty() {
				vals[i] = u
			}
		}
		k := eval.KeyString(vals[:len(cols)]...)
		if i, ok := index[k]; ok {
			rows[i][len(cols)] = rows[i][len(cols)].(int64) + eval.WeightOf(r)
			continue
		}
		vals[len(cols)] = eval.WeightOf(r)
		index[k] = len(rows)
		rows = append(rows, vals)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	cols = append(cols, eval.WeightCol)

	var issue func(ctx context.Context, part [][]any) ([]Row, error)
	emulated := dbs.ShipEmulated(q.Store)
	if emulated {
		inner, agg, ok := q.ShipInner()
		if !ok {
			return nil, fmt.Errorf("executor: shipped query on %s without an aggregation", q.Store)
		}
		pinner := &physical.StoreQuery{StoreQuery: *inner, Input: q.Input} // Input marks it as a non-leaf query
		issue = func(_ context.Context, part [][]any) ([]Row, error) { return e.emulateShip(q, pinner, agg, cols, part) }
	} else {
		sh, ok := e.be.(Shipper)
		if !ok {
			return nil, fmt.Errorf("executor: backend cannot ship rows to %s", q.Store)
		}
		nq, err := e.lower(q)
		if err != nil {
			return nil, err
		}
		pos, err := shipPositions(nq.ShipCols, cols)
		if err != nil {
			return nil, err
		}
		issue = func(ctx context.Context, part [][]any) ([]Row, error) {
			ordered := make([][]any, len(part))
			for i, r := range part {
				o := make([]any, len(pos))
				for j, p := range pos {
					o[j] = r[p]
				}
				ordered[i] = o
			}
			raw, err := sh.RunShipped(ctx, nq, Shipment{Rows: ordered})
			if err != nil {
				return nil, err
			}
			return toRows(nq, raw)
		}
	}

	size := e.opt.ShipChunkRows
	if size <= 0 {
		size = len(rows)
	}
	var parts [][][]any
	for i := 0; i < len(rows); i += size {
		parts = append(parts, rows[i:min(i+size, len(rows))])
	}
	results := make([][]Row, len(parts))
	g, ctx := errgroup.WithContext(e.ctx)
	g.SetLimit(max(e.opt.StoreFetchParallelism, 1))
	for i, part := range parts {
		i, part := i, part
		g.Go(func() error {
			if !emulated { // in emulation the inner query counts its own round trips
				e.trip(q, q.Store)
			}
			rs, err := issue(ctx, part)
			results[i] = rs
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	var out []Row
	for _, rs := range results {
		out = append(out, rs...)
	}
	return out, nil
}

// handShip sends the input rows as they are (duplicates too) to a hand-built query and passes on only the rows the store returns.
func (e *Executor) handShip(q *physical.StoreQuery, in []Row) ([]Row, error) {
	if q.HandScan {
		leaf := *q
		leaf.Input, leaf.ShipMode, leaf.HandScan = nil, ir.ShipNone, false
		rows, err := e.storeQuery(q, &leaf, nil)
		if err != nil {
			return nil, err
		}
		return append(slices.Clip(in), rows...), nil
	}
	nq, err := e.lower(q)
	if err != nil {
		return nil, err
	}
	if q.Store == kind.Neo4j {
		return e.handGraph(q, nq, in)
	}
	sh, ok := e.be.(Shipper)
	if !ok {
		return nil, fmt.Errorf("executor: backend cannot ship rows to %s", q.Store)
	}
	if q.HandGroup != "" {
		in = slices.Clone(in)
		slices.SortFunc(in, func(a, b Row) int { return id.Compare(a[q.HandGroup].(id.UUID), b[q.HandGroup].(id.UUID)) })
	}
	var parts [][][]any
	var part [][]any
	for i, r := range in {
		vals := make([]any, len(nq.ShipCols))
		for j, c := range nq.ShipCols {
			vals[j] = r[c]
		}
		part = append(part, vals)
		next := i+1 == len(in)
		if !next && q.HandChunk > 0 && len(part) >= q.HandChunk {
			next = q.HandGroup == "" || in[i+1][q.HandGroup] != r[q.HandGroup]
		}
		if next {
			parts = append(parts, part)
			part = nil
		}
	}
	results := make([][]Row, len(parts))
	g, ctx := errgroup.WithContext(e.ctx)
	g.SetLimit(max(e.opt.StoreFetchParallelism, 1))
	for i, part := range parts {
		i, part := i, part
		g.Go(func() error {
			e.trip(q, q.Store)
			raw, err := sh.RunShipped(ctx, nq, Shipment{Rows: part})
			if err != nil {
				return err
			}
			results[i], err = toRows(nq, raw)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	var out []Row
	for _, rs := range results {
		out = append(out, rs...)
	}
	return out, nil
}

// handGraph issues a hand-built traversal. The distinct values of each input id column go as parameter
// `<column>Ids`; results join the input rows on the start column Keys[0] (to carry the start's values).
func (e *Executor) handGraph(q *physical.StoreQuery, nq *stores.Query, in []Row) ([]Row, error) {
	params := map[string]any{}
	bySrc := map[id.UUID]Row{}
	for _, c := range nq.ShipCols {
		seen := map[id.UUID]bool{}
		list := []any{}
		for _, r := range in {
			if u, ok := r[c].(id.UUID); ok && !u.Empty() && !seen[u] {
				seen[u] = true
				list = append(list, id.Stored(u))
				if c == q.Keys[0] {
					bySrc[u] = r
				}
			}
		}
		params[c+"Ids"] = list
	}
	cp := *nq
	body := *nq.TextBody()
	body.Params = params
	cp.Body = &body
	e.trip(q, q.Store)
	raw, err := e.be.Run(e.ctx, &cp, nil)
	if err != nil {
		return nil, err
	}
	rows, err := toRows(nq, raw)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if u, ok := r[q.Keys[0]].(id.UUID); ok {
			for k, v := range bySrc[u] {
				if _, has := r[k]; !has && v != nil { // in columnar batches missing columns are present as nil
					r[k] = v
				}
			}
		}
	}
	return rows, nil
}

// emulateShip emulates a table join for stores without one (ClickHouse, Cassandra, kvs): the table rows are
// input rows for the query under the aggregation, issued with the id boundary (IN on the keys and a join on
// them; a left outer join for value fetches), and the coordinator aggregates the joined rows (with shipped
// columns and weights). The store's work equals a value fetch; the difference is that each table chunk is aggregated.
func (e *Executor) emulateShip(owner physical.Node, inner *physical.StoreQuery, agg *ir.Aggregation, cols []string, part [][]any) ([]Row, error) {
	in := make([]Row, len(part))
	for i, vals := range part {
		r := Row{}
		for j, c := range cols {
			r[c] = vals[j]
		}
		eval.SetWeight(r, eval.WeightOf(r))
		in[i] = r
	}
	joined, err := e.storeQuery(owner, inner, in)
	if err != nil {
		return nil, err
	}
	return runCoord(ir.WithInput(agg, &ir.Boundary{}), [][]Row{joined})
}

// shipPositions maps the table column order a native query wants (want) to positions in the executor's columns (have).
func shipPositions(want, have []string) ([]int, error) {
	pos := make([]int, len(want))
	for i, w := range want {
		pos[i] = -1
		for j, h := range have {
			if h == w {
				pos[i] = j
			}
		}
		if pos[i] < 0 {
			return nil, fmt.Errorf("executor: shipped column %s is not provided", w)
		}
	}
	return pos, nil
}
