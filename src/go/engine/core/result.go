package core

import "time"

// StepMetric は演算子1つ分の計測。3エンジンで別々だった Metrics を統合したもの。
// 未計測フィールドの番兵値: InRows = -1（bulk のみ計測）。
type StepMetric struct {
	Step     int           // 実行順序（葉→根）
	Op       string        // 演算子種別（EntityScan, Expand, ...）
	Duration time.Duration // その演算子の実行時間
	InRows   int           // 入力行数（bulk のみ。未計測は -1）
	OutRows  int           // 出力（中間結果）行数
}

// Result は 1 クエリ実行の結果と計測。旧 bench.ExecResult / bulk.Result / volcano.Result を統合。
//
// 計測は PlanTime（parse+plan）と ExecTime（実行）を分離して持つ。
//   - 旧 stream の TotalLatency（parse 込み）は PlanTime + ExecTime（＝ TotalLatency()）に対応。
//   - 旧 bulk/volcano の Latency（parse 除外）は ExecTime に対応。
//
// エンジン固有フィールドの番兵値: RoundTrips = 0（volcano のみ）, VectorWidth = 0（vectorized のみ）。
type Result struct {
	Rows        []map[string]interface{}
	PlanTime    time.Duration
	ExecTime    time.Duration
	Steps       []StepMetric
	RoundTrips  int64  // volcano のみ（他 0）
	VectorWidth int    // vectorized のみ（他 0）
	Engine      string // 実行したエンジン名
}

// RowCount は結果行数。
func (r Result) RowCount() int { return len(r.Rows) }

// TotalLatency は parse+plan+実行の合計（旧 stream の TotalLatency と同義）。
func (r Result) TotalLatency() time.Duration { return r.PlanTime + r.ExecTime }

// SumStepTime は演算子時間の合計。
func (r Result) SumStepTime() time.Duration {
	var s time.Duration
	for _, st := range r.Steps {
		s += st.Duration
	}
	return s
}
