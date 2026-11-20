package physical

// graph.go places traversals on Neo4j and list fetches on the store of their element values.

import (
	"polystore_database/src/go/query/ir"

	"polystore_database/src/go/base/kind"
)

// addGraph adds a Neo4j operator (a traversal), merged into the open Neo4j query when it has the needed id columns.
func (pl *planner) addGraph(st *state, op ir.Node, need []string) error {
	o := st.open
	if o != nil && o.store == kind.Neo4j && identAll(o, need) {
		pl.extend(st, op)
		return nil
	}
	pl.boundary(st, kind.Neo4j, op, need, "")
	return nil
}

// addVarLength adds a variable-length expansion. Segment conditions on non-Neo4j columns make the query
// return the segment, the element values are fetched, and the coordinator evaluates the condition
// (PathPost), or the expansion runs hop by hop with the conditions checked at each hop (PathStepwise, StepExpand).
func (pl *planner) addVarLength(st *state, x *ir.VarLengthExpand) error {
	if pl.opt.PathPredicates == PathStepwise {
		if se, ok := pl.stepExpand(x); ok {
			pl.close(st)
			se.Input = st.phys
			st.phys = se
			return nil
		}
	}
	var keep, moved []ir.SegPred
	for _, sp := range x.SegPreds {
		labels := pl.cat.ElemLabels(x.Types, sp.Part)
		inGraph := true
		for _, prop := range elemProps(sp.Pred, sp.Var) {
			k, err := pl.listStore(labels, sp.Part, prop)
			if err != nil {
				return err
			}
			if k != kind.Neo4j {
				inGraph = false
			}
		}
		if inGraph {
			keep = append(keep, sp)
		} else {
			moved = append(moved, sp)
		}
	}
	y := *x
	y.SegPreds = keep
	if err := pl.addGraph(st, &y, []string{x.From}); err != nil {
		return err
	}
	seg := ir.SegColumn(x.Seg)
	for _, sp := range moved {
		labels := pl.cat.ElemLabels(x.Types, sp.Part)
		for _, prop := range elemProps(sp.Pred, sp.Var) {
			if err := pl.add(st, &ir.GetListProperty{Input: nil, Source: seg, Part: sp.Part, Prop: prop, ElemLabels: labels}); err != nil {
				return err
			}
		}
		fn := "nodes"
		if sp.Part == ir.SegRels {
			fn = "relationships"
		}
		pred := &ir.Quant{Kind: "all", Var: sp.Var,
			List: &ir.Call{Fn: fn, Args: []ir.Expr{&ir.Column{Key: seg}}}, Pred: sp.Pred}
		if err := pl.add(st, &ir.Filter{Pred: pred}); err != nil {
			return err
		}
	}
	return nil
}

