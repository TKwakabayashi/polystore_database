package physical

// prefilter.go is PathPrefilter: a condition on the nodes of a path or segment whose values are outside Neo4j is
// evaluated first over all entities of the element labels in their stores (a set plan per condition), and the
// traversal keeps the paths whose elements are in the resulting id set (ir.InSet, `x.uuid IN $set` in Neo4j).
//
// In a WHERE (a Filter keeps rows whose predicate is true), all(x IN l WHERE c(x)) is true exactly when every
// element is in T = {x : c(x) is true}, and any(...) when some element is; null and false are both dropped, so
// three-valued logic agrees. none(c) is all(NOT c). single is not rewritten (it is null when some c(x) is null,
// which membership in T cannot tell), nor are relationships (no relationship scan outside Neo4j).

import (
	"fmt"
	"polystore_database/src/go/query/ir"
	"strings"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/logical"
)

// setSpec is a condition to evaluate before the main plan: pred on element variable v over the entities labels.
type setSpec struct {
	name   string
	labels []string
	v      string
	pred   ir.Expr
}

type prefilterer struct {
	pl    *planner
	sets  []*setSpec
	byKey map[string]string // labels and condition -> set name (the same condition twice reads one set)
}

// prefilter returns a copy of the plan with the covered conditions reading id sets, and the sets to compute.
// pl must have indexed root (path and segment element labels).
func (pl *planner) prefilter(root ir.Node) (ir.Node, []*setSpec) {
	p := &prefilterer{pl: pl, byKey: map[string]string{}}
	out := p.rewrite(root)
	if len(p.sets) > 0 {
		out = logical.Prune(out) // value lists the rewritten conditions no longer read
	}
	return out, p.sets
}

// rewrite copies the operators down to the leaves (the logical plan may be planned again), rewriting conditions.
func (p *prefilterer) rewrite(n ir.Node) ir.Node {
	switch x := n.(type) {
	case nil:
		return nil
	case *ir.NodeHashJoin:
		y := *x
		y.Left, y.Right = p.rewrite(x.Left), p.rewrite(x.Right)
		return &y
	case *ir.Optional:
		y := *x
		y.Outer, y.Inner = p.rewrite(x.Outer), p.rewrite(x.Inner)
		return &y
	}
	c := ir.Child(n)
	if c == nil {
		return n
	}
	cp := ir.WithInput(n, p.rewrite(c))
	switch y := cp.(type) {
	case *ir.VarLengthExpand:
		preds := append([]ir.SegPred(nil), y.SegPreds...)
		for i, sp := range preds {
			if sp.Part != ir.SegNodes {
				continue
			}
			labels := p.pl.cat.Entities(p.pl.cat.ElemLabels(y.Types, sp.Part))
			if name, ok := p.setFor(labels, sp.Var, sp.Pred); ok {
				preds[i].Pred = &ir.InSet{Var: sp.Var, Elem: true, Set: name}
			}
		}
		y.SegPreds = preds
	case *ir.Filter:
		if q, ok := y.Pred.(*ir.Quant); ok {
			if nq, ok := p.quant(q); ok {
				y.Pred = nq
			}
		}
	}
	return cp
}

// quant rewrites a quantifier over nodes(path) that is a whole Filter predicate.
func (p *prefilterer) quant(q *ir.Quant) (ir.Expr, bool) {
	if q.Kind == "single" {
		return nil, false
	}
	call, ok := q.List.(*ir.Call)
	if !ok || call.Fn != "nodes" || len(call.Args) != 1 {
		return nil, false
	}
	src, ok := call.Args[0].(*ir.Column)
	if !ok {
		return nil, false
	}
	labels := p.pl.cat.Entities(p.pl.listLabels(src.Key, ir.SegNodes))
	pred, quant := q.Pred, q.Kind
	if quant == "none" {
		pred, quant = &ir.Un{Op: "NOT", X: pred}, "all"
	}
	name, ok := p.setFor(labels, q.Var, pred)
	if !ok {
		return nil, false
	}
	return &ir.Quant{Kind: quant, Var: q.Var, List: q.List, Pred: &ir.InSet{Var: q.Var, Elem: true, Set: name}}, true
}

