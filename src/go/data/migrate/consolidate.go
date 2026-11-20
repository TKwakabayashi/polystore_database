package migrate

import (
	"fmt"
	"sort"
	"strings"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores/conn"
)

// ConsolidationPlan says how to bring every property back to Neo4j.
type ConsolidationPlan struct {
	// Replicas already have their values in Neo4j; only the record changes.
	Replicas []string
	// Moves copy values back to Neo4j, one per (owner, source database).
	Moves []MigrationConfig
}

// PlanConsolidation lists what Consolidate does for record md. exclude holds "Label.prop" or
// "TYPE.prop" entries that keep their current placement.
func PlanConsolidation(md *catalog.MappingDictionary, exclude []string) ConsolidationPlan {
	skip := map[string]bool{}
	for _, e := range exclude {
		skip[strings.TrimSpace(e)] = true
	}
	var plan ConsolidationPlan
	type key struct {
		obj   catalog.ObjectType
		owner string
		src   kind.Kind
	}
	moves := map[key][]string{}
	visit := func(obj catalog.ObjectType, owner string, props map[string]*catalog.PropertyMapping) {
		for name, p := range props {
			if p == nil || skip[owner+"."+name] {
				continue
			}
			var kinds []kind.Kind
			inGraph := false
			for _, d := range p.Databases {
				k, err := kind.ParseKind(d)
				if err != nil {
					continue
				}
				kinds = append(kinds, k)
				inGraph = inGraph || k == kind.Neo4j
			}
			switch {
			case len(kinds) == 0 || len(kinds) == 1 && inGraph:
			case inGraph:
				plan.Replicas = append(plan.Replicas, owner+"."+name)
			default:
				k := key{obj, owner, kinds[0]}
				moves[k] = append(moves[k], name)
			}
		}
	}
	for label, e := range md.Entities {
		visit(catalog.ObjEntity, label, e.Properties)
	}
	for typ, r := range md.Relationships {
		visit(catalog.ObjRelationship, typ, r.Properties)
	}
	sort.Strings(plan.Replicas)
	for k, props := range moves {
		sort.Strings(props)
		plan.Moves = append(plan.Moves, MigrationConfig{ObjType: k.obj, Entity: k.owner, Properties: props, Mode: ModeOf(k.src, kind.Neo4j)})
	}
	sort.Slice(plan.Moves, func(i, j int) bool {
		a, b := plan.Moves[i], plan.Moves[j]
		if a.Entity != b.Entity {
			return a.Entity < b.Entity
		}
		return a.Mode < b.Mode
	})
	return plan
}

// Consolidate brings every property not in exclude back to Neo4j and records it there.
// Properties whose values are only elsewhere are migrated (deleteSource removes them from the
// source afterwards); replicas only have their record changed.
func Consolidate(cfg conn.Config, exclude []string, deleteSource bool) (ConsolidationPlan, error) {
	md, err := catalog.LoadRecord(cfg.MappingPath())
	if err != nil {
		return ConsolidationPlan{}, err
	}
	plan := PlanConsolidation(md, exclude)
	for _, m := range plan.Moves {
		m.MappingPath, m.DeleteSource = cfg.MappingPath(), deleteSource
		if cfg.Mongo != nil {
			m.MongoDbName = cfg.Mongo.DBName
		}
		if _, err := MigrateData(m, cfg); err != nil {
			return plan, fmt.Errorf("consolidate %s %v: %w", m.Entity, m.Properties, err)
		}
	}
	// A copy records Neo4j next to the source; the consolidated record names Neo4j only.
	toGraph := append([]string{}, plan.Replicas...)
	for _, m := range plan.Moves {
		for _, p := range m.Properties {
			toGraph = append(toGraph, m.Entity+"."+p)
		}
	}
	if len(toGraph) == 0 {
		return plan, nil
	}
	// Reload: the migrations above rewrote the record.
	if md, err = catalog.LoadRecord(cfg.MappingPath()); err != nil {
		return plan, err
	}
	for _, r := range toGraph {
		owner, prop, _ := strings.Cut(r, ".")
		obj := catalog.ObjEntity
		if _, ok := md.Entities[owner]; !ok {
			obj = catalog.ObjRelationship
		}
		if err := md.UpdateDatastore(obj, owner, []string{prop}, kind.Neo4j.String()); err != nil {
			return plan, err
		}
	}
	return plan, md.SaveMappingDictionary(cfg.MappingPath())
}
