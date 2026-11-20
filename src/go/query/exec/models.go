package exec

// models.go chooses the iteration model. The meaning (step) is shared; only the row flow differs:
//
//	bulk        materialize every operator from the leaves (bulk, here)
//	volcano     pull iterators over row batches ([]Row) of width 1 (one store round trip per input row; pull.go)
//	vectorized  like volcano with width ExecOptions.VectorSize (pull.go)
//	stream      push: a goroutine and channel per operator; store queries run on parallel workers (push.go)
//	vecstream   pull with an exchange (a driver pulling children serially and W workers); output rechunked to the width (pull.go)
//
// Operators that query stores (StoreQuery / ListFetch) and row-independent coordinator operators
// (Filter / Projection / ProjectPath / RelUniqueFilter) work per batch. Aggregation, sorting,
// deduplication and joins wait for all rows (pipeline breakers, blocking). Limit counts as rows flow (limit.go).

import (
	"fmt"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
)

// ExecuteModel evaluates a physical plan with the iteration model ExecOptions.Model.
func (e *Executor) ExecuteModel(n physical.Node) ([]Row, error) {
	return e.ExecuteWithArgs(n, nil)
}

// ExecuteWithArgs runs a plan with rows given to its ArgRows leaves (for opbench, which measures single
// operators; in ordinary plans ArgRows appear only inside OPTIONAL and are filled with outer rows).
func (e *Executor) ExecuteWithArgs(n physical.Node, args map[*physical.ArgRows][]Row) ([]Row, error) {
	e.reset(n)
	e.mu.Lock()
	for a, rows := range args {
		e.args[a] = rows
	}
	e.mu.Unlock()
	n, err := e.unwrapSets(n)
	if err != nil {
		return nil, err
	}
	switch model := e.opt.Model; model {
	case ModelBulk, "":
		return e.bulk(n)
	case ModelVolcano:
		return e.pull(n, 1, 0)
	case ModelVectorized:
		return e.pull(n, e.opt.VectorWidth(), 0)
	case ModelVecStream:
		return e.pull(n, e.opt.VectorWidth(), exchangeWorkers)
	case ModelStream:
		return e.push(n, e.opt.VectorWidth())
	default:
		return nil, fmt.Errorf("executor: unknown model %q", model)
	}
}

// streamWorkers / exchangeWorkers are the parallel store queries of stream and vecstream.
const (
	streamWorkers   = 4
	exchangeWorkers = 4
)

// blocking reports whether an operator waits for all rows. A shipped aggregation (ShipAggregate) too: it sends
// all input in ShipChunkRows-row tables (per batch the tables would shrink to the iteration width, adding round trips and partial rows).
func blocking(n physical.Node) bool {
	if q, ok := n.(*physical.StoreQuery); ok {
		return q.ShipMode == ir.ShipAggregate || q.ShipMode == ir.ShipHand
	}
	c, ok := n.(*physical.Coord)
	if !ok {
		return false
	}
	switch c.Op.(type) {
	case *ir.Aggregation, *ir.Sort, *ir.Distinct, *ir.NodeHashJoin:
		return true
	}
	return false
}

func chunk(rows []Row, width int) [][]Row {
	var out [][]Row
	for len(rows) > 0 {
		n := len(rows)
		if n > width {
			n = width
		}
		out = append(out, rows[:n:n])
		rows = rows[n:]
	}
	return out
}

// bulk evaluates by materializing everything from the leaves.
func (e *Executor) bulk(n physical.Node) ([]Row, error) {
	if o, ok := n.(*physical.OptionalApply); ok {
		// The inner side depends on outer rows: evaluate the outer side, pass rows to the leaf, then the inner side.
		outer, err := e.bulk(o.Outer)
		if err != nil {
			return nil, err
		}
		e.setArg(o.Arg, outer)
		inner, err := e.bulk(o.Inner)
		if err != nil {
			return nil, err
		}
		return e.step(n, [][]Row{outer, inner})
	}
	var inputs [][]Row
	for _, c := range n.Children() {
		rows, err := e.bulk(c)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, rows)
	}
	return e.step(n, inputs)
}
