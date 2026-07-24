//go:build integration

// workload_nonempty_test は合成 mini-SNB（seed.go）を投入した上で、全ワークロードが
// baseline エンジン（stream / graph 配置）で 1 行以上返すことを検証する。
// ハードコード params のドリフト（実在しない id で 0 件になる等）を検知する回帰テスト。
//
//	実行（隔離スタック例。datastore/env/citest.env で up 済みとする）:
//	  POLYSTORE_CONFIG=/abs/path/config.citest.json \
//	    go test -tags integration ./integration/ -run WorkloadNonempty -v
package integration

import (
	"context"
	"testing"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/workload"
)

// nonemptySkip は「データ次第で 0 件が正当」なワークロードの除外リスト（現状なし）。
var nonemptySkip = map[string]bool{}

// TestWorkloadNonempty はロード済みデータ（合成 mini-SNB は POLYSTORE_SEED=1 で TestMain が投入）
// に対し、全ワークロードが 1 行以上返すことを検証する。実データに対して流せば、実在しない id で
// 0 件になるハードコード params のドリフトを検知できる。
func TestWorkloadNonempty(t *testing.T) {
	cfg := loadCfg(t)
	ctx := context.Background()

	for _, name := range workload.AllWorkloadNames() {
		name := name
		t.Run(name, func(t *testing.T) {
			if nonemptySkip[name] {
				t.Skipf("%s は 0 件許容のため skip", name)
			}
			def := workload.Registry[name]
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)

			r, err := bench.RunEngine(ctx, cfg, "stream", cypher, params)
			if err != nil {
				t.Fatalf("[%s] 実行失敗: %v", name, err)
			}
			if len(r.Rows) == 0 {
				t.Errorf("[%s] 0 件（params が seed データと不一致の可能性）", name)
			} else {
				t.Logf("[%s] rows=%d", name, len(r.Rows))
			}
		})
	}
}
