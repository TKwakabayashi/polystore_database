package logical_plan

import (
	"polystore_database/src/go/plan"
)

// PushdownMode は集約 pushdown の方針。ソース上で SelectedPushdown を書き換えて切替える。
type PushdownMode int

const (
	PushdownAuto        PushdownMode = iota // 単一ストアに解決できれば委譲、散在ならエンジン
	PushdownForceEngine                     // 常にコーディネータ（自作エンジン）
)

// ★ ここを書き換えて pushdown 方針を切替える（実験用）。
// ベンチマーク等から実行時に上書きも可能（var）。
var SelectedPushdown = PushdownAuto

// MaybePushdown は、集約クエリの参照プロパティが単一ストアに解決できる場合に
// そのストアへ委譲する StorePushdown を返す。できなければ既存のコーディネータ木を返す。
//   - 判定はマッピング依存なので、migrate でプロパティ配置を変えると自動でフォールバックへ切替わる。
//   - 集約が無いクエリは対象外（baseline 比較のため既存挙動を維持）。
func (l *QueryPlannerListener) MaybePushdown(query string, params map[string]string) plan.PlanNode {
	if SelectedPushdown == PushdownForceEngine {
		return l.planRoot
	}

	hasAgg := false
	for _, it := range l.returnItems {
		if it.IsAggregate {
			hasAgg = true
			break
		}
	}
	if !hasAgg {
		return l.planRoot
	}

	refStores := l.collectRefStores()
	hasTraversal := len(l.relInfo) > 0

	if hasTraversal {
		// traversal は graph 専用。全参照が graph なら Cypher を丸ごと委譲。
		if storesOnly(refStores, "graph") {
			return &plan.StorePushdown{Store: "graph", Query: query, Params: params, Items: l.returnItems}
		}
		return l.planRoot
	}

	// traversal 無し（単一スキャン）。参照プロパティが解決する単一ストアへ委譲を試みる。
	target, ok := singleStore(refStores)
	if !ok {
		return l.planRoot // 複数ストアに散在
	}
	if target == "" {
		// 参照プロパティ無し（count(*) のみ）。スキャン先のストアを対象にする。
		target = l.baseScanStore()
	}
	if target == "graph" {
		return &plan.StorePushdown{Store: "graph", Query: query, Params: params, Items: l.returnItems}
	}
	// 非graph 単一ストア: 生成器が対応可能なら委譲、不可なら安全にフォールバック。
	if op := l.buildNonGraphPushdown(target); op != nil {
		return op
	}
	return l.planRoot
}

// singleStore は集合が単一ストアか（空集合＝count(*)のみは ("",true)）を返す。
func singleStore(s map[string]bool) (string, bool) {
	switch len(s) {
	case 0:
		return "", true
	case 1:
		for k := range s {
			return k, true
		}
	}
	return "", false
}

// collectRefStores は WHERE/inline フィルタ・group キー・集約引数・order キーが
// 参照するストアの集合を返す（解決不能・未設定は "graph" 扱い）。
func (l *QueryPlannerListener) collectRefStores() map[string]bool {
	stores := map[string]bool{}
	add := func(s string) {
		if s == "" || s == "unknown" {
			s = "graph"
		}
		stores[s] = true
	}

	for _, conds := range l.symbolCondMapping {
		for _, c := range conds {
			add(c.DataStore)
		}
	}
	for _, it := range l.returnItems {
		if it.IsAggregate {
			if it.Agg != nil && it.Agg.Prop != "" {
				add(l.propStore(it.Agg.Alias, it.Agg.Prop))
			}
			continue
		}
		for _, p := range it.Props {
			if p != "" {
				add(l.propStore(it.Alias, p))
			}
		}
	}
	for _, oi := range l.orderItems {
		if oi.Alias != "" && oi.Prop != "" {
			add(l.propStore(oi.Alias, oi.Prop))
		}
	}
	return stores
}

// propStore は alias.prop の物理ストアをマッピング辞書から引く（不明時は "graph"）。
func (l *QueryPlannerListener) propStore(alias, prop string) string {
	var objType plan.ObjectType
	var label string
	if idx, ok := l.symbolEntTable[alias]; ok {
		objType = plan.Entity
		if len(l.entityInfo[idx].labels) > 0 {
			label = l.entityInfo[idx].labels[0]
		}
	} else if idx, ok := l.symbolRelTable[alias]; ok {
		objType = plan.Relationship
		label = l.relInfo[idx].label
	}
	store, _, err := l.mappingDictionary.LookupMappingDictionary(objType, label, prop)
	if err != nil || store == "" {
		return "graph"
	}
	return store
}

// baseScanStore は planRoot を下って EntityScan のストアを返す。
func (l *QueryPlannerListener) baseScanStore() string {
	op := l.planRoot
	for op != nil {
		if es, ok := op.(*plan.EntityScan); ok {
			if es.DataStore == "" {
				return "graph"
			}
			return es.DataStore
		}
		ch := op.Children()
		if len(ch) == 0 {
			break
		}
		op = ch[0]
	}
	return "graph"
}

// storesOnly は集合が target のみ（空集合含む）かを返す。
func storesOnly(s map[string]bool, target string) bool {
	for k := range s {
		if k != target {
			return false
		}
	}
	return true
}

// buildNonGraphPushdown は非graph 単一ストア（traversal 無しの単一エンティティ）向けに
// ネイティブ集約委譲用の StorePushdown を構築する。ストアが対応できない形なら nil（フォールバック）。
//   - relational(MySQL) / document(Mongo): GROUP BY / ORDER BY / LIMIT / DISTINCT に対応。
//   - columnar(Cassandra): CQL 制約により全体集約のみ（GROUP BY / ORDER BY / DISTINCT は不可）。
//   - kvs(LevelDB): ネイティブ集約なし → 常に nil。
func (l *QueryPlannerListener) buildNonGraphPushdown(store string) plan.PlanNode {
	if store == "kvs" {
		return nil
	}
	// 単一エンティティであること（カルテシアン積などは対象外）。
	if len(l.entityInfo) != 1 {
		return nil
	}
	ent := l.entityInfo[0]
	label := ""
	if len(ent.labels) > 0 {
		label = ent.labels[0]
	}
	if label == "" {
		return nil
	}

	var groupKeys []plan.GroupKey
	var aggs []plan.AggregateItem
	hasDistinct := false
	for _, it := range l.returnItems {
		if it.IsAggregate {
			aggs = append(aggs, *it.Agg)
			if it.Agg.Distinct {
				hasDistinct = true
			}
			continue
		}
		prop := ""
		if len(it.Props) > 0 {
			prop = it.Props[0]
		}
		groupKeys = append(groupKeys, plan.GroupKey{Alias: it.Alias, Prop: prop, OutName: it.Name})
	}

	// Cassandra は GROUP BY / ORDER BY / DISTINCT を汎用にサポートしないため全体集約のみ許可。
	if store == "columnar" {
		if len(groupKeys) > 0 || len(l.orderItems) > 0 || hasDistinct {
			return nil
		}
	}

	return &plan.StorePushdown{
		Store:      store,
		Table:      label,
		Filters:    l.symbolCondMapping[ent.alias],
		GroupKeys:  groupKeys,
		Aggs:       aggs,
		OrderItems: l.orderItems,
		Limit:      l.limitNum,
		Items:      l.returnItems,
	}
}
