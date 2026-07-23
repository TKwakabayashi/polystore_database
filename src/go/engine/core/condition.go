package core

import "polystore_database/src/go/plan"

// 条件演算子（plan.ConditionType）→ 各ストアのネイティブ演算子。
// 3エンジンで重複していたものを1本化（bulk 版は CondGreaterEq/CondLessEq が欠落して
// いたため、完全版の stream/volcano 版を採用した）。access 層と pushdown 双方が使う。

func SQLOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "<>"
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	case plan.CondGreaterEq:
		return ">="
	case plan.CondLessEq:
		return "<="
	default:
		return "="
	}
}

func MongoOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "$eq"
	case plan.CondNeq:
		return "$ne"
	case plan.CondGreater:
		return "$gt"
	case plan.CondLess:
		return "$lt"
	case plan.CondGreaterEq:
		return "$gte"
	case plan.CondLessEq:
		return "$lte"
	default:
		return "$eq"
	}
}

func CQLOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "!="
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	case plan.CondGreaterEq:
		return ">="
	case plan.CondLessEq:
		return "<="
	default:
		return "="
	}
}
