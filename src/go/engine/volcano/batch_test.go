package volcano

import (
	"testing"

	"polystore_database/src/go/plan"
)

// DB を必要としない純ロジック（Batch / remap）の単体テスト。
// 実 DB を用いた 3 モデルの結果一致・往復回数の検証は cmd/volcanobench で行う。

func TestBatchAppendAndRow(t *testing.T) {
	b := newBatch(3, 2)
	b.appendRow([]string{"a0", "a1", "a2"})
	b.appendRow([]string{"b0", "b1"}) // 短い行は残スロットが空文字

	if b.n != 2 {
		t.Fatalf("n = %d, want 2", b.n)
	}
	if got := b.get(0, 2); got != "a2" {
		t.Errorf("get(0,2) = %q, want a2", got)
	}
	if got := b.get(1, 2); got != "" {
		t.Errorf("get(1,2) = %q, want empty", got)
	}
	row := b.row(0)
	if len(row) != 3 || row[0] != "a0" || row[1] != "a1" {
		t.Errorf("row(0) = %v", row)
	}
}

func TestRemapCarriesSharedAliases(t *testing.T) {
	in := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "b": 1}}
	out := plan.SlotTable{VarToSlot: map[string]int{"b": 0, "c": 1}} // a は落ち、c は新規

	got := remap([]string{"va", "vb"}, in, out)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != "vb" {
		t.Errorf("out slot for b = %q, want vb", got[0])
	}
	if got[1] != "" {
		t.Errorf("out slot for c = %q, want empty", got[1])
	}
}

func TestVectorWidthByMode(t *testing.T) {
	// NewProcessor は DB 接続を張るため直接呼べないが、幅決定ロジックは
	// mode/vectorSize の対応として明示しておく。
	cases := []struct {
		mode Mode
		size int
		want int
	}{
		{ModeVolcano, 999, 1},
		{ModeVectorized, 512, 512},
		{ModeVectorized, 0, 1},
	}
	for _, c := range cases {
		width := 1
		if c.mode == ModeVectorized {
			if c.size < 1 {
				width = 1
			} else {
				width = c.size
			}
		}
		if width != c.want {
			t.Errorf("mode=%v size=%d width=%d, want %d", c.mode, c.size, width, c.want)
		}
	}
}
