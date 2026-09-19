package planner

import (
	"fmt"
	"strconv"

	collections "polystore_database/src/go/collections"
	plan "polystore_database/src/go/plan"
	stk "polystore_database/src/go/store"
)

// multistage.go は WITH 多段クエリ（Phase B2）の structured プラン構築。
// settings.MultiStage=true のときだけ使われる（既定 false では ParseQuery が verbatim へ委譲）。
//
// 現段階（v1）で扱う形: WITH の projection 項目が **すべて bare 束縛**（例 `WITH f`）の
// ステージ連鎖（`MATCH … WITH f MATCH (f)-… RETURN …`）。ステージ境界は RowSource（Row→Record）で
// 連結する。WITH に集約・値列（`f.id AS x`）・post-WITH WHERE を含む形、および式列は未対応で、
// buildMultiStagePlan がエラーを返し、呼び出し側（ParseQuery）が graph verbatim へフォールバックする。
// これにより「対応済みの形だけ engine 多段・それ以外は従来どおり verbatim」を保証する。

// stageData は WITH で閉じた 1 ステージのスナップショット（instance の record 状態＋WITH 項目）。
type stageData struct {
	entityInfo        []*EntityInfo
	symbolEntTable    map[string]int
	relInfo           []*RelInfo
	symbolRelTable    map[string]int
	sorceEntMapping   map[string][]int
	targetEntMapping  map[string][]int
	symbolCondMapping map[string][]*plan.ConditionNode
	crossFilters      []*plan.ConditionNode
	pathFilters       map[string][]plan.VarLengthFilter

	carriedIn  map[string]bool   // 前段 WITH から RowSource でスロット供給される束縛
	carriedVal map[string]bool   // 前段 WITH から RowSource で Vals 供給される値/集約列
	items      []plan.ReturnItem // この WITH（または最終 RETURN）の projection 項目
	distinct   bool
	order      []plan.OrderItem
	limit      int
	postWhere  *plan.ConditionNode // post-WITH WHERE（Row 世界）
	isFinal    bool
}

// snapshotStage は現ステージ（instance の record 状態＋WITH 項目バッファ）を stages へ積み、
// instance の record 状態をリセットして次ステージを carried 束縛で種付けする（ExitWithClause から呼ぶ）。
func (l *QueryPlannerListener) snapshotStage() {
	st := &stageData{
		entityInfo:        l.entityInfo,
		symbolEntTable:    l.symbolEntTable,
		relInfo:           l.relInfo,
		symbolRelTable:    l.symbolRelTable,
		sorceEntMapping:   l.sorceEntMapping,
		targetEntMapping:  l.targetEntMapping,
		symbolCondMapping: l.symbolCondMapping,
		crossFilters:      l.crossFilters,
		pathFilters:       l.pathFiltersByRelAlias,
		carriedIn:         l.carriedBinds,
		carriedVal:        l.carriedVals,
		items:             l.withItemsBuf,
		distinct:          l.withDistinct,
		order:             l.withOrderBuf,
		limit:             l.withLimitBuf,
		postWhere:         l.curPostWhere,
	}
	l.stages = append(l.stages, st)

	// 次ステージへ持ち越す: bare 束縛→スロット、値/集約→Vals（いずれも clean 出力名で参照）。
	nextBinds := map[string]bool{}
	nextVals := map[string]bool{}
	for _, it := range st.items {
		if isBareBinding(it) {
			nextBinds[it.Alias] = true
		} else {
			nextVals[it.Name] = true
		}
	}
	l.resetStageState(nextBinds, nextVals)
}

