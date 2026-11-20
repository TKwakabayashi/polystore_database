package ir

import (
	"fmt"
	"strings"
)

// Node is a logical operator: meaning only, no slots or stores (decided by physical planning).
type Node interface {
	Children() []Node
	// Describe prints the operator alone (no children) on one deterministic line, for goldens.
	Describe() string
}

// Direction is a relationship's direction from From to To.
type Direction int

const (
	Out  Direction = iota // (From)-[]->(To)
	In                    // (From)<-[]-(To)
	Both                  // (From)-[]-(To)
)

// Reverse is the direction with From and To swapped.
func (d Direction) Reverse() Direction {
	switch d {
	case Out:
		return In
	case In:
		return Out
	default:
		return Both
	}
}

// Pattern renders a relationship pattern with its direction ("-[r:T]->", ...).
func (d Direction) Pattern(body string) string {
	switch d {
	case Out:
		return "-[" + body + "]->"
	case In:
		return "<-[" + body + "]-"
	default:
		return "-[" + body + "]-"
	}
}

// ObjKind is the kind of a property's owner.
type ObjKind int

const (
	NodeObj ObjKind = iota
	RelObj
)

func (k ObjKind) String() string {
	if k == RelObj {
		return "rel"
	}
	return "node"
}

// HopRange is a variable-length range. Max < 0 is unbounded; Min defaults to 1.
type HopRange struct {
	Min int
	Max int
}

// Cypher returns the range in Cypher notation (`*1..3`, `*2..`, `*3`).
func (h HopRange) Cypher() string {
	switch {
	case h.Max < 0:
		return fmt.Sprintf("*%d..", h.Min)
	case h.Min == h.Max:
		return fmt.Sprintf("*%d", h.Min)
	default:
		return fmt.Sprintf("*%d..%d", h.Min, h.Max)
	}
}

// Empty reports whether no length satisfies the range.
func (h HopRange) Empty() bool { return h.Max >= 0 && h.Min > h.Max }

// Constrain intersects the range with "hops op k" (op is = < <= > >=); ok=false for an op it cannot express (<>).
func (h HopRange) Constrain(op string, k int) (HopRange, bool) {
	lo, hi := h.Min, h.Max
	tighten := func(newLo, newHi int) {
		if newLo > lo {
			lo = newLo
		}
		if newHi >= 0 && (hi < 0 || newHi < hi) {
			hi = newHi
		}
	}
	switch op {
	case "=":
		tighten(k, k)
	case "<":
		tighten(lo, k-1)
	case "<=":
		tighten(lo, k)
	case ">":
		tighten(k+1, -1)
	case ">=":
		tighten(k, -1)
	default:
		return h, false
	}
	return HopRange{Min: lo, Max: hi}, true
}

// SegPart selects the elements of a variable-length segment.
type SegPart int

const (
	SegNodes SegPart = iota // the segment's inner nodes (endpoints excluded)
	SegRels                 // all relationships of the segment
)

func (p SegPart) String() string {
	if p == SegRels {
		return "rels"
	}
	return "nodes"
}

// SegPred is a condition every element of a segment must meet: `all(Var IN part WHERE Pred)`. Element
// properties in Pred are ElemProp{Var, ...}.
type SegPred struct {
	Part SegPart
	Var  string
	Pred Expr
}

func (s SegPred) String() string {
	return fmt.Sprintf("all(%s IN %s WHERE %s)", s.Var, s.Part, s.Pred)
}

// ================================
// Operators that touch stores (placed by the physical planner)
// ================================

// EntityScan scans the entities with Labels and binds Alias (a leaf).
type EntityScan struct {
	Alias  string
	Labels []string
}

// GetProperty adds Alias's property Prop as value column "Alias.Prop" (a left outer join: null if absent).
// Labels / Kind identify the owner (the mapping lookup keys).
type GetProperty struct {
	Input  Node
	Alias  string
	Prop   string
	Kind   ObjKind
	Labels []string
}

