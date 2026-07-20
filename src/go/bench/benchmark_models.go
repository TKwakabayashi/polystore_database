package bench

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	bulk "polystore_database/src/go/engine/bulk"
	volcano "polystore_database/src/go/engine/volcano"
	"polystore_database/src/go/migrator"
	planner "polystore_database/src/go/planner"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// modelOrder は計測する実行モデルの既定順。
var modelOrder = []string{"stream", "bulk", "volcano", "vectorized"}

// RunModelBenchmark は 1 つ以上のワークロードについて次を計測し、long 形式 CSV へ追記する。
//   - baseline: Neo4j へ直接クエリを発行した場合（system=neo4j, model=-, placement=graph）
//   - custom  : データ配置(placement) × 実行モデル(model) を変えた自作システム
//
// 出力は「1 行 = 1 計測」の long 形式で、列は
//
//	run, workload, system, model, placement, vector, rows, latency_ms
//
// placement は graph/rdb/doc/col/kvs、model は stream/bulk/volcano/vectorized。
// 非graph placement では該当プロパティを対象ストアへコピー(DeleteSource=false)し、mapping を
// 更新してから、その配置に対して全モデルを計測する（migration は placement ごとに 1 回）。
// 計測後は mapping を graph へ戻す。実験終了時（異常時含む）も必ず graph 状態へ復元する。
//
// 計測公平性:
//   - stream は RunCustom（Warmup=%d, Trials=%d を内部で実施）。
//   - bulk/volcano/vectorized は Warmup 回を捨ててから Trials 回の平均を測る。
//   - pushdown は engine に固定（Q9 等の非集約クエリでは元々 no-op だが明示）。
//   - pprof（block/mutex 記録）は計測歪みを避けるため無効化。
func RunModelBenchmark(ctx context.Context, cfg storage.Config, workloads, placements, models []string, vectorSize int, out string) (err error) {
	if vectorSize < 1 {
		vectorSize = 1
	}

	// pprof の block/mutex 記録は計測を歪めるため無効化。
	prevProfile := ProfileCustomRun
	ProfileCustomRun = false
	defer func() { ProfileCustomRun = prevProfile }()

	// pushdown は engine に固定（非集約クエリでは no-op だが実験の再現性のため明示）。
	prevPushdown := planner.SelectedPushdown
	planner.SelectedPushdown = planner.PushdownForceEngine
	defer func() { planner.SelectedPushdown = prevPushdown }()

	// mapping スナップショット（graph 状態）。各 placement 後と実験終了時に復元する。
	snapshot, rerr := os.ReadFile(cfg.MappingPath)
	if rerr != nil {
		return fmt.Errorf("mapping 読込失敗: %w", rerr)
	}
	restoreMapping := func() error { return os.WriteFile(cfg.MappingPath, snapshot, 0644) }
	defer restoreMapping()

	w, closeCSV, cerr := openModelCSV(out)
	if cerr != nil {
		return cerr
	}
	defer closeCSV()

	runID := time.Now().Format("2006-01-02T15:04:05")
	fmt.Printf("=== Model benchmark run %s → %s (vectorSize=%d) ===\n", runID, out, vectorSize)

	for _, name := range workloads {
		def, ok := workload.Registry[name]
		if !ok {
			fmt.Printf("[skip] 未知のワークロード %q\n", name)
			continue
		}
		cypher, params, migs := def(migrator.ModeGraphToRdb, true) // migs.Mode は placement ごとに上書き

		// --- baseline: Neo4j 直接発行（graph 状態）---
		if cfg.Neo4j != nil {
			if r, nerr := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); nerr != nil {
				fmt.Printf("[%s] neo4j-direct エラー: %v\n", name, nerr)
			} else {
				_ = w.Write(modelRow(runID, name, "neo4j", "-", "graph", "", r.RowCount(), toMs(r)))
				liveModelLine(name, "neo4j", "-", "graph", r.RowCount(), toMs(r))
			}
		}

		// --- custom: placement × model ---
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

			for _, model := range models {
				rows, ms, rerr := runModelOnce(ctx, cfg, model, cypher, params, vectorSize)
				if rerr != nil {
					fmt.Printf("[%s|%s|%s] custom エラー: %v\n", name, place, model, rerr)
					continue
				}
				vec := ""
				if model == "vectorized" {
					vec = strconv.Itoa(vectorSize)
				}
				_ = w.Write(modelRow(runID, name, "custom", model, place, vec, rows, ms))
				liveModelLine(name, "custom/"+model, "-", place, rows, ms)
			}

			// placement を graph に戻す（mapping 復元）。
			if place != "graph" {
				if rerr := restoreMapping(); rerr != nil {
					return fmt.Errorf("mapping 復元失敗: %w", rerr)
				}
			}
		}
		w.Flush() // ワークロード単位で逐次フラッシュ（中断耐性）
	}

	// 最終的に graph へ戻す（保険）。
	return restoreMapping()
}