// resetStageState は instance の record 状態を初期化し、carried 束縛を pre-bound エンティティとして
// 種付けする（次の MATCH が (f) を参照したら再利用され、起点は RowSource になる）。
func (l *QueryPlannerListener) resetStageState(carried, carriedVals map[string]bool) {
	l.entityInfo = nil
	l.relInfo = nil
	l.symbolEntTable = map[string]int{}
	l.symbolRelTable = map[string]int{}
	l.sorceEntMapping = map[string][]int{}
	l.targetEntMapping = map[string][]int{}
	l.symbolCondMapping = map[string][]*plan.ConditionNode{}
	l.crossFilters = nil
	l.pathFiltersByRelAlias = map[string][]plan.VarLengthFilter{}
	l.varLenRelByPathVar = map[string]string{}
	l.lastNodeAlias = ""
	l.withItemsBuf = nil
	l.withOrderBuf = nil
	l.withLimitBuf = 0
	l.withDistinct = false
	l.curPostWhere = nil
	l.carriedBinds = carried
	l.carriedVals = carriedVals

	// carried 束縛を先頭に登録（getStartEntity が起点＝carried を返すよう）。carried 値は
	// スカラなのでエンティティ登録しない（RowSource が Vals で供給し、Projection が MergeVals する）。
	for a := range carried {
		l.entityInfo = append(l.entityInfo, &EntityInfo{alias: a})
		l.symbolEntTable[a] = len(l.entityInfo) - 1
	}
}

// isBareBinding は WITH 項目が bare 束縛（`WITH f`）かを返す（集約/coalesce/プロパティ/式ではない）。
func isBareBinding(it plan.ReturnItem) bool {
	if it.IsAggregate || it.IsCoalesce || it.Alias == "" {
		return false
	}
	return len(it.Props) == 0 || (len(it.Props) == 1 && it.Props[0] == "")
}

// BuildMultiStagePlan は WITH 多段の structured プランを構築する。未対応の形は error を返し、
// 呼び出し側が graph verbatim へフォールバックする。対応する形:
//   - Shape B（集約/射影）: 単一 WITH（束縛/値/集約）＋ post-WITH WHERE ＋ 最終 RETURN（後続 MATCH なし）。
//     record→Projection→[Aggregate]→RowProject→[RowFilter]→[Sort]→[Limit]→Return。
//   - Shape A（連鎖）: bare 束縛の WITH ＋ 後続 MATCH のステージ連鎖（RowSource で連結）。
func (l *QueryPlannerListener) BuildMultiStagePlan() (plan.PlanNode, error) {
	// 最終ステージ（instance fields）の項目を loadStage 前に退避。
	finalItems := l.returnItems
	finalOrder := l.orderItems
	finalLimit := l.limitNum

	// WITH 計算列（CASE/算術）は RowProject が評価して下流へ流せる（passthrough / RowFilter / 最終 RETURN）。
	// ただし次の 2 形は Aggregate 演算子が Row 上の式/スカラ値列を扱えないため未対応 → verbatim:
	//   (a) 同一ステージで集約と計算式を併用（グループ化後の per-row 式は曖昧）
	//   (b) 前段 WITH の値/式列を後段の集約引数にする（scalar 値列の集約）
	valNames := map[string]bool{} // WITH が出力する値/式列名（bare 束縛・集約以外）
	for _, st := range l.stages {
		hasAgg, hasExpr := false, false
		for _, it := range st.items {
			if it.IsAggregate {
				hasAgg = true
			}
			if it.Expr != nil {
				hasExpr = true
			}
			if it.Expr != nil || (!isBareBinding(it) && !it.IsAggregate) {
				valNames[it.Name] = true
			}
		}
		if hasAgg && hasExpr {
			return nil, fmt.Errorf("multi-stage: aggregate mixed with computed expression not supported")
		}
	}
	checkAggs := func(items []plan.ReturnItem) error {
		for _, it := range items {
			if it.IsAggregate && it.Agg != nil && it.Agg.Prop == "" && valNames[it.Agg.Alias] {
				return fmt.Errorf("multi-stage: aggregate over computed value column not supported")
			}
		}
		return nil
	}
	for _, st := range l.stages {
		if err := checkAggs(st.items); err != nil {
			return nil, err
		}
	}
	if err := checkAggs(finalItems); err != nil {
		return nil, err
	}

	// Shape B: 単一 WITH（集約/値を含む）＋ 最終に新規 MATCH 無し ＋ 最終 RETURN が WITH 出力名のみ参照。
	// （最終が carried 束縛のプロパティを再材料化する形＝Shape B 不可。集約後の再材料化は未対応→verbatim）。
	if len(l.stages) == 1 && l.finalHasNoMatch() {
		st := l.stages[0]
		// Shape B は「集約/値を含む WITH ＋ 最終が WITH 出力名のみ参照」のとき RowSource を省ける効率形。
		// それ以外（bare 束縛のみ・最終が carried 束縛を再材料化）は一般チェーンに任せる。
		if withHasAggOrValue(st.items) && shapeBSafe(st.items, finalItems, finalOrder) {
			return l.buildAggProjectStage(st, finalItems, finalOrder, finalLimit)
		}
	}

	// 一般チェーン（RowSource で連結・値/集約持ち越し・再材料化に対応）。
	return l.buildBindingChain(finalItems, finalOrder, finalLimit)
}

