package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"

	"polystore_database/src/go/data_setup"
	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	executor "polystore_database/src/go/stream_exec"
	workloads "polystore_database/src/go/test"
)

// workloadDef はワークロード定義関数のシグネチャ。
// cypher 本体・パラメータ・（今回は使わない）migration 設定を返す。
type workloadDef func(migrator.MigrationMode, bool) (string, map[string]string, []migrator.MigrationConfig)

// 名前引きできるワークロード一覧（test パッケージの公開関数をそのまま登録）。
var workloadRegistry = map[string]workloadDef{
	"Q2":  workloads.DefineWorkloadQ2,
	"Q8":  workloads.DefineWorkloadQ8,
	"Q9":  workloads.DefineWorkloadQ9,
	"Q11": workloads.DefineWorkloadQ11,
	"IS1": workloads.DefineWorkloadIS1,
	"IS2": workloads.DefineWorkloadIS2,
	"IS3": workloads.DefineWorkloadIS3,
	"IS4": workloads.DefineWorkloadIS4,
	"IS5": workloads.DefineWorkloadIS5,
	"IS6": workloads.DefineWorkloadIS6,
}

func main() {
	var (
		workloadName = flag.String("workload", "IS4", "実行するワークロード名 (Q2,Q8,Q9,Q11,IS1..IS6)")
		configPath   = flag.String("config", "../config/config.json", "データストア設定ファイル(JSON)")
	)
	flag.Parse()

	def, ok := workloadRegistry[*workloadName]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", *workloadName, availableWorkloads())
	}

	cfg, err := storage.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}

	if err := data_setup.Run(context.Background(), cfg); err != nil {
		log.Fatalf("data setup に失敗: %v", err)
	}

	// 今回は migration なしでクエリ実行のみ。mode は migration 用なので任意。
	cypher, params, _ := def(migrator.ModeKvsToGraph, false)

	ctx := context.Background()

	fmt.Printf("=== ワークロード %s 実行 ===\n", *workloadName)

	// 1. 自作システムへ接続（5ストアへ all-or-nothing 接続）
	qp, err := executor.NewQueryProcessorWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("QueryProcessor の初期化に失敗（DB は全て起動済みですか？）: %v", err)
	}
	defer qp.Close()

	// 2. Cypher を論理プランへ
	op, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		log.Fatalf("クエリのパース／プラン構築に失敗: %v", err)
	}

	// 3. ストリーム実行
	results, err := qp.ProcessQueryStream(op)
	if err != nil {
		log.Fatalf("クエリ実行に失敗: %v", err)
	}

	// 4. 結果出力
	printResults(results)
}

func printResults(rows []map[string]interface{}) {
	fmt.Printf("\n--- 結果: %d 件 ---\n", len(rows))
	for i, row := range rows {
		// キー順を安定させて表示
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Printf("[%d] ", i)
		for j, k := range keys {
			if j > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s=%v", k, row[k])
		}
		fmt.Println()
	}
}

func availableWorkloads() string {
	names := make([]string, 0, len(workloadRegistry))
	for n := range workloadRegistry {
		names = append(names, n)
	}
	sort.Strings(names)
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
