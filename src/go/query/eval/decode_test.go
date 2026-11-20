package eval

import (
	"testing"
	"time"
)

// ParseDateString returns the same as trying the three layouts in order.
func TestParseDateStringMatchesLayouts(t *testing.T) {
	old := func(s string) (time.Time, bool) {
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999", "2006-01-02"} {
			if tm, err := time.Parse(layout, s); err == nil {
				return tm.UTC(), true
			}
		}
		return time.Time{}, false
	}
	for _, s := range []string{
		"2012-02-22 14:05:58.785000", "2012-02-22 14:05:58", "2012-02-22", "2012-02-22T14:05:58.785Z",
		"2012-02-22T14:05:58+09:00", "1985-01-01", "not a date", "", "2012-02-22X",
	} {
		g, gok := ParseDateString(s)
		w, wok := old(s)
		if gok != wok || !g.Equal(w) {
			t.Errorf("%q: got %v,%v want %v,%v", s, g, gok, w, wok)
		}
	}
}