// finalHasNoMatch は最終ステージ（instance fields）が新規 MATCH を持たない（carried 束縛のみ・rel 無し）かを返す。
func (l *QueryPlannerListener) finalHasNoMatch() bool {
	if len(l.relInfo) > 0 {
		return false
	}
	for _, e := range l.entityInfo {
		if !l.carriedBinds[e.alias] {
			return false
		}
	}
	return true
}

// buildAggProjectStage は Shape B を構築する。
func (l *QueryPlannerListener) buildAggProjectStage(st *stageData, finalItems []plan.ReturnItem, finalOrder []plan.OrderItem, finalLimit int) (plan.PlanNode, error) {
	recRoot, _, err := l.buildStageRecord(st, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	l.loadStage(st)
	top, err := l.buildProjAgg(recRoot, st.items, st.order)
	if err != nil {
		return nil, err
	}
	// WITH 出力を clean 名へ揃える。
	top = &plan.RowProject{Input: top, Cols: rowProjColsFor(st.items)}
	// post-WITH WHERE → RowFilter。
	if st.postWhere != nil {
		e, err := condToExpr(st.postWhere)
		if err != nil {
			return nil, err // 未対応述語 → verbatim
		}
		top = &plan.RowFilter{Input: top, Cond: e}
	}
	// 最終 Sort/Limit/Return（clean 名を読む）。
	if len(finalOrder) > 0 {
		top = &plan.Sort{Input: top, OrderItems: finalOrder}
	}
	if finalLimit > 0 {
		top = &plan.Limit{Input: top, Count: finalLimit}
	}
	return &plan.Return{Input: top, Items: finalItems}, nil
}

// buildBindingChain は一般のステージ連鎖を構築する（bare 束縛・値列・集約を含む WITH に対応）。
// 各中間ステージ: record→Projection→[Aggregate]→RowProject（clean 名）→[RowFilter(post-WITH WHERE)]。
// ステージ境界は RowSource（bare 束縛→スロット / 値・集約→Vals）で連結し、後段 Projection が MergeVals する。
// 最終ステージ: record（carried あれば RowSource 起点）→Projection（新規材料化＋carried 値 passthrough）→[Sort][Limit]→Return。
func (l *QueryPlannerListener) buildBindingChain(finalItems []plan.ReturnItem, finalOrder []plan.OrderItem, finalLimit int) (plan.PlanNode, error) {
	final := &stageData{
		entityInfo: l.entityInfo, symbolEntTable: l.symbolEntTable,
		relInfo: l.relInfo, symbolRelTable: l.symbolRelTable,
		sorceEntMapping: l.sorceEntMapping, targetEntMapping: l.targetEntMapping,
		symbolCondMapping: l.symbolCondMapping, crossFilters: l.crossFilters,
		pathFilters: l.pathFiltersByRelAlias, carriedIn: l.carriedBinds, carriedVal: l.carriedVals,
		items: finalItems, order: finalOrder, limit: finalLimit, isFinal: true,
	}
	allStages := append(append([]*stageData{}, l.stages...), final)

	var prevRows plan.PlanNode
	var prevBinds, prevVals []string

	for _, st := range allStages {
		recRoot, _, err := l.buildStageRecord(st, prevRows, prevBinds, prevVals)
		if err != nil {
			return nil, err
		}
		l.loadStage(st)
		top, err := l.buildProjAgg(recRoot, st.items, st.order)
		if err != nil {
			return nil, err
		}

		if st.isFinal {
			if len(st.order) > 0 {
				top = &plan.Sort{Input: top, OrderItems: st.order}
			}
			if st.limit > 0 {
				top = &plan.Limit{Input: top, Count: st.limit}
			}
			return &plan.Return{Input: top, Items: st.items}, nil
		}

		// 中間ステージ: clean 名へ揃え、post-WITH WHERE を RowFilter で適用、次段へ Rows を渡す。
		top = &plan.RowProject{Input: top, Cols: rowProjColsFor(st.items)}
		if st.postWhere != nil {
			e, err := condToExpr(st.postWhere)
			if err != nil {
				return nil, err
			}
			top = &plan.RowFilter{Input: top, Cond: e}
		}
		prevRows = top
		prevBinds, prevVals = prevBinds[:0], prevVals[:0]
		for _, it := range st.items {
			if isBareBinding(it) {
				prevBinds = append(prevBinds, it.Name)
			} else {
				prevVals = append(prevVals, it.Name)
			}
		}
	}
	return nil, fmt.Errorf("multi-stage: no final RETURN stage")
}

// withHasAggOrValue は WITH 項目に集約または値列（プロパティ/coalesce）が含まれるかを返す。
func withHasAggOrValue(items []plan.ReturnItem) bool {
	for _, it := range items {
		if !isBareBinding(it) {
			return true
		}
	}
	return false
}

// shapeBSafe は Shape B（RowSource 無し）で最終 RETURN/ORDER が WITH 出力名だけを参照するかを返す。
// WITH 出力名 = 各 WITH 項目の Name（RowProject が生成する clean 名）。最終集約・非 WITH 出力参照は不可。
func shapeBSafe(withItems, finalItems []plan.ReturnItem, finalOrder []plan.OrderItem) bool {
	out := map[string]bool{}
	for _, it := range withItems {
		out[it.Name] = true
	}
	readKeyOK := func(alias, prop string) bool {
		key := alias
		if prop != "" {
			key = alias + "." + prop
		}
		return out[key]
	}
	for _, it := range finalItems {
		if it.IsAggregate {
			return false // 集約後の再集約は未対応
		}
		if it.IsCoalesce {
			for _, p := range it.Props {
				if !readKeyOK(it.Alias, p) {
					return false
				}
			}
			continue
		}
		p := ""
		if len(it.Props) > 0 {
			p = it.Props[0]
		}
		if !readKeyOK(it.Alias, p) {
			return false
		}
	}
	for _, oi := range finalOrder {
		key := oi.Key
		if key == "" {
			key = oi.Alias
			if oi.Prop != "" {
				key = oi.Alias + "." + oi.Prop
			}
		}
		if !out[key] {
			return false
		}
	}
	return true
}

// rowProjColsFor は WITH 項目を clean 名へ写す RowProject 列を作る。
//   - bare 束縛 f: Out=f, Src=f（uuid）
//   - 値 f.id AS fid: Out=fid, Src=f.id
//   - 集約 count(*) AS c: Out=c, Src=c（OutName）
func rowProjColsFor(items []plan.ReturnItem) []plan.RowProjCol {
	cols := make([]plan.RowProjCol, 0, len(items))
	for _, it := range items {
		switch {
		case it.Expr != nil:
			// 計算列（CASE/算術など）: Projection が材料化した参照列を使って RowProject で評価する。
			cols = append(cols, plan.RowProjCol{Out: it.Name, Expr: it.Expr})
		case it.IsAggregate:
			cols = append(cols, plan.RowProjCol{Out: it.Name, Src: it.Agg.OutName})
		case isBareBinding(it):
			cols = append(cols, plan.RowProjCol{Out: it.Name, Src: it.Alias})
		default: // 値（プロパティ／coalesce）
			src := it.Alias
			if len(it.Props) > 0 && it.Props[0] != "" {
				src = it.Alias + "." + it.Props[0]
			}
			cols = append(cols, plan.RowProjCol{Out: it.Name, Src: src})
		}
	}
	return cols
}

// loadStage は stageData を instance フィールドへ載せ替える（buildOperatorTree 等の既存ヘルパ再利用のため）。
func (l *QueryPlannerListener) loadStage(st *stageData) {
	l.entityInfo = st.entityInfo
	l.symbolEntTable = st.symbolEntTable
	l.relInfo = st.relInfo
	l.symbolRelTable = st.symbolRelTable
	l.sorceEntMapping = st.sorceEntMapping
	l.targetEntMapping = st.targetEntMapping
	l.symbolCondMapping = st.symbolCondMapping
	l.crossFilters = st.crossFilters
	l.pathFiltersByRelAlias = st.pathFilters
	l.returnItems = st.items
	l.orderItems = st.order
	l.limitNum = st.limit
}

// buildStageRecord は 1 ステージの record パイプライン（EntityScan/RowSource → Expand*/Filter* → PostFilter）
// を構築し、出力に現れる alias 群を返す。carried 束縛があれば RowSource を起点にする。
func (l *QueryPlannerListener) buildStageRecord(st *stageData, prevRows plan.PlanNode, prevBinds, prevVals []string) (plan.PlanNode, []string, error) {
	l.loadStage(st)

	// 条件メタデータ（DataStore/DataType）を埋める。
	for alias := range st.symbolCondMapping {
		for _, cond := range st.symbolCondMapping[alias] {
			l.fillConditionMetadata(cond)
		}
	}

	startEnt := l.getStartEntity()
	if startEnt == nil {
		return nil, nil, fmt.Errorf("multi-stage: empty stage")
	}

	visitedEnt := make(map[string]bool)
	visitedRel := make(map[string]bool)

	var root plan.PlanNode
	if len(st.carriedIn) > 0 {
		// carried ステージ: 起点は RowSource（前段 Row の束縛を Record スロットへ）。
		// 出力スロットは carried 束縛（この段で使う分）を割り当てる。
		outSlot := slotTableOf(carriedSlice(st.carriedIn))
		root = &plan.RowSource{
			Input:       prevRows,
			BindCols:    append([]string{}, prevBinds...),
			ValCols:     append([]string{}, prevVals...), // 値/集約列を Record.Vals へ持ち越す
			OutputSlot:  outSlot,
			OutputAlias: carriedSlice(st.carriedIn),
		}
		for a := range st.carriedIn {
			visitedEnt[a] = true
		}
	} else {
		// 通常ステージ: 起点は EntityScan（既存単段と同じ配置ロジック）。
		startFilters := st.symbolCondMapping[startEnt.alias]
		storesMap := map[stk.Kind][]*plan.ConditionNode{}
		for _, cond := range startFilters {
			storesMap[cond.DataStore] = append(storesMap[cond.DataStore], cond)
		}
		best := stk.Graph
		maxc := -1
		for sk, cs := range storesMap {
			if len(cs) > maxc {
				maxc, best = len(cs), sk
			}
		}
		root = &plan.EntityScan{Alias: startEnt.alias, Labels: startEnt.labels, DataStore: best, Filter: storesMap[best]}
		for sk, cs := range storesMap {
			if sk == best {
				continue
			}
			root = &plan.Filter{Input: root, Filter: cs, Alias: startEnt.alias, ObjType: plan.Entity, DataStore: sk}
		}
		visitedEnt[startEnt.alias] = true
	}

	root = l.buildOperatorTree(startEnt.alias, root, visitedEnt, visitedRel)
	for _, cf := range st.crossFilters {
		root = &plan.PostFilter{Input: root, Cond: cf}
	}

	// 出力 alias（この段で束縛される全 alias）。
	out := make([]string, 0, len(st.entityInfo))
	for _, e := range st.entityInfo {
		out = append(out, e.alias)
	}
	return root, out, nil
}

// bindingProjection は bare 束縛のみを持つ中間 WITH の Projection（fetch なし＝束縛 uuid のみ）を作り、
// 次段 RowSource の Row 入力にする。
func (l *QueryPlannerListener) bindingProjection(recRoot plan.PlanNode, items []plan.ReturnItem, outAliases []string) *plan.Projection {
	var units []plan.ProjectionUnit
	for _, it := range items {
		units = append(units, plan.ProjectionUnit{Alias: it.Alias, ObjType: plan.Entity})
	}
	proj := &plan.Projection{Input: recRoot, Units: units}
	// record 側スロットを確定（recRoot→proj の連鎖）。
	refineRecordChain(recRoot, proj, aliasesOf(items))
	return proj
}

// buildProjAgg は record 出力に対する Projection（材料化）＋（集約があれば）Aggregate を組み、
// その top ノード（Projection または Aggregate）を返す。Sort/Limit/Return は呼び出し側で付ける。
// items は WITH/RETURN の projection 項目、order は addColumn 用の並べ替え参照。
// instance フィールドは loadStage 済み前提。
func (l *QueryPlannerListener) buildProjAgg(recRoot plan.PlanNode, items []plan.ReturnItem, order []plan.OrderItem) (plan.PlanNode, error) {
	// 既存 BuildPlan の tail 構築を踏襲（ProjectionUnit 収集）。
	unitMap := map[string]*plan.ProjectionUnit{}
	ensureUnit := func(alias string) *plan.ProjectionUnit {
		if u, ok := unitMap[alias]; ok {
			return u
		}
		var labels []string
		var ot plan.ObjectType
		if idx, ok := l.symbolEntTable[alias]; ok {
			labels, ot = l.entityInfo[idx].labels, plan.Entity
		} else if idx, ok := l.symbolRelTable[alias]; ok {
			labels, ot = []string{l.relInfo[idx].label}, plan.Relationship
		}
		u := &plan.ProjectionUnit{Alias: alias, ObjType: ot, Labels: labels, Fetches: []plan.FetchPlan{}}
		unitMap[alias] = u
		return u
	}
	addColumn := func(alias, prop string) {
		if alias == "" || prop == "" {
			return
		}
		u := ensureUnit(alias)
		lbl := ""
		if len(u.Labels) > 0 {
			lbl = u.Labels[0]
		}
		store, dt, _ := l.mappingDictionary.LookupMappingDictionary(u.ObjType, lbl, prop)
		sk := kindOf(store)
		for i := range u.Fetches {
			if u.Fetches[i].Store == sk {
				for _, p := range u.Fetches[i].Props {
					if p == prop {
						return
					}
				}
				u.Fetches[i].Props = append(u.Fetches[i].Props, prop)
				u.Fetches[i].TypeMap[prop] = dt
				return
			}
		}
		u.Fetches = append(u.Fetches, plan.FetchPlan{Store: sk, Props: []string{prop}, TypeMap: map[string]string{prop: dt}})
	}

	var groupKeys []plan.GroupKey
	var aggs []plan.AggregateItem
	hasAgg := false
	for _, item := range items {
		if item.IsAggregate {
			hasAgg = true
			aggs = append(aggs, *item.Agg)
			addColumn(item.Agg.Alias, item.Agg.Prop)
			continue
		}
		if item.Expr != nil {
			// 計算式（CASE 等）: 参照プロパティを材料化。group key にはしない（Return/RowProject が評価）。
			for _, key := range exprColumns(item.Expr) {
				a, p := splitKey(key)
				addColumn(a, p)
			}
			continue
		}
		// carried 値（前段 WITH 由来の値/集約列）は record 材料化せず passthrough（Vals→MergeVals 済）。
		// carried 束縛はスロットにあるので通常どおり group key/unit に含める（Row へはスロットから入る）。
		if isBareBinding(item) && l.carriedVals[item.Alias] {
			continue
		}
		for _, p := range item.Props {
			addColumn(item.Alias, p)
		}
		fp := ""
		if len(item.Props) > 0 {
			fp = item.Props[0]
		}
		groupKeys = append(groupKeys, plan.GroupKey{Alias: item.Alias, Prop: fp, OutName: item.Name})
	}
	for _, oi := range order {
		addColumn(oi.Alias, oi.Prop)
	}

	units := make([]plan.ProjectionUnit, 0, len(unitMap))
	for _, u := range unitMap {
		units = append(units, *u)
	}

	// InputSlot は record 束縛のみ。carried 値（Vals 由来）はスロットではないので除外する
	//（含めると Row 生成時にスロット由来の空 uuid で MergeVals 済みの値を上書きしてしまう）。
	needed := requiredAliases(items, order)
	if len(l.carriedVals) > 0 {
		filtered := needed[:0]
		for _, a := range needed {
			if !l.carriedVals[a] {
				filtered = append(filtered, a)
			}
		}
		needed = filtered
	}
	proj := &plan.Projection{Input: recRoot, Units: units}
	refineRecordChain(recRoot, proj, needed)

	var top plan.PlanNode = proj
	if hasAgg {
		top = &plan.Aggregate{Input: proj, GroupKeys: groupKeys, Aggs: aggs}
	}
	return top, nil
}

// ---- スロット確定（ステージ内の record 連鎖）----

// refineRecordChain は proj.Input（record 連鎖）に対し、RefinePlan と同じ規約で
// InputSlot/OutputSlot を割り当てる。RowSource/EntityScan を source leaf として停止する。
func refineRecordChain(recRoot plan.PlanNode, proj *plan.Projection, needed []string) {
	requiredNext := make(collections.Set)
	for _, a := range needed {
		requiredNext.Insert(a)
	}
	var nextInputSlot plan.SlotTable
	proj.InputAlias = requiredNext.ConvertSlice()
	proj.InputSlot = buildSlots(proj.InputAlias)
	nextInputSlot = proj.InputSlot

	for op := plan.PlanNode(recRoot); op != nil; {
		switch o := op.(type) {
		case *plan.Filter:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = nextInputSlot
			for _, c := range o.Filter {
				if c.Alias != "" {
					requiredNext.Insert(c.Alias)
				}
			}
			o.InputAlias = requiredNext.ConvertSlice()
			o.InputSlot = buildSlots(o.InputAlias)
			nextInputSlot = o.InputSlot
		case *plan.PostFilter:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = nextInputSlot
			if o.Cond != nil {
				if o.Cond.Alias != "" {
					requiredNext.Insert(o.Cond.Alias)
				}
				if o.Cond.RAlias != "" {
					requiredNext.Insert(o.Cond.RAlias)
				}
			}
			o.InputAlias = requiredNext.ConvertSlice()
			o.InputSlot = buildSlots(o.InputAlias)
			nextInputSlot = o.InputSlot
		case *plan.Expand:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = nextInputSlot
			requiredNext.Remove(o.Alias)
			requiredNext.Remove(o.TargetEntity)
			requiredNext.Insert(o.SourceEntity)
			o.InputAlias = requiredNext.ConvertSlice()
			o.InputSlot = buildSlots(o.InputAlias)
			nextInputSlot = o.InputSlot
		case *plan.VarLengthExpand:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = nextInputSlot
			requiredNext.Remove(o.Alias)
			requiredNext.Remove(o.TargetEntity)
			requiredNext.Insert(o.SourceEntity)
			o.InputAlias = requiredNext.ConvertSlice()
			o.InputSlot = buildSlots(o.InputAlias)
			nextInputSlot = o.InputSlot
		case *plan.EntityScan:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = buildSlots(o.OutputAlias)
			return // source leaf
		case *plan.RowSource:
			o.OutputAlias = requiredNext.ConvertSlice()
			o.OutputSlot = buildSlots(o.OutputAlias)
			return // source leaf（前段 Row 世界へは伝播しない）
		}
		ch := op.Children()
		if len(ch) == 0 {
			break
		}
		op = ch[0]
	}
}

func buildSlots(aliases []string) plan.SlotTable {
	st := plan.SlotTable{VarToSlot: map[string]int{}, SlotToVar: make([]string, len(aliases))}
	for i, a := range aliases {
		st.VarToSlot[a] = i
		st.SlotToVar[i] = a
	}
	return st
}

func slotTableOf(aliases []string) plan.SlotTable { return buildSlots(aliases) }

func carriedSlice(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for a := range m {
		out = append(out, a)
	}
	return out
}

func aliasesOf(items []plan.ReturnItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		if it.Alias != "" {
			out = append(out, it.Alias)
		}
	}
	return out
}

