package mysql

import (
	"testing"
	"time"
)

func TestAnalyzeRoot(t *testing.T) {
	for out, want := range map[string]struct {
		d    time.Duration
		rows int64
	}{
		"-> Filter: (Person.id = 1)  (cost=1136 rows=1064) (actual time=1.24..1.81 rows=1 loops=1)\n    -> Table scan (actual time=0.02..1.45 rows=10295 loops=1)": {1810 * time.Microsecond, 1},
		"-> Rows fetched before execution  (cost=0..0 rows=1) (actual time=250e-6..250e-6 rows=1 loops=1)":                                                         {250 * time.Nanosecond, 1},
		"-> Zero rows (no matching row)  (cost=0..0 rows=0)":                                                                                                       {0, 0},
	} {
		d, rows := analyzeRoot(out)
		if d != want.d || rows != want.rows {
			t.Errorf("%q: got %v %d, want %v %d", out, d, rows, want.d, want.rows)
		}
	}
}
