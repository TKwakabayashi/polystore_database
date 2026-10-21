// Package ast はクエリ処理系（刷新版）の段階 A: ANTLR のパース木を、文法と 1 対 1 に対応する
// クエリ AST へ写し取る。ここでは計画判断を一切しない（どの演算子にするか・どのストアかは
// logical / physical の責務）。
//
// 設計: docs/operator_plan.md §4.1。
package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// Query はクエリ全体。節を出現順に並べる（MATCH / WITH / RETURN）。
type Query struct {
	Clauses []Clause
}

// Clause は MatchClause / UnwindClause / ProjectionClause。
type Clause interface{ clause() }

// MatchClause は MATCH（カンマ区切りの複数パターン）と、その直後の WHERE。
// Optional なら OPTIONAL MATCH（一致しない行は新しい変数を null にして残す。WHERE は一致の条件）。
type MatchClause struct {
	Optional bool
	Patterns []PatternPath
	Where    Expr // nil 可
}

// UnwindClause は UNWIND（リストの要素ごとに 1 行へ展開し、要素を Var に束縛する）。
type UnwindClause struct {
	Expr Expr
	Var  string
}

// ProjectionClause は WITH と RETURN の共通形。
type ProjectionClause struct {
	IsReturn bool
	Distinct bool
	Items    []ProjItem
	OrderBy  []SortItem
	Skip     *int
	Limit    *int
	Where    Expr // WITH の後ろの WHERE（RETURN では常に nil）
}

func (*MatchClause) clause()      {}
func (*ProjectionClause) clause() {}
func (*UnwindClause) clause()     {}

// PatternPath は 1 本のパターン `(p =)? (n0)-[r1]-(n1)…`。Rels[i] は Nodes[i] と Nodes[i+1] を結ぶ。
type PatternPath struct {
	PathVar string // p = ( … ) の p。無ければ空
	Nodes   []NodePat
	Rels    []RelPat
}

// NodePat はノードパターン `(v:L {k: val})`。名前が無ければ Parse が `_nN` を付ける。
type NodePat struct {
	Var    string
	Labels []string
	Props  []PropEq
}

// RelPat は関係パターン `-[v:T1|T2 *h {k: val}]->`。名前が無ければ Parse が `_rN` を付ける。
// Dir は「パターン上で左のノード → 右のノード」に対する向き。
type RelPat struct {
	Var   string
	Types []string
	Dir   Direction
	Range *HopRange // nil なら固定長 1 ホップ
	Props []PropEq
}

// PropEq はパターン内プロパティ `{k: val}` の 1 項。
type PropEq struct {
	Key   string
	Value Expr
}

// ProjItem は WITH / RETURN の 1 項目。Name は AS 別名、無ければ元の式の文字列。
type ProjItem struct {
	Expr  Expr
	Name  string
	Alias bool // AS で名前が明示されたか
}

// SortItem は ORDER BY の 1 項目。
type SortItem struct {
	Expr Expr
	Desc bool
}

// Direction はパターン上の関係の向き。
type Direction int

const (
	DirRight Direction = iota // (a)-[]->(b)
	DirLeft                   // (a)<-[]-(b)
	DirBoth                   // (a)-[]-(b)
)

func (d Direction) String() string {
	switch d {
	case DirRight:
		return "->"
	case DirLeft:
		return "<-"
	default:
		return "--"
	}
}

// HopRange は可変長のホップ範囲。Min/Max が nil なら「指定なし」（Min 省略は 1、Max 省略は上限なし）。
type HopRange struct {
	Min *int
	Max *int
}

// Cypher は範囲を Cypher の表記（`*`, `*2..`, `*..3`, `*2`, `*1..3`）で返す。
func (h HopRange) Cypher() string {
	switch {
	case h.Min == nil && h.Max == nil:
		return "*"
	case h.Max == nil:
		return fmt.Sprintf("*%d..", *h.Min)
	case h.Min == nil:
		return fmt.Sprintf("*..%d", *h.Max)
	case *h.Min == *h.Max:
		return fmt.Sprintf("*%d", *h.Min)
	default:
		return fmt.Sprintf("*%d..%d", *h.Min, *h.Max)
	}
}

// ---- 式（構文としての式。列への解決は logical の翻訳で行う）----

// Expr は AST の式ノード。
type Expr interface {
	expr()
	String() string
}

// VarRef は変数参照 `a`。
type VarRef struct{ Name string }

// PropRef はプロパティ参照 `a.p`。
type PropRef struct {
	Var  string
	Prop string
}

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

// Call は関数呼び出し（coalesce / size / length / nodes / relationships …）。Fn は小文字。
type Call struct {
	Fn   string
	Args []Expr
}

// Agg は集約（count/sum/avg/min/max/collect）。Arg が nil なら count(*)。
type Agg struct {
	Fn       string
	Distinct bool
	Arg      Expr
}

