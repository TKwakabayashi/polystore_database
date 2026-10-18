package executor

import (
	"fmt"
	"math"
	"strings"

	"polystore_database/src/go/id"
	"polystore_database/src/go/logical"
)

// eval.go は行の上で論理式を評価する（Cypher の意味。null は伝播し、AND/OR は 3 値論理）。

// elemCtx は量化子の反復変数 → 対象のリスト列と位置。
type elemCtx map[string]elemPos

type elemPos struct {
	source string
	part   logical.SegPart
	index  int
}

// Eval は行 r の上で式 e を評価する。
func Eval(e logical.Expr, r Row) (any, error) { return eval(e, r, nil) }

func eval(e logical.Expr, r Row, ec elemCtx) (any, error) {
	switch x := e.(type) {
	case *logical.Column:
		v, ok := r[x.Key]
		if !ok {
			return nil, fmt.Errorf("executor: column %s not in row", x.Key)
		}
		return v, nil
	case *logical.Lit:
		return normalize(x.Val), nil
	case *logical.ElemProp:
		pos, ok := ec[x.Var]
		if !ok {
			return nil, fmt.Errorf("executor: unbound iteration variable %s", x.Var)
		}
		list, _ := r[logical.ListPropKey(pos.source, pos.part, x.Prop)].([]any)
		if pos.index >= len(list) {
			return nil, nil
		}
		return list[pos.index], nil
	case *logical.Un:
		v, err := eval(x.X, r, ec)
		if err != nil {
			return nil, err
		}
		switch x.Op {
		case "IS NULL":
			return v == nil, nil
		case "IS NOT NULL":
			return v != nil, nil
		case "NOT":
			if v == nil {
				return nil, nil
			}
			b, ok := v.(bool)
			if !ok {
				return nil, fmt.Errorf("executor: NOT on %T", v)
			}
			return !b, nil
		case "-":
			switch n := v.(type) {
			case nil:
				return nil, nil
			case int64:
				return -n, nil
			case float64:
				return -n, nil
			}
			return nil, fmt.Errorf("executor: negation of %T", v)
		}
	case *logical.Bin:
		return evalBin(x, r, ec)
	case *logical.List:
		out := make([]any, len(x.Elems))
		for i, el := range x.Elems {
			v, err := eval(el, r, ec)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case *logical.Case:
		for _, w := range x.Whens {
			c, err := eval(w.When, r, ec)
			if err != nil {
				return nil, err
			}
			if b, ok := c.(bool); ok && b {
				return eval(w.Then, r, ec)
			}
		}
		if x.Else == nil {
			return nil, nil
		}
		return eval(x.Else, r, ec)
	case *logical.Call:
		return evalCall(x, r, ec)
	case *logical.Quant:
		return evalQuant(x, r, ec)
	}
	return nil, fmt.Errorf("executor: expression %T", e)
}

func evalBin(x *logical.Bin, r Row, ec elemCtx) (any, error) {
	l, err := eval(x.L, r, ec)
	if err != nil {
		return nil, err
	}
	// AND / OR は 3 値論理（片側が決まれば他方の null を問わない）。
	if x.Op == "AND" || x.Op == "OR" {
		rv, err := eval(x.R, r, ec)
		if err != nil {
			return nil, err
		}
		lb, lok := l.(bool)
		rb, rok := rv.(bool)
		if x.Op == "AND" {
			switch {
			case (lok && !lb) || (rok && !rb):
				return false, nil
			case lok && rok:
				return true, nil
			}
			return nil, nil
		}
		switch {
		case (lok && lb) || (rok && rb):
			return true, nil
		case lok && rok:
			return false, nil
		}
		return nil, nil
	}
	rv, err := eval(x.R, r, ec)
	if err != nil {
		return nil, err
	}
	switch x.Op {
	case "=", "<>":
		if l == nil || rv == nil {
			return nil, nil
		}
		c, ok := compare(l, rv)
		if !ok {
			return x.Op == "<>", nil // 型が違えば等しくない
		}
		return (c == 0) == (x.Op == "="), nil
	case "<", ">", "<=", ">=":
		c, ok := compare(l, rv)
		if !ok {
			return nil, nil
		}
		switch x.Op {
		case "<":
			return c < 0, nil
		case ">":
			return c > 0, nil
		case "<=":
			return c <= 0, nil
		}
		return c >= 0, nil
	case "IN":
		list, ok := rv.([]any)
		if !ok {
			return nil, nil
		}
		if l == nil {
			return nil, nil
		}
		sawNull := false
		for _, el := range list {
			if el == nil {
				sawNull = true
				continue
			}
			if c, ok := compare(l, el); ok && c == 0 {
				return true, nil
			}
		}
		if sawNull {
			return nil, nil
		}
		return false, nil
	case "+":
		if ls, ok := l.(string); ok {
			if rs, ok := rv.(string); ok {
				return ls + rs, nil
			}
		}
		if ll, ok := l.([]any); ok {
			if rl, ok := rv.([]any); ok {
				return append(append([]any{}, ll...), rl...), nil
			}
		}
		return arith(x.Op, l, rv)
	case "-", "*", "/", "%":
		return arith(x.Op, l, rv)
	}
	return nil, fmt.Errorf("executor: operator %s", x.Op)
}

func arith(op string, a, b any) (any, error) {
	if a == nil || b == nil {
		return nil, nil
	}
	ai, aInt := a.(int64)
	bi, bInt := b.(int64)
	if aInt && bInt {
		switch op {
		case "+":
			return ai + bi, nil
		case "-":
			return ai - bi, nil
		case "*":
			return ai * bi, nil
		case "/":
			if bi == 0 {
				return nil, fmt.Errorf("executor: division by zero")
			}
			return ai / bi, nil
		case "%":
			if bi == 0 {
				return nil, fmt.Errorf("executor: division by zero")
			}
			return ai % bi, nil
		}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if !aok || !bok {
		return nil, fmt.Errorf("executor: %s on %T and %T", op, a, b)
	}
	switch op {
	case "+":
		return af + bf, nil
	case "-":
		return af - bf, nil
	case "*":
		return af * bf, nil
	case "/":
		return af / bf, nil
	}
	return math.Mod(af, bf), nil
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, true
	}
	return 0, false
}

func evalCall(x *logical.Call, r Row, ec elemCtx) (any, error) {
	if x.Fn == "length" || x.Fn == "nodes" || x.Fn == "relationships" {
		src, ok := x.Args[0].(*logical.Column)
		if !ok {
			return nil, fmt.Errorf("executor: %s argument", x.Fn)
		}
		lv, _ := r[src.Key].(*ListVal)
		if lv == nil {
			return nil, nil
		}
		switch x.Fn {
		case "length":
			return int64(len(lv.Rels)), nil
		case "nodes":
			return uuidList(lv.Nodes), nil
		}
		return uuidList(lv.Rels), nil
	}
	args := make([]any, len(x.Args))
	for i, a := range x.Args {
		v, err := eval(a, r, ec)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	switch x.Fn {
	case "coalesce":
		for _, a := range args {
			if a != nil {
				return a, nil
			}
		}
		return nil, nil
	case "size":
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case []any:
			return int64(len(v)), nil
		case string:
			return int64(len([]rune(v))), nil
		}
	case "abs":
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case int64:
			if v < 0 {
				return -v, nil
			}
			return v, nil
		case float64:
			return math.Abs(v), nil
		}
	case "floor", "ceil", "round":
		f, ok := toFloat(args[0])
		if !ok {
			return nil, nil
		}
		switch x.Fn {
		case "floor":
			return math.Floor(f), nil
		case "ceil":
			return math.Ceil(f), nil
		}
		return math.Round(f), nil
	case "tointeger":
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case int64:
			return v, nil
		case float64:
			return int64(v), nil
		case string:
			var n int64
			if _, err := fmt.Sscan(v, &n); err == nil {
				return n, nil
			}
			return nil, nil
		}
	case "tofloat":
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case int64:
			return float64(v), nil
		case float64:
			return v, nil
		case string:
			var f float64
			if _, err := fmt.Sscan(v, &f); err == nil {
				return f, nil
			}
			return nil, nil
		}
	case "tostring":
		if args[0] == nil {
			return nil, nil
		}
		return fmt.Sprint(args[0]), nil
	case "tolower", "toupper":
		s, ok := args[0].(string)
		if !ok {
			return nil, nil
		}
		if x.Fn == "tolower" {
			return strings.ToLower(s), nil
		}
		return strings.ToUpper(s), nil
	case "head":
		if l, ok := args[0].([]any); ok && len(l) > 0 {
			return l[0], nil
		}
		return nil, nil
	}
	return nil, fmt.Errorf("executor: function %s(%T)", x.Fn, firstOrNil(args))
}

