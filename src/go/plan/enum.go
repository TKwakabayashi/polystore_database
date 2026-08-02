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

// （旧 plan.DataStore enum は未使用のため削除。ストア種別の語彙は store.Kind に一本化した。）

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
	CondEq        ConditionType = iota // =
	CondNeq                            // !=
	CondGreater                        // >
	CondLess                           // <
	CondGreaterEq                      // >=
	CondLessEq                         // <=

	CondAnd   // &&
	CondOr    // ||
	CondNot   // !
	CondParen // ()

	CondAll    //
	CondNone   //
	CondAny    //
	CondSingle //
)

type AggFunc int

const (
	AggCount AggFunc = iota
	AggSum
	AggAvg
	AggMin
	AggMax
)

func (a AggFunc) String() string {
	switch a {
	case AggCount:
		return "count"
	case AggSum:
		return "sum"
	case AggAvg:
		return "avg"
	case AggMin:
		return "min"
	case AggMax:
		return "max"
	default:
		return "unknown"
	}
}

type OrderDir int

const (
	OrderAsc OrderDir = iota
	OrderDesc
)
