package core

import (
	"testing"

	"polystore_database/src/go/plan"
)

func TestBuildStageRecords(t *testing.T) {
	outSlot := plan.SlotTable{VarToSlot: map[string]int{"friend": 0}, SlotToVar: []string{"friend"}}
	rows := []map[string]any{
		{"friend": "0001", "c": int64(3)},
		{"friend": "0002", "c": int64(5)},
	}
	recs := BuildStageRecords(rows, []string{"friend"}, []string{"c"}, outSlot)
	if len(recs) != 2 {
		t.Fatalf("want 2 records, got %d", len(recs))
	}
	if recs[0].Slots[0].String() != "0001" {
		t.Errorf("bind slot: got %q want 0001", recs[0].Slots[0].String())
	}
	if recs[0].Vals["c"] != int64(3) {
		t.Errorf("carried val: got %v want 3", recs[0].Vals["c"])
	}
	if recs[1].Vals["c"] != int64(5) {
		t.Errorf("carried val row2: got %v want 5", recs[1].Vals["c"])
	}
}

func TestBuildStageRecordsNoVals(t *testing.T) {
	outSlot := plan.SlotTable{VarToSlot: map[string]int{"p": 0}, SlotToVar: []string{"p"}}
	recs := BuildStageRecords([]map[string]any{{"p": "abc"}}, []string{"p"}, nil, outSlot)
	if len(recs) != 1 || recs[0].Vals != nil {
		t.Fatalf("expected 1 record with nil Vals, got %+v", recs)
	}
}

func TestMergeVals(t *testing.T) {
	row := map[string]any{"friend.id": int64(7)}
	MergeVals(row, map[string]any{"c": int64(3), "xyCount": int64(9)})
	if row["c"] != int64(3) || row["xyCount"] != int64(9) || row["friend.id"] != int64(7) {
		t.Errorf("merge failed: %+v", row)
	}
}
