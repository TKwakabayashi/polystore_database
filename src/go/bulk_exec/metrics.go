package bulk_executor

import "time"

// Metrics は演算子 1 つ分の計測（逐次実行なので排他不要）。
type Metrics struct {
	StepNum  int           // 実行ツリー上の割り当て番号（葉→根で加算）
	OpType   string        // 演算子種別
	InRows   int           // その演算子への入力行数（EntityScan は 0）
	OutRows  int           // その演算子が出力した行数
	Duration time.Duration // その演算子が「自身の処理」に費やした時間（子の実行は除外）
}

// recordOp は step の演算子計測を記録する。
func (p *Processor) recordOp(step int, opType string, dur time.Duration, in, out int) {
	m, ok := p.metrics[step]
	if !ok {
		m = &Metrics{StepNum: step, OpType: opType}
		p.metrics[step] = m
	}
	m.InRows += in
	m.OutRows += out
	m.Duration += dur
}

// countRoundTrip は DB への 1 往復（1 クエリ / 1 点 Get など）を計上する。
func (p *Processor) countRoundTrip() { p.roundTrips++ }

// RoundTrips は総 DB 往復回数を返す。
func (p *Processor) RoundTrips() int64 { return p.roundTrips }

// StepMetrics は StepNum 昇順の演算子計測一覧を返す。
func (p *Processor) StepMetrics() []Metrics {
	out := make([]Metrics, 0, len(p.metrics))
	for step := 1; step <= len(p.metrics); step++ {
		if m, ok := p.metrics[step]; ok {
			out = append(out, *m)
		}
	}
	return out
}
