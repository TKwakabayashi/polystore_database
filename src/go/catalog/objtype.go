package catalog

// ObjectType is the kind of a property's owner (entity or relationship): it selects the mapping's
// entities or relationships.
type ObjectType int

const (
	ObjEntity ObjectType = iota
	ObjRelationship
)

func (ot ObjectType) String() string {
	switch ot {
	case ObjEntity:
		return "Entity"
	case ObjRelationship:
		return "Relationship"
	}
	return ""
}
