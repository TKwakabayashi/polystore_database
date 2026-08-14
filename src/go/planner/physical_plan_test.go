package planner

import (
	"fmt"
	"strings"
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// physical_plan_test は BuildPushdownPlan が生成する**物理プランの形**を固定する golden 風テスト。
// 等価性テスト（integration）は結果しか見ないため、プラン形状の意図しない変化はここで検知する。
//
// 描画規約:
//   - 1 行 1 ノード。子はインデント。
//   - StoreFragment は Plan（委譲する論理サブプラン）を "Plan⇒" 配下に、Integrate は Inputs を展開する。

func renderPlan(n plan.PlanNode, indent string, sb *strings.Builder) {
	if n == nil {
		return
	}
	fmt.Fprintf(sb, "%s%s\n", indent, nodeLabel(n))
	switch o := n.(type) {
	case *plan.StoreFragment:
		if o.Plan != nil {
			fmt.Fprintf(sb, "%s  Plan⇒\n", indent)
			renderPlan(o.Plan, indent+"    ", sb)
		}
	default:
		for _, c := range n.Children() {
			renderPlan(c, indent+"  ", sb)
		}
	}
}

// nodeLabel はノード種別を安定した短い表記にする（String() は slot map を含み順序が不定なため使わない）。
func nodeLabel(n plan.PlanNode) string {
	switch o := n.(type) {
	case *plan.StoreFragment:
		kind := "plan"
		if o.Verbatim != "" {
			kind = "verbatim"
		}
		return fmt.Sprintf("StoreFragment[%s %s %s]", o.Store, o.Emits, kind)
	case *plan.Integrate:
		var stores []string
		for _, u := range o.Units {
			for _, f := range u.Fetches {
				if len(f.Props) > 0 {
					stores = append(stores, f.Store.String())
				}
			}
		}
		return fmt.Sprintf("Integrate[keys=%d stores=%s]", len(o.Keys), strings.Join(stores, "+"))
	case *plan.Projection:
		var stores []string
		for _, u := range o.Units {
			for _, f := range u.Fetches {
				if len(f.Props) > 0 {
					stores = append(stores, f.Store.String())
				}
			}
		}
		return "Projection[" + strings.Join(stores, "+") + "]"
	case *plan.Return:
		return "Return"
	case *plan.Limit:
		return "Limit"
	case *plan.Sort:
		return "Sort"
	case *plan.Aggregate:
		return "Aggregate"
	case *plan.EntityScan:
		return fmt.Sprintf("EntityScan[%s %s]", o.Alias, o.DataStore)
	case *plan.Filter:
		return fmt.Sprintf("Filter[%s %s]", o.Alias, o.DataStore)
	case *plan.Expand:
		return fmt.Sprintf("Expand[%s->%s]", o.SourceEntity, o.TargetEntity)
	case *plan.VarLengthExpand:
		return fmt.Sprintf("VarLengthExpand[%s->%s]", o.SourceEntity, o.TargetEntity)
	}
	return fmt.Sprintf("%T", n)
}

func assertShape(t *testing.T, root plan.PlanNode, want string) {
	t.Helper()
	var sb strings.Builder
	renderPlan(root, "", &sb)
	got := strings.TrimRight(sb.String(), "\n")
	want = strings.TrimRight(strings.TrimLeft(want, "\n"), "\n")
	if got != want {
		t.Errorf("物理プラン形状が不一致\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// ---- プラン構築ヘルパ（各テストが同じ論理木から始められるようにする）----

func slots(aliases ...string) plan.SlotTable {
	st := plan.SlotTable{VarToSlot: map[string]int{}, SlotToVar: append([]string{}, aliases...)}
	for i, a := range aliases {
		st.VarToSlot[a] = i
	}
	return st
}

// traversalPlan は (p)-[r]->(friend) の record パイプライン＋Projection＋Return を組む。
// scanStore / fetches で配置を変えて各戦略を誘発する。
func traversalPlan(scanStore store.Kind, fetches []plan.FetchPlan, agg bool) (*plan.Return, *plan.Projection) {
	scan := &plan.EntityScan{
		Alias: "p", Labels: []string{"Person"}, DataStore: scanStore,
		OutputSlot: slots("p"),
	}
	exp := &plan.Expand{
		Alias: "r", RelLabel: "KNOWS", Dir: plan.Outgoing,
		SourceEntity: "p", TargetEntity: "friend", TargetLabels: []string{"Person"},
		Input: scan, OutputSlot: slots("friend"),
	}
	proj := &plan.Projection{
		InputSlot: slots("friend"), InputAlias: []string{"friend"},
		Units: []plan.ProjectionUnit{{Alias: "friend", Labels: []string{"Person"}, Fetches: fetches}},
		Input: exp,
	}
	var top plan.PlanNode = proj
	if agg {
		top = &plan.Aggregate{
			Aggs:  []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input: proj,
		}
	}
	items := []plan.ReturnItem{{Name: "friend.name", Alias: "friend", Props: []string{"name"}}}
	if agg {
		items = []plan.ReturnItem{{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}}}
	}
	return &plan.Return{Items: items, Input: top}, proj
}

// ---- 各戦略の物理プラン形状 ----

// 戦略1: 全 graph の集約 → クエリ全体を graph へ verbatim 委譲（単一ノードに畳まれる）。
func TestPhysicalPlanWholeGraphVerbatim(t *testing.T) {
	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}, true)
	got := BuildPushdownPlan(root, "MATCH (p)-[r:KNOWS]->(friend) RETURN count(*)", nil)
	assertShape(t, got, `
StoreFragment[graph result verbatim]
  Plan⇒
    Return
      Aggregate
        Projection[graph]
          Expand[p->friend]
            EntityScan[p graph]`)
}

// 戦略2: クロスストア部分融合 → Projection.Input が record-mode フラグメントに置換される。
func TestPhysicalPlanCrossStorePartialFusion(t *testing.T) {
	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{
		{Store: store.Graph, Props: []string{"id"}},
		{Store: store.Relational, Props: []string{"name"}},
	}, false)
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Projection[graph+relational]
    StoreFragment[graph bindings plan]
      Plan⇒
        Expand[p->friend]
          EntityScan[p graph]`)
}

// 戦略3: 一般セグメンタ → ストア交替（rdb scan → graph expand）がランごとのフラグメント連鎖になる。
func TestPhysicalPlanGeneralSegmenter(t *testing.T) {
	prev := settings.GeneralSegmentation
	settings.GeneralSegmentation = true
	defer func() { settings.GeneralSegmentation = prev }()

	root, _ := traversalPlan(store.Relational, []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}, false)
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Projection[graph]
    StoreFragment[graph bindings plan]
      Plan⇒
        Expand[p->friend]
          StoreFragment[relational bindings plan]
            Plan⇒
              EntityScan[p relational]`)
}

