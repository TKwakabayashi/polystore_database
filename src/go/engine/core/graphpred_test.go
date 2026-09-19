package core

import (
	"testing"

	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

func TestBuildCypherPredicateInIsNull(t *testing.T) {
	cases := []struct {
		name string
		cond *plan.ConditionNode
		want string
	}{
		{"in", &plan.ConditionNode{Type: plan.CondIn, Alias: "o", Property: "type", Values: []string{"Company", "University"}, DataType: "string"}, "o.type IN $val0"},
		{"isnull", &plan.ConditionNode{Type: plan.CondIsNull, Alias: "m", Property: "content"}, "m.content IS NULL"},
		{"isnotnull", &plan.ConditionNode{Type: plan.CondIsNotNull, Alias: "m", Property: "content"}, "m.content IS NOT NULL"},
		{"cmp", &plan.ConditionNode{Type: plan.CondGreater, Alias: "m", Property: "length", Value: "5", DataType: "int"}, "m.length > $val0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params := map[string]interface{}{}
			seq := 0
			got, err := BuildCypherPredicate(c.cond.Alias, c.cond, params, &seq)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestBuildPathFilterClauses(t *testing.T) {
	filters := []plan.VarLengthFilter{
		{Target: plan.TargetRelationship, Quant: plan.CondAll, IterVar: "r",
			Cond: &plan.ConditionNode{Type: plan.CondGreater, Alias: "r", Property: "creationDate", Value: "2010", DataType: "string"}},
		{Target: plan.TargetGlobalPath, Func: "length", CmpOp: plan.CondLessEq, Value: "3"},
	}
	params := map[string]interface{}{}
	seq := 0
	got, err := BuildPathFilterClauses("vp", filters, params, &seq)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 clauses, got %d: %v", len(got), got)
	}
	if got[0] != "all(r IN relationships(vp) WHERE r.creationDate > $val0)" {
		t.Errorf("element clause: %q", got[0])
	}
	if got[1] != "length(vp) <= $val1" {
		t.Errorf("global clause: %q", got[1])
	}
}

func TestEvalVarCmp(t *testing.T) {
	inSlot := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "b": 1}}
	rec := Record{Slots: []id.UUID{id.UUID("0001"), id.UUID("0002")}}
	tests := []struct {
		op   plan.ConditionType
		want bool
	}{
		{plan.CondLess, true}, {plan.CondGreater, false}, {plan.CondNeq, true}, {plan.CondEq, false}, {plan.CondLessEq, true},
	}
	for _, tt := range tests {
		cond := &plan.ConditionNode{Type: plan.CondVarCmp, Alias: "a", Property: "id", RAlias: "b", RProperty: "id", CmpOp: tt.op}
		if got := EvalVarCmp(cond, rec, inSlot); got != tt.want {
			t.Errorf("op %v: got %v want %v", tt.op, got, tt.want)
		}
	}
}
