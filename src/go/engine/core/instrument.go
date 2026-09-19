package core

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// TraceFlow を true にすると各演算子のデータフロー（バッチ/行/クエリ/wall）を収集する。
// stream(push) / vecstream(pull) 共通のグローバルトグル。off 時 RecordFlow は即 return でゼロコスト。
var TraceFlow bool

// FlowMetric は 1 演算子分のデータフロー計測。StepMetric（出力行数＋合算Duration）を補完し、
// 「入出力バッチ数・入出力行数・DBクエリ数・実時間(wall)」を可視化する。
//   - BatchesIn/Out: 受け取った／下流へ流したバッチ数。再チャンクで Out>In になる。
//   - RowsIn/Out: 入力／出力の総行数。Expand の fan-out 倍率、Filter の選択率がここに出る。
//   - Queries: 発行した DB クエリ数（IN-batch 単位。expand/filter=入力バッチ数、scan=1、
//     projection=バッチ×fetch 数）。kvs の per-key Get は各回計上。
//   - Wall: 最初の処理開始〜最後の処理終了の実時間（並列オーバーラップを含む実占有時間）。
//     Duration（StepMetric、全ワーカー合算）と違い Wall は実際の経過時間。
type FlowMetric struct {
	Step                  int
	Op                    string
	BatchesIn, BatchesOut int64
	RowsIn, RowsOut       int64
	Queries               int64
	WallStart, WallEnd    time.Time
}

// Wall は演算子の実占有時間。
func (m FlowMetric) Wall() time.Duration {
	if m.WallStart.IsZero() || m.WallEnd.IsZero() {
		return 0
	}
	return m.WallEnd.Sub(m.WallStart)
}

// Instr は実行エンジン共通の計測器（DB往復・演算子時間・データフロー）。
// stream(push) と vecstream(pull) が同一コードを使い、メトリクスを定義上一致させるための共有型。
// RecordOp/RecordFlow/CountRoundTrip は複数 goroutine（ワーカー）から同時に呼ばれる。
type Instr struct {
	mu         sync.Mutex
	roundTrips int64 // atomic
	steps      map[int]*StepMetric
	flow       map[int]*FlowMetric
}

// NewInstr は空の計測器を返す。
func NewInstr() *Instr {
	return &Instr{
		steps: make(map[int]*StepMetric),
		flow:  make(map[int]*FlowMetric),
	}
}

// Reset は試行間で計測をクリアする。
func (i *Instr) Reset() {
	i.mu.Lock()
	i.steps = make(map[int]*StepMetric)
	i.flow = make(map[int]*FlowMetric)
	i.mu.Unlock()
	atomic.StoreInt64(&i.roundTrips, 0)
}

// CountRoundTrip は DB への 1 往復（1 クエリ / 1 点 Get など）を計上する。
func (i *Instr) CountRoundTrip() { atomic.AddInt64(&i.roundTrips, 1) }

// RoundTrips は総 DB 往復回数を返す。
func (i *Instr) RoundTrips() int64 { return atomic.LoadInt64(&i.roundTrips) }

// RecordOp は step の演算子計測を累積する（Duration 合算・OutRows 加算。InRows は -1 番兵）。
// 並列実行下では Duration は全ワーカー合算のため実時間を超え得る（実時間は Result.ExecTime）。
func (i *Instr) RecordOp(step int, op string, dur time.Duration, rows int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	m := i.steps[step]
	if m == nil {
		m = &StepMetric{Step: step, Op: op, InRows: -1}
		i.steps[step] = m
	}
	m.Duration += dur
	m.OutRows += rows
}

// StepMetrics は step 昇順の演算子計測一覧を返す。
func (i *Instr) StepMetrics() []StepMetric {
	i.mu.Lock()
	defer i.mu.Unlock()
	keys := make([]int, 0, len(i.steps))
	for k := range i.steps {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := make([]StepMetric, 0, len(keys))
	for _, k := range keys {
		out = append(out, *i.steps[k])
	}
	return out
}

// RecordFlow は step のフロー計測を累積する（TraceFlow 時のみ）。複数ワーカーから同時に呼ばれる。
func (i *Instr) RecordFlow(step int, op string, batIn, batOut, rowIn, rowOut, queries int64, t0, t1 time.Time) {
	if !TraceFlow {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	m := i.flow[step]
	if m == nil {
		m = &FlowMetric{Step: step, Op: op}
		i.flow[step] = m
	}
	m.BatchesIn += batIn
	m.BatchesOut += batOut
	m.RowsIn += rowIn
	m.RowsOut += rowOut
	m.Queries += queries
	if !t0.IsZero() && (m.WallStart.IsZero() || t0.Before(m.WallStart)) {
		m.WallStart = t0
	}
	if t1.After(m.WallEnd) {
		m.WallEnd = t1
	}
}

// FlowMetrics は step 昇順のフロー計測一覧を返す（TraceFlow=false なら空）。
func (i *Instr) FlowMetrics() []FlowMetric {
	i.mu.Lock()
	defer i.mu.Unlock()
	keys := make([]int, 0, len(i.flow))
	for k := range i.flow {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := make([]FlowMetric, 0, len(keys))
	for _, k := range keys {
		out = append(out, *i.flow[k])
	}
	return out
}

// FormatFlow はフロー表を文字列化する（葉→根の step 順）。0 は "-" で表示して構造を読みやすくする。
func FormatFlow(title string, flows []FlowMetric) string {
	var b strings.Builder
	fmt.Fprintf(&b, "=== %s ===\n", title)
	fmt.Fprintf(&b, "%-4s %-16s %6s %6s %9s %9s %6s %10s %8s\n",
		"step", "op", "batIn", "batOut", "rowIn", "rowOut", "Q", "fanout", "wall")
	dash := func(v int64) string {
		if v == 0 {
			return "-"
		}
		return fmt.Sprintf("%d", v)
	}
	for _, m := range flows {
		fanout := "-"
		if m.RowsIn > 0 {
			fanout = fmt.Sprintf("%.2fx", float64(m.RowsOut)/float64(m.RowsIn))
		}
		fmt.Fprintf(&b, "%-4d %-16s %6s %6s %9s %9s %6s %10s %8s\n",
			m.Step, m.Op, dash(m.BatchesIn), dash(m.BatchesOut),
			dash(m.RowsIn), dash(m.RowsOut), dash(m.Queries), fanout,
			m.Wall().Round(time.Millisecond))
	}
	return b.String()
}
