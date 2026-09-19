package plan

import (
	"fmt"
	"strings"
)

// expr.go は汎用式 IR（Phase B1）。WITH/RETURN の計算列・WHERE の複合条件・CASE・算術・関数を
// 表現する。評価器は engine/core（EvalExpr）が Row 値（map[string]any）上で行う。plan は値表現に
// 依存しない純粋な構文木のみを持つ。
//
// 列参照（Column.Key）の規約は Projection が生成する wide row のキーに一致させる:
//   - プロパティ: "alias.prop"
//   - 束縛 UUID:  "alias"
//   - 計算列/集約: その出力名（AS 別名）

// Expr は式ノードの共通インタフェース。
type Expr interface {
	exprNode()
	String() string
}

// Column は row のキーで値を引く列参照。
type Column struct{ Key string }

// Lit はリテラル値（nil/bool/int64/float64/string）。
type Lit struct{ Val any }

// Bin は二項演算。
type Bin struct {
	Op   BinOp
	L, R Expr
}

// Un は単項演算。
type Un struct {
	Op UnOp
	X  Expr
}

// Call は関数呼び出し（coalesce/size/head/toInteger/toFloat/floor 等）。
type Call struct {
	Fn   string
	Args []Expr
}

// Case は CASE 式。Subject が非 nil なら単純 CASE（Subject を各 When と比較）、nil なら探索 CASE。
type Case struct {
	Subject Expr
	Whens   []CaseWhen
	Else    Expr // 省略時 nil（= null）
}

type CaseWhen struct{ When, Then Expr }

// List はリストリテラル。
type List struct{ Elems []Expr }

func (*Column) exprNode() {}
func (*Lit) exprNode()    {}
func (*Bin) exprNode()    {}
func (*Un) exprNode()     {}
func (*Call) exprNode()   {}
func (*Case) exprNode()   {}
func (*List) exprNode()   {}

// BinOp は二項演算子。
type BinOp int

const (
	OpAdd BinOp = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
	OpAnd
	OpOr
	OpIn // L IN R（R はリスト）
)

func (o BinOp) String() string {
	switch o {
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpMod:
		return "%"
	case OpEq:
		return "="
	case OpNe:
		return "<>"
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpIn:
		return "IN"
	default:
		return "?"
	}
}

// UnOp は単項演算子。
type UnOp int

const (
	OpNeg       UnOp = iota // -x
	OpNot                   // NOT x
	OpIsNull                // x IS NULL
	OpIsNotNull             // x IS NOT NULL
)

func (o UnOp) String() string {
	switch o {
	case OpNot:
		return "NOT"
	case OpIsNull:
		return "IS NULL"
	case OpIsNotNull:
		return "IS NOT NULL"
	default:
		return "-"
	}
}

// ---- String（決定的ダンプ用）----

func (e *Column) String() string { return e.Key }

func (e *Lit) String() string {
	if e.Val == nil {
		return "null"
	}
	if s, ok := e.Val.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", e.Val)
}

func (e *Bin) String() string {
	return fmt.Sprintf("(%s %s %s)", e.L.String(), e.Op.String(), e.R.String())
}

func (e *Un) String() string {
	if e.Op == OpIsNull || e.Op == OpIsNotNull {
		return fmt.Sprintf("(%s %s)", e.X.String(), e.Op.String())
	}
	return fmt.Sprintf("%s(%s)", e.Op.String(), e.X.String())
}

func (e *Call) String() string {
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = a.String()
	}
	return fmt.Sprintf("%s(%s)", e.Fn, strings.Join(args, ", "))
}

func (e *Case) String() string {
	var b strings.Builder
	b.WriteString("CASE")
	if e.Subject != nil {
		b.WriteString(" " + e.Subject.String())
	}
	for _, w := range e.Whens {
		b.WriteString(fmt.Sprintf(" WHEN %s THEN %s", w.When.String(), w.Then.String()))
	}
	if e.Else != nil {
		b.WriteString(" ELSE " + e.Else.String())
	}
	b.WriteString(" END")
	return b.String()
}

func (e *List) String() string {
	elems := make([]string, len(e.Elems))
	for i, x := range e.Elems {
		elems[i] = x.String()
	}
	return "[" + strings.Join(elems, ", ") + "]"
}
