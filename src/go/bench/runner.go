package bench

import (
	"context"
	"fmt"
	"log"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// 出力形式・実行対象は settings.Format / settings.RunTarget で切り替える（型・定数も settings）。

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

	if settings.RunTarget == settings.TargetCustom || settings.RunTarget == settings.TargetBoth {
		r, err := RunEngine(ctx, cfg, settings.Engine, cypher, params)
		if err != nil {
			log.Fatalf("%v", err)
		}
		output("Custom System", r)
	}

	if settings.RunTarget == settings.TargetNeo4j || settings.RunTarget == settings.TargetBoth {
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
	if r, err := RunEngine(ctx, cfg, settings.Engine, cypher, params); err != nil {
		log.Printf("自作システム 実行エラー: %v", err)
	} else {
		output("Custom System", r)
	}
}

// output は settings.Format に従って結果を出力する（既存/自作の両方に適用）。
func output(title string, r core.Result) {
	switch settings.Format {
	case settings.FormatTiming:
		PrintTiming(title, r)
	case settings.FormatDetail:
		PrintTiming(title, r)
		PrintDetail(title, r)
	default: // FormatRows
		PrintRows(r)
	}
}
