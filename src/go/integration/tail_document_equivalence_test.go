//go:build integration

// tail_document_equivalence_test は tail pushdown の document(Mongo) 対応の等価性を、非破壊 migration で検証する。
//
// シナリオ（TP1: all-graph traversal で author ごとの返信数を firstName/lastName で GROUP BY）:
//
//  1. 全 graph 配置で engine 全集約の結果を baseline として取得。
//
//  2. tail 参照プロパティ Person.{firstName,lastName} を Mongo へ COPY（DeleteSource=false）し mapping 更新。
//     → seed フィルタ Person.id は graph のまま、tail の参照列だけ document に解決する。
//
//  3. ON（TailPushdown=true）でプランに TailPushdown{Store:document} が現れることを確認。
//
//  4. ON(document tail pushdown) と OFF(engine 全集約) の結果、および baseline との一致を検証。
//
//  5. mapping を復元（Neo4j 不変。Mongo に一時コピーが残るのみ＝非破壊）。
//
//     実行: cwd=src/go・docker スタック up 状態で
//     go test -tags integration ./integration/ -run TailDocumentEquivalence -v
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

// findTailPushdown は論理木から最初の tail 委譲形 StoreFragment を返す（無ければ nil）。
// tail 委譲形は「Plan に束縛フラグメント（EmitBindings）が入れ子」という構造で判別する。
func findTailPushdown(n plan.PlanNode) *plan.StoreFragment {
	if n == nil {
		return nil
	}
	if f, ok := n.(*plan.StoreFragment); ok {
		if _, isTail := plan.LowerTail(f.Plan); isTail {
			return f
		}
	}
	for _, c := range n.Children() {
		if r := findTailPushdown(c); r != nil {
			return r
		}
	}
	return nil
}

func TestTailDocumentEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	def := workload.Registry["TP1"]
	cypher, params, _ := def(migrator.ModeGraphToDoc, true)

	prevPush, prevTail := settings.Pushdown, settings.TailPushdown
	defer func() { settings.Pushdown = prevPush; settings.TailPushdown = prevTail }()

	// 1. 全 graph・engine 全集約の baseline。
	settings.Pushdown = settings.PushdownForceEngine
	settings.TailPushdown = false
	baseline, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[baseline] %v", err)
	}
	if len(baseline) == 0 {
		t.Fatalf("baseline が空（TP1 の personId が sf1 に存在しない？）")
	}

	// 2. mapping スナップショット → 終了時に必ず復元。Person.{firstName,lastName} を Mongo へ COPY。
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
		Mode:         migrator.ModeGraphToDoc,
		DeleteSource: false, // graph 側は保持（非破壊）
	}}
	if _, err := bench.RunMigration(ctx, cfg, migs); err != nil {
		t.Fatalf("migration(copy→document) 失敗: %v", err)
	}

	// 3. ON で TailPushdown{Store:document} が発火することを確認。
	settings.Pushdown = settings.PushdownAuto
	settings.TailPushdown = true
	root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tp := findTailPushdown(root)
	if tp == nil {
		t.Fatalf("tail pushdown が発火していない（TailPushdown ノード不在）")
	}
	if tp.Store != store.Document {
		t.Fatalf("TailPushdown.Store = %s, want document", tp.Store)
	}

	// 4. ON(document tail pushdown) を実行。
	onRows, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[ON tail] %v", err)
	}

	// クロスストア配置での OFF(engine 全集約)。
	settings.TailPushdown = false
	settings.Pushdown = settings.PushdownForceEngine
	offRows, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[OFF engine] %v", err)
	}

	on, off, base := canonRows(onRows), canonRows(offRows), canonRows(baseline)
	if !reflect.DeepEqual(on, off) {
		t.Errorf("ON(document tail) と OFF(engine) が不一致\n%s", firstDiff(off, on))
	}
	if !reflect.DeepEqual(on, base) {
		t.Errorf("document tail 結果が全 graph baseline と不一致\n%s", firstDiff(base, on))
	}
	t.Logf("document tail pushdown: ON≡OFF≡baseline rows=%d 一致", len(onRows))
}
