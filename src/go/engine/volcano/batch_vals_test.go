package volcano

import (
	"testing"

	uid "polystore_database/src/go/id"
)

// Batch の carried 値（vals）の backfill 挙動を検証する（appendRow で始まり appendRowVals が入る場合、
// 既存行が nil で backfill され整合が保たれること）。
func TestBatchValsBackfill(t *testing.T) {
	b := newBatch(1, 4)
	b.appendRow([]uid.UUID{uid.UUID("a")})                          // vals 無し
	b.appendRowVals([]uid.UUID{uid.UUID("b")}, map[string]any{"c": int64(5)}) // ここで backfill
	b.appendRow([]uid.UUID{uid.UUID("d")})                          // vals 有り batch に nil 追加
	if b.n != 3 {
		t.Fatalf("n=%d want 3", b.n)
	}
	if b.valsAt(0) != nil {
		t.Errorf("row0 vals should be nil (backfilled), got %v", b.valsAt(0))
	}
	if b.valsAt(1)["c"] != int64(5) {
		t.Errorf("row1 vals[c]=%v want 5", b.valsAt(1))
	}
	if b.valsAt(2) != nil {
		t.Errorf("row2 vals should be nil, got %v", b.valsAt(2))
	}
}

// carried 値を一切使わない場合、vals は nil のまま（ゼロ負担）。
func TestBatchValsNilWhenUnused(t *testing.T) {
	b := newBatch(1, 2)
	b.appendRow([]uid.UUID{uid.UUID("x")})
	if b.vals != nil {
		t.Errorf("vals should stay nil when unused, got %v", b.vals)
	}
	if b.valsAt(0) != nil {
		t.Errorf("valsAt should be nil")
	}
}
