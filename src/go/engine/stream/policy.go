package stream

import (
	"context"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type ExecMode int

const (
	ExecFixed   ExecMode = iota // 固定ワーカー数（演算子ごと）
	ExecDynamic                 // システム全体で1つのセマフォ
)

type OpKind string

const (
	OpEntityScan      OpKind = "EntityScan"
	OpExpand          OpKind = "Expand"
	OpVarLengthExpand OpKind = "VarLengthExpand"
	OpFilter          OpKind = "Filter"
	OpProjection      OpKind = "Projection"
)

// OpConcurrency は ExecFixed 用：演算子種別ごとのワーカー数。
type OpConcurrency struct {
	Workers int
}

func (c OpConcurrency) workers() int {
	if c.Workers < 1 {
		return 1
	}
	return c.Workers
}

// ExecPolicy は実行並行戦略。
//   - ExecFixed:   PerOp[op].Workers（演算子ごとのワーカー数）
//   - ExecDynamic: GlobalMaxConcurrency（システム全体の同時DBアクセス上限。共有セマフォ）
type ExecPolicy struct {
	Mode                 ExecMode
	PerOp                map[OpKind]OpConcurrency
	Default              OpConcurrency
	GlobalMaxConcurrency int
	VectorWidth          int // record パイプラインの batch 幅（scan 払い出し・emit 再チャンクの単一の真実）
}

func (p ExecPolicy) For(op OpKind) OpConcurrency {
	if c, ok := p.PerOp[op]; ok {
		return c
	}
	return p.Default
}

func (p ExecPolicy) globalMax() int {
	if p.GlobalMaxConcurrency < 1 {
		return 1
	}
	return p.GlobalMaxConcurrency
}

// vectorWidth は record パイプラインの batch 幅（<1 なら 1）。vecstream と同じ VectorWidth ノブ。
func (p ExecPolicy) vectorWidth() int {
	if p.VectorWidth < 1 {
		return 1
	}
	return p.VectorWidth
}

// runBatches は並行戦略とリソース R の生存管理を担う。特定DBに非依存。
//   - process: 1バッチを処理して生成レコードを返す（emit はしない / DBアクセスはこの中）。
//   - 動的モードは sem（システム全体で共有）を DB 区間だけ acquire/release し、
//     解放後に runBatches が emit するためデッドロックしない。
func runBatches[R any](
	ctx context.Context,
	policy ExecPolicy,
	sem *semaphore.Weighted, // 動的モードのみ使用。全演算子で同一インスタンスを共有
	op OpKind,
	qp *Processor, // 計測の記録先（vecstream と同一意味論）
	step int, // 葉→根で採番した演算子番号
	inputStream <-chan []Record,
	outputStream chan<- []Record,
	newRes func() R,
	closeRes func(R),
	process func(res R, batch []Record) ([]Record, error),
) (int, error) {
	var total int64
	var g errgroup.Group

	emit := func(rows []Record) {
		chunk := policy.vectorWidth()
		for len(rows) > 0 {
			n := len(rows)
			if n > chunk {
				n = chunk
			}
			outputStream <- rows[:n:n]
			rows = rows[n:]
		}
	}

	// record は 1 バッチ処理の計測（Duration 合算・フロー）を記録する。vecstream の exchange と同型:
	// 1 入力バッチ = 1 DB クエリ、出力は VectorWidth 幅へ再チャンク（batOut）。
	record := func(inRows, outRows int, t0, t1 time.Time) {
		qp.recordOp(step, string(op), t1.Sub(t0), outRows)
		vw := policy.vectorWidth()
		batOut := 0
		if outRows > 0 {
			batOut = (outRows + vw - 1) / vw
		}
		qp.recordFlow(step, string(op), 1, int64(batOut), int64(inRows), int64(outRows), 1, t0, t1)
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
				t0 := time.Now()
				out, err := process(res, b) // DB区間（permit 保持）
				t1 := time.Now()
				closeRes(res)
				sem.Release(1) // emit 前に解放
				if err != nil {
					return err
				}
				record(len(b), len(out), t0, t1)
				emit(out) // permit 無しで送信
				atomic.AddInt64(&total, int64(len(out)))
				return nil
			})
		}

	default: // ExecFixed
		c := policy.For(op)
		batchChan := make(chan []Record, c.workers())
		for i := 0; i < c.workers(); i++ {
			g.Go(func() error {
				res := newRes() // ワーカー単位で使い回す
				defer closeRes(res)
				var werr error
				for b := range batchChan {
					t0 := time.Now()
					out, err := process(res, b)
					t1 := time.Now()
					if err != nil {
						if werr == nil {
							werr = err
						}
						continue
					}
					record(len(b), len(out), t0, t1)
					emit(out)
					atomic.AddInt64(&total, int64(len(out)))
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
