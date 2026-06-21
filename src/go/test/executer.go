package test

import (
	"context"
	"fmt"
	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/storage"
	executor "polystore_database/src/go/stream_exec"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Trials はベンチ計測の試行回数（この回数だけ実行して平均を取る）。手動で切り替え。
const Trials = 10

// average は exec を Trials 回実行し、TotalLatency（と Steps）を平均した ExecResult を返す。
// Rows は最後の試行のものを保持する（件数は試行間で一定の前提）。
func average(exec func() (ExecResult, error)) (ExecResult, error) {
	if Trials <= 0 {
		return ExecResult{}, fmt.Errorf("Trials must be >= 1")
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
