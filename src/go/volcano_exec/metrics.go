package volcano_executor

import "time"

// Metrics は演算子 1 つ分の計測（pull 実行は単一 goroutine なので排他不要）。
type Metrics struct {
	StepNum  int           // 実行ツリー上の割り当て番号（葉→根で加算）
	OpType   string        // 演算子種別
	Duration time.Duration // その演算子が「自身の処理」に費やした累積時間（子の Next は除外）
	RowCount int           // その演算子が出力した行数の累積
}

// recordOp は step の演算子計測を累積する。
func (p *Processor) recordOp(step int, opType string, dur time.Duration, rows int) {
	m, ok := p.metrics[step]
	if !ok {
		m = &Metrics{StepNum: step, OpType: opType}
		p.metrics[step] = m
	}
	m.Duration += dur
	m.RowCount += rows
}

// countRoundTrip は DB への 1 往復（1 クエリ / 1 点 Get など）を計上する。
func (p *Processor) countRoundTrip() { p.roundTrips++ }

// RoundTrips は総 DB 往復回数を返す（仮説検証の一次証拠）。
func (p *Processor) RoundTrips() int64 { return p.roundTrips }

// StepMetrics は StepNum 昇順の演算子計測一覧を返す。
func (p *Processor) StepMetrics() []Metrics {
	out := make([]Metrics, 0, len(p.metrics))
	// StepNum は 1..len で連番付与しているので順に取り出す。
	for step := 1; step <= len(p.metrics); step++ {
		if m, ok := p.metrics[step]; ok {
			out = append(out, *m)
		}
	}
	return out
}
