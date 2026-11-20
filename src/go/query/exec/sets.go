package exec

// sets.go: id sets (physical.WithSets). The set plans run first (in bulk); store queries and coordinator
// operators that read a set (ir.InSet) get a copy of their operators with the set's value bound.

import (
	"fmt"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
)

// unwrapSets runs the set plans of a WithSets root and returns the plan to run next (n itself otherwise).
func (e *Executor) unwrapSets(n physical.Node) (physical.Node, error) {
	w, ok := n.(*physical.WithSets)
	if !ok {
		return n, nil
	}
	for _, s := range w.Sets {
		var ids []id.UUID
		for _, p := range s.Plans {
			rows, err := e.bulk(p)
			if err != nil {
				return nil, fmt.Errorf("set %s: %w", s.Name, err)
			}
			for _, r := range rows {
				if u, ok := r[s.Col].(id.UUID); ok && !u.Empty() {
					ids = append(ids, u)
				}
			}
		}
		e.mu.Lock()
		e.sets[s.Name] = ir.NewIDSet(ids)
		e.mu.Unlock()
	}
	return w.Input, nil
}

// bindExpr returns e with the values of the sets it reads (e itself when it reads none).
func (e *Executor) bindExpr(x ir.Expr) (ir.Expr, error) {
	if !ir.HasSets(x) {
		return x, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	var missing string
	out := ir.Transform(x, func(y ir.Expr) ir.Expr {
		s, ok := y.(*ir.InSet)
		if !ok {
			return y
		}
		v, ok := e.sets[s.Set]
		if !ok {
			missing = s.Set
			return y
		}
		cp := *s
		cp.Bound = v
		return &cp
	})
	if missing != "" {
		return nil, fmt.Errorf("executor: id set %s was not computed", missing)
	}
	return out, nil
}

// bindNode returns a copy of the operator tree with the sets bound in Filter and segment conditions (n itself when none are read).
func (e *Executor) bindNode(n ir.Node) (ir.Node, error) {
	if !readsSets(n) {
		return n, nil
	}
	switch x := n.(type) {
	case *ir.NodeHashJoin:
		l, err := e.bindNode(x.Left)
		if err != nil {
			return nil, err
		}
		r, err := e.bindNode(x.Right)
		if err != nil {
			return nil, err
		}
		y := *x
		y.Left, y.Right = l, r
		return &y, nil
	case *ir.Optional:
		o, err := e.bindNode(x.Outer)
		if err != nil {
			return nil, err
		}
		in, err := e.bindNode(x.Inner)
		if err != nil {
			return nil, err
		}
		y := *x
		y.Outer, y.Inner = o, in
		return &y, nil
	}
	cp := n
	if c := ir.Child(n); c != nil {
		bc, err := e.bindNode(c)
		if err != nil {
			return nil, err
		}
		cp = ir.WithInput(n, bc) // a copy
	}
	switch y := cp.(type) {
	case *ir.Filter:
		p, err := e.bindExpr(y.Pred)
		if err != nil {
			return nil, err
		}
		y.Pred = p
	case *ir.VarLengthExpand:
		preds := make([]ir.SegPred, len(y.SegPreds))
		for i, sp := range y.SegPreds {
			p, err := e.bindExpr(sp.Pred)
			if err != nil {
				return nil, err
			}
			preds[i] = ir.SegPred{Part: sp.Part, Var: sp.Var, Pred: p}
		}
		y.SegPreds = preds
	}
	return cp, nil
}

// readsSets reports whether a Filter or segment condition in the tree reads an id set.
func readsSets(n ir.Node) bool {
	if n == nil {
		return false
	}
	switch x := n.(type) {
	case *ir.Filter:
		if ir.HasSets(x.Pred) {
			return true
		}
	case *ir.VarLengthExpand:
		for _, sp := range x.SegPreds {
			if ir.HasSets(sp.Pred) {
				return true
			}
		}
	}
	for _, c := range n.Children() {
		if readsSets(c) {
			return true
		}
	}
	return false
}

// coordOp returns a coordinator operator with its sets bound (bound once per run).
func (e *Executor) coordOp(c *physical.Coord) (ir.Node, error) {
	if !readsSets(c.Op) {
		return c.Op, nil
	}
	e.mu.Lock()
	op, ok := e.boundOps[c]
	e.mu.Unlock()
	if ok {
		return op, nil
	}
	op, err := e.bindNode(c.Op)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.boundOps[c] = op
	e.mu.Unlock()
	return op, nil
}
