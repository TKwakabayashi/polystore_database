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

// RunNeo4j は Cypher を Neo4j へ直接実行する（既存システムのベースライン）。
func RunNeo4j(ctx context.Context, cfg storage.Neo4jConfig, cypher string, params map[string]interface{}) (ExecResult, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return ExecResult{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

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
}

// RunCustom は自作システムで parse＋ストリーム実行する。
func RunCustom(ctx context.Context, cfg storage.Config, cypher string, params map[string]string) (ExecResult, error) {
	qp, err := executor.NewQueryProcessorWithConfig(ctx, cfg)
	if err != nil {
		return ExecResult{}, fmt.Errorf("QueryProcessor の初期化に失敗（DB は全て起動済みですか？）: %w", err)
	}
	defer qp.Close()

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
