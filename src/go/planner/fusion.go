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

	// 1. クエリ全体が単一ストアに解決 → 全体委譲（row-mode）。
	//    集約クエリは AggregationPushdown、非集約クエリは ProjectionPushdown（＝ RETURN 列を
	//    ストアのネイティブ SELECT/RETURN へ畳み込み、materialize の往復を無くす）で個別に切り替える。
	delegate := (a.hasAggregate && settings.AggregationPushdown) ||
		(!a.hasAggregate && settings.ProjectionPushdown)
	if delegate && ok {
		// 全 graph（traversal 可）→ 原 Cypher を丸ごと委譲（baseline と同一発行）。
		// Plan には論理木を持たせ、Verbatim は「その既知の忠実な lowering」として添える。
		if single == store.Graph {
			return &plan.StoreFragment{
				Store:       store.Graph,
				Plan:        root,
				Emits:       plan.EmitResult,
				Verbatim:    query,
				Params:      params,
				OutputAlias: returnAliases(a.returnNode),
			}
		}
		// 非 graph 単一ストア（traversal 無し）→ ネイティブ集約へ委譲（Plan を lowering して発行）。
		if single != store.Graph && !a.hasTraversal && fragmentSupported(single, root) {
			return &plan.StoreFragment{
				Store:       single,
				Plan:        root,
				Emits:       plan.EmitResult,
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
	//    融合クエリ生成は各エンジン（core.BuildGraphRecordCypher）が担い、生成不能なら Plan を通常実行。
	if proj := findProjection(root); proj != nil {
		switch {
		case recordPipelineAllGraph(proj.Input) && projectionHasNonGraph(proj):
			proj.Input = &plan.StoreFragment{
				Store:       store.Graph,
				Plan:        proj.Input,
				Emits:       plan.EmitBindings,
				OutputSlot:  proj.InputSlot,
				OutputAlias: proj.InputAlias,
			}

		// 3. 一般セグメンタ: record パイプラインがストアをまたぐ場合、隣接同一ストアの最長ランへ
		//    分割して融合する（境界は下位ランのフラグメントを入れ子にして表現）。
		case settings.GeneralSegmentation:
			proj.Input = segmentRecordPipeline(proj.Input)
		}
	}

	// 4. 統合の明示化: materialize するカラムが 2 ストア以上に散る Projection を Integrate へ置換する。
	//    実行は同一（ID 材料化）だが、統合が起きる場所がプラン上に現れる。
	//    record パイプラインの処理（戦略2/3）の後に適用する。
	if settings.ExplicitIntegrate {
		explicitIntegrate(root)
	}

	// 融合できなかった部分はコーディネータ木のまま（結果は等価）。
	return root
}

// explicitIntegrate は tail 直下の Projection が複数ストアから materialize する場合、
// それを Integrate（統合演算子）へ置き換える。親ノードの Input を差し替える必要があるため、
// 根から辿って親を保持しながら走査する。
func explicitIntegrate(root plan.PlanNode) {
	var parent plan.PlanNode
	for n := root; n != nil; {
		if p, ok := n.(*plan.Projection); ok {
			if parent == nil || len(projectionStores(p)) < 2 {
				return
			}
			setTailInput(parent, integrateFromProjection(p))
			return
		}
		ch := n.Children()
		if len(ch) == 0 {
			return
		}
		parent, n = n, ch[0]
	}
}

// projectionStores は Projection が materialize するストア集合を返す。
func projectionStores(p *plan.Projection) map[store.Kind]bool {
	stores := map[store.Kind]bool{}
	for _, u := range p.Units {
		for _, f := range u.Fetches {
			if len(f.Props) > 0 {
				stores[f.Store] = true
			}
		}
	}
	return stores
}

// integrateFromProjection は Projection を等価な Integrate へ変換する
// （結合キーは束縛 UUID＝KeyID、必要カラムは各 Fetch のプロパティ）。
func integrateFromProjection(p *plan.Projection) *plan.Integrate {
	ig := &plan.Integrate{
		Units:      p.Units,
		InputAlias: p.InputAlias,
		InputSlot:  p.InputSlot,
		OutputSlot: p.InputSlot,
	}
	if p.Input != nil {
		ig.Inputs = []plan.PlanNode{p.Input}
	}
	for _, u := range p.Units {
		hasProps := false
		for _, f := range u.Fetches {
			for _, prop := range f.Props {
				ig.Needed = append(ig.Needed, plan.ColumnRef{Alias: u.Alias, Prop: prop})
				hasProps = true
			}
		}
		if hasProps {
			ig.Keys = append(ig.Keys, plan.IntegrateKey{
				Kind: plan.KeyID,
				Refs: []plan.ColumnRef{{Alias: u.Alias}},
			})
		}
	}
	return ig
}

// setTailInput は tail 演算子の入力を差し替える（Projection → Integrate の置換用）。
func setTailInput(parent plan.PlanNode, in plan.PlanNode) {
	switch o := parent.(type) {
	case *plan.Return:
		o.Input = in
	case *plan.Limit:
		o.Input = in
	case *plan.Sort:
		o.Input = in
	case *plan.Aggregate:
		o.Input = in
	}
}

// ===== 一般セグメンタ（capability 駆動） =====

// segmentRecordPipeline は record パイプライン（Projection.Input 以下）を「隣接する同一ストアの
// 最長ラン」へ分割し、融合可能なランを StoreFragment（Emits:Bindings）へ包んで返す。
// ラン境界は「上位ランの Plan 連鎖の末端に下位ランのフラグメントが入れ子になる」形で表現し、
// 実行側は境界フラグメントの束縛 uuid を IN-list として上位ランのクエリへ注入する。
//
// 実行順序は変えない（隣接演算子をまとめるだけ）。融合できないランはそのままの論理演算子として残り、
// コーディネータが従来どおり実行する（結果は等価）。
//
// 現状 lowering があるのは graph ラン（core.BuildGraphRecordCypher）のみ。非 graph ランは
// フラグメント化しても lowering 不能でフォールバックするため、無駄な入れ子を避けて包まない。
func segmentRecordPipeline(sub plan.PlanNode) plan.PlanNode {
	// 葉→根の順に演算子を並べる。
	var chain []plan.PlanNode
	for n := sub; n != nil; {
		chain = append(chain, n)
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	// 隣接同一ストアでランに区切る。
	type run struct {
		store store.Kind
		ops   []plan.PlanNode
	}
	var runs []run
	for _, op := range chain {
		k, ok := recordOpStore(op)
		if !ok {
			return sub // 想定外の演算子 → 融合しない
		}
		if len(runs) > 0 && runs[len(runs)-1].store == k {
			runs[len(runs)-1].ops = append(runs[len(runs)-1].ops, op)
			continue
		}
		runs = append(runs, run{store: k, ops: []plan.PlanNode{op}})
	}
	// 単一ランでも 2 演算子以上なら畳む価値がある（例: all-graph traversal を 1 Cypher へ）。
	// 1 演算子だけのランは既存の scan 実行と往復数が変わらないため包まない。
	if len(runs) == 0 || (len(runs) == 1 && len(runs[0].ops) < 2) {
		return sub
	}

	// 葉→根へ、各ランをフラグメントへ包む。lower は直前ランのフラグメント（＝境界）。
	//
	// 全ランを包むのが要点: 上位ランの lowering が「連鎖の末端に別ストアの素の演算子」を見ると
	// 必ず生成失敗するため、境界は常にフラグメントとして現れる必要がある。lowering を持たない
	// ストアのランはフラグメントのまま Plan を通常実行してフォールバックする（結果は等価）。
	var lower plan.PlanNode
	for _, r := range runs {
		top := r.ops[len(r.ops)-1] // ラン内で最も根に近い演算子
		if lower != nil {
			setRecordInput(r.ops[0], lower) // ラン先頭の入力を下位ランのフラグメントへ繋ぎ直す
		}
		if !runSupported(r.store, r.ops) {
			return sub // capability を満たさないランがある → 融合しない
		}
		lower = &plan.StoreFragment{
			Store:      r.store,
			Plan:       top,
			Emits:      plan.EmitBindings,
			OutputSlot: recordOutputSlot(top),
		}
	}
	return lower
}

// recordOpStore は record 演算子のアクセス先ストアを返す（traversal は graph 固定）。
func recordOpStore(op plan.PlanNode) (store.Kind, bool) {
	switch o := op.(type) {
	case *plan.EntityScan:
		return o.DataStore, true
	case *plan.Filter:
		return o.DataStore, true
	case *plan.Expand, *plan.VarLengthExpand:
		return store.Graph, true
	}
	return store.Graph, false
}

// runSupported はラン内の全演算子を対象ストアが capability 上ネイティブ実行できるかを返す。
func runSupported(k store.Kind, ops []plan.PlanNode) bool {
	for _, op := range ops {
		var cap plan.OpCapability
		switch op.(type) {
		case *plan.EntityScan, *plan.Filter:
			cap = plan.CapFilter
		case *plan.Expand:
			cap = plan.CapExpand
		case *plan.VarLengthExpand:
			cap = plan.CapVarExpand
		default:
			return false
		}
		if !plan.Supports(k, cap) {
			return false
		}
	}
	return true
}

// setRecordInput は record 演算子の入力を差し替える（ラン境界の結線）。
func setRecordInput(op plan.PlanNode, in plan.PlanNode) {
	switch o := op.(type) {
	case *plan.Filter:
		o.Input = in
	case *plan.Expand:
		o.Input = in
	case *plan.VarLengthExpand:
		o.Input = in
	}
}

// recordOutputSlot は record 演算子の出力スロット表を返す（フラグメントの出力束縛）。
func recordOutputSlot(op plan.PlanNode) plan.SlotTable {
	switch o := op.(type) {
	case *plan.EntityScan:
		return o.OutputSlot
	case *plan.Filter:
		return o.OutputSlot
	case *plan.Expand:
		return o.OutputSlot
	case *plan.VarLengthExpand:
		return o.OutputSlot
	}
	return plan.SlotTable{}
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

// buildTailPushdown は tail pushdown が適用可能なら tail を包む StoreFragment を、不可なら nil を返す。
// 適用条件:
//   - record パイプライン（Projection.Input 以下）が all-graph（traversal は graph で完結）。
//   - tail に Aggregate が存在する（sort/limit のみは対象外）。
//   - tail が参照する全プロパティが単一の非 graph ストア S に解決できる。
//   - S が Aggregate/Project（＋必要なら GroupBy/Sort/Limit）を capability 上サポートする。
//   - staging するエンティティ（group/prop-agg の alias）の永続テーブル（Label）が判明する（plan.LowerTail）。
//
// 生成する形（＝tail pushdown 形状の唯一の判別子）:
//
//	StoreFragment{S, Emits:Result, Plan: Return→…→Projection→StoreFragment{graph, Emits:Bindings}}
//
// 平坦フィールドは持たず、実行側は plan.LowerTail(Plan) で必要情報を導出する。
// 未対応エンジン／未対応ストアは Plan をそのまま通常実行すればよい（結果は等価）。
func buildTailPushdown(root plan.PlanNode) *plan.StoreFragment {
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
	if findSort(root) != nil && !plan.Supports(s, plan.CapSort) {
		return nil
	}
	if findLimit(root) != nil && !plan.Supports(s, plan.CapLimit) {
		return nil
	}

	// record パイプラインを束縛フラグメントへ包み、tail の中に入れ子にする。
	proj.Input = &plan.StoreFragment{
		Store:       store.Graph,
		Plan:        proj.Input,
		Emits:       plan.EmitBindings,
		OutputSlot:  proj.InputSlot,
		OutputAlias: proj.InputAlias,
	}
	frag := &plan.StoreFragment{
		Store: s,
		Plan:  root,
		Emits: plan.EmitResult,
	}
	// 導出可能性（staging テーブル解決）を plan 時に確認しておく。不可なら融合しない。
	if _, ok := plan.LowerTail(frag.Plan); !ok {
		proj.Input = proj.Input.(*plan.StoreFragment).Plan // 包みを戻す
		return nil
	}
	return frag
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

// findAggregate / findSort / findLimit / findReturn は tail チェーン（根→第1子）を辿って各ノードを返す。
// findReturn は現在この経路では未使用（Return 項目の取得は plan.LowerTail が担う）。一般セグメンタで
// tail を畳む判定に使う想定のため helper 一式として残す。
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
