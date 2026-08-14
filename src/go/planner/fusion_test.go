package planner

import (
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// 集約クエリで参照が全て graph → クエリ全体を graph へ委譲（Verbatim）。
func TestBuildPushdownPlanGraphWhole(t *testing.T) {
	root := &plan.Return{
		Items: []plan.ReturnItem{{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}}},
		Input: &plan.Aggregate{
			Aggs: []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: &plan.Projection{
				Units: []plan.ProjectionUnit{{Alias: "n", Fetches: []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}}},
				Input: &plan.EntityScan{Alias: "n", Labels: []string{"Person"}, DataStore: store.Graph},
			},
		},
	}
	got := BuildPushdownPlan(root, "MATCH (n:Person) RETURN count(n)", nil)
	frag, ok := got.(*plan.StoreFragment)
	if !ok {
		t.Fatalf("got %T, want *plan.StoreFragment", got)
	}
	if frag.Store != store.Graph || frag.Verbatim == "" {
		t.Errorf("frag = %+v, want graph raw", frag)
	}
}

// AggregationPushdown=OFF → 集約クエリでも委譲せず engine（コーディネータ木）へフォールバック。
func TestBuildPushdownPlanAggregationToggleOff(t *testing.T) {
	prev := settings.AggregationPushdown
	settings.AggregationPushdown = false
	defer func() { settings.AggregationPushdown = prev }()

	root := &plan.Return{
		Items: []plan.ReturnItem{{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}}},
		Input: &plan.Aggregate{
			Aggs: []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: &plan.Projection{
				Units: []plan.ProjectionUnit{{Alias: "n", Fetches: []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}}},
				Input: &plan.EntityScan{Alias: "n", Labels: []string{"Person"}, DataStore: store.Graph},
			},
		},
	}
	got := BuildPushdownPlan(root, "MATCH (n:Person) RETURN count(n)", nil)
	if _, ok := got.(*plan.StoreFragment); ok {
		t.Errorf("AggregationPushdown=false should fall back, got StoreFragment")
	}
}

// 非集約 graph クエリ → engine（コーディネータ木）へフォールバック（現行挙動の維持）。
func TestBuildPushdownPlanGraphNonAggFallback(t *testing.T) {
	root := &plan.Return{
		Items: []plan.ReturnItem{{Name: "n.name", Alias: "n", Props: []string{"name"}}},
		Input: &plan.Projection{
			Units: []plan.ProjectionUnit{{Alias: "n", Fetches: []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}}},
			Input: &plan.EntityScan{Alias: "n", Labels: []string{"Person"}, DataStore: store.Graph},
		},
	}
	got := BuildPushdownPlan(root, "MATCH (n:Person) RETURN n.name", nil)
	if _, ok := got.(*plan.StoreFragment); ok {
		t.Errorf("non-aggregate graph query should fall back, got StoreFragment")
	}
}

// 単一 relational ストアの集約 → 融合フラグメント（Plan に Return 含む論理木）。
func TestBuildPushdownPlanRelationalAggregate(t *testing.T) {
	root := &plan.Return{
		Items: []plan.ReturnItem{
			{Name: "gender", Alias: "p", Props: []string{"gender"}},
			{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}},
		},
		Input: &plan.Aggregate{
			GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "gender", OutName: "gender"}},
			Aggs:      []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: &plan.Projection{
				Units: []plan.ProjectionUnit{{Alias: "p", Fetches: []plan.FetchPlan{{Store: store.Relational, Props: []string{"gender"}}}}},
				Input: &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Relational},
			},
		},
	}
	got := BuildPushdownPlan(root, "irrelevant", nil)
	frag, ok := got.(*plan.StoreFragment)
	if !ok {
		t.Fatalf("got %T, want *plan.StoreFragment", got)
	}
	if frag.Store != store.Relational || frag.Verbatim != "" {
		t.Errorf("frag = %+v, want relational fused (no raw)", frag)
	}
	// Plan は Return を含む論理木全体（lowering は実行時に engine/core が行う）。
	if frag.Plan != plan.PlanNode(root) {
		t.Errorf("frag.Plan = %T, want 元の論理木全体", frag.Plan)
	}
	if frag.Emits != plan.EmitResult {
		t.Errorf("frag.Emits = %v, want EmitResult", frag.Emits)
	}
}

