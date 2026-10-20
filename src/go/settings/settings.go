// Package settings は「内部挙動の切り替え」を1か所に集約する。
//
// ここは CLI フラグではなくソースコードで切り替える設定の唯一の置き場所。
// 値を変えたら再ビルドして反映する。プロジェクト内 import を持たない葉パッケージ
// （将来 store.Kind 導入時に store のみ import する予定）なので、どの層からも参照できる。
//
// ベンチマーク実行系だけは、スイープのためにこれらの var を実行時に一時上書きする
// （唯一の合法な実行時書き換え箇所。前後で必ず元値へ戻す）。
package settings

// ===== 実行エンジン =====
// -mode run で使う実行モデル。P4 のエンジンレジストリで消費される。
type EngineKind string

const (
	EngineStream     EngineKind = "stream"
	EngineBulk       EngineKind = "bulk"
	EngineVolcano    EngineKind = "volcano"
	EngineVectorized EngineKind = "vectorized"
	EngineVecStream  EngineKind = "vecstream" // push × 列 Batch × parallel
)

var (
	Engine     = EngineStream // ★ -mode run の実行モデル（刷新版 executor の反復モデル）
	VectorSize = 1024         // vectorized のベクトル長
)

// MaterializeChunkSize は ID 材料化で 1 クエリへ載せる識別子数の希望値（0 = ストア既定を使う）。
// 実測スイープ用の全体上書きノブ。実効値は executor の chunkSize が
// min(この値 or ストア既定, ストアの物理上限 store.LimitsOf(k).MaxInList) で clamp する。
var MaterializeChunkSize = 0

// UniqueKeyProps は「等値条件で 1 件に絞れる」とみなすプロパティ名（エンティティの一意キー）。
// 刷新版の翻訳（logical.Translate）が、パターンの起点を選ぶときに統計の代わりに使う。
var UniqueKeyProps = []string{"id"}

// StoreFetchParallelism は、1 つの演算子が IN-list をチャンクに分けて同じストアへ投げるときの
// 同時発行数（刷新版の executor）。1 なら直列。
var StoreFetchParallelism = 4

// WeightedRows は、可変長の展開を含む graph のクエリの結果を、同じ出力の組ごとに件数（重み）付きの
// 1 行へまとめて返させるか（刷新版）。経路ごとの重複で中間行が爆発する形（Q11 など）を抑える。結果は不変。
var WeightedRows = true

// GraphFetchParallelism は graph（Neo4j）へのチャンクの同時発行数。graph のチャンクは 5 万件の起点から
// 可変長トラバーサルを走らせることがあり、並列にすると Neo4j がメモリ不足で落ちる（ldbc_sf1 の BI9 で
// heap 4G / pagecache 1G の構成が OOM-kill された）ので既定は直列。
var GraphFetchParallelism = 1

// ===== 計測（bench） =====
const (
	Warmup = 3  // 計測から除外する先頭ウォームアップ回数
	Trials = 10 // Warmup の後に実行し平均を取る計測回数
)

var (
	// ★ bench / bench-models のスイープ軸（旧 -placements/-models フラグ）。
	// BenchPlacements は P3b で []store.Kind へ型付け予定（現状は文字列）。
	BenchPlacements = []string{"graph", "rdb", "doc", "col", "kvs"}
	BenchModels     = []string{"stream", "bulk", "volcano", "vectorized", "vecstream"}

	MigrationDeleteSource = true // ★ 移行成功後にソース側を削除する（旧 -delete フラグ）
)

// ===== verify（-mode verify：配置×pushdown×エンジン×バッチの性能検証マトリクス）=====
var (
	// VerifyEngines は verify マトリクスで横断計測する自作エンジン（Neo4j 直接は別途基準として測る）。
	VerifyEngines = []string{"stream", "bulk", "vecstream"}
	// VerifyBatchSizes は VectorSize を振るスイープ点。stream/vecstream のバッチ幅に効く
	// （bulk はバッチ非依存のため 1 回のみ計測）。
	VerifyBatchSizes = []int{256, 1024, 4096}
)

// ===== 出力 =====
type OutputFormat string

const (
	FormatRows   OutputFormat = "rows"   // 結果を1件ずつ
	FormatTiming OutputFormat = "timing" // 全体実行時間+件数
	FormatDetail OutputFormat = "detail" // 演算子ごとの時間・中間件数（bulk用）
)

type Target string

const (
	TargetCustom Target = "custom" // 自作システムのみ
	TargetNeo4j  Target = "neo4j"  // Neo4j のみ
	TargetBoth   Target = "both"   // 両方（比較）
)

var (
	Format    = FormatTiming // ★ -mode run の出力形式（旧 SelectedFormat）
	RunTarget = TargetCustom // ★ -mode run の実行対象（旧 SelectedTarget）
)

// ===== プロファイル =====
// 採取スコープ。profile パッケージが参照し、有効なスコープだけ pprof を採取する。
type ProfileScope string

const (
	ScopeRun       ProfileScope = "run"       // プロセス全体
	ScopeEngine    ProfileScope = "engine"    // クエリ実行のみ（旧 ProfileCustomRun 相当）
	ScopeMigration ProfileScope = "migration" // 移行処理
	ScopePlanner   ProfileScope = "planner"   // parse + plan
	ScopeSetup     ProfileScope = "setup"     // データセットアップ
)

// ★ 有効化したいスコープを true に。空 = プロファイル無効。
// 既定は engine（旧 ProfileCustomRun=true 相当）。bench 実行系は計測歪みを避けるため
// 実行中だけ engine スコープを false に一時上書きする。
var ProfileScopes = map[ProfileScope]bool{
	ScopeEngine: true,
}

// ===== 生成物の出力先 =====
// cwd = src/go での実行を前提とした相対パス（リポジトリルートの results/）。
var ResultsDir = "../../results"
