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
	Engine     = EngineStream // ★ -mode run の実行モデル（P4 で有効化）
	VectorSize = 1024         // vectorized のベクトル長
)

// ===== プランナ：集約 pushdown 方針 =====
type PushdownMode int

const (
	PushdownAuto        PushdownMode = iota // 単一ストアに解決できれば委譲、散在ならエンジン
	PushdownForceEngine                     // 常にコーディネータ（自作エンジン）
)

var Pushdown = PushdownAuto // ★ pushdown 方針（旧 planner.SelectedPushdown）

// 以下は Pushdown==PushdownAuto のときだけ意味を持つ独立トグル（projection / aggregation を個別制御）。
var (
	// TailPushdown: traversal 由来の中間 UUID を単一の非 graph ストア（現状 relational/MySQL）の
	// 一時テーブルへロードし、RETURN 句 tail（Projection/Aggregate/GroupBy/Sort/Limit）を
	// そのストアのネイティブエンジンで実行する実験的パス。OFF なら従来通り tail を自作エンジンで計算する。
	// tail が参照する全プロパティが単一ストアに解決でき、そのストアが必要な能力を持つ場合のみ発火する。
	TailPushdown = false
	// AggregationPushdown: 集約をストアへ委譲するか。OFF なら委譲せず engine で計算する
	// （fusion の BuildPushdownPlan が判定に使用済み）。
	AggregationPushdown = true
	// ProjectionPushdown: 同一ストアの projection 列をネイティブクエリの SELECT/RETURN へ
	// 畳み込むか。projection-only 畳み込み（P6）で有効化予定。現状は未配線（no-op）。
	ProjectionPushdown = true
)

// IntegrateRowThreshold は統合演算子の実行時方式選択の閾値
// （hash join / IN-list 押し込み / nested-loop の分岐）。
var IntegrateRowThreshold = 1024

// ===== 計測（bench） =====
const (
	Warmup = 3  // 計測から除外する先頭ウォームアップ回数
	Trials = 10 // Warmup の後に実行し平均を取る計測回数
)

var (
	// ★ bench / bench-models のスイープ軸（旧 -placements/-pushdowns/-models フラグ）。
	// BenchPlacements は P3b で []store.Kind へ型付け予定（現状は文字列）。
	BenchPlacements = []string{"graph", "rdb", "doc", "col", "kvs"}
	BenchPushdowns  = []string{"auto", "engine"}
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
