package bench

import (
	"context"
	"fmt"
	executor "polystore_database/src/go/engine/stream"
	planner "polystore_database/src/go/planner"
	"polystore_database/src/go/profile"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// 計測回数は settings.Warmup / settings.Trials で切り替える。

// average は exec を Warmup 回（結果は破棄）→ Trials 回 実行し、
// 後半 Trials 回の TotalLatency（と Steps）を平均した ExecResult を返す。
// Rows は最後の試行のものを保持する（件数は試行間で一定の前提）。
func average(exec func() (ExecResult, error)) (ExecResult, error) {
	if settings.Trials <= 0 {
		return ExecResult{}, fmt.Errorf("Trials must be >= 1")
	}

	// ウォームアップ（計測に含めない）
	for i := 0; i < settings.Warmup; i++ {
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

	for i := 0; i < settings.Trials; i++ {
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
		TotalLatency: sumLatency / time.Duration(settings.Trials),
	}
	if sumSteps != nil {
		avg.Steps = make([]StepMetric, len(sumSteps))
		for j := range sumSteps {
			avg.Steps[j] = StepMetric{
				Name:     stepNames[j],
				Duration: sumSteps[j] / time.Duration(settings.Trials),
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
		// pushdown 経路（runGraphPushdown）と同じオートコミット session.Run に揃える。
		// 管理トランザクション(ExecuteRead)のオーバーヘッドを両者から除き、公平に比較する。
		res, err := session.Run(ctx, cypher, params)
		if err != nil {
			return ExecResult{}, err
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
		if err := res.Err(); err != nil {
			return ExecResult{}, err
		}
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

	// プロファイル（接続確立後・計測対象だけを囲む）。ScopeEngine が有効なときだけ採取。
	// bench 実行系は ScopeEngine を一時的に無効化するので、この場合 no-op になる。
	defer profile.Start(settings.ScopeEngine, "custom").Stop()

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
