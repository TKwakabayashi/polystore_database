package main

import (
	"context"
	"flag"
	"log"
	"strings"

	bench "polystore_database/src/go/bench"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/setup"
	"polystore_database/src/go/storage"
	workloads "polystore_database/src/go/workload"
)

// CLI は「何を実行するか」だけを受け取る。実行モデル・pushdown 方針・配置/実行モデルの
// スイープ軸・ベクトル長・ソース削除・出力形式・プロファイル等の「内部挙動」は settings
// パッケージ（編集→再ビルド）で切り替える。
func main() {
	var (
		mode       = flag.String("mode", "run", "実行モード: setup | migrate | run | workflow | bench | bench-models")
		workload   = flag.String("workload", "Q11", "ワークロード名（bench はカンマ区切り複数 or all）")
		configPath = flag.String("config", "../../config/config.json", "設定ファイル(JSON)")
		migMode    = flag.String("migmode", "graph_to_rdb", "移行モード（a_to_b）: migrate / workflow で使用")
		outPath    = flag.String("out", "../../results/bench/bench_results.csv", "bench: 結果CSVの出力先（追記）")
	)
	flag.Parse()

	cfg, err := storage.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}
	ctx := context.Background()

	switch *mode {
	case "setup":
		// Neo4j 整備＋他4ストア初期化（最初に1回）
		if err := setup.Run(ctx, cfg); err != nil {
			log.Fatalf("data setup に失敗: %v", err)
		}

	case "migrate":
		// 単体マイグレーション（-migmode の方向で移行し、時間・件数を出力）
		def := lookup(*workload)
		mode := migrator.MigrationMode(*migMode)
		_, _, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode // def が Mode 未設定の定義(IS4等)でも確実に効かせる
			migs[i].DeleteSource = settings.MigrationDeleteSource
		}
		res, err := bench.RunMigration(ctx, cfg, migs)
		if err != nil {
			log.Fatalf("migration に失敗: %v", err)
		}
		bench.PrintMigration(res)

	case "workflow":
		// migration → クエリ実行(Neo4j比較) まで通しで（-migmode の方向で移行）
		def := lookup(*workload)
		mode := migrator.MigrationMode(*migMode)
		cypher, params, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode
			migs[i].DeleteSource = settings.MigrationDeleteSource
		}
		bench.RunWorkflow(ctx, *workload, cfg, cypher, params, migs)

	case "bench":
		// baseline(Neo4j直) と placement×pushdown の自作システムを計測し CSV へ追記。最後に graph へ戻す。
		wls := workloads.AllWorkloadNames()
		if *workload != "all" && *workload != "" {
			wls = splitCSV(*workload)
		}
		if err := bench.RunBenchmark(ctx, cfg, wls, *outPath); err != nil {
			log.Fatalf("bench に失敗: %v", err)
		}

	case "bench-models":
		// baseline(Neo4j直) と placement×実行モデルの自作システムを計測し long 形式 CSV へ追記。
		// 配置/実行モデル/ベクトル長は settings.BenchPlacements / BenchModels / VectorSize で切替。
		// 例: -mode bench-models -workload Q9 -out ../../results/bench/q9_models.csv
		wls := workloads.AllWorkloadNames()
		if *workload != "all" && *workload != "" {
			wls = splitCSV(*workload)
		}
		if err := bench.RunModelBenchmark(ctx, cfg, wls, *outPath); err != nil {
			log.Fatalf("bench-models に失敗: %v", err)
		}

	default: // "run"
		// クエリ実行のみ（settings.RunTarget / settings.Format に従う）
		bench.RunWorkloadByName(ctx, *workload, cfg)
	}
}

// splitCSV はカンマ区切り文字列を空白トリムしてスライス化する（空要素は除外）。
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func lookup(name string) func(migrator.MigrationMode, bool) (string, map[string]string, []migrator.MigrationConfig) {
	def, ok := workloads.Registry[name]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", name, workloads.AvailableWorkloads())
	}
	return def
}
