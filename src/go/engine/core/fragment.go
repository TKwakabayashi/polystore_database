package core

import (
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// fragment.go は「委譲する論理サブプラン（StoreFragment.Plan）→ ネイティブクエリ生成に必要な
// 正規化仕様（FragmentSpec）」への lowering を一本化する。物理演算子に平坦フィールドを持たせず、
// 意味の源泉であるサブプランを走査して都度導出する。
//
// 各ストアのネイティブクエリ生成（SQL/Mongo/CQL）は pushdown.go の Build* が FragmentSpec を消費する。

// ProjCol は非集約の投影列（物理カラム名 → 出力名）。Items が空の projection-only 経路で使う。
type ProjCol struct {
	Prop    string
	OutName string
}

// FragmentSpec は StoreFragment を各ストアのネイティブクエリへ翻訳するための正規化中間表現。
//
// 生成規則:
//   - Verbatim != "": graph 全体委譲。原 Cypher をそのまま発行する（Build* は通らない）。
//   - len(Items) > 0: RETURN 項目駆動で SELECT を組む（順序＝出力順。集約/非集約が混在）。
//   - それ以外:       Projections 駆動の純投影（GROUP BY なし）。
type FragmentSpec struct {
	Store store.Kind

	// graph 全体委譲用（Plan の既知の忠実な lowering）。
	Verbatim string
	Params   map[string]string

	Table      string
	Filters    []*plan.ConditionNode
	Items      []plan.ReturnItem // RETURN 項目（SELECT 生成の主）
	GroupKeys  []plan.GroupKey
	Aggs       []plan.AggregateItem
	OrderItems []plan.OrderItem
	Limit      int

	Projections []ProjCol // Items が無い projection-only 経路用
}

// LowerFragment は StoreFragment.Plan（委譲する論理サブプラン）を走査して FragmentSpec に正規化する。
// 期待する並び（根→葉）: [Return]→[Limit]→[Sort]→[Aggregate]→[Projection]→[Filter…]→EntityScan。
// graph 全体委譲（Verbatim）は Expand を含み得るため、走査結果ではなく Verbatim をそのまま使う。
func LowerFragment(f *plan.StoreFragment) FragmentSpec {
	spec := FragmentSpec{
		Store:    f.Store,
		Verbatim: f.Verbatim,
		Params:   f.Params,
	}
	if spec.Verbatim != "" {
		return spec
	}

	for n := f.Plan; n != nil; {
		switch op := n.(type) {
		case *plan.Return:
			spec.Items = op.Items
		case *plan.Limit:
			spec.Limit = op.Count
		case *plan.Sort:
			spec.OrderItems = append(spec.OrderItems, op.OrderItems...)
		case *plan.Aggregate:
			spec.Aggs = append(spec.Aggs, op.Aggs...)
			spec.GroupKeys = append(spec.GroupKeys, op.GroupKeys...)
		case *plan.Projection:
			for _, u := range op.Units {
				for _, fe := range u.Fetches {
					if fe.Store != f.Store {
						continue // 別ストアの列は統合（材料化）側で扱う
					}
					for _, p := range fe.Props {
						spec.Projections = append(spec.Projections, ProjCol{Prop: p, OutName: u.Alias + "." + p})
					}
				}
			}
		case *plan.Filter:
			spec.Filters = append(spec.Filters, op.Filter...)
		case *plan.EntityScan:
			if spec.Table == "" && len(op.Labels) > 0 {
				spec.Table = op.Labels[0]
			}
			spec.Filters = append(spec.Filters, op.Filter...)
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	return spec
}
