package plan

import "sort"

// lowering.go は「委譲する論理サブプランから、ネイティブクエリ生成に必要な情報を導出する」
// 純粋なプラン木解析。平坦フィールドを物理演算子に持たせる代わりに、意味の源泉である
// サブプラン（StoreFragment.Plan）を走査して都度導出する。
//
// ネイティブクエリ（SQL/Mongo/CQL/Cypher）の生成そのものは各エンジンが行い、ここでは
// ストア非依存の仕様だけを返す。

// TailEntity は tail を単一ストアで実行する際に一時領域へ staging する束縛エンティティ
// （alias）と、その永続テーブル/コレクション（Label）を表す。実行時に uuid をキーに
// 永続テーブルへ JOIN してプロパティを引く。
type TailEntity struct {
	Alias string   // 束縛エイリアス（例 "author"）
	Table string   // 永続テーブル/コレクション（例 "Person"）。JOIN 先
	Props []string // この alias から tail が参照するプロパティ（SELECT/GROUP BY 用）
}

// TailSpec は tail（Projection/Aggregate/Sort/Limit/Return）を単一ストアのネイティブクエリへ
// 翻訳するのに必要な情報。StoreFragment.Plan から LowerTail で導出する。
type TailSpec struct {
	Source    *StoreFragment // 束縛 UUID を供給する下位フラグメント（Emits==EmitBindings）
	InputSlot SlotTable      // alias → Source 出力 Record のスロット番号

	Entities   []TailEntity
	Return     []ReturnItem
	GroupKeys  []GroupKey
	Aggs       []AggregateItem
	OrderItems []OrderItem
	Limit      int
}

// LowerTail は論理サブプランが「tail を単一ストアへ委譲できる形」かを判定し、可能なら仕様を返す。
// 期待する形（根→葉）: Return→[Limit]→[Sort]→Aggregate→Projection→StoreFragment(EmitBindings)。
//
// この構造自体が「tail pushdown 形状」の唯一の判別子になる（＝入れ子の束縛フラグメントの有無）。
// 未対応エンジンや条件を満たさない場合は ok=false を返し、呼び出し側は Plan を通常実行すればよい。
func LowerTail(p PlanNode) (TailSpec, bool) {
	var spec TailSpec
	var proj *Projection
	var agg *Aggregate

	for n := p; n != nil; {
		switch op := n.(type) {
		case *Return:
			spec.Return = op.Items
		case *Limit:
			spec.Limit = op.Count
		case *Sort:
			spec.OrderItems = append(spec.OrderItems, op.OrderItems...)
		case *Aggregate:
			agg = op
			spec.GroupKeys = append(spec.GroupKeys, op.GroupKeys...)
			spec.Aggs = append(spec.Aggs, op.Aggs...)
		case *Projection:
			proj = op
		case *StoreFragment:
			if op.Emits == EmitBindings {
				spec.Source = op
			}
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}

	// tail 委譲には集約・Projection・束縛フラグメント（source）が揃っている必要がある。
	if agg == nil || proj == nil || spec.Source == nil {
		return TailSpec{}, false
	}
	// source は Projection の直下でなければならない（間に engine 演算子が挟まらない）。
	if f, ok := proj.Input.(*StoreFragment); !ok || f != spec.Source {
		return TailSpec{}, false
	}
	spec.InputSlot = proj.InputSlot

	entities, ok := tailEntities(proj, agg, spec.Return)
	if !ok {
		return TailSpec{}, false
	}
	spec.Entities = entities
	return spec, true
}

// tailEntities は staging すべきエンティティ（group key の alias ＋ prop/distinct を持つ agg の alias ＋
// 非集約 RETURN 項目の alias）を集め、各 alias の永続テーブル（Projection unit の Label）を解決する。
// いずれかの alias でテーブルが不明なら (_, false)。
func tailEntities(p *Projection, agg *Aggregate, ret []ReturnItem) ([]TailEntity, bool) {
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
	for _, it := range ret {
		if !it.IsAggregate && it.Alias != "" {
			need[it.Alias] = true
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

	// alias 順にソートして決定的にする（列番号 c0/c1… の割り当てを安定させる）。
	aliases := make([]string, 0, len(need))
	for alias := range need {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	entities := make([]TailEntity, 0, len(aliases))
	for _, alias := range aliases {
		table := labelOf[alias]
		if table == "" {
			return nil, false // JOIN 先テーブルが不明
		}
		entities = append(entities, TailEntity{Alias: alias, Table: table, Props: propsOf[alias]})
	}
	return entities, true
}
