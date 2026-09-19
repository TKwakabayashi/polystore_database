package bench

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// RunVerifyMatrix は「データ配置 × pushdown × エンジン × バッチサイズ」の性能検証マトリクスを
// 計測し、2 種類の CSV を出力する。
//
//   - Summary CSV（<out>）: 全エンジン横断の行数 + 全体実行時間。
//     列: run, workload, engine, placement, pushdown, batch, rows, latency_ms, neo4j_ms, speedup, status
//     engine ∈ {neo4j, <VerifyEngines...>}。neo4j は全 graph 配置での直接クエリ（OLAP を
//     Neo4j エンジンで実行するケースの基準 + 性能ベースライン）。speedup = neo4j_ms / latency_ms。
//
//   - Detail CSV（<out> の拡張子前に "_detail" を挿入）: stream/vecstream のみの演算子別内訳。
//     列: run, workload, engine, placement, pushdown, batch, step, operator, op_duration_ms, out_rows, roundtrips
//     ボトルネックとなる演算子を配置・バッチ別に特定するための Result.Steps + RoundTrips。
//
// スイープ軸は settings から取得（VerifyEngines / VerifyBatchSizes / BenchPlacements /
// BenchPushdowns）。非 graph 配置では該当プロパティを対象ストアへコピー（DeleteSource=false）し
// mapping を更新してから、その配置に対して全 (pushdown × engine × batch) を計測する。計測後は
// mapping を graph へ戻す。実験終了時（異常時含む）も必ず graph 状態へ復元する。
//
// 計測公平性:
//   - 各エンジンは RunEngine（Warmup + Trials 平均）。
//   - バッチ幅は settings.VectorSize を一時上書き（前後で復元）。bulk / neo4j は非依存のため 1 回。
//   - pprof（block/mutex 記録）は計測歪みを避けるため無効化。
func RunVerifyMatrix(ctx context.Context, cfg storage.Config, workloads []string, out string) (err error) {
	placements := settings.BenchPlacements
	pushdowns := settings.BenchPushdowns
	engines := settings.VerifyEngines
	batchSizes := settings.VerifyBatchSizes
	if len(batchSizes) == 0 {
		batchSizes = []int{settings.VectorSize}
	}

	// pprof の block/mutex 記録は計測を歪めるため、verify 中は engine スコープを無効化。
	prevProfile := settings.ProfileScopes[settings.ScopeEngine]
	settings.ProfileScopes[settings.ScopeEngine] = false
	defer func() { settings.ProfileScopes[settings.ScopeEngine] = prevProfile }()

	// pushdown / VectorSize はループ内で一時上書きするため元値を退避（bench と同じ作法）。
	prevPushdown := settings.Pushdown
	prevVector := settings.VectorSize
	defer func() {
		settings.Pushdown = prevPushdown
		settings.VectorSize = prevVector
	}()

	// mapping スナップショット（graph 状態）。各 placement 後と実験終了時に復元する。
	snapshot, rerr := os.ReadFile(cfg.MappingPath)
	if rerr != nil {
		return fmt.Errorf("mapping 読込失敗: %w", rerr)
	}
	restoreMapping := func() error { return os.WriteFile(cfg.MappingPath, snapshot, 0644) }
	defer restoreMapping()

	sumW, closeSum, cerr := openVerifyCSV(out, verifySummaryHeader)
	if cerr != nil {
		return cerr
	}
	defer closeSum()

	detailPath := detailPathFor(out)
	detW, closeDet, derr := openVerifyCSV(detailPath, verifyDetailHeader)
	if derr != nil {
		return derr
	}
	defer closeDet()

	runID := time.Now().Format("2006-01-02T15:04:05")
	fmt.Printf("=== Verify matrix run %s → %s (+ %s) ===\n", runID, out, detailPath)

	for _, name := range workloads {
		def, ok := workload.Registry[name]
		if !ok {
			fmt.Printf("[skip] 未知のワークロード %q\n", name)
			continue
		}
		cypher, params, migs := def(migrator.ModeGraphToRdb, true) // migs.Mode は placement ごとに上書き

		// --- baseline: 全 graph 配置での Neo4j 直接発行（OLAP の Neo4j エンジン実行 = 性能基準）---
		var neo4jMs float64
		if cfg.Neo4j != nil {
			if r, nerr := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); nerr != nil {
				fmt.Printf("[%s] neo4j-direct エラー: %v\n", name, nerr)
			} else {
				neo4jMs = toMs(r)
				_ = sumW.Write(verifyRow(runID, name, "neo4j", "graph", "-", "-", r.RowCount(), neo4jMs, neo4jMs, "ok"))
				liveVerifyLine(name, "neo4j", "graph", "-", "-", r.RowCount(), neo4jMs, "ok")
			}
		}

		// --- custom: placement × pushdown × engine × batch ---
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
				if _, merr := RunMigration(ctx, cfg, pmigs); merr != nil {
					fmt.Printf("[%s→%s] migration エラー: %v\n", name, place, merr)
					restoreMapping()
					continue
				}
			}

			for _, pd := range pushdowns {
				pdMode, ok := pushdownModeMap[pd]
				if !ok {
					fmt.Printf("[%s] 未知の pushdown %q（auto/engine）\n", name, pd)
					continue
				}
				settings.Pushdown = pdMode

				for _, eng := range engines {
					for _, bs := range batchSizesFor(eng, batchSizes) {
						settings.VectorSize = bs
						batchLabel := batchLabelFor(eng, bs)

						r, rerr := RunEngine(ctx, cfg, settings.EngineKind(eng), cypher, params)
						if rerr != nil {
							// vecstream 等のエンジンギャップは skip 記録して継続（アボートしない）。
							_ = sumW.Write(verifyRow(runID, name, eng, place, pd, batchLabel, 0, 0, neo4jMs, "skip"))
							liveVerifyLine(name, eng, place, pd, batchLabel, 0, 0, "skip: "+shortErr(rerr))
							continue
						}
						lat := toMs(r)
						_ = sumW.Write(verifyRow(runID, name, eng, place, pd, batchLabel, r.RowCount(), lat, neo4jMs, "ok"))
						liveVerifyLine(name, eng, place, pd, batchLabel, r.RowCount(), lat, "ok")

						// Detail: stream/vecstream のみ演算子別内訳を吐く。
						if eng == "stream" || eng == "vecstream" {
							writeDetail(detW, runID, name, eng, place, pd, batchLabel, r)
						}
					}
				}
			}

			// placement を graph に戻す（mapping 復元）。
			if place != "graph" {
				if rerr := restoreMapping(); rerr != nil {
					return fmt.Errorf("mapping 復元失敗: %w", rerr)
				}
			}
		}
		sumW.Flush() // ワークロード単位で逐次フラッシュ（中断耐性）
		detW.Flush()
	}

	// 最終的に graph へ戻す（保険）。
	return restoreMapping()
}

