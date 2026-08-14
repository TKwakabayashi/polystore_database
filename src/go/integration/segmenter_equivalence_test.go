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

	// 4. 全 4 モデルで ON（融合）と OFF（コーディネータ）をクロスストア配置で実行し、
	//    それぞれが全 graph baseline と一致することを確認。
	base := canonRows(baseline)
	for _, model := range []settings.EngineKind{"stream", "bulk", "volcano", "vectorized"} {
		model := model
		t.Run(string(model), func(t *testing.T) {
			settings.Pushdown = settings.PushdownAuto
			settings.GeneralSegmentation = true
			onRows, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[ON segmented] %v", err)
			}
			settings.Pushdown = settings.PushdownForceEngine
			settings.GeneralSegmentation = false
			offRows, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[OFF engine] %v", err)
			}
			on, off := canonRows(onRows), canonRows(offRows)
			if !reflect.DeepEqual(on, off) {
				t.Errorf("ON(融合) と OFF(コーディネータ) が不一致\n%s", firstDiff(off, on))
			}
			if !reflect.DeepEqual(on, base) {
				t.Errorf("融合結果が全 graph baseline と不一致\n%s", firstDiff(base, on))
			}
			t.Logf("ON≡OFF≡baseline rows=%d 一致", len(onRows))
		})
	}
}

// TestSegmenterAllGraphFusion は「全 graph・非集約」クエリ（従来は融合対象外）が
// 一般セグメンタで単一 graph フラグメント（1 Cypher）へ畳まれ、結果が OFF と一致することを検証する。
// 配置変更は不要（全 graph のまま）なので migration も要らない。
func TestSegmenterAllGraphFusion(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	def := workload.Registry["IS3"] // (p:Person {id})-[r:KNOWS]-(friend) 非集約・全 graph
	cypher, params, _ := def(migrator.ModeGraphToRdb, true)

	prevPush, prevSeg := settings.Pushdown, settings.GeneralSegmentation
	defer func() { settings.Pushdown = prevPush; settings.GeneralSegmentation = prevSeg }()

	// OFF（コーディネータ）。
	settings.Pushdown = settings.PushdownForceEngine
	settings.GeneralSegmentation = false
	offRows, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[OFF] %v", err)
	}
	if len(offRows) == 0 {
		t.Fatalf("OFF の結果が空")
	}

	// ON: 単一 graph ランへ畳まれることをプラン上で確認。
	settings.Pushdown = settings.PushdownAuto
	settings.GeneralSegmentation = true
	root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	stores := fragmentStores(root)
	if len(stores) != 1 || stores[0] != store.Graph {
		t.Fatalf("全 graph 融合が発火していない（fragment stores=%v）", stores)
	}

	off := canonRows(offRows)
	for _, model := range []settings.EngineKind{"stream", "bulk", "volcano", "vectorized"} {
		model := model
		t.Run(string(model), func(t *testing.T) {
			settings.Pushdown = settings.PushdownAuto
			settings.GeneralSegmentation = true
			onRows, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[ON fused] %v", err)
			}
			if !reflect.DeepEqual(canonRows(onRows), off) {
				t.Errorf("融合結果が OFF と不一致\n%s", firstDiff(off, canonRows(onRows)))
			}
			t.Logf("all-graph fused ≡ OFF rows=%d 一致", len(onRows))
		})
	}
}

// hasIntegrate は論理木に統合演算子 Integrate が含まれるかを返す。
func hasIntegrate(n plan.PlanNode) bool {
	if n == nil {
		return false
	}
	if _, ok := n.(*plan.Integrate); ok {
		return true
	}
	for _, c := range n.Children() {
		if hasIntegrate(c) {
			return true
		}
	}
	return false
}

// TestExplicitIntegrateEquivalence は「複数ストアから materialize する Projection が Integrate へ
// 明示化されても結果が変わらない」ことを、非破壊 migration によるクロスストア配置で検証する。
// IS3 の friend.firstName/lastName を rdb へ移すと materialize が graph＋rdb に散り、統合が発生する。
func TestExplicitIntegrateEquivalence(t *testing.T) {
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

	prevPush, prevInt := settings.Pushdown, settings.ExplicitIntegrate
	defer func() { settings.Pushdown = prevPush; settings.ExplicitIntegrate = prevInt }()

	// 全 graph・OFF の baseline。
	settings.Pushdown = settings.PushdownForceEngine
	settings.ExplicitIntegrate = false
	baseline, err := runRows(ctx, cfg, "bulk", cypher, params)
	if err != nil {
		t.Fatalf("[baseline] %v", err)
	}
	if len(baseline) == 0 {
		t.Fatalf("baseline が空")
	}

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
		DeleteSource: false,
	}}
	if _, err := bench.RunMigration(ctx, cfg, migs); err != nil {
		t.Fatalf("migration(copy) 失敗: %v", err)
	}

	// ON: プランに Integrate が現れることを確認。
	settings.Pushdown = settings.PushdownAuto
	settings.ExplicitIntegrate = true
	root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !hasIntegrate(root) {
		t.Fatalf("統合演算子 Integrate が明示化されていない")
	}

	base := canonRows(baseline)
	for _, model := range []settings.EngineKind{"stream", "bulk", "volcano", "vectorized"} {
		model := model
		t.Run(string(model), func(t *testing.T) {
			settings.Pushdown = settings.PushdownAuto
			settings.ExplicitIntegrate = true
			onRows, err := runRows(ctx, cfg, model, cypher, params)
			if err != nil {
				t.Fatalf("[ON integrate] %v", err)
			}
			if !reflect.DeepEqual(canonRows(onRows), base) {
				t.Errorf("Integrate 明示化で結果が変化\n%s", firstDiff(base, canonRows(onRows)))
			}
			t.Logf("explicit Integrate ≡ baseline rows=%d 一致", len(onRows))
		})
	}
}

