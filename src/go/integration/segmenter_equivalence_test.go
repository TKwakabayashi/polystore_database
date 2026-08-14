//go:build integration

// segmenter_equivalence_test は Phase B の一般セグメンタ（record パイプラインを隣接同一ストアの
// 最長ランへ分割して融合）の等価性を、非破壊 migration で検証する。
//
// シナリオ（AGG1: (p:Person {id})<-[:HAS_CREATOR]-(m)<-[:REPLY_OF]-(c)-[:HAS_CREATOR]->(author)）:
//
//  1. 全 graph 配置で OFF（コーディネータ）結果を baseline として取得。
//
//  2. Person.id を MySQL へ COPY（DeleteSource=false）し mapping を rdb へ更新。
//     → seed フィルタ p.id が rdb に解決し EntityScan(rdb) となる一方、traversal は graph のまま。
//     ＝ record パイプラインが rdb → graph とストア交替する（一般セグメンタの対象）。
//
//  3. ON（GeneralSegmentation=true）でプランがランごとの StoreFragment 連鎖に分割されることを確認。
//
//  4. ON（融合）と OFF（コーディネータ）の結果、および baseline との一致を検証。
//
//  5. mapping を復元（Neo4j 不変。MySQL に一時コピーが残るのみ＝非破壊）。
//
//     実行: cwd=src/go・docker スタック up 状態で
//     go test -tags integration ./integration/ -run SegmenterEquivalence -v
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
	"polystore_database/src/go/store"
	"polystore_database/src/go/workload"
)

// fragmentStores は論理木に現れる record-mode StoreFragment のストアを葉→根の順で返す。
func fragmentStores(n plan.PlanNode) []store.Kind {
	var out []store.Kind
	var walk func(plan.PlanNode)
	walk = func(x plan.PlanNode) {
		if x == nil {
			return
		}
		if f, ok := x.(*plan.StoreFragment); ok && f.Emits == plan.EmitBindings {
			walk(f.Plan)
			out = append(out, f.Store)
			return
		}
		for _, c := range x.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

func TestSegmenterEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	def := workload.Registry["AGG1"]
	cypher, params, _ := def(migrator.ModeGraphToRdb, true)

	prevPush, prevSeg := settings.Pushdown, settings.GeneralSegmentation
	defer func() { settings.Pushdown = prevPush; settings.GeneralSegmentation = prevSeg }()

	// 1. 全 graph・OFF の baseline。
	settings.Pushdown = settings.PushdownForceEngine
	settings.GeneralSegmentation = false
	baseline, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[baseline] %v", err)
	}
	if len(baseline) == 0 {
		t.Fatalf("baseline が空（AGG1 の personId が sf1 に存在しない？）")
	}

	// 2. mapping スナップショット → 終了時に必ず復元。Person.id を rdb へ COPY。
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
		Properties:   []string{"id"},
		Mode:         migrator.ModeGraphToRdb,
		DeleteSource: false, // graph 側は保持（非破壊）
	}}
	if _, err := bench.RunMigration(ctx, cfg, migs); err != nil {
		t.Fatalf("migration(copy) 失敗: %v", err)
	}

	// 3. ON でランごとのフラグメント連鎖（rdb → graph）に分割されることを確認。
	settings.Pushdown = settings.PushdownAuto
	settings.GeneralSegmentation = true
	root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	stores := fragmentStores(root)
	if len(stores) < 2 {
		t.Fatalf("一般セグメンタが分割していない（fragment stores=%v）", stores)
	}
	if stores[0] != store.Relational || stores[len(stores)-1] != store.Graph {
		t.Fatalf("ラン構成 = %v, want 先頭 relational・末尾 graph", stores)
	}
	t.Logf("セグメント構成（葉→根）: %v", stores)

	// 4. ON（融合）と OFF（コーディネータ）をクロスストア配置で実行。
	onRows, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[ON segmented] %v", err)
	}
	settings.Pushdown = settings.PushdownForceEngine
	settings.GeneralSegmentation = false
	offRows, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[OFF engine] %v", err)
	}

	on, off, base := canonRows(onRows), canonRows(offRows), canonRows(baseline)
	if !reflect.DeepEqual(on, off) {
		t.Errorf("ON(融合) と OFF(コーディネータ) が不一致\n%s", firstDiff(off, on))
	}
	if !reflect.DeepEqual(on, base) {
		t.Errorf("融合結果が全 graph baseline と不一致\n%s", firstDiff(base, on))
	}
	t.Logf("general segmenter: ON≡OFF≡baseline rows=%d 一致", len(onRows))
}
