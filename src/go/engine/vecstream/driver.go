package vecstream

import (
	"context"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// runBatches は 1 演算子分の並行駆動を担う（engine/stream.runBatches の *Batch 版）。
// 特定 DB に非依存。process が 1 バッチを処理して 1 バッチを返し（DB 往復はこの中）、
// runBatches が結果を VectorWidth 幅へ再チャンクして下流へ emit する。
//
// engine/stream との差分:
//   - 流れる単位が []Record → *Batch。
//   - emit の固定 chunk=2000 → split(out, policy.vectorWidth()) による VectorWidth 再チャンク。
//   - バッチごとに rec(dur, rows) で演算子計測を記録（recordOp は mutex で保護）。
//
// 動的モードは sem（全演算子で共有）を DB 区間だけ acquire/release し、解放後に emit する
// ためデッドロックしない。空バッチ (n==0) は EOF と紛れないよう下流へ流さない。
func runBatches[R any](
	ctx context.Context,
	policy ExecPolicy,
	sem *semaphore.Weighted, // 動的モードのみ使用。全演算子で同一インスタンスを共有
	op OpKind,
	inputStream <-chan *Batch,
	outputStream chan<- *Batch,
	newRes func() R,
	closeRes func(R),
	process func(res R, in *Batch) (*Batch, error),
	rec func(dur time.Duration, rows int),
) (int, error) {
	var total int64
	var g errgroup.Group

	emit := func(out *Batch, dur time.Duration) {
		rows := 0
		if out != nil {
			rows = out.n
		}
		rec(dur, rows)
		for _, sub := range split(out, policy.vectorWidth()) {
			outputStream <- sub
		}
		atomic.AddInt64(&total, int64(rows))
	}

	switch policy.Mode {
	case ExecDynamic:
		if sem == nil {
			sem = semaphore.NewWeighted(int64(policy.globalMax()))
		}
		for batch := range inputStream {
			b := batch
			if err := sem.Acquire(ctx, 1); err != nil { // permit 未保持でブロック
				break
			}
			g.Go(func() error {
				res := newRes()
				start := time.Now()
				out, err := process(res, b) // DB区間（permit 保持）
				dur := time.Since(start)
				closeRes(res)
				sem.Release(1) // emit 前に解放
				if err != nil {
					return err
				}
				emit(out, dur) // permit 無しで送信
				return nil
			})
		}

	default: // ExecFixed
		c := policy.For(op)
		batchChan := make(chan *Batch, c.workers())
		for i := 0; i < c.workers(); i++ {
			g.Go(func() error {
				res := newRes() // ワーカー単位で使い回す
				defer closeRes(res)
				var werr error
				for b := range batchChan {
					start := time.Now()
					out, err := process(res, b)
					dur := time.Since(start)
					if err != nil {
						if werr == nil {
							werr = err
						}
						continue
					}
					emit(out, dur)
				}
				return werr
			})
		}
		for batch := range inputStream {
			batchChan <- batch
		}
		close(batchChan)
	}

	return int(total), g.Wait()
}

// noRes / noResClose は process が自前でセッションを確保する演算子（Expand 等）向けの
// 空リソース（engine/volcano と同様、Neo4j セッションは process スコープに閉じる）。
func noRes() struct{}       { return struct{}{} }
func noResClose(_ struct{}) {}
