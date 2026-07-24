package core

import (
	"testing"

	"polystore_database/src/go/plan"
)

// SQLOp/MongoOp/CQLOp は 3 ストアのネイティブ演算子への変換。全 6 演算子を網羅する。
// 直近バグ（bulk graph access が >= / <= を扱えず = にフォールバック）の回帰ガードとして
// CondGreaterEq / CondLessEq を明示ケース化する。
func TestStoreOperators(t *testing.T) {
	cases := []struct {
		name            string
		t               plan.ConditionType
		sql, mongo, cql string
	}{
		{"eq", plan.CondEq, "=", "$eq", "="},
		{"neq", plan.CondNeq, "<>", "$ne", "!="},
		{"gt", plan.CondGreater, ">", "$gt", ">"},
		{"lt", plan.CondLess, "<", "$lt", "<"},
		{"ge", plan.CondGreaterEq, ">=", "$gte", ">="},
		{"le", plan.CondLessEq, "<=", "$lte", "<="},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SQLOp(c.t); got != c.sql {
				t.Errorf("SQLOp(%v) = %q, want %q", c.t, got, c.sql)
			}
			if got := MongoOp(c.t); got != c.mongo {
				t.Errorf("MongoOp(%v) = %q, want %q", c.t, got, c.mongo)
			}
			if got := CQLOp(c.t); got != c.cql {
				t.Errorf("CQLOp(%v) = %q, want %q", c.t, got, c.cql)
			}
		})
	}
}

// 論理演算子など未対応の型は既定へフォールバックする（= / $eq / =）。
func TestStoreOperatorsFallback(t *testing.T) {
	if got := SQLOp(plan.CondAnd); got != "=" {
		t.Errorf("SQLOp(fallback) = %q, want =", got)
	}
	if got := MongoOp(plan.CondAnd); got != "$eq" {
		t.Errorf("MongoOp(fallback) = %q, want $eq", got)
	}
	if got := CQLOp(plan.CondAnd); got != "=" {
		t.Errorf("CQLOp(fallback) = %q, want =", got)
	}
}