// ---- post-WITH WHERE を plan.Expr へ変換（RowFilter 用）----

// condToExpr は ConditionNode を Row 世界の Expr へ変換する。post-WITH WHERE は WITH 出力の
// clean 名（集約 OutName・値の AS 名・束縛 alias）を参照するため、列キーは alias[.prop] とする。
// 未対応（IS NULL / 変数間比較 / 量化子など）は error を返し、呼び出し側が verbatim へフォールバックする。
func condToExpr(c *plan.ConditionNode) (plan.Expr, error) {
	if c == nil {
		return nil, fmt.Errorf("condToExpr: nil condition")
	}
	switch c.Type {
	case plan.CondAnd, plan.CondOr:
		le, err := condToExpr(c.Left)
		if err != nil {
			return nil, err
		}
		re, err := condToExpr(c.Right)
		if err != nil {
			return nil, err
		}
		op := plan.OpAnd
		if c.Type == plan.CondOr {
			op = plan.OpOr
		}
		return &plan.Bin{Op: op, L: le, R: re}, nil
	case plan.CondNot:
		x, err := condToExpr(c.Child)
		if err != nil {
			return nil, err
		}
		return &plan.Un{Op: plan.OpNot, X: x}, nil
	case plan.CondParen:
		return condToExpr(c.Child)
	case plan.CondEq, plan.CondNeq, plan.CondGreater, plan.CondLess, plan.CondGreaterEq, plan.CondLessEq:
		op, ok := binCmpOf(c.Type)
		if !ok {
			return nil, fmt.Errorf("condToExpr: bad comparison")
		}
		return &plan.Bin{Op: op, L: colExpr(c.Alias, c.Property), R: litOf(c.Value)}, nil
	case plan.CondIn:
		elems := make([]plan.Expr, len(c.Values))
		for i, v := range c.Values {
			elems[i] = litOf(v)
		}
		return &plan.Bin{Op: plan.OpIn, L: colExpr(c.Alias, c.Property), R: &plan.List{Elems: elems}}, nil
	case plan.CondIsNull:
		return &plan.Un{Op: plan.OpIsNull, X: colExpr(c.Alias, c.Property)}, nil
	case plan.CondIsNotNull:
		return &plan.Un{Op: plan.OpIsNotNull, X: colExpr(c.Alias, c.Property)}, nil
	case plan.CondVarCmp:
		// 変数間比較 a.x OP b.y。両辺とも列参照（"alias.prop" / 束縛 "alias"）。
		op, ok := binCmpOf(c.CmpOp)
		if !ok {
			return nil, fmt.Errorf("condToExpr: bad var comparison op")
		}
		return &plan.Bin{Op: op, L: colExpr(c.Alias, c.Property), R: colExpr(c.RAlias, c.RProperty)}, nil
	default:
		return nil, fmt.Errorf("condToExpr: unsupported condition type %v (post-WITH WHERE)", c.Type)
	}
}

