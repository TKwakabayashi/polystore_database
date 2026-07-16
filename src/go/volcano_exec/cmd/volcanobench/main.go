// Command volcanobench は pull 型実行系(volcano_exec)を独立バイナリとして実行する。
// 既存 main.go を変更せずに Volcano / Vectorized を試すためのエントリ。
//
// 例:
//
//	go run ./volcano_exec/cmd/volcanobench -workload Q11 -mode vectorized -vsize 512
//	go run ./volcano_exec/cmd/volcanobench -workload Q11 -mode volcano
//	go run ./volcano_exec/cmd/volcanobench -workload Q11 -mode sweep   # N を掃引して往復回数を比較
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/test"
	volcano "polystore_database/src/go/volcano_exec"
)

func main() {
	var (
		mode       = flag.String("mode", "vectorized", "実行モデル: volcano | vectorized | sweep")
		workload   = flag.String("workload", "Q11", "ワークロード名 (test.Registry のキー)")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		vsize      = flag.Int("vsize", 512, "Vectorized のベクトル長")
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

	switch *mode {
	case "volcano":
		res, err := volcano.RunVolcano(ctx, cfg, cypher, params, *trials)
		if err != nil {
			log.Fatalf("実行に失敗: %v", err)
		}
		volcano.PrintResult(fmt.Sprintf("%s/Volcano", *workload), res)

	case "sweep":
		// Volcano と Vectorized(N 掃引) を並べて往復回数の償却を観察する。
		vres, err := volcano.RunVolcano(ctx, cfg, cypher, params, *trials)
		if err != nil {
			log.Fatalf("Volcano 実行に失敗: %v", err)
		}
		volcano.PrintResult(fmt.Sprintf("%s/Volcano", *workload), vres)
		for _, n := range []int{8, 64, 512, 2048, 8192, 32768} {
			r, err := volcano.RunVectorized(ctx, cfg, cypher, params, n, *trials)
			if err != nil {
				log.Fatalf("Vectorized(N=%d) 実行に失敗: %v", n, err)
			}
			volcano.PrintResult(fmt.Sprintf("%s/Vectorized N=%d", *workload, n), r)
		}

	default: // vectorized
		res, err := volcano.RunVectorized(ctx, cfg, cypher, params, *vsize, *trials)
		if err != nil {
			log.Fatalf("実行に失敗: %v", err)
		}
		volcano.PrintResult(fmt.Sprintf("%s/Vectorized N=%d", *workload, *vsize), res)
	}
}
