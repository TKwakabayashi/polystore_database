package test

import (
	"context"
	"fmt"
	"os"
	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/storage"
	executor "polystore_database/src/go/stream_exec"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ベンチ計測の回数（マクロ定数。手動で切り替え）。
//   - Warmup: 計測から除外する先頭ウォームアップ回数（キャッシュ/JIT 等の暖機）
//   - Trials: Warmup の後に実行し、平均を取る計測回数
const (
	Warmup = 3
	Trials = 10
)

// average は exec を Warmup 回（結果は破棄）→ Trials 回 実行し、
// 後半 Trials 回の TotalLatency（と Steps）を平均した ExecResult を返す。
// Rows は最後の試行のものを保持する（件数は試行間で一定の前提）。
func average(exec func() (ExecResult, error)) (ExecResult, error) {
	if Trials <= 0 {
		return ExecResult{}, fmt.Errorf("Trials must be >= 1")
	}

	// ウォームアップ（計測に含めない）
	for i := 0; i < Warmup; i++ {
		if _, err := exec(); err != nil {
			return ExecResult{}, err
		}
	}

	var (
		sumLatency time.Duration
		last       ExecResult
		sumSteps   []time.Duration // 演算子ごとの時間合計（bulk導入後に有効）
		stepNames  []string
		stepRows   []int
	)

	for i := 0; i < Trials; i++ {
		r, err := exec()
		if err != nil {
			return ExecResult{}, err
		}
		sumLatency += r.TotalLatency
		last = r

		// Steps の集計（index 揃えで加算）
		if len(r.Steps) > 0 {
			if sumSteps == nil {
				sumSteps = make([]time.Duration, len(r.Steps))
				stepNames = make([]string, len(r.Steps))
				stepRows = make([]int, len(r.Steps))
			}
			for j, st := range r.Steps {
				if j < len(sumSteps) {
					sumSteps[j] += st.Duration
					stepNames[j] = st.Name
					stepRows[j] = st.Rows
				}
			}
		}
	}

	avg := ExecResult{
		Rows:         last.Rows,
		TotalLatency: sumLatency / time.Duration(Trials),
	}
	if sumSteps != nil {
		avg.Steps = make([]StepMetric, len(sumSteps))
		for j := range sumSteps {
			avg.Steps[j] = StepMetric{
				Name:     stepNames[j],
				Duration: sumSteps[j] / time.Duration(Trials),
				Rows:     stepRows[j],
			}
		}
	}
	return avg, nil
}

// RunNeo4j は Cypher を Neo4j へ直接実行する（既存システムのベースライン）。Trials 回の平均。
func RunNeo4j(ctx context.Context, cfg storage.Neo4jConfig, cypher string, params map[string]interface{}) (ExecResult, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return ExecResult{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	// cypher = "CYPHER runtime=parallel\n" + cypher
	return average(func() (ExecResult, error) {
		start := time.Now()
		rowsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			res, err := tx.Run(ctx, cypher, params)
			if err != nil {
				return nil, err
			}
			var rows []map[string]interface{}
			for res.Next(ctx) {
				rec := res.Record()
				row := make(map[string]interface{}, len(rec.Keys))
				for i, key := range rec.Keys {
					row[key] = rec.Values[i]
				}
				rows = append(rows, row)
			}
			return rows, res.Err()
		})
		if err != nil {
			return ExecResult{}, err
		}
		rows, _ := rowsAny.([]map[string]interface{})
		return ExecResult{Rows: rows, TotalLatency: time.Since(start)}, nil
	})
}

// RunCustom は自作システムで parse＋ストリーム実行する。Trials 回の平均。
func RunCustom(ctx context.Context, cfg storage.Config, cypher string, params map[string]string) (ExecResult, error) {
	qp, err := executor.NewQueryProcessorWithConfig(ctx, cfg)
	if err != nil {
		return ExecResult{}, fmt.Errorf("QueryProcessor の初期化に失敗（DB は全て起動済みですか？）: %w", err)
	}
	defer qp.Close()
	// --- ここからプロファイル（接続確立後・計測対象だけを囲む）---
	fcpu, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(fcpu)
	runtime.SetBlockProfileRate(1) // 1ns = 全ブロックイベント記録
	runtime.SetMutexProfileFraction(1)
	defer func() {
		pprof.StopCPUProfile()
		fcpu.Close()

		dump := func(name string) {
			f, _ := os.Create(name + ".prof")
			defer f.Close()
			pprof.Lookup(name).WriteTo(f, 0)
		}
		runtime.GC() // heap のライブ集合を正確に
		dump("heap")
		dump("allocs")
		dump("block")
		dump("mutex")
	}()
	return average(func() (ExecResult, error) {
		qp.Reset() // 試行ごとに中間状態をクリア

		start := time.Now()
		op, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
		if err != nil {
			return ExecResult{}, fmt.Errorf("クエリのパース／プラン構築に失敗: %w", err)
		}
		results, err := qp.ProcessQueryStream(op)
		elapsed := time.Since(start)
		if err != nil {
			return ExecResult{}, fmt.Errorf("クエリ実行に失敗: %w", err)
		}
		// Steps は stream 版では未計測。bulk 導入時に qp のメトリクスから埋める。
		return ExecResult{Rows: results, TotalLatency: elapsed}, nil
	})
}

// toValuedParams は string params を Neo4j 用の typed params に変換する。
func toValuedParams(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		if n, err := strconv.Atoi(v); err == nil {
			out[k] = n
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = t
		} else {
			out[k] = v
		}
	}
	return out
}
