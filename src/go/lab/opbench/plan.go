package opbench

import (
	"encoding/json"
	"fmt"
	"math"
	"polystore_database/src/go/query/ir"

	"polystore_database/src/go/catalog"
)

// Variable names in plans: the input vertex is a, an expansion's destination b.
const (
	inAlias  = "a"
	outAlias = "b"
	// barrierVar is the column of the Unwind([1]) put before the target with at=coordinator. Unwind always runs on the
	// coordinator and keeps rows unchanged (a one-element list), so the target after it is not merged or absorbed into a store query.
	barrierVar = "_barrier"
)

// BuildLogical builds the logical plan of experiment e under condition c: leaf (Argument(a) or EntityScan(a)) ->
// earlier part -> target. The target is the root (and the root physical step contains it).
func BuildLogical(e *Experiment, c Condition) (ir.Node, error) {
	var leaf ir.Node = &ir.Argument{Cols: []string{inAlias}, Labels: map[string][]string{inAlias: labels(e.Label)}}
	if c.Input == "scan" {
		leaf = &ir.EntityScan{Alias: inAlias, Labels: []string{e.Label}}
	}
	get := func(in ir.Node, prop string) ir.Node {
		return &ir.GetProperty{Input: in, Alias: inAlias, Prop: prop, Kind: ir.NodeObj, Labels: []string{e.Label}}
	}
	col := func(prop string) ir.Expr { return &ir.Column{Key: inAlias + "." + prop} }
	barrier := func(in ir.Node) ir.Node {
		if c.At != "coordinator" {
			return in
		}
		return &ir.Unwind{Input: in, List: &ir.List{Elems: []ir.Expr{&ir.Lit{Val: int64(1)}}}, Var: barrierVar}
	}

	switch e.Op {
	case "EntityScan":
		if e.Cmp == "" {
			return leaf, nil
		}
		pred, err := e.predicate(col(e.Prop))
		if err != nil {
			return nil, err
		}
		return &ir.Filter{Input: get(leaf, e.Prop), Pred: pred}, nil

	case "GetProperty":
		return get(leaf, e.Prop), nil

	case "Projection":
		in := leaf
		if e.Cmp != "" {
			pred, err := e.predicate(col(e.Prop))
			if err != nil {
				return nil, err
			}
			in = &ir.Filter{Input: get(in, e.Prop), Pred: pred}
		}
		proj := &ir.Projection{}
		for _, p := range e.Columns {
			if e.Cmp == "" || p != e.Prop {
				in = get(in, p)
			}
			proj.Items = append(proj.Items, ir.ProjItem{Name: p, Expr: col(p)})
		}
		proj.Input = in
		return proj, nil

	case "Expand":
		return &ir.Expand{Input: leaf, From: inAlias, Rel: "r", To: outAlias, Types: []string{e.Rel},
			Dir: direction(e.Dir), ToLabels: labels(e.To)}, nil

	case "VarLengthExpand":
		// max < 0 is unbounded (as in ir.HopRange).
		return &ir.VarLengthExpand{Input: leaf, From: inAlias, Seg: "s", To: outAlias, Types: []string{e.Rel},
			Dir: direction(e.Dir), ToLabels: labels(e.To), Range: ir.HopRange{Min: e.Min, Max: e.Max}}, nil

	case "Filter":
		pred, err := e.predicate(col(e.Prop))
		if err != nil {
			return nil, err
		}
		return &ir.Filter{Input: barrier(get(leaf, e.Prop)), Pred: pred}, nil

	case "Aggregation":
		in := leaf
		agg := &ir.Aggregation{}
		if e.Prop != "" {
			in = get(in, e.Prop)
			agg.Keys = []ir.ProjItem{{Name: "key", Expr: col(e.Prop)}}
		}
		item := ir.AggItem{Name: "value", Fn: e.Agg}
		if e.AggProp != "" {
			if e.AggProp != e.Prop {
				in = get(in, e.AggProp)
			}
			item.Arg = col(e.AggProp)
		}
		agg.Aggs = []ir.AggItem{item}
		agg.Input = barrier(in)
		return agg, nil

	case "Sort":
		return &ir.Sort{Input: barrier(get(leaf, e.Prop)), Keys: []ir.SortKey{{Expr: col(e.Prop), Desc: e.Desc}}}, nil

	case "Limit":
		in := leaf
		if e.Prop != "" {
			in = get(in, e.Prop)
		}
		return &ir.Limit{Input: barrier(in), Count: e.Count}, nil

	case "Distinct":
		proj := &ir.Projection{Input: get(leaf, e.Prop), Items: []ir.ProjItem{{Name: "value", Expr: col(e.Prop)}}}
		return &ir.Distinct{Input: barrier(proj)}, nil
	}
	return nil, fmt.Errorf("opbench: unsupported op %q", e.Op)
}

// predicate builds "column cmp value" (IN takes an array).
func (e *Experiment) predicate(left ir.Expr) (ir.Expr, error) {
	switch e.Cmp {
	case "=", "<>", "<", ">", "<=", ">=":
		return &ir.Bin{Op: e.Cmp, L: left, R: &ir.Lit{Val: jsonValue(e.Value)}}, nil
	case "IN":
		vs, ok := e.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("the value of IN is an array: %v", e.Value)
		}
		l := &ir.List{}
		for _, v := range vs {
			l.Elems = append(l.Elems, &ir.Lit{Val: jsonValue(v)})
		}
		return &ir.Bin{Op: "IN", L: left, R: l}, nil
	}
	return nil, fmt.Errorf("unsupported cmp %q", e.Cmp)
}

// jsonValue makes integral JSON numbers int64 (the physical planner types literals from the mapping).
func jsonValue(v any) any {
	if f, ok := v.(float64); ok && f == math.Trunc(f) && math.Abs(f) < 1<<53 {
		return int64(f)
	}
	return v
}

func direction(d string) ir.Direction {
	switch d {
	case "in":
		return ir.In
	case "both":
		return ir.Both
	}
	return ir.Out
}

func labels(l string) []string {
	if l == "" {
		return nil
	}
	return []string{l}
}

// Props returns the properties the experiment moves between stores.
func (e *Experiment) Props() []string {
	var out []string
	for _, p := range append([]string{e.Prop, e.AggProp}, e.Columns...) {
		if p != "" && !contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// PlaceMapping returns a copy of base with only the experiment's properties placed in store (base itself for Neo4j);
// the statistics carry over.
func PlaceMapping(base *catalog.MappingDictionary, e *Experiment, store string) (*catalog.MappingDictionary, error) {
	if store == "graph" || len(e.Props()) == 0 {
		return base, nil
	}
	b, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	var md catalog.MappingDictionary
	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}
	md.Stats = base.Stats
	ent := entityOf(&md, e.Label)
	if ent == nil {
		return nil, fmt.Errorf("no entity %s in the mapping", e.Label)
	}
	for _, p := range e.Props() {
		pm, ok := ent.Properties[p]
		if !ok {
			return nil, fmt.Errorf("no %s.%s in the mapping", e.Label, p)
		}
		pm.Databases = []string{store}
	}
	return &md, nil
}

// entityOf returns the entity of a label, sub-labels included (Post -> Message).
func entityOf(md *catalog.MappingDictionary, label string) *catalog.EntityMapping {
	if ent, ok := md.Entities[label]; ok {
		return ent
	}
	for _, ent := range md.Entities {
		if contains(ent.Labels, label) {
			return ent
		}
	}
	return nil
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
