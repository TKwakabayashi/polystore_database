// Package catalog holds what the planner knows about an instance: where each property is placed
// (the record mapping and runtime placements), property types, unique keys, the uuid storage form,
// and the statistics (catalog/stats).
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog/stats"
	"sort"
)

// ================================
// Structure for mapping dictionary
// ================================

type MappingDictionary struct {
	Labels        map[string][]string             `json:"labels"`
	Entities      map[string]*EntityMapping       `json:"entities"`
	Relationships map[string]*RelationshipMapping `json:"relationships"`

	// UUIDStorage is how the instance stores uuids ("text" or "binary"; empty is text).
	UUIDStorage string `json:"uuid_storage,omitempty"`

	// Stats are the planner statistics (stats.json next to mapping.json; nil if absent). Not written to mapping.json.
	Stats *stats.Stats `json:"-"`
}

type EntityMapping struct {
	Labels     []string                     `json:"labels"`
	Properties map[string]*PropertyMapping  `json:"properties"`
	References map[string]*ReferenceMapping `json:"references"`
}

type RelationshipMapping struct {
	Labels     []string                    `json:"labels"`
	Properties map[string]*PropertyMapping `json:"properties"`
}

type PropertyMapping struct {
	Type      string   `json:"type"`
	Databases []string `json:"database"`
	// Unique marks a key: equality with a literal selects at most one entity.
	Unique bool `json:"unique,omitempty"`
}

type ReferenceMapping struct {
	TargetEntities []string `json:"targetEntities"`
	Databases      []string `json:"database"`
}

func LoadMappingDictionary(path string) (*MappingDictionary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping dictionary: %v", err)
	}
	var mappingDictionary MappingDictionary
	if err := json.Unmarshal(data, &mappingDictionary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mapping dictionary: %v", err)
	}
	st, err := stats.Load(stats.PathFor(path))
	if err != nil {
		return nil, err
	}
	mappingDictionary.Stats = st
	return &mappingDictionary, nil
}

func (md *MappingDictionary) SaveMappingDictionary(path string) error {
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (md *MappingDictionary) LookupMappingDictionary(objType ObjectType, label string, propertyName string) (store string, datatype string, err error) {
	if md == nil {
		return "unknown_store", "unknown_type", fmt.Errorf("mapping dictionary not found")
	}

	switch objType {
	case ObjEntity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Databases[0], prop.Type, nil
			}
		}
	case ObjRelationship:
		if rel, ok := md.Relationships[label]; ok {
			if prop, ok := rel.Properties[propertyName]; ok {
				return prop.Databases[0], prop.Type, nil
			}
		}
	default:
	}

	return "unknown_store", "unknown_type", fmt.Errorf("property '%s' not found for %s '%s'", propertyName, objType.String(), label)
}

func (md *MappingDictionary) GetPropertyDataType(objType ObjectType, label string, propertyName string) (string, error) {
	switch objType {
	case ObjEntity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Type, nil
			}
		}
	case ObjRelationship:
		if rel, ok := md.Relationships[label]; ok {
			if prop, ok := rel.Properties[propertyName]; ok {
				return prop.Type, nil
			}
		}
	default:
		return "", fmt.Errorf("invalid ObjectType: %s", objType.String())
	}

	return "", fmt.Errorf("property '%s' not found for %s '%s'", propertyName, objType.String(), label)
}

func (md *MappingDictionary) GetDatastores(objType ObjectType, label string, propertyName string) ([]string, error) {
	switch objType {
	case ObjEntity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Databases, nil
			}
		}
	case ObjRelationship:
		if rel, ok := md.Relationships[label]; ok {
			if prop, ok := rel.Properties[propertyName]; ok {
				return prop.Databases, nil
			}
		}
	default:
		return nil, fmt.Errorf("invalid dataType: %s", objType)
	}

	return nil, fmt.Errorf("property '%s' not found for %s '%s'", propertyName, objType.String(), label)
}

