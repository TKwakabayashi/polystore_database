package bulk_executor

import (
	"testing"

	"polystore_database/src/go/plan"
)

func TestUniqueSlot(t *testing.T) {
	in := []Record{{"a"}, {"b"}, {"a"}, {"c"}, {"b"}}
	got := uniqueSlot(in, 0)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("uniqueSlot len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("uniqueSlot[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUniqueSlotShortRow(t *testing.T) {
	// idx が行長を超える行はスキップされる。
	in := []Record{{"x", "y"}, {"z"}}
	got := uniqueSlot(in, 1)
	if len(got) != 1 || got[0] != "y" {
		t.Errorf("uniqueSlot short row = %v, want [y]", got)
	}
}

func TestRemap(t *testing.T) {
	inSlot := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "b": 1}}
	outSlot := plan.SlotTable{VarToSlot: map[string]int{"b": 0, "a": 1}}
	got := remap([]string{"va", "vb"}, inSlot, outSlot)
	if len(got) != 2 || got[0] != "vb" || got[1] != "va" {
		t.Errorf("remap = %v, want [vb va]", got)
	}
}

func TestRemapDropsUnmappedAlias(t *testing.T) {
	// 出力に無い alias は落ち、入力に無い alias は空文字のまま。
	inSlot := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "b": 1}}
	outSlot := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "c": 1}}
	got := remap([]string{"va", "vb"}, inSlot, outSlot)
	if len(got) != 2 || got[0] != "va" || got[1] != "" {
		t.Errorf("remap = %v, want [va \"\"]", got)
	}
}

func TestMaterializeRaw(t *testing.T) {
	p := &Processor{}
	recs := []Record{{"u1", "u2"}}
	out := p.materializeRaw(recs)
	if len(out) != 1 {
		t.Fatalf("materializeRaw len = %d, want 1", len(out))
	}
	if out[0]["slot0"] != "u1" || out[0]["slot1"] != "u2" {
		t.Errorf("materializeRaw row = %v", out[0])
	}
}
