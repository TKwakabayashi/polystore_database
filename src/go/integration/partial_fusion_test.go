//go:build integration

// partial_fusion_test は P3(3b) クロスストア部分融合の等価性を、非破壊 migration で検証する。
//
// シナリオ（IS3: (p:Person {id})-[r:KNOWS]-(friend:Person) を traversal し friend 列を返す）:
//
//  1. 全 graph 配置で OFF（コーディネータ）結果を baseline として取得。
//
//  2. Person.firstName/lastName を MySQL へ COPY（DeleteSource=false）し mapping を rdb へ更新。
//     → traversal は graph のまま、friend.firstName/lastName の materialize だけ rdb になる。
//
//  3. ON でプランに record-mode StoreFragment（融合）が現れることを確認。
//
//  4. ON(融合) と OFF(コーディネータ) の結果、さらに baseline との一致を検証。
//
//  5. mapping を復元（Neo4j データは不変。MySQL に一時コピーが残るのみ＝bench-models と同じ非破壊）。
//
//     実行: cwd=src/go・docker スタック up 状態で
//     go test -tags integration ./integration/ -run PartialFusion -v
package integration

import (
	"context"
	"os"
	"reflect"
	"testing"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// hasRecordFragment は論理木に record-mode StoreFragment（部分融合）が含まれるかを返す。
func hasRecordFragment(n plan.PlanNode) bool {
	if n == nil {
		return false
	}
	if f, ok := n.(*plan.StoreFragment); ok && f.AsRecords {
		return true
	}
	for _, c := range n.Children() {
		if hasRecordFragment(c) {
			return true
		}
	}
	return false
}

func TestPartialFusionEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	def := workload.Registry["IS3"]
	cypher, params, _ := def(migrator.ModeGraphToRdb, true)

	prev := settings.Pushdown
	defer func() { settings.Pushdown = prev }()

	// 1. 全 graph・OFF の baseline。
	settings.Pushdown = settings.PushdownForceEngine
	baseline, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[baseline] %v", err)
	}
	if len(baseline) == 0 {
		t.Fatalf("baseline が空（IS3 の personId が sf1 に存在しない？）")
	}

	// 2. mapping スナップショット → 終了時に必ず復元。Person.firstName/lastName を rdb へ COPY。
	snap, err := os.ReadFile(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping 読込: %v", err)
	}
	defer func() {
		if werr := os.WriteFile(cfg.MappingPath, snap, 0644); werr != nil {
			t.Errorf("mapping 復元失敗: %v", werr)
		}
	}()

	migs := []migrator.MigrationConfig{{
		ObjType:      plan.Entity,
		Entity:       "Person",
		Properties:   []string{"firstName", "lastName"},
		Mode:         migrator.ModeGraphToRdb,
		DeleteSource: false, // graph 側は保持（非破壊）
	}}
	if _, err := bench.RunMigration(ctx, cfg, migs); err != nil {
		t.Fatalf("migration(copy) 失敗: %v", err)
	}

	// 3. ON で部分融合（record-mode StoreFragment）が発火することを確認。
	settings.Pushdown = settings.PushdownAuto
	root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !hasRecordFragment(root) {
		t.Fatalf("クロスストア部分融合が発火していない（record-mode StoreFragment 不在）")
	}

	// 4. 全 4 モデルで ON(融合) と OFF(コーディネータ) をクロスストア配置で実行し、
	//    それぞれが全 graph baseline と一致することを確認（bulk は融合実行、他は融合 or フォールバック）。
	base := canonRows(baseline)
	for _, model := range []settings.EngineKind{"stream", "bulk", "volcano", "vectorized"} {
		model := model
		t.Run(string(model), func(t *testing.T) {
			settings.Pushdown = settings.PushdownAuto
			onCross, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[ON cross] %v", err)
			}
			settings.Pushdown = settings.PushdownForceEngine
			offCross, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[OFF cross] %v", err)
			}
			on, off := canonRows(onCross), canonRows(offCross)
			if !reflect.DeepEqual(on, off) {
				t.Errorf("ON(融合) と OFF(コーディネータ) が不一致\n%s", firstDiff(off, on))
			}
			if !reflect.DeepEqual(on, base) {
				t.Errorf("融合クロスストア結果が全 graph baseline と不一致\n%s", firstDiff(base, on))
			}
			t.Logf("ON≡OFF≡baseline rows=%d 一致", len(onCross))
		})
	}
}
