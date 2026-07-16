// Command bulkbench は全件マテリアライズ実行系(bulk_exec)を独立バイナリとして実行する。
// 既存 main.go を変更せずに、各演算子が「何件に対して」「どれだけの時間」かかったかを測る。
//
// 例:
//
//	go run ./bulk_exec/cmd/bulkbench -workload Q11 -trials 3             # 既定=engine: 演算子(段階)別に計測
//	go run ./bulk_exec/cmd/bulkbench -workload Q11 -pushdown auto        # 集約を単一ストアへ委譲（Pushdown 1 演算子のみ）
//
// 注: -pushdown auto は集約全体を 1 ストアへ委譲するため、計測は Pushdown 演算子 1 つに集約される。
// 各段階（Scan/Filter/Expand/Projection/Aggregate/…）の内訳を見たい場合は既定の engine を使う。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	bulk "polystore_database/src/go/bulk_exec"
	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/test"
)

// pushdown モード名 → planner.PushdownMode（test/benchmark.go と同じ命名）。
var pushdownModeMap = map[string]planner.PushdownMode{
	"auto":   planner.PushdownAuto,
	"engine": planner.PushdownForceEngine,
}

func main() {
	var (
		workload   = flag.String("workload", "Q11", "ワークロード名 (test.Registry のキー)")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		trials     = flag.Int("trials", 1, "試行回数(平均)")
		pushdown   = flag.String("pushdown", "engine", "集約 pushdown 方針: engine(段階別に計測) | auto(単一ストアへ委譲)")
	)
	flag.Parse()

	pm, ok := pushdownModeMap[*pushdown]
	if !ok {
		log.Fatalf("未知の pushdown モード %q (利用可能: auto | engine)", *pushdown)
	}
	planner.SelectedPushdown = pm

	cfg, err := storage.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}
	ctx := context.Background()

	def, ok := workloads.Registry[*workload]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", *workload, workloads.AvailableWorkloads())
	}
	cypher, params, _ := def(migrator.ModeKvsToGraph, false)

	res, err := bulk.RunBulk(ctx, cfg, cypher, params, *trials)
	if err != nil {
		log.Fatalf("実行に失敗: %v", err)
	}
	bulk.PrintResult(fmt.Sprintf("%s/Bulk (pushdown=%s)", *workload, *pushdown), res)
}
