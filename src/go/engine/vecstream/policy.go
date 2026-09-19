package vecstream

// ExecMode は並行度の決め方（engine/stream.ExecPolicy と同じ意味論）。
type ExecMode int

const (
	ExecFixed   ExecMode = iota // 固定ワーカー数（演算子ごと）
	ExecDynamic                 // システム全体で1つのセマフォ
)

// OpKind は並行度を設定する演算子の種別。
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

// ExecPolicy は実行並行戦略 ＋ ベクトル幅。engine/stream.ExecPolicy に VectorWidth を足したもの。
//   - ExecFixed:   PerOp[op].Workers（演算子ごとのワーカー数）
//   - ExecDynamic: GlobalMaxConcurrency（システム全体の同時DBアクセス上限。共有セマフォ）
//   - VectorWidth: scan の払い出し幅 ＝ emit の再チャンク幅（単一の真実）。
type ExecPolicy struct {
	Mode                 ExecMode
	PerOp                map[OpKind]OpConcurrency
	Default              OpConcurrency
	GlobalMaxConcurrency int
	VectorWidth          int
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

// vectorWidth はベクトル幅（<1 なら 1）。
func (p ExecPolicy) vectorWidth() int {
	if p.VectorWidth < 1 {
		return 1
	}
	return p.VectorWidth
}

// workersFor は op の exchange ワーカー数（＝再利用する Neo4j セッション本数）。
// pull+exchange では安定したワーカーが session を使い回すため、演算子ごとの固定数を返す。
// システム全体の同時 DB 数は別途 GlobalMaxConcurrency の共有セマフォが抑える。
func (p ExecPolicy) workersFor(op OpKind) int {
	return p.For(op).workers()
}

// bufferFor は exchange 出力チャネルのバッファ長。有界にして pull の需要駆動性
// （backpressure）を保つ（無界にすると push に退化しメモリが膨らむ）。
func (p ExecPolicy) bufferFor(op OpKind) int {
	w := p.workersFor(op)
	if w < 1 {
		return 1
	}
	return w
}
