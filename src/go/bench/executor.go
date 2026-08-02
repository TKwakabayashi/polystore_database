package bench

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"polystore_database/src/go/engine"
	_ "polystore_database/src/go/engine/all" // 全エンジンの init() 登録を集約
	"polystore_database/src/go/engine/core"
	planner "polystore_database/src/go/planner"
	"polystore_database/src/go/profile"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// RunEngine は指定エンジンで cypher を実行する（Warmup 回捨て → Trials 回平均）。
// 試行ごとに parse(PlanTime) と実行(ExecTime) を分離集計し、core.Result に格納する。
//   - 旧 stream の TotalLatency（parse 込み）は TotalLatency()（＝PlanTime+ExecTime）に対応。
//   - 旧 bulk/volcano の Latency（parse 除外）は ExecTime に対応。
func RunEngine(ctx context.Context, cfg storage.Config, kind settings.EngineKind, cypher string, params map[string]string) (core.Result, error) {
	eng, err := engine.New(kind)
	if err != nil {
		return core.Result{}, err
	}
	inst, err := eng.Open(ctx, cfg)
	if err != nil {
		return core.Result{}, fmt.Errorf("エンジン初期化に失敗（DB は全て起動済みですか？）: %w", err)
	}
	defer inst.Close()

	// プロファイル（接続確立後・計測対象だけを囲む）。ScopeEngine 有効時のみ採取。
	defer profile.Start(settings.ScopeEngine, string(kind)).Stop()

	if settings.Trials <= 0 {
		return core.Result{}, fmt.Errorf("Trials must be >= 1")
	}

	// ウォームアップ（計測に含めない）
	for i := 0; i < settings.Warmup; i++ {
		op, perr := planner.ParseQuery(cypher, cfg.MappingPath, params)
		if perr != nil {
			return core.Result{}, fmt.Errorf("クエリのパース／プラン構築に失敗: %w", perr)
		}
		if _, rerr := inst.Run(op); rerr != nil {
			return core.Result{}, fmt.Errorf("クエリ実行に失敗: %w", rerr)
		}
	}

	var sumPlan, sumExec time.Duration
	var last core.Result
	for i := 0; i < settings.Trials; i++ {
		t0 := time.Now()
		op, perr := planner.ParseQuery(cypher, cfg.MappingPath, params)
		planT := time.Since(t0)
		if perr != nil {
			return core.Result{}, fmt.Errorf("クエリのパース／プラン構築に失敗: %w", perr)
		}
		r, rerr := inst.Run(op)
		if rerr != nil {
			return core.Result{}, fmt.Errorf("クエリ実行に失敗: %w", rerr)
		}
		sumPlan += planT
		sumExec += r.ExecTime
		last = r
	}
	last.PlanTime = sumPlan / time.Duration(settings.Trials)
	last.ExecTime = sumExec / time.Duration(settings.Trials)
	return last, nil
}

// RunNeo4j は Cypher を Neo4j へ直接実行する（既存システムのベースライン）。Trials 回の平均。
func RunNeo4j(ctx context.Context, cfg storage.Neo4jConfig, cypher string, params map[string]interface{}) (core.Result, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return core.Result{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	if settings.Trials <= 0 {
		return core.Result{}, fmt.Errorf("Trials must be >= 1")
	}
	one := func() (core.Result, error) {
		start := time.Now()
		// pushdown 経路（runGraphPushdown）と同じオートコミット session.Run に揃える。
		res, err := session.Run(ctx, cypher, params)
		if err != nil {
			return core.Result{}, err
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
			return core.Result{}, err
		}
		return core.Result{Rows: rows, ExecTime: time.Since(start), Engine: "neo4j"}, nil
	}
	for i := 0; i < settings.Warmup; i++ {
		if _, err := one(); err != nil {
			return core.Result{}, err
		}
	}
	var sum time.Duration
	var last core.Result
	for i := 0; i < settings.Trials; i++ {
		r, err := one()
		if err != nil {
			return core.Result{}, err
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
