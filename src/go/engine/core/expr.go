package core

import (
	"fmt"
	"math"
	"strings"

	"polystore_database/src/go/plan"
)

// expr.go は汎用式 IR（plan.Expr）を Row 値（map[string]any）上で評価する（Phase B1）。
// WITH/RETURN の計算列・複合 WHERE・CASE・算術・関数の実行基盤。全エンジンが共有する。
//
// 値表現: nil / bool / int64 / float64 / string / []any。数値は int64↔float64 を必要に応じ昇格。
// Cypher に倣い null は伝播し、比較は null 混在で null（unknown）、AND/OR は三値論理。

// EvalExpr は式 e を row 上で評価する。未知の構文は error。
func EvalExpr(e plan.Expr, row map[string]any) (any, error) {
	switch x := e.(type) {
	case *plan.Column:
		return row[x.Key], nil // 欠損キーは nil
	case *plan.Lit:
		return x.Val, nil
	case *plan.List:
		out := make([]any, len(x.Elems))
		for i, el := range x.Elems {
			v, err := EvalExpr(el, row)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case *plan.Un:
		return evalUnary(x, row)
	case *plan.Bin:
		return evalBinary(x, row)
	case *plan.Call:
		return evalCall(x, row)
	case *plan.Case:
		return evalCase(x, row)
	default:
		return nil, fmt.Errorf("EvalExpr: unsupported expr %T", e)
	}
}

// Truthy は WHERE 判定用。true のみ通過（nil/false/非bool は false）。
func Truthy(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func evalUnary(x *plan.Un, row map[string]any) (any, error) {
	v, err := EvalExpr(x.X, row)
	if err != nil {
		return nil, err
	}
	// IS NULL / IS NOT NULL は null を吸収して bool を返す（null 伝播しない）。
	switch x.Op {
	case plan.OpIsNull:
		return v == nil, nil
	case plan.OpIsNotNull:
		return v != nil, nil
	}
	if v == nil {
		return nil, nil
	}
	switch x.Op {
	case plan.OpNot:
		if b, ok := v.(bool); ok {
			return !b, nil
		}
		return nil, nil
	case plan.OpNeg:
		if i, ok := asInt(v); ok {
			return -i, nil
		}
		if f, ok := asFloat(v); ok {
			return -f, nil
		}
		return nil, fmt.Errorf("unary minus on non-numeric %T", v)
	}
	return nil, fmt.Errorf("bad unary op")
}

func evalBinary(x *plan.Bin, row map[string]any) (any, error) {
	// AND/OR は短絡＋三値論理のため、先に左右を評価しないケースを処理する。
	l, err := EvalExpr(x.L, row)
	if err != nil {
		return nil, err
	}
	switch x.Op {
	case plan.OpAnd:
		if l != nil && !boolOr(l, true) {
			return false, nil // false AND _ = false
		}
		r, err := EvalExpr(x.R, row)
		if err != nil {
			return nil, err
		}
		return threeValAnd(l, r), nil
	case plan.OpOr:
		if l != nil && boolOr(l, false) {
			return true, nil // true OR _ = true
		}
		r, err := EvalExpr(x.R, row)
		if err != nil {
			return nil, err
		}
		return threeValOr(l, r), nil
	}

	r, err := EvalExpr(x.R, row)
	if err != nil {
		return nil, err
	}

	switch x.Op {
	case plan.OpIn:
		return inList(l, r), nil
	case plan.OpEq, plan.OpNe, plan.OpLt, plan.OpLe, plan.OpGt, plan.OpGe:
		return compare(x.Op, l, r), nil
	case plan.OpAdd, plan.OpSub, plan.OpMul, plan.OpDiv, plan.OpMod:
		return arith(x.Op, l, r)
	}
	return nil, fmt.Errorf("bad binary op %v", x.Op)
}

func evalCall(x *plan.Call, row map[string]any) (any, error) {
	args := make([]any, len(x.Args))
	for i, a := range x.Args {
		v, err := EvalExpr(a, row)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	switch strings.ToLower(x.Fn) {
	case "coalesce":
		for _, a := range args {
			if a != nil {
				return a, nil
			}
		}
		return nil, nil
	case "size", "length":
		if len(args) == 1 {
			switch v := args[0].(type) {
			case []any:
				return int64(len(v)), nil
			case string:
				return int64(len([]rune(v))), nil
			case nil:
				return nil, nil
			}
		}
		return nil, fmt.Errorf("%s: bad arg", x.Fn)
	case "head":
		if len(args) == 1 {
			if v, ok := args[0].([]any); ok {
				if len(v) == 0 {
					return nil, nil
				}
				return v[0], nil
			}
			if args[0] == nil {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("head: bad arg")
	case "last":
		if len(args) == 1 {
			if v, ok := args[0].([]any); ok {
				if len(v) == 0 {
					return nil, nil
				}
				return v[len(v)-1], nil
			}
			if args[0] == nil {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("last: bad arg")
	case "tointeger":
		if len(args) == 1 {
			if args[0] == nil {
				return nil, nil
			}
			if i, ok := asInt(args[0]); ok {
				return i, nil
			}
			if f, ok := asFloat(args[0]); ok {
				return int64(f), nil
			}
		}
		return nil, nil
	case "tofloat":
		if len(args) == 1 {
			if args[0] == nil {
				return nil, nil
			}
			if f, ok := asFloat(args[0]); ok {
				return f, nil
			}
		}
		return nil, nil
	case "abs":
		if len(args) == 1 && args[0] != nil {
			if i, ok := asInt(args[0]); ok {
				if i < 0 {
					return -i, nil
				}
				return i, nil
			}
			if f, ok := asFloat(args[0]); ok {
				return math.Abs(f), nil
			}
		}
		return nil, nil
	case "floor":
		if len(args) == 1 && args[0] != nil {
			if f, ok := asFloat(args[0]); ok {
				return math.Floor(f), nil
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("EvalExpr: unsupported function %q", x.Fn)
	}
}

func evalCase(x *plan.Case, row map[string]any) (any, error) {
	var subj any
	var err error
	if x.Subject != nil {
		if subj, err = EvalExpr(x.Subject, row); err != nil {
			return nil, err
		}
	}
	for _, w := range x.Whens {
		wv, err := EvalExpr(w.When, row)
		if err != nil {
			return nil, err
		}
		var match bool
		if x.Subject != nil {
			match = valueEq(subj, wv) // 単純 CASE: Subject = When
		} else {
			match = Truthy(wv) // 探索 CASE: When が真
		}
		if match {
			return EvalExpr(w.Then, row)
		}
	}
	if x.Else != nil {
		return EvalExpr(x.Else, row)
	}
	return nil, nil
}

// ---- 値ヘルパ ----

func asInt(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	default:
		return 0, false
	}
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}

func isFloat(v any) bool {
	switch v.(type) {
	case float64, float32:
		return true
	}
	return false
}

func arith(op plan.BinOp, l, r any) (any, error) {
	if l == nil || r == nil {
		return nil, nil
	}
	// どちらかが float なら float 演算、両方 int なら int 演算。
	if isFloat(l) || isFloat(r) {
		lf, ok1 := asFloat(l)
		rf, ok2 := asFloat(r)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("arith on non-numeric")
		}
		switch op {
		case plan.OpAdd:
			return lf + rf, nil
		case plan.OpSub:
			return lf - rf, nil
		case plan.OpMul:
			return lf * rf, nil
		case plan.OpDiv:
			if rf == 0 {
				return nil, nil
			}
			return lf / rf, nil
		case plan.OpMod:
			return math.Mod(lf, rf), nil
		}
	}
	li, ok1 := asInt(l)
	ri, ok2 := asInt(r)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arith on non-numeric")
	}
	switch op {
	case plan.OpAdd:
		return li + ri, nil
	case plan.OpSub:
		return li - ri, nil
	case plan.OpMul:
		return li * ri, nil
	case plan.OpDiv:
		if ri == 0 {
			return nil, nil
		}
		return li / ri, nil
	case plan.OpMod:
		if ri == 0 {
			return nil, nil
		}
		return li % ri, nil
	}
	return nil, fmt.Errorf("bad arith op")
}

// compare は比較演算。null 混在は null（unknown）を返す。bool は等値のみ、数値/文字列は順序比較。
func compare(op plan.BinOp, l, r any) any {
	if l == nil || r == nil {
		return nil
	}
	c, ok := cmpValues(l, r)
	if !ok {
		// 型不一致（例: string vs number）は等値のみ判定可、順序は null。
		switch op {
		case plan.OpEq:
			return false
		case plan.OpNe:
			return true
		default:
			return nil
		}
	}
	switch op {
	case plan.OpEq:
		return c == 0
	case plan.OpNe:
		return c != 0
	case plan.OpLt:
		return c < 0
	case plan.OpLe:
		return c <= 0
	case plan.OpGt:
		return c > 0
	case plan.OpGe:
		return c >= 0
	}
	return nil
}

// cmpValues は同種の値を比較して -1/0/1 を返す。比較不能なら ok=false。
func cmpValues(l, r any) (int, bool) {
	// 数値同士
	if (isNum(l)) && (isNum(r)) {
		lf, _ := asFloat(l)
		rf, _ := asFloat(r)
		switch {
		case lf < rf:
			return -1, true
		case lf > rf:
			return 1, true
		default:
			return 0, true
		}
	}
	// 文字列同士
	if ls, ok := l.(string); ok {
		if rs, ok := r.(string); ok {
			return strings.Compare(ls, rs), true
		}
	}
	// bool 同士
	if lb, ok := l.(bool); ok {
		if rb, ok := r.(bool); ok {
			if lb == rb {
				return 0, true
			}
			return 1, true // 不等（順序は未定義だが等値判定に使う）
		}
	}
	return 0, false
}

func isNum(v any) bool {
	switch v.(type) {
	case int, int32, int64, float32, float64:
		return true
	}
	return false
}

func valueEq(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	c, ok := cmpValues(a, b)
	return ok && c == 0
}

func inList(l, r any) any {
	list, ok := r.([]any)
	if !ok {
		return nil
	}
	found := false
	sawNull := false
	for _, item := range list {
		if l == nil || item == nil {
			sawNull = true
			continue
		}
		if valueEq(l, item) {
			found = true
			break
		}
	}
	if found {
		return true
	}
	if l == nil || sawNull {
		return nil // 三値: 見つからず null 混在なら unknown
	}
	return false
}

// boolOr は bool 値を取り出す（非 bool は def）。
func boolOr(v any, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func threeValAnd(l, r any) any {
	lb, lok := l.(bool)
	rb, rok := r.(bool)
	if lok && !lb || rok && !rb {
		return false
	}
	if !lok || !rok {
		return nil
	}
	return lb && rb
}

func threeValOr(l, r any) any {
	lb, lok := l.(bool)
	rb, rok := r.(bool)
	if lok && lb || rok && rb {
		return true
	}
	if !lok || !rok {
		return nil
	}
	return lb || rb
}
