package executor

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/settings"
)

// models.go は反復モデルの駆動部。意味（step）は共通で、行の流し方だけが違う。
//
//	bulk        葉から全件を実体化（run.go の bulk）
//	volcano     pull 型の反復子。列指向 Batch、幅 1（入力 1 行ごとにストアへ 1 往復）
//	vectorized  volcano と同じで幅 settings.VectorSize
//	stream      push 型。演算子ごとの goroutine とチャネル。ストアへの発行はワーカー並列
//	vecstream   pull 型＋exchange（子を直列に pull する driver と W ワーカー）。出力は幅で再チャンク
//
// ストアへ問い合わせる演算子（StoreQuery / ListFetch）と、行ごとに独立なコーディネータ演算子
// （Filter / Projection / ProjectPath / RelUniqueFilter）はバッチ単位で処理できる。集約・並べ替え・
// 重複除去・結合は全件が揃ってから処理する（pipeline breaker）。Limit は件数を数えながら流す。

// ExecuteModel は物理計画を指定の反復モデルで評価する。
func (e *Executor) ExecuteModel(n physical.Node, model settings.EngineKind) ([]Row, error) {
	return e.ExecuteWithArgs(n, model, nil)
}

// ExecuteWithArgs は、計画の葉 ArgRows に外から行を与えて実行する（演算子を単体で測る実験基盤 opbench 用。
// 通常の計画では ArgRows は OPTIONAL の内側にしか現れず、外側の行で埋まる）。
func (e *Executor) ExecuteWithArgs(n physical.Node, model settings.EngineKind, args map[*physical.ArgRows][]Row) ([]Row, error) {
	e.reset(n)
	e.mu.Lock()
	for a, rows := range args {
		e.args[a] = rows
	}
	e.mu.Unlock()
	switch model {
	case settings.EngineBulk, "":
		return e.bulk(n)
	case settings.EngineVolcano:
		return e.pull(n, 1, 0)
	case settings.EngineVectorized:
		return e.pull(n, vectorWidth(), 0)
	case settings.EngineVecStream:
		return e.pull(n, vectorWidth(), exchangeWorkers)
	case settings.EngineStream:
		return e.push(n, vectorWidth())
	}
	return nil, fmt.Errorf("executor: unknown model %q", model)
}

func vectorWidth() int {
	if settings.VectorSize < 1 {
		return 1
	}
	return settings.VectorSize
}

// streamWorkers / exchangeWorkers はストアへの並列発行数（旧 stream / vecstream の既定と同じ 4）。
const (
	streamWorkers   = 4
	exchangeWorkers = 4
)

// blocking は全件が揃ってから処理する演算子か。
func blocking(n physical.Node) bool {
	c, ok := n.(*physical.Coord)
	if !ok {
		return false
	}
	switch c.Op.(type) {
	case *logical.Aggregation, *logical.Sort, *logical.Distinct, *logical.NodeHashJoin:
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

// ================================
// 列指向 Batch
// ================================

// Batch は列指向のバッチ（volcano / vectorized / vecstream の単位）。Data[c][i] が列 Cols[c] の i 行目。
type Batch struct {
	Cols []string
	Data [][]any
	n    int
}

// Len は行数。
func (b *Batch) Len() int { return b.n }

func toBatch(rows []Row) *Batch {
	b := &Batch{n: len(rows)}
	idx := map[string]int{}
	for _, r := range rows {
		for k := range r {
			if _, ok := idx[k]; !ok {
				idx[k] = len(b.Cols)
				b.Cols = append(b.Cols, k)
			}
		}
	}
	b.Data = make([][]any, len(b.Cols))
	for c := range b.Data {
		b.Data[c] = make([]any, len(rows))
	}
	for i, r := range rows {
		for k, v := range r {
			b.Data[idx[k]][i] = v
		}
	}
	return b
}

func (b *Batch) rows() []Row {
	out := make([]Row, b.n)
	for i := range out {
		r := make(Row, len(b.Cols))
		for c, k := range b.Cols {
			r[k] = b.Data[c][i]
		}
		out[i] = r
	}
	return out
}

// ================================
// pull（volcano / vectorized / vecstream）
// ================================

type iterator interface {
	// next は次のバッチ（nil で終わり）。
	next() (*Batch, error)
	close()
}

// pull は根から反復子を組み、末尾まで引き出す。workers > 0 ならストアへの発行を exchange で並列化する。
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
		out = append(out, b.rows()...)
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
	case len(kids) == 0: // 葉（入力を持たない StoreQuery）: 1 回発行して幅で払い出す
		return &leafIter{e: e, n: n, width: width}, nil
	case blocking(n):
		return &blockingIter{e: e, n: n, kids: kids, width: width}, nil
	case isStoreAccess(n) && workers > 0:
		return newExchange(ctx, e, n, kids[0], width, workers), nil
	}
	return &mapIter{e: e, n: n, kid: kids[0], limit: limitOf(n)}, nil
}

func isStoreAccess(n physical.Node) bool {
	switch n.(type) {
	case *physical.StoreQuery, *physical.ListFetch:
		return true
	}
	return false
}

// leafIter は葉の結果を 1 回で取り、幅ずつ払い出す。
type leafIter struct {
	e     *Executor
	n     physical.Node
	width int
	rows  [][]Row
	done  bool
}

func (it *leafIter) next() (*Batch, error) {
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
	b := toBatch(it.rows[0])
	it.rows = it.rows[1:]
	return b, nil
}

