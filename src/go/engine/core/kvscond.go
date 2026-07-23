package core

import (
	"cmp"
	"fmt"
	"strconv"
	"time"

	"polystore_database/src/go/plan"
)

// KVS 条件評価ヘルパ。3エンジンで重複していたものを1本化
// （bulk 版は CondGreaterEq/CondLessEq が欠落していたため完全版の stream 版を採用）。

func FindPrimaryEqCondition(filters []*plan.ConditionNode) *plan.ConditionNode {
	for _, c := range filters {
		if c != nil && c.Type == plan.CondEq {
			return c
		}
	}
	return nil
}

func EvalConditionKVS(actual interface{}, cond *plan.ConditionNode) bool {
	if actual == nil {
		return false
	}
	switch cond.DataType {
	case "int", "integer", "long":
		var a int64
		switch v := actual.(type) {
		case int64:
			a = v
		case int32:
			a = int64(v)
		case int:
			a = int64(v)
		default:
			a, _ = strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		}
		b, _ := strconv.ParseInt(cond.Value, 10, 64)
		return compareOrderedKVS(a, b, cond.Type)

	case "float", "double":
		var a float64
		switch v := actual.(type) {
		case float64:
			a = v
		case float32:
			a = float64(v)
		default:
			a, _ = strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		}
		b, _ := strconv.ParseFloat(cond.Value, 64)
		return compareOrderedKVS(a, b, cond.Type)

	case "datetime", "date":
		a, ok := actual.(time.Time)
		if !ok {
			return false
		}
		var b time.Time
		var err error
		if cond.DataType == "date" {
			b, err = time.Parse("2006-01-02", cond.Value)
		} else {
			b, err = time.Parse(time.RFC3339, cond.Value)
		}
		if err != nil {
			return false
		}
		return compareTimeKVS(a, b, cond.Type)

	default:
		return compareOrderedKVS(fmt.Sprintf("%v", actual), cond.Value, cond.Type)
	}
}

func compareOrderedKVS[T cmp.Ordered](a, b T, op plan.ConditionType) bool {
	switch op {
	case plan.CondEq:
		return a == b
	case plan.CondNeq:
		return a != b
	case plan.CondGreater:
		return a > b
	case plan.CondLess:
		return a < b
	case plan.CondGreaterEq:
		return a >= b
	case plan.CondLessEq:
		return a <= b
	default:
		return false
	}
}

func compareTimeKVS(a, b time.Time, op plan.ConditionType) bool {
	switch op {
	case plan.CondEq:
		return a.Equal(b)
	case plan.CondNeq:
		return !a.Equal(b)
	case plan.CondGreater:
		return a.After(b)
	case plan.CondLess:
		return a.Before(b)
	case plan.CondGreaterEq:
		return !a.Before(b)
	case plan.CondLessEq:
		return !a.After(b)
	default:
		return false
	}
}
