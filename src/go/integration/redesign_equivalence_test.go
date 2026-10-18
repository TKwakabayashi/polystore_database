//go:build integration

// redesign_equivalence_test は刷新版のクエリ処理系（ast → logical → physical → native → executor）が
// graph 配置で Neo4j に原文をそのまま投げた結果（オラクル）と同じ行集合を返すことを確かめる。
//
//	実行: cwd=src/go で docker スタックを up した状態で
//	  go test -tags integration ./integration/ -run RedesignEquivalence -v
//	設定は POLYSTORE_CONFIG（未設定なら ../../config/config.json）から読む。
//
// 構造化できないワークロードは処理系自身が verbatim に落とすので一致は自明だが、実行エラーの検出のため残す。
// 非 graph 配置の一致は移行が要るので、配置×ワークロードの全面確認は bench-models 側で行う。
package integration

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/executor"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

func TestRedesignEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	md, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping load (%s): %v", cfg.MappingPath, err)
	}
	ctx := context.Background()
	reg, err := storage.NewRegistryFor(ctx, cfg, storage.Graph)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := executor.NewStoreBackend(reg)

	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			cypher, params, _ := workload.Registry[name](migrator.ModeGraphToRdb, false)
			oracle, err := bench.RunNeo4j(ctx, *cfg.Neo4j, cypher, typedForNeo4j(params))
			if err != nil {
				t.Skipf("Neo4j（オラクル）が失敗したため skip: %v", err)
			}
			got, err := executor.Run(ctx, be, md, cypher, params)
			if err != nil {
				t.Fatalf("redesign の実行に失敗: %v", err)
			}
			want, have := canonRedesign(oracle.Rows), canonRedesign(got.Rows)
			if strings.Join(want, "\n") != strings.Join(have, "\n") {
				t.Errorf("[%s] 行集合がオラクルと不一致（oracle=%d, redesign=%d）\n%s",
					got.Engine, len(want), len(have), firstRedesignDiff(want, have))
				return
			}
			t.Logf("[%s] rows=%d 一致（round trips=%d, plan=%v, exec=%v）",
				got.Engine, len(have), got.RoundTrips, got.PlanTime, got.ExecTime)
		})
	}
}

// typedForNeo4j は bench.RunNeo4j に渡す typed params（整数・RFC3339 時刻・文字列）。
func typedForNeo4j(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && fmt.Sprint(n) == v {
			out[k] = n
		} else if tm, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = tm
		} else {
			out[k] = v
		}
	}
	return out
}

// canonRedesign は行を順序非依存で比べられる文字列にする（数値・時刻の表現をそろえる）。
func canonRedesign(rows []map[string]interface{}) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%s;", k, canonValue(row[k]))
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return out
}

func canonValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case int:
		return fmt.Sprint(int64(x))
	case int32:
		return fmt.Sprint(int64(x))
	case int64:
		return fmt.Sprint(x)
	case float32:
		return fmt.Sprint(float64(x))
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano)
	case []interface{}:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = canonValue(e)
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	return fmt.Sprint(v)
}

func firstRedesignDiff(want, got []string) string {
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			return fmt.Sprintf("  行 %d:\n    oracle  : %s\n    redesign: %s", i, want[i], got[i])
		}
	}
	return fmt.Sprintf("  長さのみ相違: oracle=%d redesign=%d", len(want), len(got))
}
