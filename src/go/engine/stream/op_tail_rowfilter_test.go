package stream

import (
	"testing"

	"polystore_database/src/go/plan"
)

// streamRowFilter を DB 無しで検証する（post-WITH WHERE 相当）。合成 Row チャネルに
// c > 4 の Expr を適用し、通過行だけが出ることを確認する。
func TestStreamRowFilter(t *testing.T) {
	in := make(chan []Row, 1)
	in <- []Row{
		{"friend": "a", "c": int64(3)},
		{"friend": "b", "c": int64(5)},
		{"friend": "d", "c": int64(9)},
	}
	close(in)

	out := make(chan []Row, 1)
	cond := &plan.Bin{Op: plan.OpGt, L: &plan.Column{Key: "c"}, R: &plan.Lit{Val: int64(4)}}
	n := streamRowFilter(&plan.RowFilter{Cond: cond}, in, out)
	close(out)

	if n != 2 {
		t.Fatalf("emitted=%d want 2", n)
	}
	var got []Row
	for batch := range out {
		got = append(got, batch...)
	}
	if len(got) != 2 || got[0]["friend"] != "b" || got[1]["friend"] != "d" {
		t.Errorf("unexpected survivors: %+v", got)
	}
}