func (md *MappingDictionary) CheckDatastore(objType ObjectType, label string, propertyNames []string, expectedStore string) bool {
	for _, propName := range propertyNames {
		stores, err := md.GetDatastores(objType, label, propName)
		if err != nil {
			return false
		}

		found := false
		for _, s := range stores {
			if sameStore(s, expectedStore) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (md *MappingDictionary) UpdateDatastore(objType ObjectType, label string, propertyNames []string, newDatastore string) error {
	newDataList := []string{newDatastore}

	for _, propName := range propertyNames {
		var targetProps map[string]*PropertyMapping

		switch objType {
		case ObjEntity:
			if entity, ok := md.Entities[label]; ok {
				targetProps = entity.Properties
			}
		case ObjRelationship:
			if rel, ok := md.Relationships[label]; ok {
				targetProps = rel.Properties
			}
		default:
			return fmt.Errorf("invalid dataType: %s", objType.String())
		}

		if targetProps != nil {
			if prop, ok := targetProps[propName]; ok {
				prop.Databases = newDataList
			} else {
				return fmt.Errorf("property '%s' not found for %s '%s'", propName, objType.String(), label)
			}
		} else {
			return fmt.Errorf("%s '%s' not found", objType.String(), label)
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	newSlice := []string{}
	for _, s := range slice {
		if s != item {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// UniqueKeys returns the names of the properties marked unique on any entity, sorted.
func (md *MappingDictionary) UniqueKeys() []string {
	if md == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range md.Entities {
		for name, p := range e.Properties {
			if p != nil && p.Unique && !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	sort.Strings(out)
	return out
}

// sameStore compares placement names, accepting legacy data-model names ("graph" = "neo4j").
func sameStore(a, b string) bool {
	if a == b {
		return true
	}
	ka, ea := kind.ParseKind(a)
	kb, eb := kind.ParseKind(b)
	return ea == nil && eb == nil && ka == kb
}

// LoadRecord loads an instance's record mapping and sets the process-wide uuid storage from it.
func LoadRecord(path string) (*MappingDictionary, error) {
	md, err := LoadMappingDictionary(path)
	if err != nil {
		return nil, err
	}
	s, err := id.ParseStorage(md.UUIDStorage)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	id.SetStorage(s)
	return md, nil
}

// AddDatastore records that the properties also have their values in newDatastore (a copy:
// the other databases keep theirs).
func (md *MappingDictionary) AddDatastore(objType ObjectType, label string, propertyNames []string, newDatastore string) error {
	for _, name := range propertyNames {
		dbs, err := md.GetDatastores(objType, label, name)
		if err != nil {
			return err
		}
		has := false
		for _, d := range dbs {
			has = has || sameStore(d, newDatastore)
		}
		if !has {
			if err := md.setDatastores(objType, label, name, append(append([]string{}, dbs...), newDatastore)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (md *MappingDictionary) setDatastores(objType ObjectType, label, name string, dbs []string) error {
	var props map[string]*PropertyMapping
	switch objType {
	case ObjEntity:
		if e, ok := md.Entities[label]; ok {
			props = e.Properties
		}
	case ObjRelationship:
		if r, ok := md.Relationships[label]; ok {
			props = r.Properties
		}
	}
	p, ok := props[name]
	if !ok {
		return fmt.Errorf("property %s.%s not found", label, name)
	}
	p.Databases = dbs
	return nil
}

// Clone returns a deep copy (the statistics are shared: they are read-only).
func (md *MappingDictionary) Clone() *MappingDictionary {
	b, err := json.Marshal(md)
	if err != nil {
		panic(err) // the mapping is plain data
	}
	var cp MappingDictionary
	if err := json.Unmarshal(b, &cp); err != nil {
		panic(err)
	}
	cp.Stats = md.Stats
	return &cp
}

// Placed returns a runtime placement: a copy of md where the given properties are read from
// target only. The record (md) is not changed.
func (md *MappingDictionary) Placed(objType ObjectType, label string, propertyNames []string, target string) (*MappingDictionary, error) {
	cp := md.Clone()
	for _, name := range propertyNames {
		if err := cp.setDatastores(objType, label, name, []string{target}); err != nil {
			return nil, err
		}
	}
	return cp, nil
}
