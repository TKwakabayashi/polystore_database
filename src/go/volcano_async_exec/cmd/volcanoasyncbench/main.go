// Command volcanoasyncbench は非同期 pull 型実行系(volcano_async_exec)を独立バイナリとして実行する。
// 既存 main.go / volcano_exec を変更せずに、pull × 並行 の組み合わせを試すためのエントリ。
//
// 例:
//
//	go run ./volcano_async_exec/cmd/volcanoasyncbench -workload Q11 -mode vectorized -vsize 512 -async exchange
//	go run ./volcano_async_exec/cmd/volcanoasyncbench -workload Q11 -mode volcano -async prefetch
//	go run ./volcano_async_exec/cmd/volcanoasyncbench -workload Q11 -mode sweep    # N を掃引
//	go run ./volcano_async_exec/cmd/volcanoasyncbench -workload Q11 -mode matrix   # 同期比較用の 2x2
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/test"
	async "polystore_database/src/go/volcano_async_exec"
)

func main() {
	var (
		mode       = flag.String("mode", "vectorized", "処理粒度: volcano | vectorized | sweep | matrix")
		asyncMode  = flag.String("async", "exchange", "並行化方式: exchange | prefetch")
		workload   = flag.String("workload", "Q11", "ワークロード名 (test.Registry のキー)")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		vsize      = flag.Int("vsize", 512, "Vectorized のベクトル長")
		trials     = flag.Int("trials", 1, "試行回数(平均)")
		workers    = flag.Int("workers", 0, "ExecFixed の演算子あたりワーカー数 (0 なら ExecDynamic)")
		globalMax  = flag.Int("globalmax", 8, "ExecDynamic のシステム全体 同時DB上限")
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

	am := async.AsyncExchange
	switch *asyncMode {
	case "prefetch":
		am = async.AsyncPrefetch
	case "exchange":
		am = async.AsyncExchange
	default:
		log.Fatalf("未知の並行化方式 %q (exchange | prefetch)", *asyncMode)
	}

	policy := buildPolicy(*workers, *globalMax)

	switch *mode {
	case "volcano":
		res, err := async.RunVolcanoAsync(ctx, cfg, cypher, params, am, policy, *trials)
		if err != nil {
			log.Fatalf("実行に失敗: %v", err)
		}
		async.PrintResult(fmt.Sprintf("%s/Volcano+%s", *workload, am), res)

	case "sweep":
		vres, err := async.RunVolcanoAsync(ctx, cfg, cypher, params, am, policy, *trials)
		if err != nil {
			log.Fatalf("Volcano 実行に失敗: %v", err)
		}
		async.PrintResult(fmt.Sprintf("%s/Volcano+%s", *workload, am), vres)
		for _, n := range []int{8, 64, 512, 2048, 8192, 32768} {
			r, err := async.RunVectorizedAsync(ctx, cfg, cypher, params, n, am, policy, *trials)
			if err != nil {
				log.Fatalf("Vectorized(N=%d) 実行に失敗: %v", n, err)
			}
			async.PrintResult(fmt.Sprintf("%s/Vectorized N=%d +%s", *workload, n, am), r)
		}

	case "matrix":
		// 処理粒度 × 並行化方式 の 2x2。往復数は粒度だけで決まり、
		// 実行時間は並行化方式でも変わる、という分離を 1 回で観察する。
		for _, m := range []struct {
			name string
			run  func(async.AsyncMode) (async.Result, error)
		}{
			{"Volcano", func(a async.AsyncMode) (async.Result, error) {
				return async.RunVolcanoAsync(ctx, cfg, cypher, params, a, policy, *trials)
			}},
			{fmt.Sprintf("Vectorized N=%d", *vsize), func(a async.AsyncMode) (async.Result, error) {
				return async.RunVectorizedAsync(ctx, cfg, cypher, params, *vsize, a, policy, *trials)
			}},
		} {
			for _, a := range []async.AsyncMode{async.AsyncPrefetch, async.AsyncExchange} {
				r, err := m.run(a)
				if err != nil {
					log.Fatalf("%s/%s 実行に失敗: %v", m.name, a, err)
				}
				async.PrintResult(fmt.Sprintf("%s/%s+%s", *workload, m.name, a), r)
			}
		}

	default: // vectorized
		res, err := async.RunVectorizedAsync(ctx, cfg, cypher, params, *vsize, am, policy, *trials)
		if err != nil {
			log.Fatalf("実行に失敗: %v", err)
		}
		async.PrintResult(fmt.Sprintf("%s/Vectorized N=%d +%s", *workload, *vsize, am), res)
	}
}

// buildPolicy は workers>0 なら ExecFixed（演算子ごと固定）、それ以外は
// ExecDynamic（全体セマフォ）を組む。ExecDynamic の既定値は stream_exec と同じ。
func buildPolicy(workers, globalMax int) async.ExecPolicy {
	if workers > 0 {
		return async.ExecPolicy{
			Mode:    async.ExecFixed,
			Default: async.OpConcurrency{Workers: workers},
			PerOp: map[async.OpKind]async.OpConcurrency{
				async.OpExpand:          {Workers: workers},
				async.OpVarLengthExpand: {Workers: workers},
				async.OpFilter:          {Workers: workers},
				async.OpProjection:      {Workers: workers},
			},
		}
	}
	p := async.DefaultExecPolicy()
	p.GlobalMaxConcurrency = globalMax
	return p
}
