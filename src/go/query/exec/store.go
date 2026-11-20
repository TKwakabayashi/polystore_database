package exec

// store.go: issuing store queries by input ids and joining their results to the input rows.

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
)

// resultSet is a store's result. Native results stay as tuples ordered by column position, without a Row (map)
// per result (values are written straight into output rows when joining). Leaf results become rows.
type resultSet struct {
	cols   []string
	pos    map[string]int
	tuples [][]any
	rows   []Row
}

func newTupleSet(cols []string, tuples [][]any) *resultSet {
	pos := make(map[string]int, len(cols))
	for i, c := range cols {
		pos[c] = i
	}
	return &resultSet{cols: cols, pos: pos, tuples: tuples}
}

func (s *resultSet) len() int {
	if s.rows != nil {
		return len(s.rows)
	}
	return len(s.tuples)
}

// get returns column c of row i (nil if absent).
func (s *resultSet) get(i int, c string) any {
	if s.rows != nil {
		return s.rows[i][c]
	}
	if j, ok := s.pos[c]; ok {
		return s.tuples[i][j]
	}
	return nil
}

// mergeInto writes every column of row i into nr.
func (s *resultSet) mergeInto(i int, nr Row) {
	if s.rows != nil {
		for k, v := range s.rows[i] {
			nr[k] = v
		}
		return
	}
	for j, c := range s.cols {
		nr[c] = s.tuples[i][j]
	}
}

// width estimates the columns per row (to size output maps up front).
func (s *resultSet) width() int {
	if s.rows != nil {
		if len(s.rows) > 0 {
			return len(s.rows[0])
		}
		return 0
	}
	return len(s.cols)
}

// weight returns row i's weight (column _w; 1 when absent).
func (s *resultSet) weight(i int) int64 {
	if w, ok := s.get(i, eval.WeightCol).(int64); ok && w > 0 {
		return w
	}
	return 1
}

// toRows turns the result into rows (when passed on without a join).
func (s *resultSet) toRows() []Row {
	if s.rows != nil {
		return s.rows
	}
	out := make([]Row, len(s.tuples))
	for i, t := range s.tuples {
		r := make(Row, len(s.cols))
		for j, c := range s.cols {
			r[c] = t[j]
		}
		out[i] = r
	}
	return out
}

// storeQuery issues the native query and joins it to the input rows on the id columns; round trips count for owner.
func (e *Executor) storeQuery(owner physical.Node, q *physical.StoreQuery, in []Row) ([]Row, error) {
	issue := func(keys []Key) (*resultSet, error) {
		nq, err := e.lower(q)
		if err != nil {
			return nil, err
		}
		tuples, err := e.run(owner, nq, keys)
		if err != nil {
			return nil, err
		}
		return newTupleSet(nq.Columns, tuples), nil
	}
	if q.Leaf() {
		rs, err := issue(nil)
		if err != nil {
			return nil, err
		}
		return rs.toRows(), nil
	}
	if q.ShipMode == ir.ShipHand {
		return e.handShip(q, in)
	}
	if q.ShipMode != ir.ShipNone {
		return e.shipQuery(q, in)
	}

	// Collect the distinct input id tuples. A key is the uuids' raw bytes (UUIDKey); lookups with m[string(buf)]
	// do not copy, so rows allocate only for new tuples.
	var keys []Key
	seen := map[string]struct{}{}
	var buf []byte
	for _, r := range in {
		var ok bool
		if buf, ok = eval.UUIDKey(buf, r, q.Keys); !ok {
			continue
		}
		if _, dup := seen[string(buf)]; dup {
			continue
		}
		seen[string(buf)] = struct{}{}
		k := make(Key, len(q.Keys))
		for i, c := range q.Keys {
			k[i] = r[c].(id.UUID)
		}
		keys = append(keys, k)
	}
	results := &resultSet{rows: []Row{}}
	if len(keys) > 0 {
		var err error
		results, err = issue(keys)
		if err != nil {
			return nil, err
		}
	}
	if q.Replace {
		return results.toRows(), nil
	}
	// Index the results by id tuple (row numbers, no Row per result).
	byKey := make(map[string][]int, results.len())
	for i := 0; i < results.len(); i++ {
		buf = buf[:0]
		ok := true
		for _, c := range q.Keys {
			u, isID := results.get(i, c).(id.UUID)
			if !isID || u.Empty() {
				ok = false
				break
			}
			buf = append(buf, u[:]...)
		}
		if ok {
			byKey[string(buf)] = append(byKey[string(buf)], i)
		}
	}

	// Join per input row (input duplicates kept). A value fetch is a left outer join (null when absent).
	fetch := q.IsFetch()
	var out []Row
	for _, r := range in {
		var ok bool
		buf, ok = eval.UUIDKey(buf, r, q.Keys)
		var matches []int
		if ok {
			matches = byKey[string(buf)]
		}
		if !ok || len(matches) == 0 {
			if fetch {
				nr := r.Clone()
				for _, c := range q.Outputs {
					if _, has := nr[c]; !has {
						nr[c] = nil
					}
				}
				out = append(out, nr)
			}
			continue
		}
		for _, m := range matches {
			nr := make(Row, len(r)+results.width()) // room for the result's columns without regrowing the map
			for k, v := range r {
				nr[k] = v
			}
			results.mergeInto(m, nr)
			eval.SetWeight(nr, eval.WeightOf(r)*results.weight(m)) // input weight x result weight
			out = append(out, nr)
		}
	}
	return out, nil
}

