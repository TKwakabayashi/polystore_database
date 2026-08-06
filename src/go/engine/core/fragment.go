package core

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
)

// NOTE(P5): 現在 row-mode 実行は検証済みの StorePushdown ブリッジ（BuildRelationalSQL 等）に寄せて
// いるため、この FragmentSpec / LowerFragment / BuildRelationalSQLFromSpec は live 経路では未使用。
// projection-only 列のネイティブ SELECT 畳み込み（P6・ProjectionPushdown）で採用する足場として温存する。
//
// FragmentSpec は StoreFragment を各ストアのネイティブクエリへ翻訳するための正規化中間表現。
// projection-only（非集約）と集約の両方を表現できる。集約経路は現行 StorePushdown と等価。
//
// 生成規則:
//   - len(Aggs)==0: 純投影（GROUP BY なし）。Projections を SELECT する。
//   - len(Aggs)>0:  集約。GroupKeys を SELECT + GROUP BY し、Aggs を集約式にする。
type FragmentSpec struct {
	Table       string
	Filters     []*plan.ConditionNode
	Projections []ProjCol // 非集約の SELECT 列（集約時は無視）
	Aggs        []plan.AggregateItem
	GroupKeys   []plan.GroupKey
	OrderItems  []plan.OrderItem
	Limit       int
}

// ProjCol は非集約の投影列（物理カラム名 → 出力名）。
type ProjCol struct {
	Prop    string
	OutName string
}

// LowerFragment は StoreFragment.Ops の部分木を走査して FragmentSpec に正規化する。
// 融合パスが組む部分木の並びは（根→葉）: [Limit]→[Sort]→[Aggregate]→[Projection]→[Filter…]→EntityScan。
// graph 全体委譲（RawQuery）は本関数を通らない（Expand を含むため）。
func LowerFragment(f *plan.StoreFragment) FragmentSpec {
	var spec FragmentSpec
	for n := f.Ops; n != nil; {
		switch op := n.(type) {
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
						continue // 別ストアの列は統合演算子側で材料化する
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

// BuildRelationalSQLFromSpec は FragmentSpec から MySQL 用 SQL を生成する。
// 集約有無で SELECT / GROUP BY の形を切り替える（純投影は GROUP BY なし）。
func BuildRelationalSQLFromSpec(s FragmentSpec) (string, []interface{}) {
	var selects, groupCols []string
	if len(s.Aggs) > 0 {
		for _, g := range s.GroupKeys {
			selects = append(selects, sqlIdent(g.Prop)+" AS "+sqlIdent(g.OutName))
			groupCols = append(groupCols, sqlIdent(g.Prop))
		}
		for _, a := range s.Aggs {
			selects = append(selects, sqlAggExpr(a)+" AS "+sqlIdent(a.OutName))
		}
	} else {
		for _, p := range s.Projections {
			selects = append(selects, sqlIdent(p.Prop)+" AS "+sqlIdent(p.OutName))
		}
	}

	var where []string
	var args []interface{}
	for _, c := range s.Filters {
		if c == nil {
			continue
		}
		where = append(where, sqlIdent(c.Property)+" "+SQLOp(c.Type)+" ?")
		v, _ := codec.ConvertToNativeType(c.Value, c.DataType)
		args = append(args, v)
	}

	q := "SELECT " + strings.Join(selects, ", ") + " FROM " + sqlIdent(s.Table)
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	if len(groupCols) > 0 {
		q += " GROUP BY " + strings.Join(groupCols, ", ")
	}
	if len(s.OrderItems) > 0 {
		var ords []string
		for _, oi := range s.OrderItems {
			dir := "ASC"
			if oi.Direction == plan.OrderDesc {
				dir = "DESC"
			}
			ords = append(ords, sqlIdent(oi.Key)+" "+dir)
		}
		q += " ORDER BY " + strings.Join(ords, ", ")
	}
	if s.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", s.Limit)
	}
	return q, args
}

// NOTE: document(Mongo) / columnar(CQL) の spec ビルダは P3（各エンジンの fragment 実行の
// 配線）と同時に追加する。現行 StorePushdown 版（BuildMongoPipeline / BuildColumnarCQL）は
// P5 の退役までそのまま残す。