// runModelOnce は指定モデルで 1 ワークロードを計測し、(行数, レイテンシms) を返す。
// stream は RunCustom が Warmup+Trials を内部で行う。bulk/volcano/vectorized は
// Warmup 回を捨ててから Trials 回平均を測る。
func runModelOnce(ctx context.Context, cfg storage.Config, model, cypher string, params map[string]string, vectorSize int) (int, float64, error) {
	switch model {
	case "stream":
		r, err := RunCustom(ctx, cfg, cypher, params)
		if err != nil {
			return 0, 0, err
		}
		return r.RowCount(), toMs(r), nil

	case "bulk":
		if Warmup > 0 {
			if _, err := bulk.RunBulk(ctx, cfg, cypher, params, Warmup); err != nil {
				return 0, 0, err
			}
		}
		r, err := bulk.RunBulk(ctx, cfg, cypher, params, Trials)
		if err != nil {
			return 0, 0, err
		}
		return r.RowCount(), durToMs(r.Latency), nil

	case "volcano":
		if Warmup > 0 {
			if _, err := volcano.RunVolcano(ctx, cfg, cypher, params, Warmup); err != nil {
				return 0, 0, err
			}
		}
		r, err := volcano.RunVolcano(ctx, cfg, cypher, params, Trials)
		if err != nil {
			return 0, 0, err
		}
		return r.RowCount(), durToMs(r.Latency), nil

	case "vectorized":
		if Warmup > 0 {
			if _, err := volcano.RunVectorized(ctx, cfg, cypher, params, vectorSize, Warmup); err != nil {
				return 0, 0, err
			}
		}
		r, err := volcano.RunVectorized(ctx, cfg, cypher, params, vectorSize, Trials)
		if err != nil {
			return 0, 0, err
		}
		return r.RowCount(), durToMs(r.Latency), nil

	default:
		return 0, 0, fmt.Errorf("未知のモデル %q（stream/bulk/volcano/vectorized）", model)
	}
}

// openModelCSV は追記モードで開き、新規ファイルなら long 形式のヘッダを書く。
func openModelCSV(path string) (*csv.Writer, func(), error) {
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
		_ = w.Write([]string{"run", "workload", "system", "model", "placement", "vector", "rows", "latency_ms"})
		w.Flush()
	}
	return w, func() { w.Flush(); f.Close() }, nil
}

// modelRow は long 形式 CSV の 1 レコードを組む。
func modelRow(runID, workload, system, model, placement, vector string, rows int, ms float64) []string {
	return []string{
		runID, workload, system, model, placement, vector,
		strconv.Itoa(rows), strconv.FormatFloat(ms, 'f', 3, 64),
	}
}

// durToMs は time.Duration をミリ秒(float)に変換する。
func durToMs(d time.Duration) float64 { return float64(d.Microseconds()) / 1000.0 }

// liveModelLine は 1 計測を標準出力へライブ表示する。
func liveModelLine(workload, system, model, placement string, rows int, ms float64) {
	fmt.Printf("  %-6s %-16s %-6s %-6s rows=%-5d %10.3f ms\n", workload, system, model, placement, rows, ms)
}
