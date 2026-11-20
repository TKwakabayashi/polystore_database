package ir

import (
	"fmt"
	"slices"
	"strings"

	"polystore_database/src/go/base/kind"
)

// Boundary marks the leaf of a delegated subtree (StoreQuery.Ops) or of a coordinator operator: the rows
// received from the boundary (the physical operator below). Cols are the id columns received (the IN $ids keys).
type Boundary struct {
	Cols []string
	// Labels maps a column to its label (node) or type (relationship) candidates; Neo4j uses them for the uuid index (no candidates: match without a label).
	Labels map[string][]string

	// Columns the leaf receives on a shipped boundary (ShipMode other than none); compilers build the table from them.
	Ship   []string // input id columns shipped with Cols
	Rid    bool     // receives row number RidCol (ShipFilter / ShipProject)
	Unnest string   // list column received unnested (a path; positions in PosCol). Unused for now
}

func (*Boundary) Children() []Node { return nil }
func (a *Boundary) Describe() string {
	s := fmt.Sprintf("Argument(%s", strings.Join(a.Cols, ", "))
	if len(a.Ship) > 0 {
		s += "; ship " + strings.Join(a.Ship, ", ")
	}
	if a.Rid {
		s += "; " + RidCol
	}
	if a.Unnest != "" {
		s += "; unnest " + a.Unnest
	}
	return s + ")"
}

// ShipMode is how a boundary is passed and how the store's result is used.
type ShipMode int

const (
	// ShipNone passes the id tuples (Keys) with IN / UNWIND and joins the result back to the input rows on the keys.
	ShipNone ShipMode = iota
	// ShipAggregate passes a table of (Keys, Ship) tuples with summed weights; the store returns partially
	// aggregated rows as they are (not joined back; the final aggregation above combines the chunks).
	ShipAggregate
	// ShipFilter passes (RidCol, Keys, Ship) and gets back the row numbers that match (a semi-join). Compile-only.
	ShipFilter
	// ShipProject passes (RidCol, Keys, Ship[, PosCol]) and gets values per row number (joined back). Compile-only.
	ShipProject
	// ShipHand is a hand-built experiment plan: the input rows go as a table to a prelowered native query whose result is passed on as is.
	ShipHand
)

var shipModeNames = [...]string{"none", "aggregate", "filter", "project", "hand"}

func (m ShipMode) String() string {
	if int(m) < len(shipModeNames) {
		return shipModeNames[m]
	}
	return fmt.Sprintf("ShipMode(%d)", int(m))
}

// Internal columns of shipped boundaries.
const (
	RidCol = "_rid" // input row number
	PosCol = "_pos" // position of an unnested list element
)

// StoreQuery is one query delegated to one store (the only type a store receives).
// - Ops:     the delegated subtree. Its leaf is an EntityScan (no input) or a Boundary (input Keys via IN $ids).
// - Keys:    the id columns that match results to input rows; results are joined back on them (input duplicates kept).
// - Outputs: the columns the store returns (set by the planner's requirement analysis).
type StoreQuery struct {
	Store   kind.Kind
	Ops     Node
	Keys    []string
	Outputs []string

	// Ship are input id columns passed with Keys; ShipMode is how they are passed and how the result is used (ShipNone: plain IN).
	Ship     []string
	ShipMode ShipMode

	// Replace passes the result on instead of joining it to the input rows (a query that absorbed a `WITH DISTINCT` dropping the input columns).
	Replace bool

	// Weighted collapses equal output rows into one with their count (weight) in column WeightCol. A
	// variable-length expansion makes one row per path; when the segment is not returned those rows repeat, so
	// transfer and intermediate rows shrink from paths to distinct rows. The weight keeps the meaning (the executor handles it).
	Weighted bool
}

// WeightCol holds a row's weight (how many original rows it stands for); 1 when absent.
const WeightCol = "_w"

// IsFetch reports whether Ops is only GetProperty over a Boundary (a value fetch).
func (q *StoreQuery) IsFetch() bool {
	n := q.Ops
	seen := false
	for {
		switch x := n.(type) {
		case *GetProperty:
			seen = true
			n = x.Input
			continue
		case *Boundary:
			return seen
		}
		return false
	}
}

// ShipInner splits a shipped aggregation query (ShipAggregate) into the id-boundary query under the
// aggregation (fetches, filters, traversals) and the partial aggregation above it. The executor uses it to
// emulate shipping on stores that cannot join a table (issue the inner query with IN, join to the table rows, aggregate).
func (q *StoreQuery) ShipInner() (*StoreQuery, *Aggregation, bool) {
	agg, ok := q.Ops.(*Aggregation)
	if !ok || q.ShipMode != ShipAggregate {
		return nil, nil, false
	}
	ops := stripShip(agg.Input)
	inner := &StoreQuery{Store: q.Store, Ops: ops, Keys: q.Keys}
	inner.Outputs = append(inner.Outputs, q.Keys...)
	for _, c := range TopCols(ops) {
		if !slices.Contains(inner.Outputs, c) {
			inner.Outputs = append(inner.Outputs, c)
		}
	}
	return inner, agg, true
}

// stripShip copies a chain of unary operators with the shipped columns removed from its leaf Boundary.
func stripShip(n Node) Node {
	if a, ok := n.(*Boundary); ok {
		cp := *a
		cp.Ship, cp.Rid, cp.Unnest = nil, false, ""
		return &cp
	}
	c := Child(n)
	if c == nil {
		return n
	}
	return WithInput(n, stripShip(c))
}

// TopCols returns the columns of the subtree's top rows, in order (Projection / Aggregation replace them).
func TopCols(n Node) []string {
	var cols []string
	add := func(ks ...string) {
		for _, k := range ks {
			if !slices.Contains(cols, k) {
				cols = append(cols, k)
			}
		}
	}
	var walk func(Node)
	walk = func(n Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *NodeHashJoin:
			walk(x.Left)
			walk(x.Right)
			return
		case *Optional:
			walk(x.Outer)
			walk(x.Inner)
			return
		case *Boundary:
			add(x.Cols...)
			return
		case *EntityScan:
			add(x.Alias)
			return
		}
		walk(Child(n))
		switch x := n.(type) {
		case *GetProperty:
			add(x.Key())
		case *GetListProperty:
			add(x.Key())
		case *Expand:
			add(x.Rel, x.To)
		case *ExpandInto:
			add(x.Rel)
		case *VarLengthExpand:
			add(x.To, SegColumn(x.Seg))
		case *ProjectPath:
			add(x.Path)
		case *Projection:
			cols = nil
			for _, it := range x.Items {
				add(it.Name)
			}
		case *Aggregation:
			cols = nil
			for _, it := range x.Keys {
				add(it.Name)
			}
			for _, a := range x.Aggs {
				add(a.Name)
			}
		}
	}
	walk(n)
	return cols
}