// exprColumns は Expr が参照する列キー（Column.Key）を集める（Projection の事前材料化用）。
func exprColumns(e plan.Expr) []string {
	var out []string
	var walk func(plan.Expr)
	walk = func(e plan.Expr) {
		switch x := e.(type) {
		case *plan.Column:
			out = append(out, x.Key)
		case *plan.Bin:
			walk(x.L)
			walk(x.R)
		case *plan.Un:
			walk(x.X)
		case *plan.Call:
			for _, a := range x.Args {
				walk(a)
			}
		case *plan.Case:
			if x.Subject != nil {
				walk(x.Subject)
			}
			for _, w := range x.Whens {
				walk(w.When)
				walk(w.Then)
			}
			if x.Else != nil {
				walk(x.Else)
			}
		case *plan.List:
			for _, el := range x.Elems {
				walk(el)
			}
		}
	}
	walk(e)
	return out
}

// splitKey は列キー "alias.prop" を (alias, prop) に分ける（ドット無しは prop 空＝束縛参照）。
func splitKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

func colExpr(alias, prop string) plan.Expr {
	key := alias
	if prop != "" {
		key = alias + "." + prop
	}
	return &plan.Column{Key: key}
}

// litOf は文字列リテラルを型推論して Lit にする（int→int64, float→float64, その他 string）。
func litOf(s string) plan.Expr {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &plan.Lit{Val: i}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return &plan.Lit{Val: f}
	}
	return &plan.Lit{Val: s}
}

func binCmpOf(t plan.ConditionType) (plan.BinOp, bool) {
	switch t {
	case plan.CondEq:
		return plan.OpEq, true
	case plan.CondNeq:
		return plan.OpNe, true
	case plan.CondGreater:
		return plan.OpGt, true
	case plan.CondLess:
		return plan.OpLt, true
	case plan.CondGreaterEq:
		return plan.OpGe, true
	case plan.CondLessEq:
		return plan.OpLe, true
	}
	return 0, false
}

func requiredAliases(items []plan.ReturnItem, order []plan.OrderItem) []string {
	seen := map[string]bool{}
	var out []string
	add := func(a string) {
		if a != "" && !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	for _, it := range items {
		add(it.Alias)
	}
	for _, oi := range order {
		add(oi.Alias)
	}
	return out
}
