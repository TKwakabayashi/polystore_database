package exec

// push.go: the push model (stream): a goroutine per operator, and leaves that stream rows as they arrive.

import (
	"context"
	"golang.org/x/sync/errgroup"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/physical"
	"sync"
	"time"
)

// push streams row batches over channels with a goroutine per operator; store queries run on parallel workers.
func (e *Executor) push(root physical.Node, width int) ([]Row, error) {
	g, ctx := errgroup.WithContext(e.ctx)
	out := e.stage(ctx, g, root, width)
	var rows []Row
	g.Go(func() error {
		for b := range out {
			rows = append(rows, b...)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// Stages do not return errors after cancellation (goStage), so check the caller's cancellation here (no partial results).
	if err := e.ctx.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

// goStage registers a stage goroutine in g. Errors after the stage's ctx is cancelled are not returned: a LIMIT
// stopping upstream is a normal end, and a stop caused by another stage's failure leaves that first error in g
// (errgroup cancels its ctx only when an error is returned). push checks the caller's cancellation.
func goStage(g *errgroup.Group, ctx context.Context, f func() error) {
	g.Go(func() error {
		if err := f(); err != nil && ctx.Err() == nil {
			return err
		}
		return nil
	})
}

// withStop returns a child ctx and its cancel function (for a LIMIT stopping its upstream subtree).
func withStop(ctx context.Context) (context.Context, func()) {
	return context.WithCancel(ctx)
}

// stage returns n's output channel and registers the goroutine processing n in g.
func (e *Executor) stage(ctx context.Context, g *errgroup.Group, n physical.Node, width int) <-chan []Row {
	out := make(chan []Row, 4)
	if o, ok := n.(*physical.OptionalApply); ok {
		// The inner side depends on outer rows: collect the outer side, then run the inner pipeline synchronously.
		outerCh := e.stage(ctx, g, o.Outer, width)
		goStage(g, ctx, func() error {
			defer close(out)
			var outer []Row
			for b := range outerCh {
				outer = append(outer, b...)
			}
			e.setArg(o.Arg, outer)
			inner, err := e.push(o.Inner, width)
			if err != nil {
				return err
			}
			rows, err := e.step(o, [][]Row{outer, inner})
			if err != nil {
				return err
			}
			for _, part := range chunk(rows, width) {
				select {
				case out <- part:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
		return out
	}
	// A LIMIT stops its upstream subtree once it has enough rows (streamed leaves close their cursors); upstream runs on a child ctx.
	kidCtx := ctx
	var stopKids func()
	if limitOf(n) != nil && !blocking(n) && len(n.Children()) > 0 {
		kidCtx, stopKids = withStop(ctx) // always called by the Limit stage (g's ctx is cancelled by Wait otherwise)
	}
	var ins []<-chan []Row
	for _, c := range n.Children() {
		ins = append(ins, e.stage(kidCtx, g, c, width))
	}
	send := func(rows []Row) error {
		for _, part := range chunk(rows, width) {
			select {
			case out <- part:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	collect := func(in <-chan []Row) []Row {
		var all []Row
		for b := range in {
			all = append(all, b...)
		}
		return all
	}

	switch {
	case len(ins) == 0:
		if q, ok := e.streamableLeaf(n); ok { // pass downstream as rows arrive
			goStage(g, ctx, func() error {
				defer close(out)
				return e.streamLeaf(ctx, q, width, func(rows []Row) error {
					select {
					case out <- rows:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				})
			})
			break
		}
		goStage(g, ctx, func() error {
			defer close(out)
			rows, err := e.step(n, nil)
			if err != nil {
				return err
			}
			return send(rows)
		})
	case blocking(n):
		goStage(g, ctx, func() error {
			defer close(out)
			inputs := make([][]Row, len(ins))
			for i, in := range ins {
				inputs[i] = collect(in)
			}
			if err := ctx.Err(); err != nil { // the input stopped early (not all rows)
				return err
			}
			rows, err := e.step(n, inputs)
			if err != nil {
				return err
			}
			return send(rows)
		})
	case limitOf(n) != nil:
		l := limitOf(n)
		goStage(g, ctx, func() error {
			defer close(out)
			if stopKids != nil {
				defer stopKids()
			}
			done := false
			for b := range ins[0] {
				if done {
					continue // drain what was being sent until upstream stops and closes
				}
				var keep []Row
				keep, done = l.take(b)
				if err := send(keep); err != nil {
					return err
				}
				if done && stopKids != nil {
					stopKids()
				}
			}
			return nil
		})
	default:
		workers := 1
		if isStoreAccess(n) {
			workers = streamWorkers
		}
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			goStage(g, ctx, func() error {
				defer wg.Done()
				for b := range ins[0] {
					rows, err := e.step(n, [][]Row{b})
					if err != nil {
						return err
					}
					if err := send(rows); err != nil {
						return err
					}
				}
				return nil
			})
		}
		g.Go(func() error {
			wg.Wait()
			close(out)
			return nil
		})
	}
	return out
}

// streamableLeaf reports whether n is a leaf (a native query without input) that can stream rows as they
// arrive.
func (e *Executor) streamableLeaf(n physical.Node) (*physical.StoreQuery, bool) {
	q, ok := n.(*physical.StoreQuery)
	if !ok || !q.Leaf() {
		return nil, false
	}
	return q, true
}

// streamLeaf issues a leaf query and passes arriving rows to send width rows at a time (without keeping them
// all). Store transfer overlaps downstream work; when downstream cancels ctx (a reached LIMIT, ...) the cursor
// closes. Metrics are added once at the end; time spent waiting on send is not counted.
func (e *Executor) streamLeaf(ctx context.Context, q *physical.StoreQuery, width int, send func([]Row) error) error {
	start := time.Now()
	var waited time.Duration
	out := 0
	var expanded int64
	defer func() { e.recordCount(q, time.Since(start)-waited, 0, out, expanded) }()

	nq, err := e.lower(q)
	if err != nil {
		return err
	}
	if width < 1 {
		width = 1
	}
	flush := func(buf []Row) error {
		t := time.Now()
		err := send(buf)
		waited += time.Since(t)
		return err
	}
	buf := make([]Row, 0, width)
	err = e.stream(ctx, q, nq, nil, func(vals []any) error {
		r := make(Row, len(nq.Columns))
		for j, c := range nq.Columns {
			r[c] = vals[j]
		}
		out++
		expanded += eval.WeightOf(r)
		buf = append(buf, r)
		if len(buf) < width {
			return nil
		}
		full := buf
		buf = make([]Row, 0, width)
		return flush(full)
	})
	if err != nil {
		return err
	}
	if len(buf) > 0 {
		return flush(buf)
	}
	return nil
}
