// Package physical assigns a store to each logical operator, merges consecutive operators on the same store
// into one StoreQuery, and fetches values when they are first needed. It does not use statistics.
package physical

import (
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores/dbs"
	"sort"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
)

// Placement looks up placements in the mapping. Store choice keeps state per plan (a replicated column prefers
// the store already chosen for the same alias), so make a new Placement per plan.
type Placement struct {
	md     *catalog.MappingDictionary
	chosen map[string]kind.Kind // alias -> store already chosen
}

// NewPlacement makes a Placement from a mapping; a nil md puts every column in Neo4j.
func NewPlacement(md *catalog.MappingDictionary) *Placement {
	return &Placement{md: md, chosen: map[string]kind.Kind{}}
}

// EntityOf resolves a node label to the entity that contains it (Post -> Message: sub-labels listed in an
// entity's labels). Empty if no entity has it.
func (c *Placement) EntityOf(label string) string {
	if c.md == nil {
		return ""
	}
	if _, ok := c.md.Entities[label]; ok {
		return label
	}
	for _, name := range sortedEntities(c.md) {
		for _, l := range c.md.Entities[name].Labels {
			if l == label {
				return name
			}
		}
	}
	return ""
}

// Entities maps labels to entity names (no duplicates; unresolved labels stay).
func (c *Placement) Entities(labels []string) []string {
	var out []string
	for _, l := range labels {
		e := c.EntityOf(l)
		if e == "" {
			e = l
		}
		dup := false
		for _, x := range out {
			if x == e {
				dup = true
			}
		}
		if !dup {
			out = append(out, e)
		}
	}
	return out
}

// IsEntity reports whether a label is itself an entity (the table or collection unit outside Neo4j). Sub-labels
// (Post, ...) are not distinguished in an entity's table, so they cannot be scanned outside Neo4j.
func (c *Placement) IsEntity(label string) bool {
	if c.md == nil {
		return false
	}
	_, ok := c.md.Entities[label]
	return ok
}

// databases returns the stores holding a property, given its owner (label or type candidates).
func (c *Placement) databases(obj ir.ObjKind, labels []string, prop string) []string {
	if c.md == nil {
		return nil
	}
	if obj == ir.NodeObj {
		labels = c.Entities(labels)
	}
	for _, l := range labels {
		if obj == ir.RelObj {
			if r, ok := c.md.Relationships[l]; ok {
				if p, ok := r.Properties[prop]; ok {
					return p.Databases
				}
			}
			continue
		}
		if e, ok := c.md.Entities[l]; ok {
			if p, ok := e.Properties[prop]; ok {
				return p.Databases
			}
		}
	}
	return nil
}

// StoreOf picks the store to read a property from. Columns missing from the mapping are in Neo4j (the source of all data).
func (c *Placement) StoreOf(alias string, obj ir.ObjKind, labels []string, prop string) kind.Kind {
	var cands []kind.Kind
	for _, d := range c.databases(obj, labels, prop) {
		if k, err := kind.ParseKind(d); err == nil {
			cands = append(cands, k)
		}
	}
	if len(cands) == 0 {
		return kind.Neo4j
	}
	prev, hasPrev := c.chosen[alias]
	pick := choose(cands, prev, hasPrev)
	if !hasPrev && alias != "" {
		c.chosen[alias] = pick
	}
	return pick
}

// choose picks among replicas: the store already chosen for the alias, then Neo4j, then the first.
func choose(cands []kind.Kind, prev kind.Kind, hasPrev bool) kind.Kind {
	if hasPrev {
		for _, k := range cands {
			if k == prev {
				return k
			}
		}
	}
	for _, k := range cands {
		if k == kind.Neo4j {
			return k
		}
	}
	return cands[0]
}

