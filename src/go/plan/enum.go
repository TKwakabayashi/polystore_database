package plan

type Direction int

const (
	Incoming      Direction = iota // <-
	Outgoing                       // ->
	Bidirectional                  // - or <-->
)

type ObjectType int

const (
	Entity ObjectType = iota
	Relationship
)

func (ot ObjectType) String() string {
	var str string
	switch ot {
	case Entity:
		str = "Entity"
	case Relationship:
		str = "Relationship"
	default:
	}
	return str
}

type DataType int

const (
	Int DataType = iota
	Long
	Float
	Double
	Date
	Datetime
	String
)

func (dt DataType) String() string {
	var str string
	switch dt {
	case Int:
		str = "int"
	case Long:
		str = "long"
	case Float:
		str = "float"
	case Double:
		str = "double"
	case Date:
		str = "date"
	case Datetime:
		str = "datetime"
	case String:
		str = "string"
	default:
		// emit error
	}
	return str
}

type DataStore int

const (
	Graph DataStore = iota
	Columnar
	Relational
	Document
	Kvs
)

func (ds DataStore) String() string {
	var str string
	switch ds {
	case Graph:
		str = "graph"
	case Columnar:
		str = "columnar"
	case Relational:
		str = "relational"
	case Document:
		str = "document"
	case Kvs:
		str = "kvs"
	default:
		// emit error
	}
	return str
}

// VarLengthExpand Condition

type FilterPhase int

const (
	PhasePre FilterPhase = iota
	PhaseStep
	PhasePost
)

type FilterTarget int

const (
	TargetNode FilterTarget = iota
	TargetRelationship
	TargetGlobalPath
)

// general condition
type ConditionType int

const (
	CondEq      ConditionType = iota // =
	CondNeq                          // !=
	CondGreater                      // >
	CondLess                         // <

	CondAnd   // &&
	CondOr    // ||
	CondNot   // !
	CondParen // ()

	CondAll    //
	CondNone   //
	CondAny    //
	CondSingle //
)

type OrderDir int

const (
	OrderAsc OrderDir = iota
	OrderDesc
)
