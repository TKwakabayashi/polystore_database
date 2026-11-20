package physical

import (
	"fmt"
	"polystore_database/src/go/query/ir"
	"strings"

	"polystore_database/src/go/base/kind"
)

// Node is a physical operator.
type Node interface {
	Children() []Node
	Describe() string
}

// StoreQuery is a query delegated to one store (ir.StoreQuery) plus its input in the physical plan and the
// fields of hand-built experiment plans.
type StoreQuery struct {
	ir.StoreQuery
	Input Node

	// With ir.ShipHand, the hand-built native query (*stores.Query). HandChunk is the rows per query (0: one),
	// HandGroup a column whose equal values stay in one chunk.
	Prelowered any
	// TextRewrite rewrites the native query text after compilation (hand-built plans).
	TextRewrite func(string) string
	HandChunk   int
	HandGroup   string
	// HandScan marks a hand-built query that appends its own (input-free leaf) results to the input rows.
	HandScan bool
}

// Leaf reports whether the query has no input (it runs within the store).
func (q *StoreQuery) Leaf() bool { return q.Input == nil }

func (q *StoreQuery) Children() []Node {
	if q.Input == nil {
		return nil
	}
	return []Node{q.Input}
}

func (q *StoreQuery) Describe() string {
	kind := "StoreQuery"
	if q.IsFetch() {
		kind = "Fetch"
	}
	keys := ""
	if len(q.Keys) > 0 {
		keys = " in=(" + strings.Join(q.Keys, ", ") + ")"
	}
	w := ""
	if q.Weighted {
		w = " weighted"
	}
	if q.Replace {
		w += " replace"
	}
	if q.ShipMode != ir.ShipNone {
		w += " mode=" + q.ShipMode.String()
		if len(q.Ship) > 0 {
			w += " ship=(" + strings.Join(q.Ship, ", ") + ")"
		}
	}
	return fmt.Sprintf("%s[%s]%s out=(%s)%s", kind, q.Store, keys, strings.Join(q.Outputs, ", "), w)
}

// ListFetch reads Prop for the elements of list column Source (the union over all rows) and makes value list column Key().
type ListFetch struct {
	Input  Node
	Store  kind.Kind
	Source string
	Part   ir.SegPart
	Prop   string
	Labels []string
}

// Key is the key of the list value column it makes.
func (f *ListFetch) Key() string { return ir.ListPropKey(f.Source, f.Part, f.Prop) }

func (f *ListFetch) Children() []Node { return []Node{f.Input} }
func (f *ListFetch) Describe() string {
	return fmt.Sprintf("ListFetch[%s] %s from %s [elem:%s]", f.Store, f.Key(), f.Source, strings.Join(f.Labels, "|"))
}

// ArgRows is the inner leaf of an OptionalApply; at run time it receives the outer rows (the distinct tuples of Cols).
type ArgRows struct {
	Cols []string
}

func (*ArgRows) Children() []Node   { return nil }
func (a *ArgRows) Describe() string { return fmt.Sprintf("ArgRows(%s)", strings.Join(a.Cols, ", ")) }

// OptionalApply is OPTIONAL MATCH as a left outer apply: evaluate Outer, pass its id columns Keys to Inner's
// leaf (ArgRows), evaluate Inner and join on Keys. Outer rows without a match stay with New and Inner's columns null.
type OptionalApply struct {
	Outer Node
	Inner Node
	Arg   *ArgRows
	Keys  []string
	New   []string
	// InnerCols are the columns Inner makes (null for rows without a match); set by finalize.
	InnerCols []string
}

func (o *OptionalApply) Children() []Node { return []Node{o.Outer, o.Inner} }
func (o *OptionalApply) Describe() string {
	return fmt.Sprintf("OptionalApply[keys: %s | new: %s]", strings.Join(o.Keys, ", "), strings.Join(o.New, ", "))
}

// Coord is one logical operator run by the coordinator; Op's input is replaced by an ir.Boundary.
type Coord struct {
	Op     ir.Node
	Inputs []Node
}

func (c *Coord) Children() []Node { return c.Inputs }
func (c *Coord) Describe() string { return "Coord " + c.Op.Describe() }

// SetPlan computes id set Name before the main plan runs: the ids in column Col of the rows of Plans (one plan
// per element label; the set is their union). The main plan reads it through ir.InSet.
type SetPlan struct {
	Name  string
	Col   string
	Plans []Node
}

