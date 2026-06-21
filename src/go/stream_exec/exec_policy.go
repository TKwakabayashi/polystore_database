package stream_executor

import (
	"sync"
	"sync/atomic"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ExecMode int

const (
	ExecFixed   ExecMode = iota // 固定ワーカー数
	ExecDynamic                 // セマフォによる動的割り当て
)

// OpKind は演算子の種別。種別ごとに並行度を設定する。
type OpKind string

const (
	OpEntityScan      OpKind = "EntityScan"
	OpExpand          OpKind = "Expand"
	OpVarLengthExpand OpKind = "VarLengthExpand"
	OpFilter          OpKind = "Filter"
	OpProjection      OpKind = "Projection"
)

// OpConcurrency は 1 演算子種別あたりの並行度設定。
type OpConcurrency struct {
	Workers        int // ExecFixed: 固定ワーカー数
	MaxConcurrency int // ExecDynamic: 同時実行（=同時セッション）上限
}

func (c OpConcurrency) workers() int {
	if c.Workers < 1 {
		return 1
	}
	return c.Workers
}

func (c OpConcurrency) maxConc() int {
	if c.MaxConcurrency < 1 {
		return 1
	}
	return c.MaxConcurrency
}

// ExecPolicy は実行並行戦略。Mode で固定/動的を切り替え、
// 並行度は演算子種別ごと(PerOp)に設定する（未設定は Default）。
type ExecPolicy struct {
	Mode    ExecMode
	PerOp   map[OpKind]OpConcurrency
	Default OpConcurrency
}

// For は指定演算子種別の並行度を返す（未設定なら Default）。
func (p ExecPolicy) For(op OpKind) OpConcurrency {
	if c, ok := p.PerOp[op]; ok {
		return c
	}
	return p.Default
}

// batchFunc は 1 バッチを処理し、処理した行数を返す。
// セッションは runBatches が用意・クローズするので、fn 側で閉じてはならない。
type batchFunc func(sess neo4j.SessionWithContext, batch []Record) (int, error)

// runBatches は inputStream の各バッチを、演算子種別 op に設定された並行度で処理する。
//   - ExecFixed:   PerOp[op].Workers 個のワーカーが batchChan を消費。
//     各ワーカーはセッションを 1 本だけ開いて使い回す。
//   - ExecDynamic: バッチ毎に goroutine を起こすが、セマフォで同時数を
//     PerOp[op].MaxConcurrency に制限。各 goroutine が専用セッションを持つ。
func (qp *QueryProcessor) runBatches(op OpKind, inputStream <-chan []Record, fn batchFunc) (int, error) {
	c := qp.exec.For(op)

	var total int64
	var firstErr error
	var errOnce sync.Once
	setErr := func(e error) {
		if e != nil {
			errOnce.Do(func() { firstErr = e })
		}
	}
	newSess := func() neo4j.SessionWithContext {
		return qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	}

	switch qp.exec.Mode {
	case ExecDynamic:
		sem := make(chan struct{}, c.maxConc())
		var wg sync.WaitGroup
		for batch := range inputStream {
			sem <- struct{}{} // 上限到達時はここでブロック（負荷に応じた抑制）
			wg.Add(1)
			go func(b []Record) {
				defer wg.Done()
				defer func() { <-sem }()
				sess := newSess()
				defer sess.Close(qp.ctx)
				n, err := fn(sess, b)
				atomic.AddInt64(&total, int64(n))
				setErr(err)
			}(batch)
		}
		wg.Wait()

	default: // ExecFixed
		batchChan := make(chan []Record, c.workers())
		var wg sync.WaitGroup
		for i := 0; i < c.workers(); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sess := newSess() // ワーカー単位で 1 本だけ開いて使い回す
				defer sess.Close(qp.ctx)
				for b := range batchChan {
					n, err := fn(sess, b)
					atomic.AddInt64(&total, int64(n))
					setErr(err)
				}
			}()
		}
		for batch := range inputStream {
			batchChan <- batch
		}
		close(batchChan)
		wg.Wait()
	}

	return int(total), firstErr
}
