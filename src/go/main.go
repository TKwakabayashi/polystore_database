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
		mode       = flag.String("mode", "run", "実行モード: setup | migrate | run | workflow | verify-migrate")
		workload   = flag.String("workload", "Q11", "ワークロード名")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		migMode    = flag.String("migmode", "graph_to_rdb", "移行モード（a_to_b）: migrate / workflow / verify-migrate で使用")
		deleteSrc  = flag.Bool("delete", true, "移行成功後にソース側の該当データを削除する（migrate / workflow）")
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
		// 単体マイグレーション（-migmode の方向で移行し、時間・件数を出力）
		def := lookup(*workload)
		mode := migrator.MigrationMode(*migMode)
		_, _, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode // def が Mode 未設定の定義(IS4等)でも確実に効かせる
			migs[i].DeleteSource = *deleteSrc
		}
		res, err := workloads.RunMigration(ctx, cfg, migs)
		if err != nil {
			log.Fatalf("migration に失敗: %v", err)
		}
		workloads.PrintMigration(res)

	case "workflow":
		// migration → クエリ実行(Neo4j比較) まで通しで（-migmode の方向で移行）
		def := lookup(*workload)
		mode := migrator.MigrationMode(*migMode)
		cypher, params, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode
			migs[i].DeleteSource = *deleteSrc
		}
		workloads.RunWorkflow(ctx, *workload, cfg, cypher, params, migs)

	case "verify-migrate":
		// 【一時】migrator の動作確認（非破壊ラウンドトリップ検証）
		workloads.RunMigrationVerify(ctx, cfg, *workload, migrator.MigrationMode(*migMode))

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
