// Command bulkbench は全件マテリアライズ実行系(bulk_exec)を独立バイナリとして実行する。
// 既存 main.go を変更せずに、各演算子が「何件に対して」「どれだけの時間」かかったかを測る。
//
// 例:
//
//	go run ./bulk_exec/cmd/bulkbench -workload Q11 -trials 3
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	bulk "polystore_database/src/go/bulk_exec"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/test"
)

func main() {
	var (
		workload   = flag.String("workload", "Q11", "ワークロード名 (test.Registry のキー)")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		trials     = flag.Int("trials", 1, "試行回数(平均)")
	)
	flag.Parse()

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
	bulk.PrintResult(fmt.Sprintf("%s/Bulk", *workload), res)
}