// stepExpand plans x as a StepExpand when a segment condition reads values outside Neo4j and the shape allows
// it: typed relationships, a bounded range from at least 1, and every node the types reach having one of the
// To labels (the hop does not tell labels, so every reached node must qualify as an end).
func (pl *planner) stepExpand(x *ir.VarLengthExpand) (*StepExpand, bool) {
	if x.Range.Min < 1 || x.Range.Max < 1 || len(x.Types) == 0 || len(x.SegPreds) == 0 {
		return nil, false
	}
	inner := pl.cat.ElemLabels(x.Types, ir.SegNodes)
	if len(inner) == 0 || !pl.cat.HaveLabels(x.Types, x.InnerLabels) {
		return nil, false
	}
	if len(x.ToLabels) > 0 {
		for _, l := range inner {
			if !contains(x.ToLabels, l) {
				return nil, false
			}
		}
	}
	se := &StepExpand{Op: x}
	var gets []*ir.GetProperty
	got := map[string]bool{}
	fetched := map[string]bool{}
	off := false
	for _, sp := range x.SegPreds {
		col, obj, labels := StepTo, ir.NodeObj, pl.cat.Entities(inner)
		if sp.Part == ir.SegRels {
			col, obj, labels = StepRel, ir.RelObj, x.Types
		}
		se.Preds = append(se.Preds, StepPred{Part: sp.Part, Pred: ir.Transform(sp.Pred, elemToColumn(sp.Var, col))})
		for _, prop := range elemProps(sp.Pred, sp.Var) {
			byStore := map[kind.Kind][]string{}
			var stores []kind.Kind
			for _, l := range labels {
				k := pl.cat.StoreOf("", obj, []string{l}, prop)
				if _, ok := byStore[k]; !ok {
					stores = append(stores, k)
				}
				byStore[k] = append(byStore[k], l)
			}
			for _, k := range stores {
				key := col + "." + prop
				if k == kind.Neo4j {
					if !got[key] {
						got[key] = true
						gets = append(gets, &ir.GetProperty{Alias: col, Prop: prop, Kind: obj, Labels: byStore[k]})
					}
					continue
				}
				off = true
				if fk := key + "@" + k.String(); !fetched[fk] {
					fetched[fk] = true
					se.Fetches = append(se.Fetches, &ListFetch{Store: k, Source: col, Part: sp.Part, Prop: prop, Labels: byStore[k]})
				}
			}
		}
	}
	if !off {
		return nil, false
	}
	var frontier []string
	if from := pl.nodeLabels[x.From]; len(from) > 0 {
		frontier = append(frontier, from...)
		for _, l := range inner {
			if !contains(frontier, l) {
				frontier = append(frontier, l)
			}
		}
	}
	var ops ir.Node = &ir.Boundary{Cols: []string{StepFrom}, Labels: map[string][]string{StepFrom: frontier}}
	if len(frontier) == 0 {
		ops = &ir.Boundary{Cols: []string{StepFrom}} // the start's labels are unknown: match without a label
	}
	ops = &ir.Expand{Input: ops, From: StepFrom, Rel: StepRel, To: StepTo, Types: x.Types, Dir: x.Dir}
	outputs := []string{StepFrom, StepRel, StepTo}
	for _, g := range gets {
		g.Input = ops
		ops = g
		outputs = append(outputs, g.Key())
	}
	se.Hop = &StoreQuery{StoreQuery: ir.StoreQuery{Store: kind.Neo4j, Ops: ops, Keys: []string{StepFrom}, Outputs: outputs}}
	return se, true
}

// listStore returns the store of a list element property. When elements may have several labels, all
// must use the same store (a ListFetch reads one store); otherwise it is unsupported.
func (pl *planner) listStore(labels []string, part ir.SegPart, prop string) (kind.Kind, error) {
	obj := ir.NodeObj
	if part == ir.SegRels {
		obj = ir.RelObj
	}
	if len(labels) == 0 {
		return kind.Neo4j, nil
	}
	first := pl.cat.StoreOf("", obj, labels[:1], prop)
	for _, l := range labels[1:] {
		if k := pl.cat.StoreOf("", obj, []string{l}, prop); k != first {
			return 0, unsupported("list elements of %v keep %s in different stores", labels, prop)
		}
	}
	return first, nil
}

// listLabels returns the label candidates of the elements of a list column (path or segment column).
func (pl *planner) listLabels(source string, part ir.SegPart) []string {
	var out []string
	add := func(ls []string) {
		for _, l := range ls {
			if !contains(out, l) {
				out = append(out, l)
			}
		}
	}
	addSeg := func(seg string) { add(pl.cat.ElemLabels(pl.segTypes[seg], part)) }
	if elems, ok := pl.pathElems[source]; ok {
		for _, e := range elems {
			switch {
			case e.Seg:
				addSeg(e.Var)
			case part == ir.SegNodes && pl.nodeLabels[e.Var] != nil:
				add(pl.nodeLabels[e.Var])
			case part == ir.SegRels && pl.relTypes[e.Var] != nil:
				add(pl.relTypes[e.Var])
			}
		}
		return out
	}
	for seg := range pl.segTypes {
		if ir.SegColumn(seg) == source {
			addSeg(seg)
		}
	}
	return out
}

func elemProps(e ir.Expr, v string) []string {
	var out []string
	seen := map[string]bool{}
	ir.Walk(e, func(x ir.Expr) {
		if ep, ok := x.(*ir.ElemProp); ok && ep.Var == v && !seen[ep.Prop] {
			seen[ep.Prop] = true
			out = append(out, ep.Prop)
		}
	})
	return out
}
