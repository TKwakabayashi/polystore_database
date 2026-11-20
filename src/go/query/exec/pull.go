package exec

// pull.go: the pull models (volcano, vectorized, vecstream): iterators and the exchange of parallel workers.

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"polystore_database/src/go/query/physical"
	"sync"
)

type iterator interface {
	// next returns the next batch of rows (nil at the end; never an empty batch). Rows pass downstream without repacking.
	next() ([]Row, error)
	close()
}

// pull builds iterators from the root and drains them. With workers > 0 an exchange runs store queries in parallel.
func (e *Executor) pull(root physical.Node, width, workers int) ([]Row, error) {
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	it, err := e.iter(ctx, root, width, workers)
	if err != nil {
		return nil, err
	}
	defer it.close()
	var out []Row
	for {
		b, err := it.next()
		if err != nil {
			return nil, err
		}
		if b == nil {
			return out, nil
		}
		out = append(out, b...)
	}
}

func (e *Executor) iter(ctx context.Context, n physical.Node, width, workers int) (iterator, error) {
	if o, ok := n.(*physical.OptionalApply); ok {
		outer, err := e.iter(ctx, o.Outer, width, workers)
		if err != nil {
			return nil, err
		}
		return &optionalIter{e: e, o: o, outer: outer, ctx: ctx, width: width, workers: workers}, nil
	}
	var kids []iterator
	for _, c := range n.Children() {
		k, err := e.iter(ctx, c, width, workers)
		if err != nil {
			return nil, err
		}
		kids = append(kids, k)
	}
	switch {
	case len(kids) == 0: // leaf (a StoreQuery without input): issue once and hand out by width
		return &leafIter{e: e, n: n, width: width, ctx: ctx}, nil
	case blocking(n):
		return &blockingIter{e: e, n: n, kids: kids, width: width}, nil
	case isStoreAccess(n) && workers > 0:
		return newExchange(ctx, e, n, kids[0], width, workers), nil
	}
	return &mapIter{e: e, n: n, kid: kids[0], limit: limitOf(n)}, nil
}

func isStoreAccess(n physical.Node) bool {
	switch n.(type) {
	case *physical.StoreQuery, *physical.ListFetch, *physical.StepExpand:
		return true
	}
	return false
}

// leafIter hands out a leaf's result by width. A streamable leaf (streamableLeaf) receives from the store on a
// goroutine while handing out (transfer overlaps downstream work); close cancels it and closes the cursor. Other leaves fetch everything first.
type leafIter struct {
	e     *Executor
	n     physical.Node
	width int
	ctx   context.Context
	rows  [][]Row
	done  bool

	// the streamed leaf
	ch     chan []Row
	errc   chan error
	cancel context.CancelFunc
	ended  bool // the producer's end (errc) has been received
}

func (it *leafIter) next() ([]Row, error) {
	if q, ok := it.e.streamableLeaf(it.n); ok {
		return it.nextStreamed(q)
	}
	if !it.done {
		rows, err := it.e.step(it.n, nil)
		if err != nil {
			return nil, err
		}
		it.rows, it.done = chunk(rows, it.width), true
	}
	if len(it.rows) == 0 {
		return nil, nil
	}
	b := it.rows[0]
	it.rows = it.rows[1:]
	return b, nil
}