// offGraph reports whether g's values are not in Neo4j (fetching them would split a traversal). false for nil.
func (c *Placement) offGraph(g *ir.GetProperty) bool {
	if g == nil {
		return false
	}
	ds := c.databases(g.Kind, g.Labels, g.Prop)
	if len(ds) == 0 {
		return false
	}
	for _, d := range ds {
		if k, err := kind.ParseKind(d); err == nil && k == kind.Neo4j {
			return false
		}
	}
	return true
}

// fetchGroup groups GetProperty that one query can fetch ("" is Neo4j, where any alias fits in one query).
func (c *Placement) fetchGroup(g *ir.GetProperty) string {
	labels := g.Labels
	if g.Kind == ir.NodeObj {
		labels = c.Entities(labels)
	}
	k := c.StoreOf(g.Alias, g.Kind, labels, g.Prop)
	if k == kind.Neo4j {
		return ""
	}
	return k.String() + "/" + g.Alias
}

// ElemLabels returns label candidates of a segment's elements: inner nodes from the relationship types'
// targetEntities, relationships are the types themselves.
func (c *Placement) ElemLabels(types []string, part ir.SegPart) []string {
	if part == ir.SegRels {
		return append([]string{}, types...)
	}
	if c.md == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range sortedEntities(c.md) {
		for _, t := range types {
			if ref, ok := c.md.Entities[e].References[t]; ok {
				for _, target := range ref.TargetEntities {
					if !seen[target] {
						seen[target] = true
						out = append(out, target)
					}
				}
			}
		}
	}
	return out
}

// HaveLabels reports whether the mapping guarantees that every node at either end of a relationship of types
// has all of labels (nodes are single-labelled entities, so all ends must be the one label asked for).
func (c *Placement) HaveLabels(types, labels []string) bool {
	if len(labels) == 0 {
		return true
	}
	if c.md == nil || len(types) == 0 || len(labels) != 1 {
		return false
	}
	ends := 0
	for _, e := range sortedEntities(c.md) {
		for _, t := range types {
			ref, ok := c.md.Entities[e].References[t]
			if !ok {
				continue
			}
			for _, end := range append([]string{e}, ref.TargetEntities...) {
				if end != labels[0] {
					return false
				}
				ends++
			}
		}
	}
	return ends > 0
}

func sortedEntities(md *catalog.MappingDictionary) []string {
	out := make([]string, 0, len(md.Entities))
	for k := range md.Entities {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ================================
// Store capabilities (what a store expresses natively)
// ================================

// canExpress reports whether store k evaluates predicate e natively.
func canExpress(k kind.Kind, e ir.Expr) bool {
	db := dbs.Of(k)
	return db != nil && db.CanExpress(e)
}

// canAbsorb reports whether a leaf query on store k can run whole-row operator op.
func canAbsorb(k kind.Kind, op ir.Node) bool {
	db := dbs.Of(k)
	return db != nil && db.CanAbsorb(op)
}

// canShip reports whether a shipped boundary of mode goes to store k (Options.ShipStores). Neo4j, MySQL and
// MongoDB join the table natively; the executor emulates it for the others. The planner only makes
// ir.ShipAggregate (ShipFilter / ShipProject are compile-only).
func (o Options) canShip(k kind.Kind, mode ir.ShipMode) bool {
	if mode != ir.ShipAggregate {
		return false
	}
	return o.ShipStores[k]
}

// TypeOf returns a property's type name (the mapping's type; "" if unknown), used to type literals in native queries.
func (c *Placement) TypeOf(obj ir.ObjKind, labels []string, prop string) string {
	if c.md == nil {
		return ""
	}
	if obj == ir.NodeObj {
		labels = c.Entities(labels)
	}
	for _, l := range labels {
		if obj == ir.RelObj {
			if r, ok := c.md.Relationships[l]; ok {
				if p, ok := r.Properties[prop]; ok {
					return p.Type
				}
			}
			continue
		}
		if e, ok := c.md.Entities[l]; ok {
			if p, ok := e.Properties[prop]; ok {
				return p.Type
			}
		}
	}
	return ""
}