func (it *leafIter) close() {}

// mapIter は子の 1 バッチに演算子を適用する（ストアへの発行は入力バッチごとに 1 往復。再チャンクしない）。
type mapIter struct {
	e     *Executor
	n     physical.Node
	kid   iterator
	limit *limitState
}

type limitState struct {
	skip, count int // count < 0 は無制限
	seen        int
}

func limitOf(n physical.Node) *limitState {
	if c, ok := n.(*physical.Coord); ok {
		if l, ok := c.Op.(*logical.Limit); ok {
			return &limitState{skip: l.Skip, count: l.Count}
		}
	}
	return nil
}

// take は Limit の状態を進め、このバッチから残す行を返す（done は以降の入力が不要か）。
// 重み付きの行は重みの分だけ数え、境目の行は重みを削って残す。
func (l *limitState) take(rows []Row) (out []Row, done bool) {
	full := func() bool { return l.count >= 0 && l.seen >= l.skip+l.count }
	for _, r := range rows {
		if full() {
			return out, true
		}
		w := weightOf(r)
		lo, hi := l.seen, l.seen+int(w) // この行が占める位置 [lo, hi)
		l.seen = hi
		from, to := lo, hi
		if from < l.skip {
			from = l.skip
		}
		if l.count >= 0 && to > l.skip+l.count {
			to = l.skip + l.count
			l.seen = to
		}
		if to <= from {
			continue
		}
		if keep := int64(to - from); keep != w {
			nr := r.clone()
			setWeight(nr, keep)
			r = nr
		}
		out = append(out, r)
	}
	return out, full()
}

func (it *mapIter) next() (*Batch, error) {
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
			out, _ = it.limit.take(b.rows())
		} else {
			out, err = it.e.step(it.n, [][]Row{b.rows()})
			if err != nil {
				return nil, err
			}
		}
		if len(out) > 0 {
			return toBatch(out), nil
		}
	}
}

func (it *mapIter) close() { it.kid.close() }

// blockingIter は子を全件引き出してから演算子を適用し、幅で払い出す。
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
		out = append(out, b.rows()...)
	}
}

func (it *blockingIter) next() (*Batch, error) {
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
	b := toBatch(it.rows[0])
	it.rows = it.rows[1:]
	return b, nil
}

func (it *blockingIter) close() {
	for _, k := range it.kids {
		k.close()
	}
}

// optionalIter は外側を全件引き出してから内側の反復子を組み（内側の葉は外側の行に依存する）、
// 結合結果を幅で払い出す。
type optionalIter struct {
	e              *Executor
	o              *physical.OptionalApply
	outer          iterator
	ctx            context.Context
	width, workers int
	rows           [][]Row
	done           bool
}

func (it *optionalIter) next() (*Batch, error) {
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
	b := toBatch(it.rows[0])
	it.rows = it.rows[1:]
	return b, nil
}

func (it *optionalIter) close() { it.outer.close() }

// exchange は子を直列に pull する driver と W ワーカーで、ストアへの発行を並列にする（vecstream）。
// 出力は幅で再チャンクする。行の順序は保たない（後段の Sort が順序を決める）。
type exchange struct {
	out    chan *Batch
	errc   chan error
	cancel context.CancelFunc
	kid    iterator
	err    error
}

func newExchange(parent context.Context, e *Executor, n physical.Node, kid iterator, width, workers int) *exchange {
	ctx, cancel := context.WithCancel(parent)
	x := &exchange{out: make(chan *Batch, workers), errc: make(chan error, 1), cancel: cancel, kid: kid}
	in := make(chan []Row, workers)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { // driver: 子を直列に pull
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
			case in <- b.rows():
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
					case x.out <- toBatch(part):
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

func (x *exchange) next() (*Batch, error) {
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
	for range x.out { // 残りを捨てて goroutine を終わらせる
	}
	x.kid.close()
}

// ================================
// push（stream）
// ================================

// push は演算子ごとの goroutine で行のバッチをチャネルに流す。ストアへの発行はワーカー並列。
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
	return rows, nil
}

// stage は n の出力チャネルを返し、n を処理する goroutine を g に登録する。
func (e *Executor) stage(ctx context.Context, g *errgroup.Group, n physical.Node, width int) <-chan []Row {
	out := make(chan []Row, 4)
	if o, ok := n.(*physical.OptionalApply); ok {
		// 内側は外側の行に依存するので、外側を集めてから内側の pipeline を同期で回す。
		outerCh := e.stage(ctx, g, o.Outer, width)
		g.Go(func() error {
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
	var ins []<-chan []Row
	for _, c := range n.Children() {
		ins = append(ins, e.stage(ctx, g, c, width))
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
		g.Go(func() error {
			defer close(out)
			rows, err := e.step(n, nil)
			if err != nil {
				return err
			}
			return send(rows)
		})
	case blocking(n):
		g.Go(func() error {
			defer close(out)
			inputs := make([][]Row, len(ins))
			for i, in := range ins {
				inputs[i] = collect(in)
			}
			rows, err := e.step(n, inputs)
			if err != nil {
				return err
			}
			return send(rows)
		})
	case limitOf(n) != nil:
		l := limitOf(n)
		g.Go(func() error {
			defer close(out)
			done := false
			for b := range ins[0] {
				if done {
					continue // 上流を止めずに読み捨てる
				}
				var keep []Row
				keep, done = l.take(b)
				if err := send(keep); err != nil {
					return err
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
			g.Go(func() error {
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
