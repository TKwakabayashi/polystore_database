package vecstream

import (
	"context"
	"sync"
	"time"
)

// exchangeIterator は Volcano 論文の exchange 演算子。pull インターフェースを保ったまま、
// 内部で子を直列 pull し、W ワーカーが process を並列実行する。
//
// 構造:
//
//	child.Next() ──(driver goroutine が直列 pull)──> work ch ──> W workers: process(res, b)
//	                                                              │  (DB 往復 = ここだけ並列)
//	                                                              v
//	                                                            out ch(有界) ──> 親の Next()
//
// R はワーカーが 1 本だけ生成してバッチ間で使い回すリソース（Neo4j セッション）。
// newRes はワーカーごとに 1 回だけ呼ばれ（＝session 再利用）、closeRes でワーカー終了時に閉じる。
// 出力は split で VectorWidth 幅へ再チャンクしてから out へ流す（列幅を一律に保つ）。
//
// child.Next は driver goroutine 1 本からのみ呼ばれるので、子 Iterator は goroutine 安全でなくてよい。
// 順序は保存されない（ワーカー完了順）。ORDER BY は tail の Sort が最終的に整列する。
type exchangeIterator[R any] struct {
	p        *Processor
	child    Iterator
	opKind   OpKind
	opName   string
	step     int
	newRes   func() R
	closeRes func(R)
	process  func(res R, in *Batch) (*Batch, error)

	out    chan *Batch
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errMu sync.Mutex
	err   error
}

func newExchange[R any](
	p *Processor, child Iterator, opKind OpKind, opName string, step int,
	newRes func() R, closeRes func(R),
	process func(res R, in *Batch) (*Batch, error),
) *exchangeIterator[R] {
	return &exchangeIterator[R]{
		p: p, child: child, opKind: opKind, opName: opName, step: step,
		newRes: newRes, closeRes: closeRes, process: process,
	}
}

func (e *exchangeIterator[R]) setErr(err error) {
	if err == nil {
		return
	}
	e.errMu.Lock()
	if e.err == nil {
		e.err = err
	}
	e.errMu.Unlock()
}

func (e *exchangeIterator[R]) getErr() error {
	e.errMu.Lock()
	defer e.errMu.Unlock()
	return e.err
}

// Open は子を初期化し、driver goroutine と W ワーカーを起動する。
func (e *exchangeIterator[R]) Open(ctx context.Context) error {
	if err := e.child.Open(ctx); err != nil {
		return err
	}
	dctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	workers := e.p.exec.workersFor(e.opKind)
	e.out = make(chan *Batch, e.p.exec.bufferFor(e.opKind))
	work := make(chan *Batch, workers)

	// driver: child から直列 pull して work へ配る。
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer close(work)
		for {
			b, err := e.child.Next(dctx)
			if err != nil {
				e.setErr(err)
				return
			}
			if b == nil {
				return // EOF
			}
			select {
			case work <- b:
			case <-dctx.Done():
				return
			}
		}
	}()

	// workers: 各自 session を 1 本生成して使い回し、process を並列実行。
	var wwg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wwg.Add(1)
		e.wg.Add(1)
		go func() {
			defer wwg.Done()
			defer e.wg.Done()
			res := e.newRes() // ワーカーにつき 1 回（session 再利用）
			defer e.closeRes(res)
			for b := range work {
				// ExecDynamic の全体上限を DB 区間だけ sem で抑える（emit 前に解放）。
				if e.p.sem != nil {
					if err := e.p.sem.Acquire(dctx, 1); err != nil {
						return
					}
				}
				start := time.Now()
				out, err := e.process(res, b)
				if e.p.sem != nil {
					e.p.sem.Release(1)
				}
				if err != nil {
					e.setErr(err)
					e.cancel() // 兄弟ワーカーと driver を畳む
					return
				}
				rows := 0
				if out != nil {
					rows = out.n
				}
				e.p.recordOp(e.step, e.opName, time.Since(start), rows)
				for _, sub := range split(out, e.p.exec.vectorWidth()) {
					select {
					case e.out <- sub:
					case <-dctx.Done():
						return
					}
				}
			}
		}()
	}

	// 全ワーカー終了で out を閉じる = 親にとっての EOF。
	go func() {
		wwg.Wait()
		close(e.out)
	}()
	return nil
}

// Next は out から 1 バッチ取り出す。閉じていれば EOF (nil, nil)。
func (e *exchangeIterator[R]) Next(ctx context.Context) (*Batch, error) {
	select {
	case b, ok := <-e.out:
		if !ok {
			return nil, e.getErr()
		}
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close はパイプラインを畳んでから子を閉じる。cancel 後に out を drain して、
// 送信待ちで固まっているワーカーを解放し wg.Wait() を確実に返させる。
func (e *exchangeIterator[R]) Close(ctx context.Context) error {
	if e.cancel != nil {
		e.cancel()
		for range e.out { //nolint:revive // 送信側を解放するための drain
		}
		e.wg.Wait()
		e.cancel = nil
	}
	return e.child.Close(ctx)
}