// WithSets runs its set plans, then Input (whose store queries and coordinator operators read the sets).
// It is only at the root. The iteration models never see it: the executor runs the sets and then Input.
type WithSets struct {
	Sets  []*SetPlan
	Input Node
}

// Children lists the set plans first (they run first), then Input.
func (w *WithSets) Children() []Node {
	var out []Node
	for _, s := range w.Sets {
		out = append(out, s.Plans...)
	}
	return append(out, w.Input)
}

func (w *WithSets) Describe() string {
	names := make([]string, len(w.Sets))
	for i, s := range w.Sets {
		names[i] = "$" + s.Name
	}
	return "WithSets[" + strings.Join(names, ", ") + "]"
}

// StepExpand is a variable-length expansion run hop by hop by the coordinator (PathStepwise). For each depth it
// expands the distinct last nodes of the partial paths with Hop (one Neo4j Expand from column StepFrom), checks
// the segment conditions on the reached relationships and nodes, and extends only the partial paths whose new
// relationship passes (relationships are not repeated within a path) and whose new node passes (it becomes an
// inner node). Rows are emitted at every depth in Range, binding To and, with EmitSeg, the segment column
// (inner nodes and relationships, as Neo4j returns it). With Weighted, rows with the same input row and To
// collapse into one whose weight is the number of paths.
type StepExpand struct {
	Input Node
	Op    *ir.VarLengthExpand // From, Seg, To, Types, Dir, Range and the conditions (ElemProp form, for Describe)
	// Preds are the segment conditions on columns StepRel+"."+prop / StepTo+"."+prop.
	Preds []StepPred
	// Hop expands the frontier: Boundary(StepFrom) -> Expand((StepFrom)-[StepRel]-(StepTo)) -> the Neo4j values the conditions read.
	Hop *StoreQuery
	// Fetches read the values outside Neo4j (Source names the element column, StepRel or StepTo).
	Fetches  []*ListFetch
	EmitSeg  bool
	Weighted bool
}

// StepPred is one segment condition of a StepExpand, on the relationship (SegRels) or node (SegNodes) of a hop.
type StepPred struct {
	Part ir.SegPart
	Pred ir.Expr
}

// Columns of a StepExpand hop.
const (
	StepFrom = "_hf"
	StepRel  = "_hr"
	StepTo   = "_ht"
)

func (s *StepExpand) Children() []Node { return []Node{s.Input} }
func (s *StepExpand) Describe() string {
	d := "StepExpand" + strings.TrimPrefix(s.Op.Describe(), "VarLengthExpand")
	d = strings.Replace(d, " emit "+ir.SegColumn(s.Op.Seg), "", 1)
	if s.EmitSeg {
		d += " emit " + ir.SegColumn(s.Op.Seg)
	}
	if s.Weighted {
		d += " weighted"
	}
	return d
}

// Dump prints a physical plan deterministically; a StoreQuery's delegated subtree is expanded with `: `.
func Dump(n Node) string {
	var b strings.Builder
	var walk func(Node, int)
	walk = func(n Node, d int) {
		if n == nil {
			return
		}
		ind := strings.Repeat("  ", d)
		if w, ok := n.(*WithSets); ok {
			b.WriteString(ind + w.Describe() + "\n")
			for _, s := range w.Sets {
				b.WriteString(ind + "  set $" + s.Name + " = " + s.Col + " of\n")
				for _, p := range s.Plans {
					walk(p, d+2)
				}
			}
			walk(w.Input, d+1)
			return
		}
		b.WriteString(ind + n.Describe() + "\n")
		if q, ok := n.(*StoreQuery); ok {
			for _, line := range strings.Split(strings.TrimRight(ir.Dump(q.Ops), "\n"), "\n") {
				b.WriteString(ind + "  : " + line + "\n")
			}
		}
		if s, ok := n.(*StepExpand); ok {
			for _, p := range s.Preds {
				b.WriteString(ind + "  check " + p.Part.String() + ": " + p.Pred.String() + "\n")
			}
			b.WriteString(ind + "  hop\n")
			walk(s.Hop, d+2)
			for _, f := range s.Fetches {
				b.WriteString(ind + "  fetch " + f.Describe() + "\n")
			}
		}
		for _, c := range n.Children() {
			walk(c, d+1)
		}
	}
	walk(n, 0)
	return b.String()
}