var (
	verifySummaryHeader = []string{"run", "workload", "engine", "placement", "pushdown", "batch", "rows", "latency_ms", "neo4j_ms", "speedup", "status"}
	verifyDetailHeader  = []string{"run", "workload", "engine", "placement", "pushdown", "batch", "step", "operator", "op_duration_ms", "out_rows", "roundtrips"}
)

// batchSizesFor はエンジンごとのバッチスイープ点を返す。バッチ非依存の bulk は単一点 [-]。
func batchSizesFor(engine string, sizes []int) []int {
	switch engine {
	case "stream", "vecstream", "vectorized":
		return sizes
	default: // bulk / volcano はバッチ幅に依存しないため 1 回のみ
		return []int{sizes[0]}
	}
}

// batchLabelFor はバッチ非依存エンジンでは "-"、依存エンジンでは数値ラベルを返す。
func batchLabelFor(engine string, bs int) string {
	switch engine {
	case "stream", "vecstream", "vectorized":
		return strconv.Itoa(bs)
	default:
		return "-"
	}
}

// verifyRow は Summary CSV の 1 レコードを組む。speedup = neo4j_ms / latency_ms（不能時は空）。
func verifyRow(runID, wl, engine, placement, pushdown, batch string, rows int, latMs, neo4jMs float64, status string) []string {
	speedup := ""
	if latMs > 0 && neo4jMs > 0 {
		speedup = strconv.FormatFloat(neo4jMs/latMs, 'f', 3, 64)
	}
	return []string{
		runID, wl, engine, placement, pushdown, batch,
		strconv.Itoa(rows), strconv.FormatFloat(latMs, 'f', 3, 64),
		strconv.FormatFloat(neo4jMs, 'f', 3, 64), speedup, status,
	}
}

// writeDetail は Result.Steps を step 昇順で 1 行ずつ Detail CSV へ書く（RoundTrips は各行に併記）。
func writeDetail(w *csv.Writer, runID, wl, engine, placement, pushdown, batch string, r core.Result) {
	rt := strconv.FormatInt(r.RoundTrips, 10)
	for _, st := range r.Steps {
		_ = w.Write([]string{
			runID, wl, engine, placement, pushdown, batch,
			strconv.Itoa(st.Step), st.Op,
			strconv.FormatFloat(durToMs(st.Duration), 'f', 3, 64),
			strconv.Itoa(st.OutRows), rt,
		})
	}
}

// detailPathFor は Summary の出力パスから Detail 用パス（拡張子前に "_detail"）を作る。
func detailPathFor(out string) string {
	ext := filepath.Ext(out)
	base := strings.TrimSuffix(out, ext)
	if ext == "" {
		ext = ".csv"
	}
	return base + "_detail" + ext
}

// openVerifyCSV は追記モードで開き、新規ファイルなら header を書く。
func openVerifyCSV(path string, header []string) (*csv.Writer, func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, fmt.Errorf("出力ディレクトリ作成失敗: %w", err)
	}
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
		_ = w.Write(header)
		w.Flush()
	}
	return w, func() { w.Flush(); f.Close() }, nil
}

// liveVerifyLine は 1 計測を標準出力へライブ表示する。
func liveVerifyLine(wl, engine, placement, pushdown, batch string, rows int, ms float64, status string) {
	fmt.Printf("  %-6s %-9s %-6s pd=%-6s bat=%-5s rows=%-5d %10.3f ms  %s\n",
		wl, engine, placement, pushdown, batch, rows, ms, status)
}

// shortErr はエラー文言を 1 行・短めに丸める（ライブ表示用）。
func shortErr(err error) string {
	s := err.Error()
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return s
}