// run issues the input in chunks, up to ExecOptions.StoreFetchParallelism at a time, and orders results by
// chunk (so every model's output is deterministic). It returns decoded tuples (order of q.Columns; ids as
// id.UUID), decoding rows as they arrive (no raw result is kept).
func (e *Executor) run(owner physical.Node, q *stores.Query, keys []Key) ([][]any, error) {
	size := e.chunkSize(q.Store)
	issue := func(ctx context.Context, part []Key) ([][]any, error) {
		var out [][]any
		err := e.stream(ctx, owner, q, part, func(vals []any) error {
			out = append(out, vals)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	if keys == nil || size <= 0 || len(keys) <= size {
		return issue(e.ctx, keys)
	}
	var parts [][]Key
	for i := 0; i < len(keys); i += size {
		j := i + size
		if j > len(keys) {
			j = len(keys)
		}
		parts = append(parts, keys[i:j])
	}
	results := make([][][]any, len(parts))
	g, ctx := errgroup.WithContext(e.ctx)
	par := e.opt.StoreFetchParallelism
	if q.Store == kind.Neo4j {
		par = e.opt.GraphFetchParallelism
	}
	if par < 1 {
		par = 1
	}
	g.SetLimit(par)
	for i, part := range parts {
		i, part := i, part
		g.Go(func() error {
			rows, err := issue(ctx, part)
			results[i] = rows
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	var out [][]any
	for _, r := range results {
		out = append(out, r...)
	}
	return out, nil
}

// stream issues q once and passes arriving rows to fn (nothing kept); round trips count for owner.
// A backend that is not a Streamer passes Run's result row by row.
func (e *Executor) stream(ctx context.Context, owner physical.Node, q *stores.Query, keys []Key, fn func(vals []any) error) error {
	e.trip(owner, q.Store)
	each := func(vals []any) error {
		if len(vals) != len(q.Columns) {
			return fmt.Errorf("executor: %s returned %d columns, want %d", q.Store, len(vals), len(q.Columns))
		}
		return fn(vals)
	}
	if r, ok := e.be.(Reporter); ok && (e.opt.Breakdown || e.opt.StoreProfile) {
		rep, err := r.StreamReport(ctx, q, keys, each, e.opt.StoreProfile)
		e.report(owner, rep)
		return err
	}
	if s, ok := e.be.(Streamer); ok {
		return s.Stream(ctx, q, keys, each)
	}
	vals, err := e.be.Run(ctx, q, keys)
	if err != nil {
		return err
	}
	for _, r := range vals {
		if err := each(r); err != nil {
			return err
		}
	}
	return nil
}

// toRows turns result tuples into rows keyed by column name.
func toRows(q *stores.Query, tuples [][]any) ([]Row, error) {
	out := make([]Row, 0, len(tuples))
	for _, vals := range tuples {
		if len(vals) != len(q.Columns) {
			return nil, fmt.Errorf("executor: %s returned %d columns, want %d", q.Store, len(vals), len(q.Columns))
		}
		r := make(Row, len(vals))
		for i, c := range q.Columns {
			r[c] = vals[i]
		}
		out = append(out, r)
	}
	return out, nil
}
