package schema

// ObjectType はプロパティの持ち主の種類（エンティティか関係か）。mapping の entities / relationships の
// どちらを引くかを決める（旧 plan.ObjectType をここへ移した）。
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
