package volcano_async_executor

import (
	"sync/atomic"
	"time"
)

// Metrics は演算子 1 つ分の計測。
//
// 同期版 (volcano_exec) は単一 goroutine 前提で排他不要だったが、本パッケージでは
// 複数ワーカーが同一 step へ同時に書き込むため、Processor 側の mu で保護する。
//
// 注意: Duration は「その演算子が自身の処理に費やした時間の“総和”」であり、
// 並行実行下では実時間（wall clock）を超え得る。ワーカー数で割ればおおよその
// 実時間占有に相当する。実時間は Result.Latency を参照すること。
type Metrics struct {
	StepNum  int           // 実行ツリー上の割り当て番号（葉→根で加算）
	OpType   string        // 演算子種別
	Duration time.Duration // 自身の処理に費やした累積時間（子の Next は除外／全ワーカー合算）
	RowCount int           // その演算子が出力した行数の累積
}

// recordOp は step の演算子計測を累積する。複数ワーカーから同時に呼ばれる。
func (p *Processor) recordOp(step int, opType string, dur time.Duration, rows int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	m, ok := p.metrics[step]
	if !ok {
		m = &Metrics{StepNum: step, OpType: opType}
		p.metrics[step] = m
	}
	m.Duration += dur
	m.RowCount += rows
}

// countRoundTrip は DB への 1 往復（1 クエリ / 1 点 Get など）を計上する。
// access_*.go から複数ワーカー経由で同時に呼ばれるため atomic。
func (p *Processor) countRoundTrip() { atomic.AddInt64(&p.roundTrips, 1) }

// RoundTrips は総 DB 往復回数を返す（仮説検証の一次証拠）。
func (p *Processor) RoundTrips() int64 { return atomic.LoadInt64(&p.roundTrips) }

// StepMetrics は StepNum 昇順の演算子計測一覧を返す。
func (p *Processor) StepMetrics() []Metrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Metrics, 0, len(p.metrics))
	// StepNum は 1..len で連番付与しているので順に取り出す。
	for step := 1; step <= len(p.metrics); step++ {
		if m, ok := p.metrics[step]; ok {
			out = append(out, *m)
		}
	}
	return out
}
