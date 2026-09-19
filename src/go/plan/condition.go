package plan

import (
	"fmt"
	"strings"

	"polystore_database/src/go/store"
)

type ConditionNode struct {
	Type      ConditionType
	DataStore store.Kind // assign at BuildPlan

	Labels   []string
	Alias    string
	Property string
	Value    string
	ObjType  ObjectType
	DataType string // int, string etc.

	// CondIn 用: IN リストの候補値（Alias.Property IN Values）。
	Values []string

	// CondVarCmp 用: 右辺の alias.property とクロスエイリアス比較の実比較子。
	//   左辺は Alias.Property、右辺は RAlias.RProperty、比較は CmpOp（CondEq/CondLess/…）。
	RAlias    string
	RProperty string
	CmpOp     ConditionType

	// パス大域述語用: length(p) OP val の一時表現（Alias=path var, Type=比較子, Value=境界）。
	// リスナが検出し、query_planner が VarLengthFilter(TargetGlobalPath) へ変換する。
	PathFunc string

	Left  *ConditionNode
	Right *ConditionNode
	Child *ConditionNode
}

func (n *ConditionNode) Print() string {
	if n == nil {
		return ""
	}

	switch n.Type {
	case CondAnd:
		leftStr := n.Left.Print()
		rightStr := n.Right.Print()
		return fmt.Sprintf("(%s AND %s)", leftStr, rightStr)
	case CondOr:
		leftStr := n.Left.Print()
		rightStr := n.Right.Print()
		return fmt.Sprintf("(%s OR %s)", leftStr, rightStr)
	case CondEq:
		return fmt.Sprintf("%s.%s = %s", n.Alias, n.Property, n.Value)
	case CondNeq:
		return fmt.Sprintf("%s.%s <> %s", n.Alias, n.Property, n.Value)
	case CondGreater:
		return fmt.Sprintf("%s.%s > %s", n.Alias, n.Property, n.Value)
	case CondLess:
		return fmt.Sprintf("%s.%s < %s", n.Alias, n.Property, n.Value)
	case CondGreaterEq:
		return fmt.Sprintf("%s.%s >= %s", n.Alias, n.Property, n.Value)
	case CondLessEq:
		return fmt.Sprintf("%s.%s <= %s", n.Alias, n.Property, n.Value)
	case CondNot:
		return fmt.Sprintf("NOT %s", n.Child.Print())
	case CondParen:
		return fmt.Sprintf("(%s)", n.Child.Print())
	case CondIn:
		return fmt.Sprintf("%s.%s IN [%s]", n.Alias, n.Property, strings.Join(n.Values, ", "))
	case CondIsNull:
		return fmt.Sprintf("%s.%s IS NULL", n.Alias, n.Property)
	case CondIsNotNull:
		return fmt.Sprintf("%s.%s IS NOT NULL", n.Alias, n.Property)
	case CondVarCmp:
		return fmt.Sprintf("%s.%s %s %s.%s", n.Alias, n.Property, cmpSymbol(n.CmpOp), n.RAlias, n.RProperty)
	default:
		return "Invalid Condition"
	}
}

// cmpSymbol は比較 ConditionType を記号へ（Print/ダンプ用）。
func cmpSymbol(t ConditionType) string {
	switch t {
	case CondEq:
		return "="
	case CondNeq:
		return "<>"
	case CondGreater:
		return ">"
	case CondLess:
		return "<"
	case CondGreaterEq:
		return ">="
	case CondLessEq:
		return "<="
	default:
		return "?"
	}
}

// VarLengthFilter は可変長パス p=(a)-[:R*..]-(b) 全体に対する述語を表す。
// VarLengthExpand.Filters に格納され、graph の Cypher 生成時に WHERE へ展開される。
//
//   - 要素述語（Target=TargetNode/TargetRelationship）:
//     Quant(all/any/none/single)(IterVar IN nodes(p)|relationships(p) WHERE Cond)
//     例: all(x IN relationships(p) WHERE x.creationDate > $d)
//   - 大域述語（Target=TargetGlobalPath）:
//     Func(p) CmpOp Value（現状 Func="length" のみ）。例: length(p) <= 4
type VarLengthFilter struct {
	Phase  FilterPhase
	Target FilterTarget

	// 要素述語用
	Quant   ConditionType  // CondAll/CondAny/CondNone/CondSingle
	IterVar string         // 反復変数（例 "x"）
	Cond    *ConditionNode // 内側述語（IterVar.prop OP val）

	// 大域述語用（TargetGlobalPath）
	Func  string        // "length" 等のパス関数
	CmpOp ConditionType // length(p) との比較子
	Value string        // 比較値
}

// String は VarLengthFilter を決定的にダンプする（golden/物理テスト用）。
func (v VarLengthFilter) String() string {
	if v.Target == TargetGlobalPath {
		return fmt.Sprintf("%s(p) %s %s", v.Func, cmpSymbol(v.CmpOp), v.Value)
	}
	list := "nodes(p)"
	if v.Target == TargetRelationship {
		list = "relationships(p)"
	}
	inner := ""
	if v.Cond != nil {
		inner = v.Cond.Print()
	}
	return fmt.Sprintf("%s(%s IN %s WHERE %s)", v.Quant.CypherQuant(), v.IterVar, list, inner)
}

// ================================
// filter用helper関数
// ================================

// 外側のandを分解する関数 今回のシステムではandしかこない　andのfilterは左右で実行順序を入れ替えても結果は変わらない
func DecomposeOuterAndOp(rootCond *ConditionNode) []*ConditionNode {
	if rootCond == nil {
		return nil
	}
	// 元のroot conditionを要素が一つだけの配列に変換して入力とする
	var CondNodeList []*ConditionNode
	//
	switch rootCond.Type {
	case CondAnd:
		CondNodeList = append(CondNodeList, DecomposeOuterAndOp(rootCond.Left)...)
		CondNodeList = append(CondNodeList, DecomposeOuterAndOp(rootCond.Right)...)
	case CondParen:
		CondNodeList = append(CondNodeList, DecomposeOuterAndOp(rootCond.Child)...)
	default:
		CondNodeList = append(CondNodeList, rootCond)
	}

	return CondNodeList
}

func ConnectCondNode(conditions []*ConditionNode) *ConditionNode {
	count := len(conditions)

	if count == 0 {
		return nil
	} else if count == 1 {
		return conditions[0]
	}

	root := conditions[0]

	for i := 1; i < count; i++ {
		root = &ConditionNode{
			Type:  CondAnd,
			Left:  root,
			Right: conditions[i],
		}
	}

	return root
}