func firstOrNil(xs []any) any {
	if len(xs) == 0 {
		return nil
	}
	return xs[0]
}

func uuidList(us []id.UUID) []any {
	out := make([]any, len(us))
	for i, u := range us {
		out[i] = u
	}
	return out
}

// evalQuant は all/any/none/single(x IN nodes(src)|relationships(src) WHERE pred)。
// 要素のプロパティは GetListProperty が作った値のリスト列から同じ位置で読む。
func evalQuant(x *logical.Quant, r Row, ec elemCtx) (any, error) {
	call, ok := x.List.(*logical.Call)
	if !ok || len(call.Args) != 1 {
		return nil, fmt.Errorf("executor: quantifier list %s", x.List)
	}
	src, ok := call.Args[0].(*logical.Column)
	if !ok {
		return nil, fmt.Errorf("executor: quantifier source %s", call.Args[0])
	}
	part := logical.SegNodes
	if call.Fn == "relationships" {
		part = logical.SegRels
	}
	lv, _ := r[src.Key].(*ListVal)
	if lv == nil {
		return nil, nil
	}
	n := len(lv.Nodes)
	if part == logical.SegRels {
		n = len(lv.Rels)
	}
	inner := elemCtx{}
	for k, v := range ec {
		inner[k] = v
	}
	trues, nulls := 0, 0
	for i := 0; i < n; i++ {
		inner[x.Var] = elemPos{source: src.Key, part: part, index: i}
		v, err := eval(x.Pred, r, inner)
		if err != nil {
			return nil, err
		}
		switch b := v.(type) {
		case bool:
			if b {
				trues++
			}
		default:
			nulls++
		}
	}
	falses := n - trues - nulls
	switch x.Kind {
	case "all":
		if falses > 0 {
			return false, nil
		}
		if nulls > 0 {
			return nil, nil
		}
		return true, nil
	case "any":
		if trues > 0 {
			return true, nil
		}
		if nulls > 0 {
			return nil, nil
		}
		return false, nil
	case "none":
		if trues > 0 {
			return false, nil
		}
		if nulls > 0 {
			return nil, nil
		}
		return true, nil
	case "single":
		if nulls > 0 {
			return nil, nil
		}
		return trues == 1, nil
	}
	return nil, fmt.Errorf("executor: quantifier %s", x.Kind)
}

// truthy は Filter が行を残すか（真のときだけ）。
func truthy(v any) bool {
	b, ok := v.(bool)
	return ok && b
}