// Case は CASE 式。Subject が nil なら探索 CASE（WHEN 条件）、非 nil なら単純 CASE
// （WHEN の値と Subject を `=` で比べる。Neo4j と同じく WHEN null は一致しない）。
type Case struct {
	Subject Expr
	Whens   []CaseWhen
	Else    Expr // nil 可（= null）
}

// CaseWhen は CASE の 1 分岐。
type CaseWhen struct{ When, Then Expr }

// List はリストリテラル。
type List struct{ Elems []Expr }

// Quant はリスト量化子 `all/any/none/single(x IN list WHERE pred)`。
type Quant struct {
	Kind string // "all","any","none","single"
	Var  string
	List Expr
	Pred Expr
}

func (*VarRef) expr()  {}
func (*PropRef) expr() {}
func (*Lit) expr()     {}
func (*Bin) expr()     {}
func (*Un) expr()      {}
func (*Call) expr()    {}
func (*Agg) expr()     {}
func (*Case) expr()    {}
func (*List) expr()    {}
func (*Quant) expr()   {}

func (e *VarRef) String() string  { return e.Name }
func (e *PropRef) String() string { return e.Var + "." + e.Prop }

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

func (e *Agg) String() string {
	arg := "*"
	if e.Arg != nil {
		arg = e.Arg.String()
	}
	if e.Distinct {
		arg = "DISTINCT " + arg
	}
	return fmt.Sprintf("%s(%s)", e.Fn, arg)
}

func (e *Case) String() string {
	var b strings.Builder
	b.WriteString("CASE")
	if e.Subject != nil {
		b.WriteString(" " + e.Subject.String())
	}
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

// ---- ダンプ（golden 用の決定的表示）----

// String は Query を決定的に表示する（1 節 1 行＋インデント）。
func (q *Query) String() string {
	var b strings.Builder
	for _, c := range q.Clauses {
		switch c := c.(type) {
		case *MatchClause:
			parts := make([]string, len(c.Patterns))
			for i, p := range c.Patterns {
				parts[i] = p.String()
			}
			kw := "MATCH"
			if c.Optional {
				kw = "OPTIONAL MATCH"
			}
			fmt.Fprintf(&b, "%s %s\n", kw, strings.Join(parts, ", "))
			if c.Where != nil {
				fmt.Fprintf(&b, "  WHERE %s\n", c.Where)
			}
		case *UnwindClause:
			fmt.Fprintf(&b, "UNWIND %s AS %s\n", c.Expr, c.Var)
		case *ProjectionClause:
			kw := "WITH"
			if c.IsReturn {
				kw = "RETURN"
			}
			if c.Distinct {
				kw += " DISTINCT"
			}
			items := make([]string, len(c.Items))
			for i, it := range c.Items {
				items[i] = fmt.Sprintf("%s AS %s", it.Expr, it.Name)
			}
			fmt.Fprintf(&b, "%s %s\n", kw, strings.Join(items, ", "))
			if len(c.OrderBy) > 0 {
				ords := make([]string, len(c.OrderBy))
				for i, o := range c.OrderBy {
					dir := "ASC"
					if o.Desc {
						dir = "DESC"
					}
					ords[i] = o.Expr.String() + " " + dir
				}
				fmt.Fprintf(&b, "  ORDER BY %s\n", strings.Join(ords, ", "))
			}
			if c.Skip != nil {
				fmt.Fprintf(&b, "  SKIP %d\n", *c.Skip)
			}
			if c.Limit != nil {
				fmt.Fprintf(&b, "  LIMIT %d\n", *c.Limit)
			}
			if c.Where != nil {
				fmt.Fprintf(&b, "  WHERE %s\n", c.Where)
			}
		}
	}
	return b.String()
}

// String はパターンを Cypher に近い形で表示する。
func (p PatternPath) String() string {
	var b strings.Builder
	if p.PathVar != "" {
		b.WriteString(p.PathVar + " = ")
	}
	for i, n := range p.Nodes {
		b.WriteString(n.String())
		if i < len(p.Rels) {
			b.WriteString(p.Rels[i].String())
		}
	}
	return b.String()
}

func (n NodePat) String() string {
	s := n.Var
	for _, l := range n.Labels {
		s += ":" + l
	}
	return "(" + s + propsString(n.Props) + ")"
}

func (r RelPat) String() string {
	s := r.Var
	if len(r.Types) > 0 {
		s += ":" + strings.Join(r.Types, "|")
	}
	if r.Range != nil {
		s += r.Range.Cypher()
	}
	body := "[" + s + propsString(r.Props) + "]"
	switch r.Dir {
	case DirRight:
		return "-" + body + "->"
	case DirLeft:
		return "<-" + body + "-"
	default:
		return "-" + body + "-"
	}
}

func propsString(ps []PropEq) string {
	if len(ps) == 0 {
		return ""
	}
	parts := make([]string, len(ps))
	for i, p := range ps {
		parts[i] = p.Key + ": " + p.Value.String()
	}
	return " {" + strings.Join(parts, ", ") + "}"
}
