package exec

// metrics.go: per-operator measurements (time, rows, round trips, store call breakdown).

import (
	"fmt"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"strings"
	"time"
)

type stepAcc struct {
	op, desc, store string
	dur             time.Duration
	in, out         int
	expanded        int64

	firstRow, receive, decode time.Duration
	profiles                  []result.StoreProfile
}

// Steps returns the per-operator metrics of the last run (leaf to root; summed over batches in batched models).
func (e *Executor) Steps() []result.StepMetric {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []result.StepMetric
	for _, n := range e.order {
		if a, ok := e.acc[n]; ok {
			out = append(out, result.StepMetric{Step: len(out) + 1, Op: a.op, Desc: a.desc, Store: a.store, Duration: a.dur,
				InRows: a.in, OutRows: a.out, OutRowsExpanded: a.expanded, RoundTrips: e.trips[n],
				FirstRow: a.firstRow, Receive: a.receive, Decode: a.decode, Profiles: a.profiles})
		}
	}
	return out
}

// RoundTrips returns the number of store queries of the last run.
func (e *Executor) RoundTrips() int64 { return e.roundTrips.Load() }

// StoreTrips returns RoundTrips per store.
func (e *Executor) StoreTrips() map[string]int64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make(map[string]int64, len(e.byStore))
	for k, n := range e.byStore {
		out[k.String()] = n
	}
	return out
}

// trip counts one store query; owner is the physical operator that issued it.
func (e *Executor) trip(owner physical.Node, k kind.Kind) {
	e.roundTrips.Add(1)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.trips == nil {
		e.trips, e.byStore = map[physical.Node]int64{}, map[kind.Kind]int64{}
	}
	e.trips[owner]++
	e.byStore[k]++
}

// reset clears the metrics and compiled queries before a run.
func (e *Executor) reset(root physical.Node) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.roundTrips.Store(0)
	e.trips = map[physical.Node]int64{}
	e.byStore = map[kind.Kind]int64{}
	e.acc = map[physical.Node]*stepAcc{}
	e.lowered = map[*physical.StoreQuery]*stores.Query{}
	e.lists = map[*physical.ListFetch][]*stores.Query{}
	e.args = map[*physical.ArgRows][]Row{}
	e.sets = map[string]*ir.IDSet{}
	e.boundOps = map[*physical.Coord]ir.Node{}
	e.memos = map[*physical.StepExpand]*stepMemo{}
	e.order = nil
	var walk func(physical.Node)
	walk = func(n physical.Node) {
		for _, c := range n.Children() {
			walk(c)
		}
		e.order = append(e.order, n)
	}
	if root != nil {
		walk(root)
	}
}

// record adds one run of operator n (one batch, or everything in bulk) to the metrics.
func (e *Executor) record(n physical.Node, dur time.Duration, in int, out []Row) {
	var expanded int64
	for _, r := range out {
		expanded += eval.WeightOf(r)
	}
	e.recordCount(n, dur, in, len(out), expanded)
}

// recordCount is record without rows (adds counts taken while streaming).
func (e *Executor) recordCount(n physical.Node, dur time.Duration, in, out int, expanded int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	a, ok := e.acc[n]
	if !ok {
		a = &stepAcc{op: opName(n), desc: n.Describe(), store: storeName(n)}
		e.acc[n] = a
	}
	a.dur += dur
	a.in += in
	a.out += out
	a.expanded += expanded
}

// report adds one store call's measurements to operator n.
func (e *Executor) report(n physical.Node, r stores.Report) {
	if n == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	a, ok := e.acc[n]
	if !ok {
		a = &stepAcc{op: opName(n), desc: n.Describe(), store: storeName(n)}
		e.acc[n] = a
	}
	a.firstRow += r.FirstRow
	a.receive += r.Receive
	a.decode += r.Decode
	if r.Profile != nil {
		a.profiles = append(a.profiles, *r.Profile)
	}
}

// storeName returns the store an operator queries (empty for coordinator operators).
func storeName(n physical.Node) string {
	switch x := n.(type) {
	case *physical.StoreQuery:
		return x.Store.String()
	case *physical.ListFetch:
		return x.Store.String()
	case *physical.StepExpand:
		return x.Hop.Store.String()
	}
	return ""
}

func opName(n physical.Node) string {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if x.IsFetch() {
			return "Fetch[" + x.Store.String() + "]"
		}
		return "StoreQuery[" + x.Store.String() + "]"
	case *physical.ListFetch:
		return "ListFetch[" + x.Store.String() + "]"
	case *physical.Coord:
		return strings.TrimPrefix(fmt.Sprintf("%T", x.Op), "*ir.")
	}
	return strings.TrimPrefix(fmt.Sprintf("%T", n), "*physical.") // ArgRows / OptionalApply
}
