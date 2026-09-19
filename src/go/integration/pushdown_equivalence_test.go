//go:build integration

// pushdown_equivalence_test は「pushdown ON（fusion→StoreFragment 委譲）と OFF（コーディネータ
// エンジンで計算）で結果が一致する」ことを確認する等価性テスト。P3(3a) で ParseQuery を on/off の
// 別プラン構築経路に分岐し、集約クエリを StoreFragment へ委譲（Plan を lowering してネイティブ発行）した
// 変更の回帰ガード。
//
//	実行: cwd=src/go で docker スタックを up した状態（graph 配置）で
//	  go test -tags integration ./integration/ -run PushdownEquivalence -v
//	設定は POLYSTORE_CONFIG（未設定なら ../../config/config.json）から読む。
//
// 配置は graph 固定。ここでの ON 経路は「原 Cypher を Neo4j へ丸ごと委譲（集約も Neo4j 実行）」、
// OFF 経路は「record パイプライン＋エンジン内集約」。両者の結果 multiset 一致を検証する。
// 非 graph 配置の委譲（ネイティブ集約）の等価性は migration を要するため bench-models で確認する。
package integration

import (
	"context"
	"os"
	"reflect"
	"testing"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// pushdownEquivWorkloads は集約ワークロード（graph 配置で ON=委譲 / OFF=エンジン集約に分岐する）。
var pushdownEquivWorkloads = []string{"AGG1", "AGG6"}

// TestPushdownEquivalence は各集約ワークロードを ON / OFF の両経路で実行し、結果の行集合
// （順序非依存 multiset）が一致することを検証する。モデルは決定的な bulk を用いる。
func TestPushdownEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	// settings.Pushdown を一時上書きするため元値を退避（bench と同じ作法）。
	prev := settings.Pushdown
	defer func() { settings.Pushdown = prev }()

	for _, name := range pushdownEquivWorkloads {
		name := name
		t.Run(name, func(t *testing.T) {
			def, ok := workload.Registry[name]
			if !ok {
				t.Fatalf("未知のワークロード %q", name)
			}
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)

			settings.Pushdown = settings.PushdownForceEngine
			offRows, err := runRows(ctx, cfg, "bulk", cypher, params)
			if err != nil {
				t.Fatalf("[OFF] 実行失敗: %v", err)
			}

			settings.Pushdown = settings.PushdownAuto
			onRows, err := runRows(ctx, cfg, "bulk", cypher, params)
			if err != nil {
				t.Fatalf("[ON] 実行失敗: %v", err)
			}

			off := canonRows(offRows)
			on := canonRows(onRows)
			if len(off) != len(on) {
				t.Errorf("行数が ON/OFF で不一致: off=%d on=%d", len(off), len(on))
				return
			}
			if !reflect.DeepEqual(off, on) {
				t.Errorf("行集合が ON/OFF で不一致\n%s", firstDiff(off, on))
				return
			}
			t.Logf("ON≡OFF rows=%d 一致", len(offRows))
		})
	}
}
