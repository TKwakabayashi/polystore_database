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

	CondIn        // expr IN [v1, v2, ...]
	CondIsNull    // expr IS NULL
	CondIsNotNull // expr IS NOT NULL
	CondVarCmp    // a.x <op> b.y（クロスエイリアス比較。実比較子は ConditionNode.CmpOp）
)

// IsQuantifier は要素述語（パスフィルタ）の量化子（all/any/none/single）かを返す。
func (t ConditionType) IsQuantifier() bool {
	switch t {
	case CondAll, CondAny, CondNone, CondSingle:
		return true
	}
	return false
}

// CypherQuant は量化子 ConditionType の Cypher 関数名を返す（パスフィルタ生成用）。
func (t ConditionType) CypherQuant() string {
	switch t {
	case CondAll:
		return "all"
	case CondAny:
		return "any"
	case CondNone:
		return "none"
	case CondSingle:
		return "single"
	default:
		return ""
	}
}

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
