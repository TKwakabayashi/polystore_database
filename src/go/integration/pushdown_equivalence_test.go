//go:build integration

// pushdown_equivalence_test は「集約 pushdown（単一ストアへ委譲したネイティブ集約）」と
// 「coordinator（自作エンジンで集約）」が、非 graph 配置でも同じ結果を返すことを検証する。
//
// engine_equivalence_test（graph 配置・4 実行モデルの一致）を補完し、非 graph 配置
// （rdb/doc/col）の pushdown クエリビルダ（BuildRelationalSQL / BuildMongoPipeline /
// BuildColumnarCQL）が graph baseline と一致することを担保する。
//
//	実行（seed 済み隔離スタック例）:
//	  POLYSTORE_SEED=1 POLYSTORE_CONFIG=../../config/config.citest.json \
//	    go test -tags integration ./integration/ -run PushdownEquivalence -v
//
// 対象は traversal 無しの単一エンティティ集約（Organisation）。avg/sum は浮動小数で
// ストア間の表現差が出るため、count/min/max と GROUP BY を使う AGG6 を厳密一致の対象にする。
package integration

import (
	"context"
	"os"
	"reflect"
	"testing"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/workload"
)

// pushdownPlacements は非 graph のネイティブ集約対象ストア。
// kvs は集約非対応で auto は常に engine へフォールバックするため（auto==engine が自明）除外。
var pushdownPlacements = []struct {
	name string
	mode migrator.MigrationMode
}{
	{"rdb", migrator.ModeGraphToRdb},
	{"doc", migrator.ModeGraphToDoc},
	{"col", migrator.ModeGraphToCol},
}

func TestPushdownEquivalence(t *testing.T) {
	cfg := loadCfg(t)
	ctx := context.Background()

	// mapping スナップショット（graph 状態）。各 placement 後と終了時に必ず復元する
	// （DeleteSource=false なのでデータは graph に残る）。
	snapshot, err := os.ReadFile(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping 読込: %v", err)
	}
	restore := func() {
		if err := os.WriteFile(cfg.MappingPath, snapshot, 0o644); err != nil {
			t.Fatalf("mapping 復元: %v", err)
		}
	}
	defer restore()

	prevPushdown := settings.Pushdown
	defer func() { settings.Pushdown = prevPushdown }()

	// AGG6: GROUP BY o.type + count/min/max（厳密一致可能）。traversal 無しで単一ストアへ解決。
	const name = "AGG6"
	def, ok := workload.Registry[name]
	if !ok {
		t.Fatalf("未知のワークロード %q", name)
	}
	cypher, params, migs := def(migrator.ModeGraphToRdb, true)
	if len(migs) == 0 {
		t.Fatalf("%s に migration 設定がありません", name)
	}

	// graph baseline（mapping は graph 状態）。pushdown auto = graph へ raw cypher 委譲。
	settings.Pushdown = settings.PushdownAuto
	refRows, err := runRows(ctx, cfg, "stream", cypher, params)
	if err != nil {
		t.Fatalf("[graph baseline] 実行失敗: %v", err)
	}
	ref := canonRows(refRows)
	if len(ref) == 0 {
		t.Fatalf("[graph baseline] 0 件では等価性を検証できない（seed/データ確認）")
	}
	t.Logf("[graph] rows=%d（基準）", len(ref))

	for _, pl := range pushdownPlacements {
		pl := pl
		t.Run(name+"_"+pl.name, func(t *testing.T) {
			// 対象プロパティを非 graph ストアへコピー（graph 側は保持）。mapping も更新される。
			pmigs := make([]migrator.MigrationConfig, len(migs))
			copy(pmigs, migs)
			for i := range pmigs {
				pmigs[i].Mode = pl.mode
				pmigs[i].DeleteSource = false
			}
			if _, err := bench.RunMigration(ctx, cfg, pmigs); err != nil {
				restore()
				t.Fatalf("[%s] migration 失敗: %v", pl.name, err)
			}
			defer restore() // この placement を graph へ戻す

			// pushdown auto（ネイティブ集約: SQL/Mongo/CQL。col は非対応で engine フォールバック）。
			settings.Pushdown = settings.PushdownAuto
			autoRows, err := runRows(ctx, cfg, "stream", cypher, params)
			if err != nil {
				t.Fatalf("[%s|auto] 実行失敗: %v", pl.name, err)
			}
			auto := canonRows(autoRows)

			// pushdown engine（coordinator: 対象ストアから取得しエンジンで集約）。
			settings.Pushdown = settings.PushdownForceEngine
			engRows, err := runRows(ctx, cfg, "stream", cypher, params)
			if err != nil {
				t.Fatalf("[%s|engine] 実行失敗: %v", pl.name, err)
			}
			eng := canonRows(engRows)

			// (1) 同一配置で auto（pushdown）と engine（coordinator）が一致。
			if !reflect.DeepEqual(auto, eng) {
				t.Errorf("[%s] pushdown auto と engine が不一致\n%s", pl.name, firstDiff(eng, auto))
			}
			// (2) 非 graph の結果が graph baseline と一致。
			if !reflect.DeepEqual(auto, ref) {
				t.Errorf("[%s] pushdown auto が graph baseline と不一致\n%s", pl.name, firstDiff(ref, auto))
			}
			t.Logf("[%s] auto rows=%d / engine rows=%d / graph rows=%d 一致", pl.name, len(auto), len(eng), len(ref))
		})
	}
}