// setFor returns the set of pred over labels when the condition reads only the element's properties and some of
// them are outside Neo4j for some label.
func (p *prefilterer) setFor(labels []string, v string, pred ir.Expr) (string, bool) {
	if len(labels) == 0 || !onlyElem(pred, v) {
		return "", false
	}
	props := elemProps(pred, v)
	off := false
	for _, l := range labels {
		for _, prop := range props {
			if p.pl.cat.StoreOf("", ir.NodeObj, []string{l}, prop) != kind.Neo4j {
				off = true
			}
		}
	}
	if !off {
		return "", false
	}
	key := strings.Join(labels, "|") + "/" + v + "/" + pred.String()
	if name, ok := p.byKey[key]; ok {
		return name, true
	}
	name := fmt.Sprintf("set%d", len(p.sets))
	p.byKey[key] = name
	p.sets = append(p.sets, &setSpec{name: name, labels: labels, v: v, pred: pred})
	return name, true
}

// onlyElem reports whether e reads nothing but properties of element v and literals (no columns, ids or nested quantifiers).
func onlyElem(e ir.Expr, v string) bool {
	ok := true
	ir.Walk(e, func(x ir.Expr) {
		switch y := x.(type) {
		case *ir.Column, *ir.Quant, *ir.InSet:
			ok = false
		case *ir.ElemProp:
			if y.Var != v {
				ok = false
			}
		}
	})
	return ok
}

// elemToColumn turns element v's properties into columns of alias (alias.p).
func elemToColumn(v, alias string) func(ir.Expr) ir.Expr {
	return func(e ir.Expr) ir.Expr {
		if ep, ok := e.(*ir.ElemProp); ok && ep.Var == v {
			return &ir.Column{Key: alias + "." + ep.Prop}
		}
		return e
	}
}

// conjuncts splits e at top-level ANDs.
func conjuncts(e ir.Expr) []ir.Expr {
	if b, ok := e.(*ir.Bin); ok && b.Op == "AND" {
		return append(conjuncts(b.L), conjuncts(b.R)...)
	}
	return []ir.Expr{e}
}

// buildSet plans a set: per label, scan -> (fetch the values a term reads -> filter by the term)... -> the id
// column. Each plan is planned like a query of its own (the scan goes to the store of the first term's values).
func buildSet(s *setSpec, cat *Placement, opt Options) (*SetPlan, error) {
	opt.PathPredicates = PathPost
	alias := "_" + s.name
	sp := &SetPlan{Name: s.name, Col: alias}
	for _, l := range s.labels {
		var n ir.Node = &ir.EntityScan{Alias: alias, Labels: []string{l}}
		fetched := map[string]bool{}
		for _, term := range conjuncts(s.pred) {
			for _, prop := range elemProps(term, s.v) {
				if !fetched[prop] {
					fetched[prop] = true
					n = &ir.GetProperty{Input: n, Alias: alias, Prop: prop, Kind: ir.NodeObj, Labels: []string{l}}
				}
			}
			n = &ir.Filter{Input: n, Pred: ir.Transform(term, elemToColumn(s.v, alias))}
		}
		n = &ir.Projection{Input: n, Items: []ir.ProjItem{{Name: alias, Expr: &ir.Column{Key: alias}}}}
		pl := newPlanner(opt, cat)
		pl.index(n, nil)
		st, err := pl.build(n)
		if err != nil {
			return nil, fmt.Errorf("set %s: %w", s.name, err)
		}
		pl.close(st)
		pl.finalize(st.phys, nil, true)
		sp.Plans = append(sp.Plans, st.phys)
	}
	return sp, nil
}
