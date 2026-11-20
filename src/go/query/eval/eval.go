package eval

import (
	"fmt"
	"math"
	"polystore_database/src/go/query/ir"
	"strings"

	"polystore_database/src/go/base/id"
)

// eval.go evaluates logical expressions on a row with Cypher semantics (null propagates; AND/OR are three-valued).

// ElemCtx maps a quantifier's iteration variable to its list column and position.
type ElemCtx map[string]elemPos

type elemPos struct {
	source string
	part   ir.SegPart
	index  int
}

// Eval evaluates e on row r.
func Eval(e ir.Expr, r Row) (any, error) { return eval(e, r, nil) }

func eval(e ir.Expr, r Row, ec ElemCtx) (any, error) {
	switch x := e.(type) {
	case *ir.Column:
		v, ok := r[x.Key]
		if !ok {
			return nil, fmt.Errorf("executor: column %s not in row", x.Key)
		}
		return v, nil
	case *ir.Lit:
		return Normalize(x.Val), nil
	case *ir.ElemProp:
		pos, ok := ec[x.Var]
		if !ok {
			return nil, fmt.Errorf("executor: unbound iteration variable %s", x.Var)
		}
		list, _ := r[ir.ListPropKey(pos.source, pos.part, x.Prop)].([]any)
		if pos.index >= len(list) {
			return nil, nil
		}
		return list[pos.index], nil
	case *ir.InSet:
		return evalInSet(x, r, ec)
	case *ir.Un:
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
	case *ir.Bin:
		return evalBin(x, r, ec)
	case *ir.List:
		out := make([]any, len(x.Elems))
		for i, el := range x.Elems {
			v, err := eval(el, r, ec)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case *ir.Case:
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
	case *ir.Call:
		return evalCall(x, r, ec)
	case *ir.Quant:
		return evalQuant(x, r, ec)
	}
	return nil, fmt.Errorf("executor: expression %T", e)
}

func evalBin(x *ir.Bin, r Row, ec ElemCtx) (any, error) {
	l, err := eval(x.L, r, ec)
	if err != nil {
		return nil, err
	}
	// AND / OR are three-valued (once one side decides, a null on the other does not matter).
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
			return x.Op == "<>", nil // different types are not equal
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
		return Arith(x.Op, l, rv)
	case "-", "*", "/", "%":
		return Arith(x.Op, l, rv)
	}
	return nil, fmt.Errorf("executor: operator %s", x.Op)
}

func Arith(op string, a, b any) (any, error) {
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
	af, aok := ToFloat(a)
	bf, bok := ToFloat(b)
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

func ToFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, true
	}
	return 0, false
}

func evalCall(x *ir.Call, r Row, ec ElemCtx) (any, error) {
	if x.Fn == "length" || x.Fn == "nodes" || x.Fn == "relationships" {
		src, ok := x.Args[0].(*ir.Column)
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
		f, ok := ToFloat(args[0])
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

// evalInSet reads the id of an id column or of the quantifier variable's element and looks it up in the bound set (null for a null id).
func evalInSet(x *ir.InSet, r Row, ec ElemCtx) (any, error) {
	if x.Bound == nil {
		return nil, fmt.Errorf("executor: id set %s is not bound", x.Set)
	}
	var v any
	if x.Elem {
		pos, ok := ec[x.Var]
		if !ok {
			return nil, fmt.Errorf("executor: unbound iteration variable %s", x.Var)
		}
		lv, _ := r[pos.source].(*ListVal)
		if lv == nil {
			return nil, nil
		}
		elems := lv.Nodes
		if pos.part == ir.SegRels {
			elems = lv.Rels
		}
		if pos.index >= len(elems) {
			return nil, nil
		}
		v = elems[pos.index]
	} else {
		v = r[x.Var]
	}
	u, ok := v.(id.UUID)
	if !ok || u.Empty() {
		return nil, nil
	}
	return x.Bound.Has(u), nil
}

// evalQuant evaluates all/any/none/single(x IN nodes(src)|relationships(src) WHERE pred). Element properties
// are read at the same position from the value list columns built by GetListProperty.
func evalQuant(x *ir.Quant, r Row, ec ElemCtx) (any, error) {
	call, ok := x.List.(*ir.Call)
	if !ok || len(call.Args) != 1 {
		return nil, fmt.Errorf("executor: quantifier list %s", x.List)
	}
	src, ok := call.Args[0].(*ir.Column)
	if !ok {
		return nil, fmt.Errorf("executor: quantifier source %s", call.Args[0])
	}
	part := ir.SegNodes
	if call.Fn == "relationships" {
		part = ir.SegRels
	}
	lv, _ := r[src.Key].(*ListVal)
	if lv == nil {
		return nil, nil
	}
	n := len(lv.Nodes)
	if part == ir.SegRels {
		n = len(lv.Rels)
	}
	inner := ElemCtx{}
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

// Truthy reports whether a Filter keeps the row (true only).
func Truthy(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

// EvalElem evaluates e with variable v bound to element index of the list column (source, part),
// as a quantifier's predicate sees it.
func EvalElem(e ir.Expr, r Row, v, source string, part ir.SegPart, index int) (any, error) {
	return eval(e, r, ElemCtx{v: {source: source, part: part, index: index}})
}
