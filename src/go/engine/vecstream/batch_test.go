package vecstream

import (
	"testing"

	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// DB を必要としない純ロジック（Batch / split / remap）の単体テスト。
// 実 DB での 3 モデル結果一致・往復回数の検証は bench 経由で行う。

func TestBatchAppendAndRow(t *testing.T) {
	b := newBatch(3, 2)
	b.appendRow([]uid.UUID{"a0", "a1", "a2"})
	b.appendRow([]uid.UUID{"b0", "b1"}) // 短い行は残スロットが空

	if b.n != 2 {
		t.Fatalf("n = %d, want 2", b.n)
	}
	if got := b.get(0, 2); got != uid.UUID("a2") {
		t.Errorf("get(0,2) = %q, want a2", got)
	}
	if got := b.get(1, 2); got != uid.UUID("") {
		t.Errorf("get(1,2) = %q, want empty", got)
	}
	row := b.row(0)
	if len(row) != 3 || row[0] != "a0" || row[1] != "a1" {
		t.Errorf("row(0) = %v", row)
	}
}

func TestSplit(t *testing.T) {
	mk := func(n int) *Batch {
		b := newBatch(1, n)
		for i := 0; i < n; i++ {
			b.appendRow([]uid.UUID{uid.UUID(string(rune('A' + i)))})
		}
		return b
	}

	// nil / 空は何も返さない。
	if got := split(nil, 4); got != nil {
		t.Errorf("split(nil) = %v, want nil", got)
	}
	if got := split(newBatch(1, 0), 4); got != nil {
		t.Errorf("split(empty) = %v, want nil", got)
	}
	// width 以下は 1 個のまま。
	if got := split(mk(3), 4); len(got) != 1 || got[0].n != 3 {
		t.Fatalf("split(3,4) = %d batches", len(got))
	}
	// 端数あり: 10 を 4 幅 → 4,4,2。
	got := split(mk(10), 4)
	if len(got) != 3 {
		t.Fatalf("split(10,4) = %d batches, want 3", len(got))
	}
	wantN := []int{4, 4, 2}
	total := 0
	for i, sb := range got {
		if sb.n != wantN[i] {
			t.Errorf("batch[%d].n = %d, want %d", i, sb.n, wantN[i])
		}
		total += sb.n
	}
	if total != 10 {
		t.Errorf("total rows = %d, want 10", total)
	}
	// width<1 は 1 とみなす。
	if got := split(mk(3), 0); len(got) != 3 {
		t.Errorf("split(3,0) = %d batches, want 3", len(got))
	}
}

func TestRemapCarriesSharedAliases(t *testing.T) {
	in := plan.SlotTable{VarToSlot: map[string]int{"a": 0, "b": 1}}
	out := plan.SlotTable{VarToSlot: map[string]int{"b": 0, "c": 1}} // a は落ち、c は新規

	got := remap([]uid.UUID{"va", "vb"}, in, out)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != uid.UUID("vb") {
		t.Errorf("out slot for b = %q, want vb", got[0])
	}
	if got[1] != uid.UUID("") {
		t.Errorf("out slot for c = %q, want empty", got[1])
	}
}