// Key is the key of the value column it adds.
func (g *GetProperty) Key() string { return g.Alias + "." + g.Prop }

// GetListProperty reads Prop for each element of Part of list column Source (a path or segment column)
// and adds value list column "Source.part.Prop". ElemLabels are the elements' label candidates.
type GetListProperty struct {
	Input      Node
	Source     string
	Part       SegPart
	Prop       string
	ElemLabels []string
}

// Key is the key of the list value column it adds.
func (g *GetListProperty) Key() string { return ListPropKey(g.Source, g.Part, g.Prop) }

// ListPropKey is GetListProperty's output key.
func ListPropKey(source string, part SegPart, prop string) string {
	return source + "." + part.String() + "." + prop
}

// Expand follows relationships from bound From and binds Rel and To.
type Expand struct {
	Input    Node
	From     string
	Rel      string
	To       string
	Types    []string
	Dir      Direction
	ToLabels []string
}

// ExpandInto binds the relationship Rel between bound From and To (closing a cycle).
type ExpandInto struct {
	Input Node
	From  string
	Rel   string
	To    string
	Types []string
	Dir   Direction
}

// VarLengthExpand follows Range hops from From and binds To, one row per path. SegPreds and InnerLabels
// constrain the segment's elements; with EmitPath the segment is output as column SegColumn(Seg).
type VarLengthExpand struct {
	Input       Node
	From        string
	Seg         string
	To          string
	Types       []string
	Dir         Direction
	ToLabels    []string
	Range       HopRange
	SegPreds    []SegPred
	InnerLabels []string
	EmitPath    bool
}

// SegColumn is the segment column key of variable-length segment seg.
func SegColumn(seg string) string { return "_seg_" + seg }

// ================================
// Computation-only operators
// ================================

// Filter keeps the rows where Pred is true (three-valued). In normal form it is one AND term.
type Filter struct {
	Input Node
	Pred  Expr
}

// ProjItem is one projected column.
type ProjItem struct {
	Name string
	Expr Expr
}

// Projection selects, renames and computes columns (WITH / RETURN).
type Projection struct {
	Input Node
	Items []ProjItem
}

// PathElem is part of a path: a variable-length segment when Seg (Var is the segment), else a bound node or relationship.
type PathElem struct {
	Var string
	Seg bool
}

// ProjectPath builds path column Path from bound variables and segment columns.
type ProjectPath struct {
	Input Node
	Path  string
	Elems []PathElem
}

// AggItem is one aggregation; a nil Arg is count(*).
type AggItem struct {
	Name     string
	Fn       string
	Distinct bool
	Arg      Expr
}

// Aggregation groups by Keys and computes Aggs (implicit GROUP BY).
type Aggregation struct {
	Input Node
	Keys  []ProjItem
	Aggs  []AggItem
	// Partial marks a partial aggregation made by SplitAggregation (a final aggregation above combines the
	// partial results). It may run per input chunk, so a shipped boundary (ir.ShipAggregate) can absorb it.
	Partial bool
}

// SortKey is a sort key.
type SortKey struct {
	Expr Expr
	Desc bool
}

// Sort orders rows by Keys.
type Sort struct {
	Input Node
	Keys  []SortKey
}

// Limit skips Skip rows and keeps Count rows; Count < 0 is unlimited.
type Limit struct {
	Input Node
	Skip  int
	Count int
}

// Distinct removes duplicates over Cols (all columns if empty).
type Distinct struct {
	Input Node
	Cols  []string
}

// NodeHashJoin joins two inputs on id columns Keys; no Keys is a cross product.
type NodeHashJoin struct {
	Left  Node
	Right Node
	Keys  []string
}

// Argument is the leaf inside an Optional: it receives each outer row's id columns Cols (bound variables).
// Labels gives node label candidates for columns the plan cannot tell (opbench inputs, for example).
type Argument struct {
	Cols   []string
	Labels map[string][]string
}

