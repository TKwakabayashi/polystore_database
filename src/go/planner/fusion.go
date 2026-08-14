package planner

import (
	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// fusion.go は pushdown ON の「別プラン構築経路」。BuildPlan/RefinePlan が作った論理木
// （Return→[Limit]→[Sort]→[Aggregate]→Projection→record pipeline）を走査し、連続する
// 同一ストアアクセスの演算子を StoreFragment へ融合する。OFF 経路（現行コーディネータ木）は
// 一切変更しない。
//
// 各ノードは既に物理ストアを保持する（EntityScan.DataStore / Filter.DataStore /
// Projection.Units[].Fetches[].Store）ため、mapping 辞書を再参照せずツリーだけで判定できる。
//
// P2 では終端ケース（クエリ全体が単一フラグメントに畳める）と安全なフォールバックのみ実装する。
// 複数フラグメント＋Integrate（部分融合）は P3。

// BuildPushdownPlan は pushdown ON 経路のプランビルダ。融合できれば StoreFragment を、
// できなければ入力の論理木をそのまま返す（＝OFF と同じ挙動へフォールバック）。
func BuildPushdownPlan(root plan.PlanNode, query string, params map[string]string) plan.PlanNode {
	a := analyzePlan(root)
	single, ok := classifyStore(a.stores)

	// 1. クエリ全体が単一ストアに解決 → 全体委譲（row-mode）。集約クエリに限定（現行トリガ維持）。
	//    AggregationPushdown=OFF なら委譲せず engine 計算（集約 pushdown の個別トグル）。
	if a.hasAggregate && settings.AggregationPushdown && ok {
		// 全 graph（traversal 可）→ 原 Cypher を丸ごと委譲（baseline と同一発行）。
		if single == store.Graph {
			return &plan.StoreFragment{
				Store:       store.Graph,
				RawQuery:    query,
				Params:      params,
				OutputAlias: returnAliases(a.returnNode),
			}
		}
		// 非 graph 単一ストア（traversal 無し）→ ネイティブ集約へ委譲（既存 StorePushdown 実行にブリッジ）。
		if single != store.Graph && !a.hasTraversal && fragmentSupported(single, root) {
			return &plan.StoreFragment{
				Store:       single,
				Ops:         root,
				OutputAlias: returnAliases(a.returnNode),
			}
		}
	}

	// 1.5 tail pushdown（実験・settings.TailPushdown 有効時のみ）: all-graph traversal で集めた
	//     中間 UUID を単一の非 graph ストアの一時テーブルへロードし、RETURN 句 tail
	//     （Projection/Aggregate/GroupBy/Sort/Limit）をそのストアのネイティブエンジンで実行する。
	//     Case 2 のより特殊な変種として先に試み、条件を満たさなければ Case 2 へ落ちる。
	if settings.TailPushdown {
		if tp := buildTailPushdown(root); tp != nil {
			return tp
		}
	}

	// 2. クロスストア部分融合（record-mode）: record パイプライン（scan+filter+expand）が全て graph で、
	//    projection が別ストアの列を参照する場合、traversal を 1 本の Cypher（束縛 UUID 返却）へ融合する。
	//    その上の既存 Projection が他ストア列を ID キーで材料化（＝統合）→ engine で集約。集約有無に依らず適用。
	//    融合クエリ生成は各エンジン（core.BuildGraphRecordCypher）が担い、生成不能なら Ops を通常実行。
	if proj := findProjection(root); proj != nil {
		if recordPipelineAllGraph(proj.Input) && projectionHasNonGraph(proj) {
			proj.Input = &plan.StoreFragment{
				Store:       store.Graph,
				Ops:         proj.Input,
				AsRecords:   true,
				OutputSlot:  proj.InputSlot,
				OutputAlias: proj.InputAlias,
			}
			return root
		}
	}

	// フォールバック: コーディネータ木（record パイプラインが非 graph 混在など）。
	return root
}

// findProjection は論理木を根→葉へ辿って最初の Projection を返す（tail の直下）。
func findProjection(root plan.PlanNode) *plan.Projection {
	for n := root; n != nil; {
		if p, ok := n.(*plan.Projection); ok {
			return p
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

// recordPipelineAllGraph は record 部分木（Projection.Input 以下）が全て graph かを返す。
func recordPipelineAllGraph(sub plan.PlanNode) bool {
	for n := sub; n != nil; {
		switch op := n.(type) {
		case *plan.EntityScan:
			if op.DataStore != store.Graph {
				return false
			}
		case *plan.Filter:
			if op.DataStore != store.Graph {
				return false
			}
		case *plan.Expand, *plan.VarLengthExpand:
			// graph 固有
		default:
			return false
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	return true
}

// projectionHasNonGraph は Projection がいずれかの非 graph ストア列を材料化するかを返す。
func projectionHasNonGraph(p *plan.Projection) bool {
	for _, u := range p.Units {
		for _, f := range u.Fetches {
			if f.Store != store.Graph {
				return true
			}
		}
	}
	return false
}

// planAnalysis は融合判定に必要な論理木のサマリ。
type planAnalysis struct {
	stores       map[store.Kind]bool
	hasTraversal bool
	hasAggregate bool
	returnNode   *plan.Return
}

// analyzePlan は論理木を根→葉へ走査して参照ストア集合・traversal / 集約の有無・Return を集める。
func analyzePlan(root plan.PlanNode) planAnalysis {
	a := planAnalysis{stores: map[store.Kind]bool{}}
	for n := root; n != nil; {
		switch op := n.(type) {
		case *plan.Return:
			a.returnNode = op
		case *plan.Aggregate:
			a.hasAggregate = true
		case *plan.Projection:
			for _, u := range op.Units {
				for _, f := range u.Fetches {
					a.stores[f.Store] = true
				}
			}
		case *plan.Filter:
			a.stores[op.DataStore] = true
		case *plan.Expand:
			a.hasTraversal = true
		case *plan.VarLengthExpand:
			a.hasTraversal = true
		case *plan.EntityScan:
			a.stores[op.DataStore] = true
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	return a
}

// classifyStore は参照ストア集合が単一ストアに解決できるかを返す（空集合＝graph 扱い）。
func classifyStore(stores map[store.Kind]bool) (store.Kind, bool) {
	switch len(stores) {
	case 0:
		return store.Graph, true
	case 1:
		for k := range stores {
			return k, true
		}
	}
	return store.Graph, false
}

// fragmentSupported は ops 部分木の全演算子を single ストアがネイティブ実行できるかを返す。
// columnar/kvs の制約（GroupBy/Sort/Distinct 不可、集約不可）は capability テーブルで弾かれる。
func fragmentSupported(single store.Kind, ops plan.PlanNode) bool {
	for n := ops; n != nil; {
		switch op := n.(type) {
		case *plan.Limit:
			if !plan.Supports(single, plan.CapLimit) {
				return false
			}
		case *plan.Sort:
			if !plan.Supports(single, plan.CapSort) {
				return false
			}
		case *plan.Aggregate:
			if !plan.Supports(single, plan.CapAggregate) {
				return false
			}
			if len(op.GroupKeys) > 0 && !plan.Supports(single, plan.CapGroupBy) {
				return false
			}
			for _, ag := range op.Aggs {
				if ag.Distinct && !plan.Supports(single, plan.CapDistinct) {
					return false
				}
			}
		case *plan.Projection:
			if !plan.Supports(single, plan.CapProject) {
				return false
			}
		case *plan.Filter:
			if !plan.Supports(single, plan.CapFilter) {
				return false
			}
		case *plan.EntityScan:
			// ネイティブクエリの FROM 句にラベルが必要。
			if len(op.Labels) == 0 || op.Labels[0] == "" {
				return false
			}
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	return true
}

// buildTailPushdown は tail pushdown が適用可能なら TailPushdown を、不可なら nil を返す。
// 適用条件:
//   - record パイプライン（Projection.Input 以下）が all-graph（traversal は graph で完結）。
//   - tail に Aggregate が存在する（sort/limit のみは対象外）。
//   - tail が参照する全プロパティが単一の非 graph ストア S に解決できる。
//   - S が Aggregate/Project（＋必要なら GroupBy/Sort/Limit）を capability 上サポートする。
//   - staging するエンティティ（group/prop-agg の alias）の永続テーブル（Label）が判明する。
func buildTailPushdown(root plan.PlanNode) *plan.TailPushdown {
	proj := findProjection(root)
	if proj == nil || !recordPipelineAllGraph(proj.Input) {
		return nil
	}
	agg := findAggregate(root)
	if agg == nil {
		return nil
	}
	s, ok := tailStore(proj)
	if !ok {
		return nil
	}
	if !plan.Supports(s, plan.CapAggregate) || !plan.Supports(s, plan.CapProject) {
		return nil
	}
	if len(agg.GroupKeys) > 0 && !plan.Supports(s, plan.CapGroupBy) {
		return nil
	}
	sortNode := findSort(root)
	if sortNode != nil && !plan.Supports(s, plan.CapSort) {
		return nil
	}
	limitNode := findLimit(root)
	if limitNode != nil && !plan.Supports(s, plan.CapLimit) {
		return nil
	}
	ret := findReturn(root)
	entities, ok := tailEntities(proj, agg, ret)
	if !ok {
		return nil
	}

	src := &plan.StoreFragment{
		Store:       store.Graph,
		Ops:         proj.Input,
		AsRecords:   true,
		OutputSlot:  proj.InputSlot,
		OutputAlias: proj.InputAlias,
	}
	tp := &plan.TailPushdown{
		Store:     s,
		Input:     src,
		Fallback:  root, // 未対応エンジンは元 coordinator tail を実行（結果等価）
		InputSlot: proj.InputSlot,
		Entities:  entities,
		GroupKeys: agg.GroupKeys,
		Aggs:      agg.Aggs,
	}
	if ret != nil {
		tp.Return = ret.Items
	}
	if sortNode != nil {
		tp.OrderItems = sortNode.OrderItems
	}
	if limitNode != nil {
		tp.Limit = limitNode.Count
	}
	return tp
}

// tailStore は Projection の全 fetch（props を持つもの）が単一の非 graph ストアに解決できるか判定し、
// できればそのストアを返す。graph の fetch が混在する／非 graph が複数ある場合は (_, false)。
func tailStore(p *plan.Projection) (store.Kind, bool) {
	var s store.Kind
	found := false
	for _, u := range p.Units {
		for _, f := range u.Fetches {
			if len(f.Props) == 0 {
				continue // 純束縛（プロパティ無し）はストア制約を与えない
			}
			if f.Store == store.Graph {
				return store.Graph, false // tail プロパティが graph に残る → 単一非 graph に載らない
			}
			if !found {
				s, found = f.Store, true
			} else if f.Store != s {
				return store.Graph, false // 非 graph が複数
			}
		}
	}
	return s, found
}

// tailEntities は staging すべきエンティティ（group key の alias ＋ prop/distinct を持つ agg の alias ＋
// 非集約 RETURN 項目の alias）を集め、各 alias の永続テーブル（Projection unit の Label）を解決する。
// いずれかの alias でテーブルが不明なら (_, false)。
func tailEntities(p *plan.Projection, agg *plan.Aggregate, ret *plan.Return) ([]plan.TailEntity, bool) {
	need := map[string]bool{}
	for _, gk := range agg.GroupKeys {
		if gk.Alias != "" {
			need[gk.Alias] = true
		}
	}
	for _, a := range agg.Aggs {
		if a.Alias != "" && (a.Prop != "" || a.Distinct) {
			need[a.Alias] = true
		}
	}
	if ret != nil {
		for _, it := range ret.Items {
			if !it.IsAggregate && it.Alias != "" {
				need[it.Alias] = true
			}
		}
	}

	labelOf := map[string]string{}
	propsOf := map[string][]string{}
	for _, u := range p.Units {
		if len(u.Labels) > 0 {
			labelOf[u.Alias] = u.Labels[0]
		}
		for _, f := range u.Fetches {
			propsOf[u.Alias] = append(propsOf[u.Alias], f.Props...)
		}
	}

	entities := make([]plan.TailEntity, 0, len(need))
	for alias := range need {
		table := labelOf[alias]
		if table == "" {
			return nil, false // JOIN 先テーブルが不明
		}
		entities = append(entities, plan.TailEntity{Alias: alias, Table: table, Props: propsOf[alias]})
	}
	return entities, true
}

// findAggregate / findSort / findLimit / findReturn は tail チェーン（根→第1子）を辿って各ノードを返す。
func findAggregate(root plan.PlanNode) *plan.Aggregate {
	for n := root; n != nil; {
		if a, ok := n.(*plan.Aggregate); ok {
			return a
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

func findSort(root plan.PlanNode) *plan.Sort {
	for n := root; n != nil; {
		if s, ok := n.(*plan.Sort); ok {
			return s
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

func findLimit(root plan.PlanNode) *plan.Limit {
	for n := root; n != nil; {
		if l, ok := n.(*plan.Limit); ok {
			return l
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

func findReturn(root plan.PlanNode) *plan.Return {
	for n := root; n != nil; {
		if r, ok := n.(*plan.Return); ok {
			return r
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

// returnAliases は Return 項目の出力名を返す（統合演算子の結線用の出力束縛）。
func returnAliases(r *plan.Return) []string {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.Items))
	for _, it := range r.Items {
		names = append(names, it.Name)
	}
	return names
}
