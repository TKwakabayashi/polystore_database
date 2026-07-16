package bulk_executor

import (
	"testing"

	"polystore_database/src/go/plan"
)

func TestCompareValues(t *testing.T) {
	if compareValues(1, 2) >= 0 {
		t.Errorf("compareValues(1,2) は負であるべき")
	}
	if compareValues("b", "a") <= 0 {
		t.Errorf("compareValues(\"b\",\"a\") は正であるべき")
	}
	if compareValues(nil, 1) != -1 {
		t.Errorf("compareValues(nil,1) は -1 であるべき")
	}
	if compareValues(nil, nil) != 0 {
		t.Errorf("compareValues(nil,nil) は 0 であるべき")
	}
}

func TestToInt64(t *testing.T) {
	if toInt64(int32(5)) != 5 || toInt64(int64(7)) != 7 || toInt64(9) != 9 {
		t.Errorf("toInt64 の変換が不正")
	}
}

func TestProjectRow(t *testing.T) {
	r := Record{Slots: []string{"u1"}}
	items := []plan.ReturnItem{{Name: "name", Alias: "a", Props: []string{"name"}}}
	aliasToSlot := map[string]int{"a": 0}
	cache := map[string]map[string]map[string]interface{}{
		"a": {"u1": {"name": "Alice"}},
	}
	row := ProjectRow(r, items, aliasToSlot, cache)
	if row["name"] != "Alice" {
		t.Errorf("ProjectRow = %v, want name=Alice", row)
	}
}

func TestProjectRowCoalesce(t *testing.T) {
	r := Record{Slots: []string{"u1"}}
	items := []plan.ReturnItem{{Name: "v", Alias: "a", Props: []string{"x", "y"}, IsCoalesce: true}}
	aliasToSlot := map[string]int{"a": 0}
	cache := map[string]map[string]map[string]interface{}{
		"a": {"u1": {"y": 42}},
	}
	row := ProjectRow(r, items, aliasToSlot, cache)
	if row["v"] != 42 {
		t.Errorf("ProjectRow coalesce = %v, want v=42", row)
	}
}