// Optional is OPTIONAL MATCH (a left outer apply). For each outer row, the matches of Inner (a subtree
// starting at Argument) are joined on id columns Keys. Outer rows without a match stay, with the variables
// Inner binds (New) and the columns it makes set to null.
type Optional struct {
	Outer Node
	Inner Node
	Keys  []string
	New   []string
}

// Unwind makes one row per element of list expression List and puts the element in column Var (an empty
// list or null keeps no row). Unwinding a list of ids (collected bindings) makes Var an id column.
type Unwind struct {
	Input Node
	List  Expr
	Var   string
}

// RelUniqueFilter keeps rows whose relationships in one MATCH (id columns Rels and segment columns Segs) are all distinct.
type RelUniqueFilter struct {
	Input Node
	Rels  []string
	Segs  []string
}

// ---- Children ----

func (*EntityScan) Children() []Node        { return nil }
func (*Argument) Children() []Node          { return nil }
func (n *Optional) Children() []Node        { return []Node{n.Outer, n.Inner} }
func (n *GetProperty) Children() []Node     { return []Node{n.Input} }
func (n *GetListProperty) Children() []Node { return []Node{n.Input} }
func (n *Expand) Children() []Node          { return []Node{n.Input} }
func (n *ExpandInto) Children() []Node      { return []Node{n.Input} }
func (n *VarLengthExpand) Children() []Node { return []Node{n.Input} }
func (n *Filter) Children() []Node          { return []Node{n.Input} }
func (n *Projection) Children() []Node      { return []Node{n.Input} }
func (n *ProjectPath) Children() []Node     { return []Node{n.Input} }
func (n *Aggregation) Children() []Node     { return []Node{n.Input} }
func (n *Sort) Children() []Node            { return []Node{n.Input} }
func (n *Limit) Children() []Node           { return []Node{n.Input} }
func (n *Distinct) Children() []Node        { return []Node{n.Input} }
func (n *NodeHashJoin) Children() []Node    { return []Node{n.Left, n.Right} }
func (n *RelUniqueFilter) Children() []Node { return []Node{n.Input} }
func (n *Unwind) Children() []Node          { return []Node{n.Input} }

// ---- Describe ----

func labelSuffix(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	return ":" + strings.Join(ls, "|")
}

func (n *Argument) Describe() string {
	return fmt.Sprintf("Argument(%s)", strings.Join(n.Cols, ", "))
}

func (n *Optional) Describe() string {
	return fmt.Sprintf("Optional[keys: %s | new: %s]", strings.Join(n.Keys, ", "), strings.Join(n.New, ", "))
}

func (n *EntityScan) Describe() string {
	return fmt.Sprintf("EntityScan(%s%s)", n.Alias, labelSuffix(n.Labels))
}

func (n *GetProperty) Describe() string {
	return fmt.Sprintf("GetProperty(%s) [%s%s]", n.Key(), n.Kind, labelSuffix(n.Labels))
}

func (n *GetListProperty) Describe() string {
	return fmt.Sprintf("GetListProperty(%s) [elem%s]", n.Key(), labelSuffix(n.ElemLabels))
}

func (n *Expand) Describe() string {
	return fmt.Sprintf("Expand((%s)%s(%s%s))", n.From, n.Dir.Pattern(n.Rel+labelSuffix(n.Types)), n.To, labelSuffix(n.ToLabels))
}

func (n *ExpandInto) Describe() string {
	return fmt.Sprintf("ExpandInto((%s)%s(%s))", n.From, n.Dir.Pattern(n.Rel+labelSuffix(n.Types)), n.To)
}

func (n *VarLengthExpand) Describe() string {
	s := fmt.Sprintf("VarLengthExpand((%s)%s(%s%s))", n.From,
		n.Dir.Pattern(n.Seg+labelSuffix(n.Types)+n.Range.Cypher()), n.To, labelSuffix(n.ToLabels))
	if len(n.SegPreds) > 0 {
		parts := make([]string, len(n.SegPreds))
		for i, p := range n.SegPreds {
			parts[i] = p.String()
		}
		s += " seg[" + strings.Join(parts, " AND ") + "]"
	}
	if len(n.InnerLabels) > 0 {
		s += " inner:" + strings.Join(n.InnerLabels, ":")
	}
	if n.EmitPath {
		s += " emit " + SegColumn(n.Seg)
	}
	return s
}

