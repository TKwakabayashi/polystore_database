package exec

// executor.go: the Executor and step, the meaning of every operator shared by all iteration models.

import (
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/dbs"
	"sync"
	"sync/atomic"
	"time"
)

// Executor evaluates physical plans. The iteration model (bulk / volcano / vectorized / stream / vecstream) is
// chosen by ExecOptions.Model. Store queries, joining to input rows, value fetches and coordinator operators
// mean the same in every model (storeQuery / listFetch / runCoord); only the iteration differs.
type Executor struct {
	opt        ExecOptions
	ctx        context.Context
	be         Backend
	types      stores.TypeFn
	roundTrips atomic.Int64

	mu      sync.Mutex
	trips   map[physical.Node]int64 // round trips per operator
	byStore map[kind.Kind]int64     // round trips per store
	acc     map[physical.Node]*stepAcc
	order   []physical.Node // order of the metrics (leaf to root)
	compile compileFn
	lowered map[*physical.StoreQuery]*stores.Query
	lists   map[*physical.ListFetch][]*stores.Query
	args    map[*physical.ArgRows][]Row // outer rows for leaves inside OPTIONAL

	sets     map[string]*ir.IDSet               // id sets computed by the set plans of a WithSets root
	boundOps map[*physical.Coord]ir.Node        // coordinator operators with their sets bound
	memos    map[*physical.StepExpand]*stepMemo // condition results per element id of each StepExpand
}

// New makes an Executor. types looks up property types for compiling native queries.
func New(ctx context.Context, be Backend, types stores.TypeFn, opt ExecOptions) *Executor {
	ctx = stores.WithKVFetch(ctx, stores.KVFetch{BatchKeys: opt.KVBatchKeys, Parallelism: opt.KVFetchParallelism})
	e := &Executor{opt: opt, ctx: ctx, be: be, types: types, compile: dbs.Compile}
	if c, ok := be.(Compiler); ok {
		e.compile = c.Compile
	}
	return e
}

// Execute evaluates the plan in bulk (materializing everything).
func (e *Executor) Execute(n physical.Node) ([]Row, error) {
	e.reset(n)
	n, err := e.unwrapSets(n)
	if err != nil {
		return nil, err
	}
	return e.bulk(n)
}

// step applies one physical operator to its input rows (the same meaning in every model). inputs are per child.
func (e *Executor) step(n physical.Node, inputs [][]Row) ([]Row, error) {
	start := time.Now()
	in := 0
	for _, rs := range inputs {
		in += len(rs)
	}
	var (
		out []Row
		err error
	)
	switch x := n.(type) {
	case *physical.StoreQuery:
		var rows []Row
		if len(inputs) > 0 {
			rows = inputs[0]
		}
		out, err = e.storeQuery(x, x, rows)
	case *physical.ListFetch:
		out, err = e.listFetch(x, inputs[0])
	case *physical.Coord:
		var op ir.Node
		if op, err = e.coordOp(x); err == nil {
			out, err = runCoord(op, inputs)
		}
	case *physical.StepExpand:
		out, err = e.stepExpand(x, inputs[0])
	case *physical.ArgRows:
		out = e.argRows(x)
	case *physical.OptionalApply:
		out = leftApply(inputs[0], inputs[1], x)
	default:
		err = fmt.Errorf("executor: physical node %T", n)
	}
	if err != nil {
		return nil, err
	}
	e.record(n, time.Since(start), in, out)
	return out, nil
}

// setArg passes the outer rows' id tuples (distinct, without nulls) to an inner leaf of OPTIONAL.
func (e *Executor) setArg(a *physical.ArgRows, outer []Row) {
	seen := map[string]bool{}
	var rows []Row
	for _, r := range outer {
		nr := Row{}
		vals := make([]any, len(a.Cols))
		ok := true
		for i, c := range a.Cols {
			v := r[c]
			if v == nil {
				ok = false
				break
			}
			nr[c], vals[i] = v, v
		}
		if !ok {
			continue
		}
		if k := eval.KeyString(vals...); !seen[k] {
			seen[k] = true
			rows = append(rows, nr)
		}
	}
	e.mu.Lock()
	e.args[a] = rows
	e.mu.Unlock()
}

func (e *Executor) argRows(a *physical.ArgRows) []Row {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Row, len(e.args[a]))
	for i, r := range e.args[a] {
		out[i] = r.Clone()
	}
	return out
}

// leftApply joins to each outer row the inner rows whose id columns Keys match. An outer row without a match
// stays once with New and the inner columns null (OPTIONAL MATCH). Weights multiply.
func leftApply(outer, inner []Row, o *physical.OptionalApply) []Row {
	var buf []byte
	key := func(r Row) bool {
		buf = buf[:0]
		for _, k := range o.Keys {
			if r[k] == nil {
				return false
			}
			buf = append(eval.AppendKey(buf, r[k]), 0)
		}
		return true
	}
	by := map[string][]Row{}
	for _, r := range inner {
		if key(r) {
			by[string(buf)] = append(by[string(buf)], r)
		}
	}
	var out []Row
	for _, r := range outer {
		ok := key(r)
		var matches []Row
		if ok {
			matches = by[string(buf)]
		}
		if !ok || len(matches) == 0 {
			nr := r.Clone()
			for _, c := range o.New {
				nr[c] = nil
			}
			for _, c := range o.InnerCols {
				nr[c] = nil
			}
			out = append(out, nr)
			continue
		}
		for _, m := range matches {
			nr := r.Clone()
			for c, v := range m {
				nr[c] = v
			}
			eval.SetWeight(nr, eval.WeightOf(r)*eval.WeightOf(m))
			out = append(out, nr)
		}
	}
	return out
}

// chunkSize is the number of ids per query (0: no chunking): min(wanted, physical limit).
func (e *Executor) chunkSize(k kind.Kind) int {
	lim := kind.LimitsOf(k)
	want := lim.DefaultChunk
	if e.opt.MaterializeChunkSize > 0 {
		want = e.opt.MaterializeChunkSize
	}
	if lim.MaxInList > 0 && (want <= 0 || want > lim.MaxInList) {
		want = lim.MaxInList
	}
	if want < 0 {
		return 0
	}
	return want
}
