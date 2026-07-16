package volcano_async_executor

import (
	"context"
	"sync"
	"time"
)

// asyncDriver は Volcano 論文の exchange 演算子に相当する。
//
// 本パッケージで並行性を知っているのはこの型だけで、上流・下流から見ると
// ただの Iterator（Open/Next/Close）に見える。Filter/Expand 側は
// 「1 バッチ処理 = processFn」を提供するだけで、同期版から本体を変更していない。
//
// 構造:
//
//	child.Next() ──(直列 pull / 安価)──> work ch ──> W workers: process(batch)
//	                                                    │  (DB 往復 = ここだけ並行)
//	                                                    v
//	                                                  out ch(有界) ──> 下流の Next()
//
// child.Next() を直列に引くのは、scan が Open 時に全件取得済みで Next が
// メモリ上のスライス分割にすぎず（op_scan.go）、並行化しても得が無いため。
// 高価なのは process 内の DB 往復であり、そこだけをワーカーで並行化する。
//
// 順序: AsyncExchange では保存されない（ワーカーの完了順）。ORDER BY を含む
// クエリは tail の Sort が最終的に整列するため結果の正しさには影響しない。
// AsyncPrefetch は workers=1 なので順序が保存される。
type asyncDriver struct {
	p       *Processor
	child   Iterator
	opKind  OpKind
	step    int
	process processFn

	out    chan *Batch
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errMu sync.Mutex
	err   error
}

// newAsyncDriver は child の各バッチへ process を適用する非同期演算子を作る。
func newAsyncDriver(p *Processor, child Iterator, op OpKind, step int, fn processFn) *asyncDriver {
	return &asyncDriver{p: p, child: child, opKind: op, step: step, process: fn}
}

func (d *asyncDriver) setErr(err error) {
	if err == nil {
		return
	}
	d.errMu.Lock()
	if d.err == nil {
		d.err = err
	}
	d.errMu.Unlock()
}

func (d *asyncDriver) getErr() error {
	d.errMu.Lock()
	defer d.errMu.Unlock()
	return d.err
}

// Open は子を初期化し、パイプライン（driver goroutine + W workers）を起動する。
// out は有界なので、下流が Next を呼ばない限り先読みは buffer 分で頭打ちになる
// （= pull の需要駆動性を保ったまま、その範囲でだけ往復を重ねる）。
func (d *asyncDriver) Open(ctx context.Context) error {
	if err := d.child.Open(ctx); err != nil {
		return err
	}

	dctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	workers := d.p.policy.workersFor(d.opKind, d.p.asyncMode)
	d.out = make(chan *Batch, d.p.policy.bufferFor(d.opKind, d.p.asyncMode))
	work := make(chan *Batch, workers)

	// driver: child から直列に pull して work へ配る。
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(work)
		for {
			b, err := d.child.Next(dctx)
			if err != nil {
				d.setErr(err)
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

	// workers: process（DB 往復）を並行実行。
	var wwg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wwg.Add(1)
		d.wg.Add(1)
		go func() {
			defer wwg.Done()
			defer d.wg.Done()
			for b := range work {
				// ExecDynamic ではシステム全体の同時 DB 数をセマフォで抑える。
				// permit は DB 区間だけ保持し、送信前に必ず解放する（デッドロック回避）。
				if d.p.sem != nil {
					if err := d.p.sem.Acquire(dctx, 1); err != nil {
						return
					}
				}
				start := time.Now()
				out, err := d.process(b)
				if d.p.sem != nil {
					d.p.sem.Release(1)
				}
				if err != nil {
					d.setErr(err)
					d.cancel() // 兄弟ワーカーと driver を畳む
					return
				}
				d.p.recordOp(d.step, string(d.opKind), time.Since(start), out.n)
				if out.n == 0 {
					continue // 空バッチは EOF と紛れるので流さない
				}
				select {
				case d.out <- out:
				case <-dctx.Done():
					return
				}
			}
		}()
	}

	// 全ワーカー終了で out を閉じる = 下流にとっての EOF。
	go func() {
		wwg.Wait()
		close(d.out)
	}()
	return nil
}

// Next は out から 1 バッチ取り出す。閉じていれば EOF (nil, nil)。
// エラーはワーカー側で記録され、ここで返る。
func (d *asyncDriver) Next(ctx context.Context) (*Batch, error) {
	select {
	case b, ok := <-d.out:
		if !ok {
			return nil, d.getErr()
		}
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close はパイプラインを畳んでから子を閉じる。
// cancel 後に out を drain するのは、送信待ちで固まっているワーカーを
// 解放して wg.Wait() を確実に返させるため。
func (d *asyncDriver) Close(ctx context.Context) error {
	if d.cancel != nil {
		d.cancel()
		for range d.out { //nolint:revive // 送信側を解放するための drain
		}
		d.wg.Wait()
		d.cancel = nil
	}
	return d.child.Close(ctx)
}
