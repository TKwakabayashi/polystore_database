package core

import (
	"testing"

	"polystore_database/src/go/plan"
)

func col(k string) plan.Expr                      { return &plan.Column{Key: k} }
func lit(v any) plan.Expr                         { return &plan.Lit{Val: v} }
func bin(op plan.BinOp, l, r plan.Expr) plan.Expr { return &plan.Bin{Op: op, L: l, R: r} }

func TestEvalArithmetic(t *testing.T) {
	row := map[string]any{"a": int64(10), "b": int64(3), "f": 2.5}
	cases := []struct {
		e    plan.Expr
		want any
	}{
		{bin(plan.OpAdd, col("a"), col("b")), int64(13)},
		{bin(plan.OpSub, col("a"), col("b")), int64(7)},
		{bin(plan.OpMul, col("a"), col("b")), int64(30)},
		{bin(plan.OpDiv, col("a"), col("b")), int64(3)},
		{bin(plan.OpMod, col("a"), col("b")), int64(1)},
		{bin(plan.OpAdd, col("a"), col("f")), 12.5}, // int+float→float
		{bin(plan.OpMul, lit(int64(1)), col("a")), int64(10)},
		{bin(plan.OpDiv, col("a"), lit(int64(0))), nil}, // /0 → null
	}
	for i, c := range cases {
		got, err := EvalExpr(c.e, row)
		if err != nil {
			t.Fatalf("[%d] err: %v", i, err)
		}
		if got != c.want {
			t.Errorf("[%d] %s = %v(%T) want %v", i, c.e.String(), got, got, c.want)
		}
	}
}

func TestEvalComparisonAndLogic(t *testing.T) {
	row := map[string]any{"a": int64(5), "s": "foo"}
	cases := []struct {
		e    plan.Expr
		want any
	}{
		{bin(plan.OpGt, col("a"), lit(int64(3))), true},
		{bin(plan.OpLe, col("a"), lit(int64(5))), true},
		{bin(plan.OpEq, col("s"), lit("foo")), true},
		{bin(plan.OpLt, col("s"), lit("zzz")), true},
		{bin(plan.OpEq, col("missing"), lit(int64(1))), nil}, // null compare → null
		{bin(plan.OpAnd, lit(true), lit(false)), false},
		{bin(plan.OpAnd, lit(false), col("missing")), false}, // short-circuit false
		{bin(plan.OpOr, lit(true), col("missing")), true},    // short-circuit true
		{bin(plan.OpAnd, lit(true), col("missing")), nil},    // true AND null = null
		{bin(plan.OpOr, lit(false), col("missing")), nil},    // false OR null = null
	}
	for i, c := range cases {
		got, err := EvalExpr(c.e, row)
		if err != nil {
			t.Fatalf("[%d] err: %v", i, err)
		}
		if got != c.want {
			t.Errorf("[%d] %s = %v want %v", i, c.e.String(), got, c.want)
		}
	}
}

func TestEvalInList(t *testing.T) {
	list := &plan.List{Elems: []plan.Expr{lit("ar"), lit("hu")}}
	row := map[string]any{}
	if v, _ := EvalExpr(bin(plan.OpIn, lit("hu"), list), row); v != true {
		t.Errorf("hu IN [ar,hu] = %v want true", v)
	}
	if v, _ := EvalExpr(bin(plan.OpIn, lit("xx"), list), row); v != false {
		t.Errorf("xx IN [ar,hu] = %v want false", v)
	}
}

func TestEvalCase(t *testing.T) {
	// 探索 CASE: length のビン分け
	searched := &plan.Case{
		Whens: []plan.CaseWhen{
			{When: bin(plan.OpLt, col("len"), lit(int64(40))), Then: lit(int64(0))},
			{When: bin(plan.OpLt, col("len"), lit(int64(80))), Then: lit(int64(1))},
		},
		Else: lit(int64(2)),
	}
	for _, tc := range []struct {
		len  int64
		want any
	}{{20, int64(0)}, {60, int64(1)}, {200, int64(2)}} {
		got, _ := EvalExpr(searched, map[string]any{"len": tc.len})
		if got != tc.want {
			t.Errorf("searched CASE len=%d = %v want %v", tc.len, got, tc.want)
		}
	}
	// 単純 CASE: uni.name WHEN null THEN false ELSE true
	simple := &plan.Case{
		Subject: col("x"),
		Whens:   []plan.CaseWhen{{When: lit(nil), Then: lit(false)}},
		Else:    lit(true),
	}
	if got, _ := EvalExpr(simple, map[string]any{"x": nil}); got != false {
		t.Errorf("simple CASE null = %v want false", got)
	}
	if got, _ := EvalExpr(simple, map[string]any{"x": "MIT"}); got != true {
		t.Errorf("simple CASE non-null = %v want true", got)
	}
}

func TestEvalFunctions(t *testing.T) {
	row := map[string]any{"l": []any{int64(1), int64(2), int64(3)}, "n": nil}
	checks := []struct {
		e    plan.Expr
		want any
	}{
		{&plan.Call{Fn: "coalesce", Args: []plan.Expr{col("n"), lit("x")}}, "x"},
		{&plan.Call{Fn: "size", Args: []plan.Expr{col("l")}}, int64(3)},
		{&plan.Call{Fn: "head", Args: []plan.Expr{col("l")}}, int64(1)},
		{&plan.Call{Fn: "toInteger", Args: []plan.Expr{lit(2.9)}}, int64(2)},
		{&plan.Call{Fn: "floor", Args: []plan.Expr{lit(2.9)}}, 2.0},
	}
	for i, c := range checks {
		got, err := EvalExpr(c.e, row)
		if err != nil {
			t.Fatalf("[%d] err: %v", i, err)
		}
		if got != c.want {
			t.Errorf("[%d] %s = %v want %v", i, c.e.String(), got, c.want)
		}
	}
}

// TestEvalIsNull は IS NULL / IS NOT NULL 単項の評価（null 吸収）を検証する（T3）。
func TestEvalIsNull(t *testing.T) {
	row := map[string]any{"a": int64(5), "n": nil}
	un := func(op plan.UnOp, x plan.Expr) plan.Expr { return &plan.Un{Op: op, X: x} }
	cases := []struct {
		e    plan.Expr
		want any
	}{
		{un(plan.OpIsNull, col("a")), false},
		{un(plan.OpIsNull, col("n")), true},
		{un(plan.OpIsNull, col("missing")), true},
		{un(plan.OpIsNotNull, col("a")), true},
		{un(plan.OpIsNotNull, col("n")), false},
		{un(plan.OpIsNotNull, col("missing")), false},
	}
	for i, c := range cases {
		got, err := EvalExpr(c.e, row)
		if err != nil {
			t.Fatalf("[%d] err: %v", i, err)
		}
		if got != c.want {
			t.Errorf("[%d] %s = %v want %v", i, c.e.String(), got, c.want)
		}
	}
}