// 戦略4: 統合の明示化 → 複数ストア materialize の Projection が Integrate になる。
func TestPhysicalPlanExplicitIntegrate(t *testing.T) {
	prev := settings.ExplicitIntegrate
	settings.ExplicitIntegrate = true
	defer func() { settings.ExplicitIntegrate = prev }()

	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{
		{Store: store.Graph, Props: []string{"id"}},
		{Store: store.Relational, Props: []string{"name"}},
	}, false)
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Integrate[keys=1 stores=graph+relational]
    StoreFragment[graph bindings plan]
      Plan⇒
        Expand[p->friend]
          EntityScan[p graph]`)
}

// 単一ストア materialize では Integrate へ置換しない（統合が起きていないため）。
func TestPhysicalPlanNoIntegrateForSingleStore(t *testing.T) {
	prev := settings.ExplicitIntegrate
	settings.ExplicitIntegrate = true
	defer func() { settings.ExplicitIntegrate = prev }()

	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}, false)
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Projection[graph]
    Expand[p->friend]
      EntityScan[p graph]`)
}

// ProjectionPushdown: 非集約でも単一ストアに解決すれば丸ごと委譲する（materialize 往復を無くす）。
func TestPhysicalPlanProjectionPushdown(t *testing.T) {
	prev := settings.ProjectionPushdown
	settings.ProjectionPushdown = true
	defer func() { settings.ProjectionPushdown = prev }()

	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}, false)
	got := BuildPushdownPlan(root, "MATCH (p)-[r:KNOWS]->(friend) RETURN friend.name", nil)
	assertShape(t, got, `
StoreFragment[graph result verbatim]
  Plan⇒
    Return
      Projection[graph]
        Expand[p->friend]
          EntityScan[p graph]`)
}

// ProjectionPushdown=OFF なら非集約は委譲しない（既定挙動）。
func TestPhysicalPlanProjectionPushdownOff(t *testing.T) {
	prev := settings.ProjectionPushdown
	settings.ProjectionPushdown = false
	defer func() { settings.ProjectionPushdown = prev }()

	root, _ := traversalPlan(store.Graph, []plan.FetchPlan{{Store: store.Graph, Props: []string{"name"}}}, false)
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Projection[graph]
    Expand[p->friend]
      EntityScan[p graph]`)
}

// OFF 経路（PushdownForceEngine）は ParseQuery が分岐するため BuildPushdownPlan を呼ばない。
// ここでは「融合できない入力はそのまま返る」ことを確認する（capability 不足のケース）。
func TestPhysicalPlanFallbackKeepsCoordinatorTree(t *testing.T) {
	// columnar は CapGroupBy=false のため集約を委譲できない。
	scan := &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Columnar, OutputSlot: slots("p")}
	proj := &plan.Projection{
		InputSlot: slots("p"),
		Units:     []plan.ProjectionUnit{{Alias: "p", Fetches: []plan.FetchPlan{{Store: store.Columnar, Props: []string{"gender"}}}}},
		Input:     scan,
	}
	root := &plan.Return{
		Items: []plan.ReturnItem{
			{Name: "gender", Alias: "p", Props: []string{"gender"}},
			{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}},
		},
		Input: &plan.Aggregate{
			GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "gender", OutName: "gender"}},
			Aggs:      []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input:     proj,
		},
	}
	got := BuildPushdownPlan(root, "irrelevant", nil)
	assertShape(t, got, `
Return
  Aggregate
    Projection[columnar]
      EntityScan[p columnar]`)
}