func (it *leafIter) nextStreamed(q *physical.StoreQuery) ([]Row, error) {
	if it.ended {
		return nil, nil
	}
	if it.ch == nil {
		ctx, cancel := context.WithCancel(it.ctx)
		it.ch, it.errc, it.cancel = make(chan []Row, 2), make(chan error, 1), cancel
		go func() {
			err := it.e.streamLeaf(ctx, q, it.width, func(rows []Row) error {
				select {
				case it.ch <- rows:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			close(it.ch)
			it.errc <- err
		}()
	}
	if rows, ok := <-it.ch; ok {
		return rows, nil
	}
	it.ended = true
	it.cancel()
	return nil, <-it.errc
}

// close cancels a streamed leaf and waits for the producer goroutine to finish (the cursor closed).
func (it *leafIter) close() {
	if it.ch == nil || it.ended {
		return
	}
	it.cancel()
	for range it.ch { // discard what the producer was sending
	}
	<-it.errc
	it.ended = true
}

// mapIter applies an operator to one batch of its child (one store round trip per input batch; no rechunking).
type mapIter struct {
	e     *Executor
	n     physical.Node
	kid   iterator
	limit *limitState
}

func (it *mapIter) next() ([]Row, error) {
	for {
		if it.limit != nil && it.limit.count >= 0 && it.limit.seen >= it.limit.skip+it.limit.count {
			return nil, nil
		}
		b, err := it.kid.next()
		if err != nil || b == nil {
			return nil, err
		}
		var out []Row
		if it.limit != nil {
			out, _ = it.limit.take(b)
		} else {
			out, err = it.e.step(it.n, [][]Row{b})
			if err != nil {
				return nil, err
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}
}

func (it *mapIter) close() { it.kid.close() }

// blockingIter drains its child, applies the operator and hands out by width.
type blockingIter struct {
	e     *Executor
	n     physical.Node
	kids  []iterator
	width int
	rows  [][]Row
	done  bool
}

func drain(it iterator) ([]Row, error) {
	var out []Row
	for {
		b, err := it.next()
		if err != nil {
			return nil, err
		}
		if b == nil {
			return out, nil
		}
		out = append(out, b...)
	}
}

func (it *blockingIter) next() ([]Row, error) {
	if !it.done {
		inputs := make([][]Row, len(it.kids))
		for i, k := range it.kids {
			rows, err := drain(k)
			if err != nil {
				return nil, err
			}
			inputs[i] = rows
		}
		rows, err := it.e.step(it.n, inputs)
		if err != nil {
			return nil, err
		}
		it.rows, it.done = chunk(rows, it.width), true
	}
	if len(it.rows) == 0 {
		return nil, nil
	}
	b := it.rows[0]
	it.rows = it.rows[1:]
	return b, nil
}

func (it *blockingIter) close() {
	for _, k := range it.kids {
		k.close()
	}
}

// optionalIter drains the outer side, then builds the inner iterators (the inner leaf depends on outer rows)
// and hands out the joined rows by width.
type optionalIter struct {
	e              *Executor
	o              *physical.OptionalApply
	outer          iterator
	ctx            context.Context
	width, workers int
	rows           [][]Row
	done           bool
}

func (it *optionalIter) next() ([]Row, error) {
	if !it.done {
		outer, err := drain(it.outer)
		if err != nil {
			return nil, err
		}
		it.e.setArg(it.o.Arg, outer)
		inIt, err := it.e.iter(it.ctx, it.o.Inner, it.width, it.workers)
		if err != nil {
			return nil, err
		}
		inner, err := drain(inIt)
		inIt.close()
		if err != nil {
			return nil, err
		}
		rows, err := it.e.step(it.o, [][]Row{outer, inner})
		if err != nil {
			return nil, err
		}
		it.rows, it.done = chunk(rows, it.width), true
	}
	if len(it.rows) == 0 {
		return nil, nil
	}
	b := it.rows[0]
	it.rows = it.rows[1:]
	return b, nil
}

func (it *optionalIter) close() { it.outer.close() }

// exchange runs store queries in parallel with a driver pulling children serially and W workers (vecstream).
// Output is rechunked to the width. Row order is not kept (a later Sort decides it).
type exchange struct {
	out    chan []Row
	errc   chan error
	cancel context.CancelFunc
	kid    iterator
	err    error
}

func newExchange(parent context.Context, e *Executor, n physical.Node, kid iterator, width, workers int) *exchange {
	ctx, cancel := context.WithCancel(parent)
	x := &exchange{out: make(chan []Row, workers), errc: make(chan error, 1), cancel: cancel, kid: kid}
	in := make(chan []Row, workers)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { // driver: pull children serially
		defer close(in)
		for {
			b, err := kid.next()
			if err != nil {
				return err
			}
			if b == nil {
				return nil
			}
			select {
			case in <- b:
			case <-gctx.Done():
				return gctx.Err()
			}
		}
	})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			for rows := range in {
				res, err := e.step(n, [][]Row{rows})
				if err != nil {
					return err
				}
				for _, part := range chunk(res, width) {
					select {
					case x.out <- part:
					case <-gctx.Done():
						return gctx.Err()
					}
				}
			}
			return nil
		})
	}
	go func() {
		err := g.Wait()
		close(x.out)
		x.errc <- err
	}()
	return x
}

func (x *exchange) next() ([]Row, error) {
	b, ok := <-x.out
	if ok {
		return b, nil
	}
	if x.err == nil {
		x.err = <-x.errc
		if x.err == nil {
			x.err = errDone
		}
	}
	if x.err == errDone {
		return nil, nil
	}
	return nil, x.err
}

var errDone = fmt.Errorf("exchange: done")

func (x *exchange) close() {
	x.cancel()
	for range x.out { // discard the rest and let goroutines finish
	}
	x.kid.close()
}
