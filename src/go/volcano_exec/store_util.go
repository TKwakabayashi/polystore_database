package volcano_executor

import (
	"cmp"
	"fmt"
	"strconv"
	"time"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
)

// ---- 比較演算子の各ストア表記 ----

func mongoOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "$eq"
	case plan.CondNeq:
		return "$ne"
	case plan.CondGreater:
		return "$gt"
	case plan.CondLess:
		return "$lt"
	default:
		return "$eq"
	}
}

func sqlOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "<>"
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	default:
		return "="
	}
}

func cqlOp(t plan.ConditionType) string {
	switch t {
	case plan.CondEq:
		return "="
	case plan.CondNeq:
		return "!="
	case plan.CondGreater:
		return ">"
	case plan.CondLess:
		return "<"
	default:
		return "="
	}
}

// ---- KVS (LevelDB) 条件評価（点 Get） ----

func findPrimaryEqCondition(filters []*plan.ConditionNode) *plan.ConditionNode {
	for _, c := range filters {
		if c != nil && c.Type == plan.CondEq {
			return c
		}
	}
	return nil
}

// matchConditionsKVS は条件に現れるプロパティだけを点 Get して判定する。
func (p *Processor) matchConditionsKVS(label, uuid string, filters []*plan.ConditionNode) bool {
	for _, cond := range filters {
		if cond == nil {
			continue
		}
		p.countRoundTrip()
		valBytes, err := p.ldb.Get(codec.BuildEntityKey(label, uuid, cond.Property), nil)
		if err != nil {
			return false // プロパティ欠落 = 不一致
		}
		actual, _ := codec.ConvertToNativeType(codec.DecodeValue(valBytes, cond.DataType), cond.DataType)
		if !evalConditionKVS(actual, cond) {
			return false
		}
	}
	return true
}

func evalConditionKVS(actual interface{}, cond *plan.ConditionNode) bool {
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
	default:
		return false
	}
}