// record パイプラインが graph・projection が別ストア列を参照 → クロスストア部分融合。
// proj.Input が record-mode StoreFragment に置換され、root（Return）はそのまま返る。
func TestBuildPushdownPlanCrossStorePartialFusion(t *testing.T) {
	proj := &plan.Projection{
		InputSlot: plan.SlotTable{VarToSlot: map[string]int{"p": 0}, SlotToVar: []string{"p"}},
		Units: []plan.ProjectionUnit{{Alias: "p", Fetches: []plan.FetchPlan{
			{Store: store.Graph, Props: []string{"id"}},
			{Store: store.Relational, Props: []string{"name"}},
		}}},
		Input: &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Graph},
	}
	root := &plan.Return{
		Items: []plan.ReturnItem{{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}}},
		Input: &plan.Aggregate{
			Aggs:  []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: proj,
		},
	}
	got := BuildPushdownPlan(root, "irrelevant", nil)
	if got != plan.PlanNode(root) {
		t.Fatalf("root は維持されるべき（proj.Input のみ置換）: got %T", got)
	}
	frag, ok := proj.Input.(*plan.StoreFragment)
	if !ok {
		t.Fatalf("proj.Input = %T, want record-mode *plan.StoreFragment", proj.Input)
	}
	if frag.Emits != plan.EmitBindings || frag.Store != store.Graph {
		t.Errorf("fragment = %+v, want EmitBindings graph record-mode", frag)
	}
	if _, ok := frag.Plan.(*plan.EntityScan); !ok {
		t.Errorf("frag.Plan = %T, want original *plan.EntityScan subtree", frag.Plan)
	}
}

// record パイプラインが非 graph（scan が relational）→ 部分融合せずフォールバック。
func TestBuildPushdownPlanNonGraphRecordFallback(t *testing.T) {
	proj := &plan.Projection{
		InputSlot: plan.SlotTable{VarToSlot: map[string]int{"p": 0}, SlotToVar: []string{"p"}},
		Units: []plan.ProjectionUnit{{Alias: "p", Fetches: []plan.FetchPlan{
			{Store: store.Relational, Props: []string{"name"}},
			{Store: store.Graph, Props: []string{"bio"}},
		}}},
		Input: &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Relational},
	}
	root := &plan.Return{
		Items: []plan.ReturnItem{{Name: "p.name", Alias: "p", Props: []string{"name"}}},
		Input: proj,
	}
	BuildPushdownPlan(root, "irrelevant", nil)
	if _, ok := proj.Input.(*plan.StoreFragment); ok {
		t.Errorf("non-graph record pipeline should not be fused")
	}
}

// columnar で GROUP BY を要求 → capability 不足でフォールバック。
func TestBuildPushdownPlanColumnarGroupByFallback(t *testing.T) {
	root := &plan.Return{
		Items: []plan.ReturnItem{
			{Name: "gender", Alias: "p", Props: []string{"gender"}},
			{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}},
		},
		Input: &plan.Aggregate{
			GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "gender", OutName: "gender"}},
			Aggs:      []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: &plan.Projection{
				Units: []plan.ProjectionUnit{{Alias: "p", Fetches: []plan.FetchPlan{{Store: store.Columnar, Props: []string{"gender"}}}}},
				Input: &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Columnar},
			},
		},
	}
	got := BuildPushdownPlan(root, "irrelevant", nil)
	if _, ok := got.(*plan.StoreFragment); ok {
		t.Errorf("columnar + GROUP BY should fall back, got StoreFragment")
	}
}
