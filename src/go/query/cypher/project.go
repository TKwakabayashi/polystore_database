package cypher

// project.go translates WITH and RETURN (projection, aggregation, sort, limit, distinct).

import (
	"fmt"

	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

func (b *builder) project(c *ast.ProjectionClause) error {
	if b.plan == nil {
		return unsupported("projection without MATCH")
	}

	hasAgg := false
	for _, it := range c.Items {
		if _, ok := it.Expr.(*ast.Agg); ok {
			hasAgg = true
		}
	}
	// When aggregation or DISTINCT merges rows, ORDER BY may reference projected columns or properties of projected bindings only.
	collapsing := hasAgg || c.Distinct
	projIdent := map[string]varInfo{} // bindings projected under their own name (ORDER BY may use their properties)
	for _, it := range c.Items {
		if v, ok := it.Expr.(*ast.VarRef); ok && v.Name == it.Name {
			if vi := b.scope[v.Name]; vi.kind == VarNode || vi.kind == VarRel {
				projIdent[it.Name] = vi
			}
		}
	}

	// 1. Add the needed columns (ORDER BY properties of merged rows are fetched after merging)
	var nd needs
	for _, it := range c.Items {
		b.collectNeeds(it.Expr, convCtx{}, &nd)
	}
	if !collapsing {
		for _, o := range c.OrderBy {
			b.collectNeeds(o.Expr, convCtx{}, &nd)
		}
	}
	b.plan = b.supply(b.plan, nd)

	// 2. Convert the items
	itemNames := map[string]string{} // expression text -> item name (to resolve ORDER BY)
	outNames := map[string]bool{}
	for _, it := range c.Items {
		if outNames[it.Name] {
			return unsupported("duplicate column name %s", it.Name)
		}
		outNames[it.Name] = true
		itemNames[it.Expr.String()] = it.Name
	}

	var sortKeys []ir.SortKey
	var hidden []ir.ProjItem
	var postProps []propNeed
	resolveOrder := func(o ast.SortItem) (ir.Expr, error) {
		if name, ok := itemNames[o.Expr.String()]; ok {
			return &ir.Column{Key: name}, nil
		}
		if v, ok := o.Expr.(*ast.VarRef); ok && outNames[v.Name] {
			return &ir.Column{Key: v.Name}, nil
		}
		if onlyOutputs(o.Expr, outNames) { // an expression over projected columns only (toInteger(personId), ...)
			saved := b.scope
			b.scope = map[string]varInfo{}
			for n := range outNames {
				b.scope[n] = varInfo{kind: VarValue}
			}
			e, err := b.conv(o.Expr, convCtx{})
			b.scope = saved
			return e, err
		}
		if collapsing {
			for _, d := range b.deps(o.Expr) {
				if _, ok := projIdent[d]; !ok {
					return nil, unsupported("ORDER BY %s not in projection", o.Expr)
				}
			}
			var ond needs
			b.collectNeeds(o.Expr, convCtx{}, &ond)
			if len(ond.paths) > 0 || len(ond.listProps) > 0 {
				return nil, unsupported("ORDER BY %s not in projection", o.Expr)
			}
			postProps = append(postProps, ond.props...)
			return b.conv(o.Expr, convCtx{})
		}
		e, err := b.conv(o.Expr, convCtx{})
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("_sort%d", len(hidden))
		hidden = append(hidden, ir.ProjItem{Name: name, Expr: e})
		return &ir.Column{Key: name}, nil
	}

	trim := func() []ir.ProjItem {
		items := make([]ir.ProjItem, 0, len(c.Items))
		for _, it := range c.Items {
			items = append(items, ir.ProjItem{Name: it.Name, Expr: &ir.Column{Key: it.Name}})
		}
		return items
	}

	var finalItems []ir.ProjItem
	if hasAgg {
		agg := &ir.Aggregation{Input: b.plan}
		for _, it := range c.Items {
			if a, ok := it.Expr.(*ast.Agg); ok {
				ai := ir.AggItem{Name: it.Name, Fn: a.Fn, Distinct: a.Distinct}
				if a.Arg != nil {
					v, err := b.conv(a.Arg, convCtx{})
					if err != nil {
						return err
					}
					ai.Arg = v
				}
				agg.Aggs = append(agg.Aggs, ai)
				continue
			}
			v, err := b.conv(it.Expr, convCtx{})
			if err != nil {
				return err
			}
			agg.Keys = append(agg.Keys, ir.ProjItem{Name: it.Name, Expr: v})
		}
		b.plan = agg
		// An aggregation outputs keys then aggregates; reorder at the end if the items differ.
		natural := make([]string, 0, len(c.Items))
		for _, k := range agg.Keys {
			natural = append(natural, k.Name)
		}
		for _, a := range agg.Aggs {
			natural = append(natural, a.Name)
		}
		for i, it := range c.Items {
			if natural[i] != it.Name {
				finalItems = trim()
				break
			}
		}
	} else {
		proj := &ir.Projection{Input: b.plan}
		for _, it := range c.Items {
			v, err := b.conv(it.Expr, convCtx{})
			if err != nil {
				return err
			}
			proj.Items = append(proj.Items, ir.ProjItem{Name: it.Name, Expr: v})
		}
		b.plan = proj
	}

	for _, o := range c.OrderBy {
		e, err := resolveOrder(o)
		if err != nil {
			return err
		}
		sortKeys = append(sortKeys, ir.SortKey{Expr: e, Desc: o.Desc})
	}
	if len(hidden) > 0 {
		proj := b.plan.(*ir.Projection)
		proj.Items = append(proj.Items, hidden...)
		finalItems = trim()
	}

	// 3. Post-processing
	if c.Distinct {
		b.plan = &ir.Distinct{Input: b.plan}
	}
	seenPost := map[string]bool{}
	for _, p := range postProps {
		key := p.v + "." + p.prop
		if seenPost[key] {
			continue
		}
		seenPost[key] = true
		vi := projIdent[p.v]
		kind := ir.NodeObj
		if vi.kind == VarRel {
			kind = ir.RelObj
		}
		b.plan = &ir.GetProperty{Input: b.plan, Alias: p.v, Prop: p.prop, Kind: kind, Labels: vi.labels}
		finalItems = trim()
	}
	if len(sortKeys) > 0 {
		b.plan = &ir.Sort{Input: b.plan, Keys: sortKeys}
	}
	if c.Skip != nil || c.Limit != nil {
		l := &ir.Limit{Input: b.plan, Count: -1}
		if c.Skip != nil {
			l.Skip = *c.Skip
		}
		if c.Limit != nil {
			l.Count = *c.Limit
		}
		b.plan = l
	}
	if len(finalItems) > 0 {
		b.plan = &ir.Projection{Input: b.plan, Items: finalItems}
	}

	// 4. Replace the scope (the essence of WITH)
	newScope := map[string]varInfo{}
	newBound := map[string]bool{}
	newAvail := map[string]bool{}
	newPaths := map[string]pathDef{}
	for _, it := range c.Items {
		info := varInfo{kind: VarValue}
		switch x := it.Expr.(type) {
		case *ast.VarRef:
			vi := b.scope[x.Name]
			if vi.kind == VarNode || vi.kind == VarRel {
				newScope[it.Name] = vi
				newBound[it.Name] = true
				continue
			}
			if vi.kind == VarPath {
				return unsupported("carrying path %s across WITH", x.Name)
			}
			info = vi // carry value columns over (with element info for lists of bindings)
		case *ast.Agg:
			if v, ok := x.Arg.(*ast.VarRef); ok && x.Fn == "collect" {
				if vi := b.scope[v.Name]; vi.kind == VarNode || vi.kind == VarRel {
					el := vi
					info.elem = &el
				}
			}
		}
		newScope[it.Name] = info
		newBound[it.Name] = true
		newAvail[it.Name] = true
	}
	b.scope, b.bound, b.avail, b.paths = newScope, newBound, newAvail, newPaths

	// 5. The WHERE after WITH
	if c.Where != nil {
		var err error
		var rest []ast.Expr
		b.plan, rest, err = b.placeReady(b.plan, b.bound, splitAnd(c.Where))
		if err != nil {
			return err
		}
		if len(rest) > 0 {
			return unsupported("unplaceable WITH predicate %s", rest[0])
		}
	}
	return nil
}
