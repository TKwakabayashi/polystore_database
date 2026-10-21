package bench

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"polystore_database/src/go/executor"
	"polystore_database/src/go/profile"
	"polystore_database/src/go/result"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// RunEngine は刷新版の処理系（ast → logical → physical → native → executor）で cypher を実行する
// （Warmup 回捨て → Trials 回平均）。kind は反復モデル（bulk / volcano / vectorized / stream / vecstream）。
// 配置は cfg.MappingPath の mapping から毎回読む（bench は配置ごとに移行して mapping を書き換えるため）。
// 試行ごとに計画（PlanTime）と実行（ExecTime）を分けて集計する。
func RunEngine(ctx context.Context, cfg storage.Config, kind settings.EngineKind, cypher string, params map[string]string) (result.Result, error) {
	if settings.Trials <= 0 {
		return result.Result{}, fmt.Errorf("Trials must be >= 1")
	}
	md, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		return result.Result{}, fmt.Errorf("mapping 読込失敗: %w", err)
	}
	reg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return result.Result{}, fmt.Errorf("ストアへの接続に失敗（DB は全て起動済みですか？）: %w", err)
	}
	defer reg.Close(ctx)
	be := executor.NewStoreBackend(reg)

	// プロファイル（接続確立後・計測対象だけを囲む）。ScopeEngine 有効時のみ採取。
	defer profile.Start(settings.ScopeEngine, string(kind)).Stop()

	for i := 0; i < settings.Warmup; i++ {
		if _, err := executor.RunModel(ctx, be, md, cypher, params, kind); err != nil {
			return result.Result{}, fmt.Errorf("クエリ実行に失敗: %w", err)
		}
	}
	var sumPlan, sumExec time.Duration
	var last result.Result
	for i := 0; i < settings.Trials; i++ {
		r, err := executor.RunModel(ctx, be, md, cypher, params, kind)
		if err != nil {
			return result.Result{}, fmt.Errorf("クエリ実行に失敗: %w", err)
		}
		sumPlan += r.PlanTime
		sumExec += r.ExecTime
		last = r
	}
	last.PlanTime = sumPlan / time.Duration(settings.Trials)
	last.ExecTime = sumExec / time.Duration(settings.Trials)
	return last, nil
}

// RunNeo4j は Cypher を Neo4j へ直接実行する（既存システムのベースライン）。Trials 回の平均。
func RunNeo4j(ctx context.Context, cfg storage.Neo4jConfig, cypher string, params map[string]interface{}) (result.Result, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return result.Result{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	if settings.Trials <= 0 {
		return result.Result{}, fmt.Errorf("Trials must be >= 1")
	}
	one := func() (result.Result, error) {
		start := time.Now()
		// 刷新版の graph への発行と同じオートコミット session.Run に揃える。
		res, err := session.Run(ctx, cypher, params)
		if err != nil {
			return result.Result{}, err
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
			return result.Result{}, err
		}
		return result.Result{Rows: rows, ExecTime: time.Since(start), Engine: "neo4j"}, nil
	}
	for i := 0; i < settings.Warmup; i++ {
		if _, err := one(); err != nil {
			return result.Result{}, err
		}
	}
	var sum time.Duration
	var last result.Result
	for i := 0; i < settings.Trials; i++ {
		r, err := one()
		if err != nil {
			return result.Result{}, err
		}
		sum += r.ExecTime
		last = r
	}
	last.ExecTime = sum / time.Duration(settings.Trials)
	return last, nil
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
