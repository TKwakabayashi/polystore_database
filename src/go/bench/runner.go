package bench

import (
	"context"
	"fmt"
	"log"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// 出力形式（手動で切り替え）
type OutputFormat string

const (
	FormatRows   OutputFormat = "rows"   // 結果を1件ずつ
	FormatTiming OutputFormat = "timing" // 全体実行時間+件数
	FormatDetail OutputFormat = "detail" // 演算子ごとの時間・中間件数（bulk用）
)

// 実行対象（手動で切り替え）
type Target string

const (
	TargetCustom Target = "custom" // 自作システムのみ
	TargetNeo4j  Target = "neo4j"  // Neo4j のみ
	TargetBoth   Target = "both"   // 両方（比較）
)

// ★ ここを書き換えて 出力形式 / 実行対象 を切り替える
const (
	SelectedFormat = FormatTiming
	SelectedTarget = TargetCustom
)

// RunWorkloadByName は名前でワークロードを引き、data_setup 後に実行して出力する。
func RunWorkloadByName(ctx context.Context, name string, cfg storage.Config) {
	def, ok := workload.Registry[name]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", name, workload.AvailableWorkloads())
	}

	cypher, params, _ := def(migrator.ModeKvsToGraph, false) // migration なし
	fmt.Printf("=== ワークロード %s 実行 ===\n", name)

	// 注意: ここでは data_setup を呼ばない（呼ぶと他4ストアが初期化され、
	// 直前に移行したデータが消えてしまう）。ベースライン初期化は `-mode setup` で明示的に行う。

	if SelectedTarget == TargetCustom || SelectedTarget == TargetBoth {
		r, err := RunCustom(ctx, cfg, cypher, params)
		if err != nil {
			log.Fatalf("%v", err)
		}
		output("Custom System", r)
	}

	if SelectedTarget == TargetNeo4j || SelectedTarget == TargetBoth {
		if cfg.Neo4j == nil {
			log.Printf("Neo4j 設定が無いためスキップ")
		} else if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("Neo4j 実行エラー: %v", err)
		} else {
			output("Neo4j (Baseline)", r)
		}
	}
}

// RunWorkflow は Neo4jベースライン計測 → migration → 自作システム計測 の順で通しで行う。
// ベースラインを migration より前に測るのは、DeleteSource 有効時に移行でソース(graph)の
// 該当プロパティが削除され、後で測ると Neo4j 側が不完全なデータになってしまうのを避けるため。
func RunWorkflow(ctx context.Context, name string, cfg storage.Config, cypher string, params map[string]string, migs []migrator.MigrationConfig) {
	fmt.Printf("===== Workload %s =====\n", name)

	// 1) Neo4j ベースラインは移行前の完全な graph 上で計測する
	if cfg.Neo4j != nil {
		if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("Neo4j 実行エラー: %v", err)
		} else {
			output("Neo4j (Baseline)", r)
		}
	}

	// 2) 移行（DeleteSource 有効ならソースから該当データを削除）
	if len(migs) > 0 {
		res, err := RunMigration(ctx, cfg, migs)
		if err != nil {
			log.Printf("migration エラー: %v", err)
		}
		PrintMigration(res)
	}

	// 3) 自作システムは移行後の配置で計測
	if r, err := RunCustom(ctx, cfg, cypher, params); err != nil {
		log.Printf("自作システム 実行エラー: %v", err)
	} else {
		output("Custom System", r)
	}
}

// output は SelectedFormat に従って結果を出力する（既存/自作の両方に適用）。
func output(title string, r ExecResult) {
	switch SelectedFormat {
	case FormatTiming:
		PrintTiming(title, r)
	case FormatDetail:
		PrintTiming(title, r)
		PrintDetail(title, r)
	default: // FormatRows
		PrintRows(r)
	}
}
