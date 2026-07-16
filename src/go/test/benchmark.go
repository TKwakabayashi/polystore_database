package test

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
)

// placement 名 → graph からの移行モード（コピー方向）。
// 注: kvs(LevelDB) はネイティブ集約が無いため pushdown は常にコーディネータへフォールバックする
//
//	（auto≈engine）。結果は正しく得られる。
var placementToMode = map[string]migrator.MigrationMode{
	"rdb": migrator.ModeGraphToRdb,
	"doc": migrator.ModeGraphToDoc,
	"col": migrator.ModeGraphToCol,
	"kvs": migrator.ModeGraphToKvs,
}

// pushdown モード名 → planner.PushdownMode。
var pushdownModeMap = map[string]planner.PushdownMode{
	"auto":   planner.PushdownAuto,
	"engine": planner.PushdownForceEngine,
}

// wide 表のカラム順（各配置での自作システムのレイテンシ）。
var placementOrder = []string{"graph", "rdb", "doc", "col", "kvs"}

// latKey は (pushdown, placement) のセル座標。
type latKey struct{ pd, place string }

// RunBenchmark は各ワークロードについて次を計測し、ワイド表形式の CSV へ追記する。
//   - baseline: Neo4j へ直接クエリを発行した場合（graph 状態）
//   - custom  : データ配置(placement) × pushdown 方針を変えた自作システム
//
// 出力は「1行 = (workload, pushdown)、列 = neo4j直 + 各配置(graph/rdb/doc/col/kvs)」の
// ワイド表で、ワークロードごとに書き出す（中断しても完了済みワークロードは残る）。
//
// 非graph placement では該当プロパティを対象ストアへ「コピー」(DeleteSource=false) し、
// mapping を更新してから計測する。計測後は mapping を graph へ戻す（データは graph に保持）。
// 実験終了時（異常時含む）は必ず mapping を graph 状態へ復元する＝最終的に Neo4j へ戻す。
func RunBenchmark(ctx context.Context, cfg storage.Config, workloads, placements, pushdowns []string, outPath string) error {
	// pprof の block/mutex 記録は計測を歪めるため無効化。
	prevProfile := ProfileCustomRun
	ProfileCustomRun = false
	defer func() { ProfileCustomRun = prevProfile }()

	prevPushdown := planner.SelectedPushdown
	defer func() { planner.SelectedPushdown = prevPushdown }()

	// mapping スナップショット（graph 状態）。各 placement 後と実験終了時に復元する。
	snapshot, err := os.ReadFile(cfg.MappingPath)
	if err != nil {
		return fmt.Errorf("mapping 読込失敗: %w", err)
	}
	restoreMapping := func() error { return os.WriteFile(cfg.MappingPath, snapshot, 0644) }
	defer restoreMapping()

	w, closeCSV, err := openBenchCSV(outPath)
	if err != nil {
		return err
	}
	defer closeCSV()

	runID := time.Now().Format("2006-01-02T15:04:05")
	fmt.Printf("=== Benchmark run %s → %s ===\n", runID, outPath)

	for _, name := range workloads {
		def, ok := Registry[name]
		if !ok {
			fmt.Printf("[skip] 未知のワークロード %q\n", name)
			continue
		}
		cypher, params, migs := def(migrator.ModeGraphToRdb, true) // migs.Mode は placement ごとに上書き

		// --- baseline: Neo4j 直接発行（graph 状態）---
		var baseMs float64
		baseRows := 0
		baseOK := false
		if cfg.Neo4j != nil {
			if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
				fmt.Printf("[%s] neo4j-direct エラー: %v\n", name, err)
			} else {
				baseMs, baseRows, baseOK = toMs(r), r.RowCount(), true
				liveLine(name, "neo4j-direct", "-", "graph", baseRows, baseMs)
			}
		}

		// --- custom: placement × pushdown を計測し lat へ蓄積 ---
		lat := map[latKey]float64{}
		for _, place := range placements {
			if place != "graph" {
				mode, ok := placementToMode[place]
				if !ok {
					fmt.Printf("[%s] 未知の placement %q（graph/rdb/doc/col/kvs）\n", name, place)
					continue
				}
				if len(migs) == 0 {
					fmt.Printf("[%s] 移行対象プロパティなし → placement=%s をスキップ\n", name, place)
					continue
				}
				pmigs := make([]migrator.MigrationConfig, len(migs))
				copy(pmigs, migs)
				for i := range pmigs {
					pmigs[i].Mode = mode
					pmigs[i].DeleteSource = false // コピー（graph 側は保持）
				}
				if _, err := RunMigration(ctx, cfg, pmigs); err != nil {
					fmt.Printf("[%s→%s] migration エラー: %v\n", name, place, err)
					restoreMapping()
					continue
				}
			}

			for _, pd := range pushdowns {
				pm, ok := pushdownModeMap[pd]
				if !ok {
					fmt.Printf("[%s] 未知の pushdown %q（auto/engine）\n", name, pd)
					continue
				}
				planner.SelectedPushdown = pm
				r, err := RunCustom(ctx, cfg, cypher, params)
				if err != nil {
					fmt.Printf("[%s|%s|%s] custom エラー: %v\n", name, place, pd, err)
					continue
				}
				lat[latKey{pd, place}] = toMs(r)
				liveLine(name, "custom", pd, place, r.RowCount(), toMs(r))
			}

			// placement を graph に戻す（mapping 復元）。
			if place != "graph" {
				if err := restoreMapping(); err != nil {
					return fmt.Errorf("mapping 復元失敗: %w", err)
				}
			}
		}

		// --- このワークロードのワイド行を書き出す（pushdown ごとに1行）---
		for _, pd := range pushdowns {
			rec := []string{runID, name, pd, strconv.Itoa(baseRows), msCell(baseOK, baseMs)}
			for _, place := range placementOrder {
				v, ok := lat[latKey{pd, place}]
				rec = append(rec, msCell(ok, v))
			}
			_ = w.Write(rec)
		}
		w.Flush() // ワークロード単位で逐次フラッシュ（中断耐性）
	}

	// 最終的に graph へ戻す（保険）。
	return restoreMapping()
}

// openBenchCSV は追記モードで開き、新規ファイルならワイド表のヘッダを書く。
func openBenchCSV(path string) (*csv.Writer, func(), error) {
	needHeader := false
	if fi, err := os.Stat(path); err != nil || fi.Size() == 0 {
		needHeader = true
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("結果ファイル open 失敗: %w", err)
	}
	w := csv.NewWriter(f)
	if needHeader {
		header := []string{"run", "workload", "pushdown", "rows", "neo4j_ms"}
		for _, p := range placementOrder {
			header = append(header, p+"_ms")
		}
		_ = w.Write(header)
		w.Flush()
	}
	closeFn := func() { w.Flush(); f.Close() }
	return w, closeFn, nil
}

// toMs は ExecResult の平均レイテンシをミリ秒(float)に変換する。
func toMs(r ExecResult) float64 { return float64(r.TotalLatency.Microseconds()) / 1000.0 }

// msCell は計測ありなら "%.3f"、無ければ空文字（欠測）を返す。
func msCell(ok bool, ms float64) string {
	if !ok {
		return ""
	}
	return strconv.FormatFloat(ms, 'f', 3, 64)
}

// liveLine は1計測を標準出力へライブ表示する（ファイルはワイド表で別途書く）。
func liveLine(workload, system, pushdown, placement string, rows int, ms float64) {
	fmt.Printf("  %-6s %-12s %-6s %-6s rows=%-5d %10.3f ms\n", workload, system, pushdown, placement, rows, ms)
}
