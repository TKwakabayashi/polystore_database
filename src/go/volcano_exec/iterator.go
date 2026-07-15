// Package volcano_executor は、既存の push 型ストリーミング実行系
// (src/go/stream_exec) とは独立した、pull 型（Volcano）実行モデルを提供する。
//
// 設計:
//   - すべての演算子は Iterator (Open/Next/Close) を実装する pull 型。
//   - Next は列指向の Batch を返す。EOF は (nil, nil)。
//   - Volcano モード    : vectorWidth = 1（tuple-at-a-time）。Expand は 1 タプル = 1 往復。
//   - Vectorized モード : vectorWidth = N。Expand は 1 ベクトル = 1 往復（往復 = ⌈rows/N⌉）。
//
// 既存 stream_exec / plan / logical_plan / storage / codec には一切手を加えず、
// plan.PlanNode を消費して iterator ツリーを構築する。
package volcano_executor

import "context"

// Batch は演算子間を流れる中間結果（列指向）。
//   - cols[slot] が長さ n の列（値は uuid などの文字列スロット）。
//   - 空バッチ (n==0) は演算子間では流さない（EOF と区別するため）。
type Batch struct {
	cols [][]string
	n    int
}

// newBatch は slotCount 個の空列を持つ Batch を確保する。
func newBatch(slotCount, capRows int) *Batch {
	cols := make([][]string, slotCount)
	for i := range cols {
		cols[i] = make([]string, 0, capRows)
	}
	return &Batch{cols: cols, n: 0}
}

// slotCount は列（スロット）数。
func (b *Batch) slotCount() int { return len(b.cols) }

// appendRow は 1 行（長さ slotCount 以下の []string）を各列末尾へ追加する。
func (b *Batch) appendRow(row []string) {
	for s := range b.cols {
		var v string
		if s < len(row) {
			v = row[s]
		}
		b.cols[s] = append(b.cols[s], v)
	}
	b.n++
}

// row は i 行目を []string として復元する。
func (b *Batch) row(i int) []string {
	out := make([]string, len(b.cols))
	for s := range b.cols {
		out[s] = b.cols[s][i]
	}
	return out
}

// get は (行 i, スロット s) の値を返す。
func (b *Batch) get(i, s int) string { return b.cols[s][i] }

// Iterator は pull 型（Volcano）の演算子インターフェース。
type Iterator interface {
	// Open は自身と上流（子）を初期化する。
	Open(ctx context.Context) error
	// Next は次のバッチを返す。もう無ければ (nil, nil)。
	Next(ctx context.Context) (*Batch, error)
	// Close は資源を解放する。
	Close(ctx context.Context) error
}
