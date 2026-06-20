package plan

import (
	"fmt"
)

type ConditionNode struct {
	Type      ConditionType
	DataStore string // assign at BuildPlan

	Labels   []string
	Alias    string
	Property string
	Value    string
	ObjType  ObjectType
	DataType string // int, string etc.

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
	case CondNot:
		return fmt.Sprintf("NOT %s", n.Child.Print())
	case CondParen:
		return fmt.Sprintf("(%s)", n.Child.Print())
	default:
		return "Invalid Condition"
	}
}

type VarLengthFilter struct {
	Phase  FilterPhase
	Target FilterTarget
	Filter *Filter
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
