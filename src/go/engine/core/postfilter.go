package core

import (
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// postfilter.go は PostFilter（クロスエイリアス比較 a.id <op> b.id）のコーディネータ評価を
// 全エンジンで共有する。中間 Record は束縛 UUID（スロット）のみを保持するため、比較は
// スロットの UUID 同士を id.Compare で行う（プロパティ材料化は不要）。現状 Property は
// identity（uuid）を前提とする。

// EvalVarCmp は cond（Type==CondVarCmp）を Record 1 行に対し評価する。
// inSlot は Record のスロット割り当て（alias→index）。
func EvalVarCmp(cond *plan.ConditionNode, rec Record, inSlot plan.SlotTable) bool {
	if cond == nil {
		return true
	}
	li, lok := inSlot.VarToSlot[cond.Alias]
	ri, rok := inSlot.VarToSlot[cond.RAlias]
	if !lok || !rok || li >= len(rec.Slots) || ri >= len(rec.Slots) {
		return false
	}
	cmp := id.Compare(rec.Slots[li], rec.Slots[ri])
	switch cond.CmpOp {
	case plan.CondEq:
		return cmp == 0
	case plan.CondNeq:
		return cmp != 0
	case plan.CondGreater:
		return cmp > 0
	case plan.CondLess:
		return cmp < 0
	case plan.CondGreaterEq:
		return cmp >= 0
	case plan.CondLessEq:
		return cmp <= 0
	default:
		return false
	}
}
