package main

import (
	"context"
	"flag"
	"log"

	"polystore_database/src/go/data_setup"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/test"
)

func main() {
	var (
		mode       = flag.String("mode", "run", "実行モード: setup | migrate | run | workflow")
		workload   = flag.String("workload", "IS4", "ワークロード名")
		configPath = flag.String("config", "../config/config.json", "設定ファイル(JSON)")
	)
	flag.Parse()

	cfg, err := storage.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}
	ctx := context.Background()

	/*
		f, err := os.Create("trace.out")
		if err != nil {
			log.Fatalf("failed to create trace file: %v", err)
		}
		defer f.Close()

		// 2. トレースの開始
		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop() // プログラム終了時に必ず止める
	*/
	switch *mode {
	case "setup":
		// Neo4j 整備＋他4ストア初期化（最初に1回）
		if err := data_setup.Run(ctx, cfg); err != nil {
			log.Fatalf("data setup に失敗: %v", err)
		}

	case "migrate":
		// 単体マイグレーション（時間・件数を出力）
		def := lookup(*workload)
		_, _, migs := def(migrator.ModeKvsToGraph, false) // ←移行モードは実験に応じて選ぶ
		res, err := workloads.RunMigration(ctx, cfg, migs)
		if err != nil {
			log.Fatalf("migration に失敗: %v", err)
		}
		workloads.PrintMigration(res)

	case "workflow":
		// migration → Neo4j比較 まで通しで
		def := lookup(*workload)
		cypher, params, migs := def(migrator.ModeKvsToGraph, true)
		workloads.RunWorkflow(ctx, *workload, cfg, cypher, params, migs)

	default: // "run"
		// クエリ実行のみ（SelectedTarget/SelectedFormat に従う）
		workloads.RunWorkloadByName(ctx, *workload, cfg)
	}
}

func lookup(name string) func(migrator.MigrationMode, bool) (string, map[string]string, []migrator.MigrationConfig) {
	def, ok := workloads.Registry[name]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", name, workloads.AvailableWorkloads())
	}
	return def
}
