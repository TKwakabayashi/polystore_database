// Package logical はクエリ処理系（刷新版）の論理計画: ストアを一切考えない細粒度の演算子と、
// 行（列の集合）に対する式を定義する。
//
// 設計: docs/operator_plan.md §2〜§3。
//
// 列キーの規約（行は map ではなく列の集合として扱うが、キー名はここで統一する）:
//   - 識別列:   "a"           … 変数 a に束縛されたノード/関係の id.UUID
//   - 値列:     "a.p" / 出力名 … プロパティ値・式・集約の結果
//   - パス列:   "p"           … 識別子のリスト [n0, r1, n1, …]（ProjectPath が作る）
//   - 区間列:   "_seg_s"      … 可変長区間 s の中身（VarLengthExpand が EmitPath のとき作る）
//   - リスト値: "<list>.p"    … GetListProperty が作る値のリスト
package logical

import (
	"fmt"
	"strconv"
	"strings"
)

// Expr は行の上で評価する式。
type Expr interface {
	exprNode()
	String() string
}

// Column は行の列参照（キーは上の規約）。
type Column struct{ Key string }

// Lit はリテラル（nil / bool / int64 / float64 / string）。
type Lit struct{ Val any }

// Bin は二項演算。Op は "+","-","*","/","%","=","<>","<",">","<=",">=","AND","OR","IN"。
type Bin struct {
	Op   string
	L, R Expr
}

// Un は単項演算。Op は "-","NOT","IS NULL","IS NOT NULL"。
type Un struct {
	Op string
	X  Expr
}

// Call は関数呼び出し。パス関数 length/nodes/relationships もここで表す（引数はパス列）。
type Call struct {
	Fn   string
	Args []Expr
}

// Case は探索 CASE。
type Case struct {
	Whens []CaseWhen
	Else  Expr
}

// CaseWhen は CASE の 1 分岐。
type CaseWhen struct{ When, Then Expr }

// List はリストリテラル。
type List struct{ Elems []Expr }

// Quant はリスト量化子 all/any/none/single(Var IN List WHERE Pred)。
// Pred の中で反復変数の要素を参照するときは ElemProp を使う。
type Quant struct {
	Kind string
	Var  string
	List Expr
	Pred Expr
}

// ElemProp は量化子の反復変数 Var の要素のプロパティ（x.p）。評価時は、対応する
// 値のリスト列（GetListProperty の出力）から同じ位置の値を取る。
type ElemProp struct {
	Var  string
	Prop string
}

func (*Column) exprNode()   {}
func (*Lit) exprNode()      {}
func (*Bin) exprNode()      {}
func (*Un) exprNode()       {}
func (*Call) exprNode()     {}
func (*Case) exprNode()     {}
func (*List) exprNode()     {}
func (*Quant) exprNode()    {}
func (*ElemProp) exprNode() {}

func (e *Column) String() string   { return e.Key }
func (e *ElemProp) String() string { return e.Var + "." + e.Prop }

func (e *Lit) String() string {
	switch v := e.Val.(type) {
	case nil:
		return "null"
	case string:
		return strconv.Quote(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (e *Bin) String() string { return fmt.Sprintf("(%s %s %s)", e.L, e.Op, e.R) }

func (e *Un) String() string {
	if strings.HasPrefix(e.Op, "IS") {
		return fmt.Sprintf("(%s %s)", e.X, e.Op)
	}
	return fmt.Sprintf("%s(%s)", e.Op, e.X)
}

func (e *Call) String() string { return fmt.Sprintf("%s(%s)", e.Fn, joinExprs(e.Args)) }

func (e *Case) String() string {
	var b strings.Builder
	b.WriteString("CASE")
	for _, w := range e.Whens {
		fmt.Fprintf(&b, " WHEN %s THEN %s", w.When, w.Then)
	}
	if e.Else != nil {
		fmt.Fprintf(&b, " ELSE %s", e.Else)
	}
	b.WriteString(" END")
	return b.String()
}

func (e *List) String() string { return "[" + joinExprs(e.Elems) + "]" }

func (e *Quant) String() string {
	return fmt.Sprintf("%s(%s IN %s WHERE %s)", e.Kind, e.Var, e.List, e.Pred)
}

func joinExprs(xs []Expr) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = x.String()
	}
	return strings.Join(parts, ", ")
}

// Columns は式が参照する列キーを出現順・重複なしで返す（ElemProp は含まない）。
func Columns(e Expr) []string {
	var out []string
	seen := map[string]bool{}
	Walk(e, func(x Expr) {
		if c, ok := x.(*Column); ok && !seen[c.Key] {
			seen[c.Key] = true
			out = append(out, c.Key)
		}
	})
	return out
}

// Walk は式木を前順で訪問する。
func Walk(e Expr, f func(Expr)) {
	if e == nil {
		return
	}
	f(e)
	switch x := e.(type) {
	case *Bin:
		Walk(x.L, f)
		Walk(x.R, f)
	case *Un:
		Walk(x.X, f)
	case *Call:
		for _, a := range x.Args {
			Walk(a, f)
		}
	case *Case:
		for _, w := range x.Whens {
			Walk(w.When, f)
			Walk(w.Then, f)
		}
		Walk(x.Else, f)
	case *List:
		for _, el := range x.Elems {
			Walk(el, f)
		}
	case *Quant:
		Walk(x.List, f)
		Walk(x.Pred, f)
	}
}