// TestProjectionPushdownEquivalence は非集約クエリの RETURN 列をストアへ畳み込む
// ProjectionPushdown が、OFF（コーディネータ materialize）と同じ結果を返すことを検証する。
//   - 全 graph 配置: 原 Cypher を verbatim 委譲（Neo4j が RETURN を計算）。
//   - 非 graph 配置（traversal 無し）: ネイティブ SELECT へ畳み込み（GROUP BY を付けない事の確認も兼ねる）。
func TestProjectionPushdownEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	prevPush, prevProj := settings.Pushdown, settings.ProjectionPushdown
	defer func() { settings.Pushdown = prevPush; settings.ProjectionPushdown = prevProj }()

	// --- (1) 全 graph・traversal あり（IS3）→ verbatim 委譲 ---
	t.Run("graph_verbatim", func(t *testing.T) {
		def := workload.Registry["IS3"]
		cypher, params, _ := def(migrator.ModeGraphToRdb, true)

		settings.Pushdown = settings.PushdownForceEngine
		settings.ProjectionPushdown = false
		off, err := runRows(ctx, cfg, "bulk", cypher, params)
		if err != nil {
			t.Fatalf("[OFF] %v", err)
		}
		if len(off) == 0 {
			t.Fatalf("OFF が空")
		}

		settings.Pushdown = settings.PushdownAuto
		settings.ProjectionPushdown = true
		on, err := runRows(ctx, cfg, "bulk", cypher, params)
		if err != nil {
			t.Fatalf("[ON] %v", err)
		}
		if !reflect.DeepEqual(canonRows(on), canonRows(off)) {
			t.Errorf("projection pushdown(graph) で結果が変化\n%s", firstDiff(canonRows(off), canonRows(on)))
		}
		t.Logf("graph verbatim ≡ OFF rows=%d 一致", len(on))
	})

	// --- (2) 非 graph・traversal 無し → ネイティブ SELECT（GROUP BY を付けない事の確認も兼ねる） ---
	t.Run("relational_select", func(t *testing.T) {
		// traversal を含まない単一エンティティの非集約クエリ（Organisation.id を rdb へ移す）。
		cypher := "MATCH (o:Organisation)\nWHERE o.id >= $minId AND o.id <= $maxId\nRETURN o.id\nORDER BY o.id ASC\nLIMIT 20"
		params := map[string]string{"minId": "100", "maxId": "5000"}

		settings.Pushdown = settings.PushdownForceEngine
		settings.ProjectionPushdown = false
		baseline, err := runRows(ctx, cfg, "bulk", cypher, params)
		if err != nil {
			t.Fatalf("[baseline] %v", err)
		}
		if len(baseline) == 0 {
			t.Fatalf("baseline が空")
		}

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
			Entity:       "Organisation",
			Properties:   []string{"id"},
			Mode:         migrator.ModeGraphToRdb,
			DeleteSource: false,
		}}
		if _, err := bench.RunMigration(ctx, cfg, migs); err != nil {
			t.Fatalf("migration(copy) 失敗: %v", err)
		}

		settings.Pushdown = settings.PushdownAuto
		settings.ProjectionPushdown = true
		root, err := planner.ParseQuery(cypher, cfg.MappingPath, params)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		f, ok := root.(*plan.StoreFragment)
		if !ok || f.Store != store.Relational {
			t.Fatalf("非 graph への projection pushdown が発火していない: %T", root)
		}

		on, err := runRows(ctx, cfg, "bulk", cypher, params)
		if err != nil {
			t.Fatalf("[ON] %v", err)
		}
		if !reflect.DeepEqual(canonRows(on), canonRows(baseline)) {
			t.Errorf("projection pushdown(relational) で結果が変化\n%s", firstDiff(canonRows(baseline), canonRows(on)))
		}
		t.Logf("relational SELECT ≡ baseline rows=%d 一致", len(on))
	})
}

// TestSmallChunkEquivalence は ID 材料化のチャンク分割が結果を変えないことを検証する。
// MaterializeChunkSize を極端に小さくして「必ず複数チャンクに割れる」状態を作り、
// 分割なし（既定）と結果が一致することを確認する（漏れ・重複が無いことの実 DB 検証）。
func TestSmallChunkEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	prevPush, prevChunk := settings.Pushdown, settings.MaterializeChunkSize
	defer func() { settings.Pushdown = prevPush; settings.MaterializeChunkSize = prevChunk }()
	settings.Pushdown = settings.PushdownForceEngine // engine 経路（materialize を必ず通す）

	for _, name := range []string{"IS3", "Q11"} {
		name := name
		t.Run(name, func(t *testing.T) {
			def := workload.Registry[name]
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)

			settings.MaterializeChunkSize = 0 // ストア既定（実質分割なし）
			base, err := runRows(ctx, cfg, "bulk", cypher, params)
			if err != nil {
				t.Fatalf("[既定] %v", err)
			}
			if len(base) == 0 {
				t.Fatalf("結果が空")
			}

			settings.MaterializeChunkSize = 7 // 必ず複数チャンクに割れる
			small, err := runRows(ctx, cfg, "bulk", cypher, params)
			if err != nil {
				t.Fatalf("[chunk=7] %v", err)
			}
			if !reflect.DeepEqual(canonRows(small), canonRows(base)) {
				t.Errorf("チャンク分割で結果が変化\n%s", firstDiff(canonRows(base), canonRows(small)))
			}
			t.Logf("chunk=7 ≡ 既定 rows=%d 一致", len(small))
		})
	}
}