func (n *Filter) Describe() string { return fmt.Sprintf("Filter(%s)", n.Pred) }

func projItems(items []ProjItem) string {
	parts := make([]string, len(items))
	for i, it := range items {
		if c, ok := it.Expr.(*Column); ok && c.Key == it.Name {
			parts[i] = it.Name
		} else {
			parts[i] = fmt.Sprintf("%s AS %s", it.Expr, it.Name)
		}
	}
	return strings.Join(parts, ", ")
}

func (n *Projection) Describe() string { return fmt.Sprintf("Projection[%s]", projItems(n.Items)) }

func (n *ProjectPath) Describe() string {
	parts := make([]string, len(n.Elems))
	for i, e := range n.Elems {
		if e.Seg {
			parts[i] = SegColumn(e.Var)
		} else {
			parts[i] = e.Var
		}
	}
	return fmt.Sprintf("ProjectPath(%s = [%s])", n.Path, strings.Join(parts, ", "))
}

func (a AggItem) String() string {
	arg := "*"
	if a.Arg != nil {
		arg = a.Arg.String()
	}
	if a.Distinct {
		arg = "DISTINCT " + arg
	}
	return fmt.Sprintf("%s(%s) AS %s", a.Fn, arg, a.Name)
}

func (n *Aggregation) Describe() string {
	keys := "∅"
	if len(n.Keys) > 0 {
		keys = projItems(n.Keys)
	}
	aggs := make([]string, len(n.Aggs))
	for i, a := range n.Aggs {
		aggs[i] = a.String()
	}
	kind := "Aggregation"
	if n.Partial {
		kind = "PartialAggregation"
	}
	return fmt.Sprintf("%s[keys: %s | %s]", kind, keys, strings.Join(aggs, ", "))
}

func (n *Sort) Describe() string {
	parts := make([]string, len(n.Keys))
	for i, k := range n.Keys {
		dir := "ASC"
		if k.Desc {
			dir = "DESC"
		}
		parts[i] = k.Expr.String() + " " + dir
	}
	return fmt.Sprintf("Sort[%s]", strings.Join(parts, ", "))
}

func (n *Limit) Describe() string {
	if n.Skip > 0 {
		return fmt.Sprintf("Limit[skip %d, %d]", n.Skip, n.Count)
	}
	return fmt.Sprintf("Limit[%d]", n.Count)
}

func (n *Distinct) Describe() string {
	if len(n.Cols) == 0 {
		return "Distinct"
	}
	return fmt.Sprintf("Distinct[%s]", strings.Join(n.Cols, ", "))
}

func (n *NodeHashJoin) Describe() string {
	if len(n.Keys) == 0 {
		return "NodeHashJoin[cartesian]"
	}
	return fmt.Sprintf("NodeHashJoin[%s]", strings.Join(n.Keys, ", "))
}

func (n *Unwind) Describe() string { return fmt.Sprintf("Unwind(%s AS %s)", n.List, n.Var) }

func (n *RelUniqueFilter) Describe() string {
	parts := append([]string{}, n.Rels...)
	for _, s := range n.Segs {
		parts = append(parts, SegColumn(s))
	}
	return fmt.Sprintf("RelUniqueFilter[%s]", strings.Join(parts, ", "))
}

// Dump prints a plan tree deterministically (one operator per line, children indented).
func Dump(n Node) string {
	var b strings.Builder
	var walk func(Node, int)
	walk = func(n Node, d int) {
		if n == nil {
			return
		}
		b.WriteString(strings.Repeat("  ", d))
		b.WriteString(n.Describe())
		b.WriteByte('\n')
		for _, c := range n.Children() {
			walk(c, d+1)
		}
	}
	walk(n, 0)
	return b.String()
}
