package volcano_async_executor

// AsyncMode は並行化の方式。処理粒度（Mode: Volcano/Vectorized）とは直交する軸。
type AsyncMode int

const (
	// AsyncExchange は Volcano 論文の exchange 演算子に相当する並行化。
	// W ワーカーが process（DB 往復）を並行実行し、往復レイテンシを隠蔽する。
	// バッチの順序は保存されない（ORDER BY は tail の Sort が担保する）。
	AsyncExchange AsyncMode = iota
	// AsyncPrefetch は深さ 1 の先読み（ダブルバッファリング）。
	// ワーカーは 1 本なので順序は保存されるが、隠蔽できるのは 1 バッチ分だけ。
	// 実体は workers=1 / buffer=1 に縮退させた exchange。
	AsyncPrefetch
)

func (a AsyncMode) String() string {
	if a == AsyncPrefetch {
		return "Prefetch"
	}
	return "Exchange"
}

// ExecMode は並行度の決め方。stream_exec.ExecPolicy と同じ意味論。
type ExecMode int

const (
	ExecFixed   ExecMode = iota // 固定ワーカー数（演算子ごと）
	ExecDynamic                 // システム全体で 1 つのセマフォ
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

// ExecPolicy は実行並行戦略。stream_exec と同じ設定値を与えれば、
// push(stream) と pull(async volcano) を同一並行度で比較できる。
//   - ExecFixed:   PerOp[op].Workers（演算子ごとのワーカー数）
//   - ExecDynamic: GlobalMaxConcurrency（システム全体の同時 DB アクセス上限。共有セマフォ）
type ExecPolicy struct {
	Mode                 ExecMode
	PerOp                map[OpKind]OpConcurrency
	Default              OpConcurrency
	GlobalMaxConcurrency int
}

// DefaultExecPolicy は stream_exec.NewQueryProcessorWithConfig と同じ既定値。
// 比較の初期条件を揃えるためにミラーしている。
func DefaultExecPolicy() ExecPolicy {
	return ExecPolicy{
		Mode:                 ExecDynamic,
		Default:              OpConcurrency{Workers: 2},
		GlobalMaxConcurrency: 8,
		PerOp: map[OpKind]OpConcurrency{
			OpExpand:          {Workers: 4},
			OpVarLengthExpand: {Workers: 2},
			OpFilter:          {Workers: 4},
			OpProjection:      {Workers: 4},
		},
	}
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

// workersFor は op に割り当てるワーカー数。
//   - AsyncPrefetch では常に 1（深さ 1 の先読みに縮退）。
//   - ExecDynamic では同時 DB 数はセマフォが握るので、ワーカー数は
//     セマフォを待てる本数として globalMax を上限に取る。
func (p ExecPolicy) workersFor(op OpKind, async AsyncMode) int {
	if async == AsyncPrefetch {
		return 1
	}
	if p.Mode == ExecDynamic {
		return p.globalMax()
	}
	return p.For(op).workers()
}

// bufferFor は演算子の出力チャネル長。有界にすることで下流の需要に上流が
// 引きずられる（= pull の backpressure が効く）状態を保つ。
// 無界にすると事実上 push 型に退化するため、必ず有界にすること。
func (p ExecPolicy) bufferFor(op OpKind, async AsyncMode) int {
	if async == AsyncPrefetch {
		return 1 // 深さ 1 の先読み
	}
	return p.workersFor(op, async)
}
