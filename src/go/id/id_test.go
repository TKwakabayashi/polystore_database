package id

import "testing"

// TestCompareIsByteOrder は「Compare ＝ Bytes() の順序」不変条件を固定する。
// 内部表現を将来変える際、この規約を破ると shard 境界が取りこぼし/重複を起こす。
func TestCompareIsByteOrder(t *testing.T) {
	cases := []struct {
		a, b UUID
		want int // 符号
	}{
		{"", "a", -1},
		{"a", "a", 0},
		{"b", "a", 1},
		{"N-1", "N-2", -1},
		{"aaa", "aab", -1},
	}
	for _, c := range cases {
		got := Compare(c.a, c.b)
		if sign(got) != c.want {
			t.Errorf("Compare(%q,%q)=%d, want sign %d", c.a, c.b, got, c.want)
		}
		if Less(c.a, c.b) != (c.want < 0) {
			t.Errorf("Less(%q,%q)=%v, want %v", c.a, c.b, Less(c.a, c.b), c.want < 0)
		}
	}
}

// TestRoundTrip は境界変換の往復（Parse/FromBytes/FromAny → String/Bytes）が一致することを確認。
func TestRoundTrip(t *testing.T) {
	s := "N-abc123"
	u, _ := Parse(s)
	if u.String() != s {
		t.Errorf("Parse/String round-trip: got %q, want %q", u.String(), s)
	}
	if string(u.Bytes()) != s {
		t.Errorf("Bytes: got %q, want %q", u.Bytes(), s)
	}
	if b, _ := FromBytes([]byte(s)); b != u {
		t.Errorf("FromBytes: got %q, want %q", b, u)
	}
	if FromAny(s) != u || FromAny([]byte(s)) != u || FromAny(u) != u {
		t.Errorf("FromAny should accept string/[]byte/UUID")
	}
	if !FromAny(42).Empty() {
		t.Errorf("FromAny(non-string) should be empty")
	}
}

// TestShardBoundsCover は ShardBounds が Compare 順で全空間を過不足なく被覆することを確認する
// （隣接シャードの hi と次の lo が一致＝境界に穴も重なりもない）。
func TestShardBoundsCover(t *testing.T) {
	for _, total := range []int{1, 2, 3, 4, 8, 16} {
		lo0, _ := ShardBounds(0, total)
		if !lo0.Empty() {
			t.Errorf("total=%d: shard0 lo should be empty (下限なし), got %q", total, lo0)
		}
		_, hiLast := ShardBounds(total-1, total)
		if !hiLast.Empty() {
			t.Errorf("total=%d: last shard hi should be empty (上限なし), got %q", total, hiLast)
		}
		for i := 0; i < total-1; i++ {
			_, hi := ShardBounds(i, total)
			loNext, _ := ShardBounds(i+1, total)
			if Compare(hi, loNext) != 0 {
				t.Errorf("total=%d: shard %d hi=%x != shard %d lo=%x (境界不連続)", total, i, hi.Bytes(), i+1, loNext.Bytes())
			}
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
