package store

// limits.go はストア固有の**物理的制約**を 1 か所に集約する。
// plan/capability.go（そのストアが何をネイティブ実行できるか）と対になる
// 「どこまで一度に投げられるか」のテーブル。
//
// ここに置くのは「ストアの事実」であって設定ではない。ユーザが性能のために動かすノブは
// settings 側（MaterializeChunkSize / IntegrateRowThreshold）に置き、実効値は
// engine/core が min(希望値, 物理上限) で clamp する。

// Limits は 1 クエリへ載せられる識別子数の制約。
type Limits struct {
	// MaxInList は IN / $in に載せられる識別子数の**硬い上限**（0 = 明文化された上限なし）。
	// これを超えるとエラーになる種類の制約だけを入れる。
	MaxInList int

	// DefaultChunk はそのストアで既定とするチャンクサイズ（0 = チャンクしない）。
	// 硬い上限ではなく「これ以上は遅くなる/危ないので分割する」運用上の既定値。
	DefaultChunk int
}

// 各ストアの制約。硬い上限（MaxInList）と運用上の既定（DefaultChunk）を区別している点に注意:
//
//   - relational(MySQL): プリペアドステートメントのプレースホルダ数はプロトコル上 16bit で
//     65535 が上限。加えて max_allowed_packet がクエリ長を縛る。→ 硬い上限あり。
//   - document(Mongo): クエリ文書全体が BSON 16MB 制限に収まる必要がある。uuid 1 件あたり
//     およそ 40B なので概算 39 万件が理論上限。余裕を見て保守的に設定する。→ 硬い上限あり。
//   - columnar(Cassandra): プロトコル上の件数上限は無いが、大きな IN はコーディネータの
//     fan-out を招くアンチパターン。→ 硬い上限なし・既定値のみ。
//   - graph(Neo4j): 明文化された件数上限は無いが、大きなパラメータはメモリとプラン生成コストに
//     効く。→ 硬い上限なし・既定値のみ。
//   - kvs(LevelDB): IN の概念が無くキー直引きのため対象外。
var limits = map[Kind]Limits{
	Graph:      {MaxInList: 0, DefaultChunk: 50000},
	Relational: {MaxInList: 65535, DefaultChunk: 1000},
	Document:   {MaxInList: 200000, DefaultChunk: 50000},
	Columnar:   {MaxInList: 0, DefaultChunk: 500},
	Kvs:        {MaxInList: 0, DefaultChunk: 0},
}

// LimitsOf は store k の制約を返す（未登録は制約なし扱い）。
func LimitsOf(k Kind) Limits { return limits[k] }
