// Package ir defines the plan vocabulary shared by the planner stages and the stores: the
// store-independent logical operators and the expressions over rows. It holds types only.
//
// Column key conventions:
//   - id column:    "a"           the id.UUID bound to node or relationship variable a
//   - value column: "a.p" / alias a property value, expression or aggregate result
//   - path column:  "p"           the list of ids [n0, r1, n1, ...] (built by ProjectPath)
//   - segment:      "_seg_s"      the contents of variable-length segment s (VarLengthExpand with EmitPath)
//   - list value:   "<list>.p"    the value list built by GetListProperty
package ir

import (
	"fmt"
	"strconv"
	"strings"

	"polystore_database/src/go/base/id"
)

// Expr is an expression evaluated on a row.
type Expr interface {
	exprNode()
	String() string
}

// Column references a row column (keys follow the conventions above).
type Column struct{ Key string }

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

// Call is a function call, including the path functions length/nodes/relationships (on path columns).
type Call struct {
	Fn   string
	Args []Expr
}

// Case is a searched CASE.
type Case struct {
	Whens []CaseWhen
	Else  Expr
}

// CaseWhen is one CASE branch.
type CaseWhen struct{ When, Then Expr }

// List is a list literal.
type List struct{ Elems []Expr }

// Quant is a list quantifier all/any/none/single(Var IN List WHERE Pred). Inside Pred the iteration
// variable's elements are referenced with ElemProp.
type Quant struct {
	Kind string
	Var  string
	List Expr
	Pred Expr
}

// ElemProp is property p of the quantifier variable's element (x.p). It is read at the same position
// from the matching value list column (GetListProperty's output).
type ElemProp struct {
	Var  string
	Prop string
}

// InSet is true when the id of Var is in id set Set: Var is an id column, or with Elem the quantifier
// variable's element. A set plan computes the set before the main plan runs (physical.WithSets); the
// executor puts its value in Bound, in a copy of the expression, before compiling or evaluating it.
type InSet struct {
	Var   string
	Elem  bool
	Set   string
	Bound *IDSet
}

// IDSet is the value of an id set.
type IDSet struct {
	IDs []id.UUID
	has map[id.UUID]bool
}

// NewIDSet makes a set of ids (duplicates are dropped; the order of first appearance is kept).
func NewIDSet(ids []id.UUID) *IDSet {
	s := &IDSet{has: make(map[id.UUID]bool, len(ids))}
	for _, u := range ids {
		if !s.has[u] {
			s.has[u] = true
			s.IDs = append(s.IDs, u)
		}
	}
	return s
}

// Has reports whether u is in the set.
func (s *IDSet) Has(u id.UUID) bool { return s != nil && s.has[u] }

func (*Column) exprNode()   {}
func (*Lit) exprNode()      {}
func (*Bin) exprNode()      {}
func (*Un) exprNode()       {}
func (*Call) exprNode()     {}
func (*Case) exprNode()     {}
func (*List) exprNode()     {}
func (*Quant) exprNode()    {}
func (*ElemProp) exprNode() {}
func (*InSet) exprNode()    {}

func (e *Column) String() string   { return e.Key }
func (e *ElemProp) String() string { return e.Var + "." + e.Prop }
func (e *InSet) String() string    { return e.Var + " IN $" + e.Set }

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

// Columns returns the column keys an expression reads, in order and without duplicates (not ElemProp).
func Columns(e Expr) []string {
	var out []string
	seen := map[string]bool{}
	add := func(k string) {
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	Walk(e, func(x Expr) {
		switch c := x.(type) {
		case *Column:
			add(c.Key)
		case *InSet:
			if !c.Elem {
				add(c.Var)
			}
		}
	})
	return out
}

// Transform returns a copy of e with every subexpression replaced by f, bottom-up (f sees the copied children).
func Transform(e Expr, f func(Expr) Expr) Expr {
	if e == nil {
		return nil
	}
	var cp Expr
	switch x := e.(type) {
	case *Bin:
		cp = &Bin{Op: x.Op, L: Transform(x.L, f), R: Transform(x.R, f)}
	case *Un:
		cp = &Un{Op: x.Op, X: Transform(x.X, f)}
	case *Call:
		c := &Call{Fn: x.Fn}
		for _, a := range x.Args {
			c.Args = append(c.Args, Transform(a, f))
		}
		cp = c
	case *Case:
		c := &Case{Else: Transform(x.Else, f)}
		for _, w := range x.Whens {
			c.Whens = append(c.Whens, CaseWhen{When: Transform(w.When, f), Then: Transform(w.Then, f)})
		}
		cp = c
	case *List:
		l := &List{}
		for _, el := range x.Elems {
			l.Elems = append(l.Elems, Transform(el, f))
		}
		cp = l
	case *Quant:
		cp = &Quant{Kind: x.Kind, Var: x.Var, List: Transform(x.List, f), Pred: Transform(x.Pred, f)}
	default:
		cp = e
	}
	return f(cp)
}

// HasSets reports whether e reads an id set (InSet).
func HasSets(e Expr) bool {
	found := false
	Walk(e, func(x Expr) {
		if _, ok := x.(*InSet); ok {
			found = true
		}
	})
	return found
}

// Walk visits the expression tree in pre-order.
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
