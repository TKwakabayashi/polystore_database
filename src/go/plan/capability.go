package plan

import (
	"fmt"

	"polystore_database/src/go/store"
)

// OpCapability は「このストアはこの演算をネイティブ実行できるか」を問い合わせる単位。
// 融合パス（planner/fusion.go）と engine/core のクエリ生成可否判定が参照する。
type OpCapability int

const (
	CapFilter OpCapability = iota
	CapExpand
	CapVarExpand
	CapProject
	CapAggregate
	CapGroupBy
	CapSort
	CapLimit
	CapDistinct
)

var capNames = map[OpCapability]string{
	CapFilter:    "filter",
	CapExpand:    "expand",
	CapVarExpand: "var_expand",
	CapProject:   "project",
	CapAggregate: "aggregate",
	CapGroupBy:   "group_by",
	CapSort:      "sort",
	CapLimit:     "limit",
	CapDistinct:  "distinct",
}

func (c OpCapability) String() string {
	if s, ok := capNames[c]; ok {
		return s
	}
	return fmt.Sprintf("OpCapability(%d)", int(c))
}

// capabilities は各ストアがネイティブ実行できる演算の宣言的テーブル。
// 融合パスと engine/core はこれを唯一の真実として参照し、対応不可なら自動でフォールバックする。
//   - graph:               全演算（traversal を含む）。
//   - relational/document:  traversal 以外を汎用サポート。
//   - columnar(Cassandra):  CQL 制約により Filter/Project/Aggregate のみ（GroupBy/Sort/Distinct 不可）。
//   - kvs(LevelDB):         ネイティブ集約なし（Filter/Project のみ）。
var capabilities = map[store.Kind]map[OpCapability]bool{
	store.Graph: {
		CapFilter: true, CapExpand: true, CapVarExpand: true, CapProject: true,
		CapAggregate: true, CapGroupBy: true, CapSort: true, CapLimit: true, CapDistinct: true,
	},
	store.Relational: {
		CapFilter: true, CapProject: true, CapAggregate: true, CapGroupBy: true,
		CapSort: true, CapLimit: true, CapDistinct: true,
	},
	store.Document: {
		CapFilter: true, CapProject: true, CapAggregate: true, CapGroupBy: true,
		CapSort: true, CapLimit: true, CapDistinct: true,
	},
	store.Columnar: {
		CapFilter: true, CapProject: true, CapAggregate: true,
	},
	store.Kvs: {
		CapFilter: true, CapProject: true,
	},
}

// Supports は store k が演算 c をネイティブ実行できるかを返す。
func Supports(k store.Kind, c OpCapability) bool {
	return capabilities[k][c]
}
