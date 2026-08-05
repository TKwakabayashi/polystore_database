// Package vecstream は pull 型（Volcano Iterator）の実行モデルに exchange 演算子による
// 並列性を組み合わせ、列指向バッチ（engine/volcano の Batch）で vectorized 実行する
// 実行エンジンを提供する。
//
// 位置づけ（実行モデルの3軸）:
//
//	           駆動       粒度        並行性
//	stream     push       行 []Record parallel
//	volcano    pull       列 Batch    serial
//	vectorized pull       列 Batch    serial
//	vecstream  pull       列 Batch    parallel(exchange)   ← 本パッケージ
//
// 設計:
//   - すべての演算子は Iterator (Open/Next/Close) を実装する pull 型。親が子の Next を呼ぶ。
//   - DB 往復を伴う演算子（Filter/Expand/VarLengthExpand）は exchangeIterator で包む。
//     exchange は「子を直列 pull する driver goroutine」＋「W ワーカー」で構成され、
//     各ワーカーは Neo4j セッションを 1 本だけ生成してバッチ間で使い回す（session 再利用）。
//   - 各演算子の 1 バッチ処理（process）と access 層・tail・pushdown は engine/volcano から
//     移植（列 Batch・id.UUID・往復計測を流用）。並行化と session 生存を知るのは exchange だけ。
//   - VectorWidth を ExecPolicy に持たせ、scan 払い出しと exchange 出力の再チャンク（split）を
//     一元化する（バッチ幅を一律 VectorWidth に保つ）。
//
// 既存 engine/stream・engine/volcano には一切手を加えない（自己完結コピー）。
package vecstream

import (
	"context"

	uid "polystore_database/src/go/id"
)

// Iterator は pull 型（Volcano）の演算子インターフェース。
type Iterator interface {
	// Open は自身と上流（子）を初期化する（exchange はここで並列パイプラインを起動）。
	Open(ctx context.Context) error
	// Next は次のバッチを返す。もう無ければ (nil, nil)。空バッチ (n==0) は返さない。
	Next(ctx context.Context) (*Batch, error)
	// Close は資源（子・ワーカー・セッション）を解放する。
	Close(ctx context.Context) error
}

// Batch は演算子間を流れる中間結果（列指向 / struct-of-arrays）。
//   - cols[slot] が長さ n の列（値は uuid スロット）。
//   - 空バッチ (n==0) は演算子間では流さない（EOF と区別するため）。
//
// engine/volcano/iterator.go の Batch をコピーしたもの。1 つの Batch は生成 goroutine から
// 下流へ所有権ごと引き渡され、複数 goroutine から同時に触られることはない。
type Batch struct {
	cols [][]uid.UUID
	n    int
}

// newBatch は slotCount 個の空列を持つ Batch を確保する。
func newBatch(slotCount, capRows int) *Batch {
	cols := make([][]uid.UUID, slotCount)
	for i := range cols {
		cols[i] = make([]uid.UUID, 0, capRows)
	}
	return &Batch{cols: cols, n: 0}
}

// slotCount は列（スロット）数。
func (b *Batch) slotCount() int { return len(b.cols) }

// appendRow は 1 行（長さ slotCount 以下の []uid.UUID）を各列末尾へ追加する。
func (b *Batch) appendRow(row []uid.UUID) {
	for s := range b.cols {
		var v uid.UUID
		if s < len(row) {
			v = row[s]
		}
		b.cols[s] = append(b.cols[s], v)
	}
	b.n++
}

// row は i 行目を []uid.UUID として復元する。
func (b *Batch) row(i int) []uid.UUID {
	out := make([]uid.UUID, len(b.cols))
	for s := range b.cols {
		out[s] = b.cols[s][i]
	}
	return out
}

// get は (行 i, スロット s) の値を返す。
func (b *Batch) get(i, s int) uid.UUID { return b.cols[s][i] }

// split は b を width 行ごとの Batch 列へ再チャンクする。emit 時に出力を VectorWidth 幅へ
// 揃えるために使う（engine/stream の emit chunk=2000 に対応する列指向版）。
// 空バッチ (n==0) は返さない。width<1 は 1 とみなす。
func split(b *Batch, width int) []*Batch {
	if b == nil || b.n == 0 {
		return nil
	}
	if width < 1 {
		width = 1
	}
	if b.n <= width {
		return []*Batch{b}
	}
	sc := b.slotCount()
	out := make([]*Batch, 0, (b.n+width-1)/width)
	for start := 0; start < b.n; start += width {
		end := start + width
		if end > b.n {
			end = b.n
		}
		sub := &Batch{cols: make([][]uid.UUID, sc), n: end - start}
		for s := 0; s < sc; s++ {
			// 元スライスを共有せず切り出す（下流での append 汚染を避ける）。
			sub.cols[s] = b.cols[s][start:end:end]
		}
		out = append(out, sub)
	}
	return out
}
