package test

import (
	"context"
	"fmt"
	"log"

	"polystore_database/src/go/data_setup"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
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
	def, ok := Registry[name]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", name, AvailableWorkloads())
	}

	cypher, params, _ := def(migrator.ModeKvsToGraph, false) // migration なし
	fmt.Printf("=== ワークロード %s 実行 ===\n", name)

	if err := data_setup.Run(ctx, cfg); err != nil {
		log.Fatalf("data setup に失敗: %v", err)
	}

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

// RunWorkflow は migration → 実行（Neo4j比較）まで通しで行う。migration を伴う実験用。
func RunWorkflow(ctx context.Context, name string, cfg storage.Config, cypher string, params map[string]string, migs []migrator.MigrationConfig) {
	fmt.Printf("===== Workload %s =====\n", name)

	if len(migs) > 0 {
		res, err := RunMigration(ctx, cfg, migs)
		if err != nil {
			log.Printf("migration エラー: %v", err)
		}
		PrintMigration(res)
	}

	if cfg.Neo4j != nil {
		if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("Neo4j 実行エラー: %v", err)
		} else {
			output("Neo4j (Baseline)", r)
		}
	}

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
