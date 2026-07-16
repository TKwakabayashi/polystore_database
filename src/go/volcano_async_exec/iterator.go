// Package volcano_async_executor は、同期版 pull 実行系 (src/go/volcano_exec) の
// 非同期（並行）版を提供する。
//
// 位置づけ:
//
//	          駆動方向   処理粒度        並行性
//	stream_exec        push     vector          parallel
//	volcano_exec       pull     tuple/vector    serial
//	volcano_async_exec pull     tuple/vector    parallel   ← 本パッケージ
//
// 設計:
//   - 演算子インターフェースは同期版と同一の Iterator (Open/Next/Close) を保つ。
//     並行性を知っているのは asyncDriver（Volcano 論文の exchange 演算子に相当）だけで、
//     Filter/Expand の 1 バッチ処理そのもの（process）は同期版から変更していない。
//   - 処理粒度は同期版と同じ Mode（Volcano: vectorWidth=1 / Vectorized: vectorWidth=N）。
//   - 並行化方式は AsyncMode で切り替える:
//   - AsyncExchange : W ワーカーが process を並行実行。順序は保存されない。
//   - AsyncPrefetch : 1 ワーカーで深さ 1 の先読みのみ。順序は保存される。
//
// 既存 volcano_exec / stream_exec / plan / logical_plan / storage / codec には
// 一切手を加えず、plan.PlanNode を消費して iterator ツリーを構築する。
package volcano_async_executor

import "context"

// Batch は演算子間を流れる中間結果（列指向）。
//   - cols[slot] が長さ n の列（値は uuid などの文字列スロット）。
//   - 空バッチ (n==0) は演算子間では流さない（EOF と区別するため）。
//
// 1 つの Batch は生成した goroutine から下流へ所有権ごと引き渡され、
// 複数 goroutine から同時に触られることはない（共有せず move する）。
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

// Iterator は pull 型（Volcano）の演算子インターフェース。同期版と同一。
type Iterator interface {
	// Open は自身と上流（子）を初期化する。
	Open(ctx context.Context) error
	// Next は次のバッチを返す。もう無ければ (nil, nil)。
	Next(ctx context.Context) (*Batch, error)
	// Close は資源を解放する。
	Close(ctx context.Context) error
}

// processFn は 1 バッチを処理して 1 バッチを返す純粋関数（DB 往復はこの中）。
// asyncDriver がこれを並行実行する。同期版の *Iterator.process と同じ本体。
type processFn func(in *Batch) (*Batch, error)
