package lower

import (
	"fmt"

	"polystore_database/src/go/query/ir"
)

// ShipTable is the table's columns (order of ShipCols) and the column key -> table column name map.
type ShipTable struct {
	Cols  []string          // ShipCols (plan column keys: Keys, Ship, WeightCol / RidCol / PosCol)
	Names map[string]string // column key -> table column name
}

func NewShipTable(q *ir.StoreQuery, arg *ir.Boundary) (*ShipTable, error) {
	t := &ShipTable{Names: map[string]string{}}
	add := func(col, name string) {
		t.Cols = append(t.Cols, col)
		t.Names[col] = name
	}
	if arg.Rid {
		add(ir.RidCol, "rid")
	}
	for i, k := range q.Keys {
		add(k, fmt.Sprintf("k%d", i))
	}
	for i, s := range arg.Ship {
		add(s, fmt.Sprintf("s%d", i))
	}
	if arg.Unnest != "" {
		add(ir.PosCol, "pos")
	}
	if q.ShipMode == ir.ShipAggregate {
		add(ir.WeightCol, "w")
	}
	return t, nil
}

// ColNames returns the table column names (order of cols).
func (t *ShipTable) ColNames() []string {
	out := make([]string, len(t.Cols))
	for i, c := range t.Cols {
		out[i] = t.Names[c]
	}
	return out
}

// IsIdent reports whether a table column is an id column (not a row number, position or weight).
func (t *ShipTable) IsIdent(col string) bool {
	switch col {
	case ir.RidCol, ir.PosCol, ir.WeightCol:
		return false
	}
	_, ok := t.Names[col]
	return ok
}

// ShipParts splits a shipped query's subtree: the leaf boundary, the alias holding values with its table, and the operators above.
func ShipParts(q *ir.StoreQuery) (*ir.Boundary, string, Owner, []ir.Node, error) {
	ops := Chain(q.Ops)
	if len(ops) == 0 {
		return nil, "", Owner{}, nil, Unsupported("empty shipped query")
	}
	arg, ok := ops[0].(*ir.Boundary)
	if !ok || len(arg.Cols) != 1 {
		return nil, "", Owner{}, nil, Unsupported("shipped query must start from one argument column")
	}
	single := arg.Cols[0]
	o, ok := Owners(q.Ops)[single]
	if !ok || len(o.Labels) != 1 {
		return nil, "", Owner{}, nil, Unsupported("table of shipped column %s", single)
	}
	return arg, single, o, ops[1:], nil
}

// ShipLeaf returns the boundary at the leaf of a chain of unary operators (nil if none).
func ShipLeaf(n ir.Node) *ir.Boundary {
	for n != nil {
		if a, ok := n.(*ir.Boundary); ok {
			return a
		}
		n = ir.Child(n)
	}
	return nil
}

// HasTraversal reports whether the subtree traverses (otherwise it only fetches values).
func HasTraversal(n ir.Node) bool {
	found := false
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil || found {
			return
		}
		switch n.(type) {
		case *ir.Expand, *ir.ExpandInto, *ir.VarLengthExpand, *ir.EntityScan:
			found = true
			return
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return found
}
