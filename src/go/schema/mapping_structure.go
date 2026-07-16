package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"polystore_database/src/go/plan"
)

// ================================
// Structure for mapping dictionary
// ================================

type MappingDictionary struct {
	Labels        map[string][]string             `json:"labels"`
	Entities      map[string]*EntityMapping       `json:"entities"`
	Relationships map[string]*RelationshipMapping `json:"relationships"`
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
	return &mappingDictionary, nil
}

func (md *MappingDictionary) SaveMappingDictionary(path string) error {
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (md *MappingDictionary) LookupMappingDictionary(objType plan.ObjectType, label string, propertyName string) (store string, datatype string, err error) {
	if md == nil {
		return "unknown_store", "unknown_type", fmt.Errorf("mapping dictionary not found")
	}

	switch objType {
	case plan.Entity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Databases[0], prop.Type, nil
			}
		}
	case plan.Relationship:
		if rel, ok := md.Relationships[label]; ok {
			if prop, ok := rel.Properties[propertyName]; ok {
				return prop.Databases[0], prop.Type, nil
			}
		}
	default:
	}

	return "unknown_store", "unknown_type", fmt.Errorf("property '%s' not found for %s '%s'", propertyName, objType.String(), label)
}

func (md *MappingDictionary) GetPropertyDataType(objType plan.ObjectType, label string, propertyName string) (string, error) {
	switch objType {
	case plan.Entity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Type, nil
			}
		}
	case plan.Relationship:
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

func (md *MappingDictionary) GetDatastores(objType plan.ObjectType, label string, propertyName string) ([]string, error) {
	switch objType {
	case plan.Entity:
		if entity, ok := md.Entities[label]; ok {
			if prop, ok := entity.Properties[propertyName]; ok {
				return prop.Databases, nil
			}
		}
	case plan.Relationship:
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

func (md *MappingDictionary) CheckDatastore(objType plan.ObjectType, label string, propertyNames []string, expectedStore string) bool {
	for _, propName := range propertyNames {
		stores, err := md.GetDatastores(objType, label, propName)
		if err != nil {
			return false
		}

		found := false
		for _, s := range stores {
			if s == expectedStore {
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

func (md *MappingDictionary) UpdateDatastore(objType plan.ObjectType, label string, propertyNames []string, newDatastore string) error {
	newDataList := []string{newDatastore}

	for _, propName := range propertyNames {
		var targetProps map[string]*PropertyMapping

		switch objType {
		case plan.Entity:
			if entity, ok := md.Entities[label]; ok {
				targetProps = entity.Properties
			}
		case plan.Relationship:
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
