// Package ast copies the ANTLR parse tree into a query AST that follows the grammar one to one.
// It makes no planning decision (operators and stores are decided by logical and physical).
package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// Query is a whole query: its clauses in order (MATCH / WITH / RETURN).
type Query struct {
	Clauses []Clause
}

// Clause is a MatchClause, UnwindClause or ProjectionClause.
type Clause interface{ clause() }

// MatchClause is a MATCH (comma-separated patterns) and the WHERE right after it. Optional makes it an
// OPTIONAL MATCH (rows without a match stay with the new variables null; WHERE is part of the match).
type MatchClause struct {
	Optional bool
	Patterns []PatternPath
	Where    Expr // may be nil
}

// UnwindClause is UNWIND (one row per list element, bound to Var).
type UnwindClause struct {
	Expr Expr
	Var  string
}

// ProjectionClause is the common form of WITH and RETURN.
type ProjectionClause struct {
	IsReturn bool
	Distinct bool
	Items    []ProjItem
	OrderBy  []SortItem
	Skip     *int
	Limit    *int
	Where    Expr // the WHERE after WITH (always nil for RETURN)
}

func (*MatchClause) clause()      {}
func (*ProjectionClause) clause() {}
func (*UnwindClause) clause()     {}

// PatternPath is one pattern `(p =)? (n0)-[r1]-(n1)...`; Rels[i] connects Nodes[i] and Nodes[i+1]
// (a quantified step is a RelPat with Quantified set).
type PatternPath struct {
	PathVar string // p in p = (...); empty if none
	Nodes   []NodePat
	Rels    []RelPat
}

// NodePat is a node pattern `(v:L {k: val} WHERE c)`. Parse names anonymous nodes `_nN`.
type NodePat struct {
	Var    string
	Labels []string
	Props  []PropEq
	Where  Expr // may be nil
}

// RelPat is a relationship pattern `-[v:T1|T2 *h {k: val} WHERE c]->`. Parse names anonymous ones `_rN`.
// Dir is the direction from the left node to the right node of the pattern.
// A quantified step has only Var (named `_qN`) and Quantified.
type RelPat struct {
	Var        string
	Types      []string
	Dir        Direction
	Range      *HopRange // nil for a single hop
	Props      []PropEq
	Where      Expr // may be nil
	Quantified *QuantifiedPath
}

// QuantifiedPath is a quantified path pattern `((x)-[r]->(y) WHERE c){m,n}` between the nodes around it, its first
// node following the left node and its last node preceding the right node in each iteration. A quantified
// relationship `-[r]->{m,n}` is the group of r between two anonymous nodes. Its variables are group variables.
type QuantifiedPath struct {
	Nodes []NodePat
	Rels  []RelPat
	Where Expr // may be nil
	Range HopRange
}

// PropEq is one in-pattern property `{k: val}`.
type PropEq struct {
	Key   string
	Value Expr
}

// ProjItem is one WITH / RETURN item. Name is the AS alias, or the item's text.
type ProjItem struct {
	Expr  Expr
	Name  string
	Alias bool // whether AS named it
}

// SortItem is one ORDER BY item.
type SortItem struct {
	Expr Expr
	Desc bool
}

// Direction is a relationship's direction in the pattern.
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

// HopRange is a variable-length range. nil Min/Max are unset (Min defaults to 1, Max to unbounded).
type HopRange struct {
	Min *int
	Max *int
}

// Quantifier returns the range as a quantifier (`{1,3}`, `{2}`, `{0,}`).
func (h HopRange) Quantifier() string {
	lo, hi := "", ""
	if h.Min != nil {
		lo = strconv.Itoa(*h.Min)
	}
	if h.Max != nil {
		hi = strconv.Itoa(*h.Max)
	}
	if h.Min != nil && h.Max != nil && *h.Min == *h.Max {
		return "{" + lo + "}"
	}
	return "{" + lo + "," + hi + "}"
}

// Cypher returns the range in Cypher notation (`*`, `*2..`, `*..3`, `*2`, `*1..3`).
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

// ---- expressions (syntactic; resolved to columns by the logical translation) ----

// Expr is an AST expression node.
type Expr interface {
	expr()
	String() string
}

// VarRef is a variable reference `a`.
type VarRef struct{ Name string }

// PropRef is a property reference `a.p`.
type PropRef struct {
	Var  string
	Prop string
}

// Lit is a literal (nil / bool / int64 / float64 / string).
type Lit struct{ Val any }

// Bin is a binary operation; Op is one of "+","-","*","/","%","=","<>","<",">","<=",">=","AND","OR","IN".
type Bin struct {
	Op   string
	L, R Expr
}

// Un is a unary operation; Op is one of "-","NOT","IS NULL","IS NOT NULL".
type Un struct {
	Op string
	X  Expr
}

// Call is a function call (coalesce / size / length / nodes / relationships ...). Fn is lowercase.
type Call struct {
	Fn   string
	Args []Expr
}

// Agg is an aggregation (count/sum/avg/min/max/collect); a nil Arg is count(*).
type Agg struct {
	Fn       string
	Distinct bool
	Arg      Expr
}

// Case is a CASE expression: a searched CASE (WHEN conditions) when Subject is nil, otherwise a simple CASE
// (each WHEN value compared with Subject by `=`; as in Neo4j, WHEN null never matches).
type Case struct {
	Subject Expr
	Whens   []CaseWhen
	Else    Expr // may be nil (= null)
}

// CaseWhen is one CASE branch.
type CaseWhen struct{ When, Then Expr }

// List is a list literal.
type List struct{ Elems []Expr }

// Quant is a list quantifier `all/any/none/single(x IN list WHERE pred)`.
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

// ---- dump (deterministic, for goldens) ----

// String prints the query deterministically (one clause per line, indented).
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

// String prints the pattern close to Cypher.
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
	return "(" + s + propsString(n.Props) + whereString(n.Where) + ")"
}

func (r RelPat) String() string {
	if q := r.Quantified; q != nil {
		var b strings.Builder
		b.WriteString(" (")
		for i, n := range q.Nodes {
			b.WriteString(n.String())
			if i < len(q.Rels) {
				b.WriteString(q.Rels[i].String())
			}
		}
		b.WriteString(whereString(q.Where) + ")" + q.Range.Quantifier() + " ")
		return b.String()
	}
	s := r.Var
	if len(r.Types) > 0 {
		s += ":" + strings.Join(r.Types, "|")
	}
	if r.Range != nil {
		s += r.Range.Cypher()
	}
	body := "[" + s + propsString(r.Props) + whereString(r.Where) + "]"
	switch r.Dir {
	case DirRight:
		return "-" + body + "->"
	case DirLeft:
		return "<-" + body + "-"
	default:
		return "-" + body + "-"
	}
}

func whereString(e Expr) string {
	if e == nil {
		return ""
	}
	return " WHERE " + e.String()
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
