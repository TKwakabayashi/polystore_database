package core

import "testing"

func TestVarLengthRange(t *testing.T) {
	cases := []struct {
		min, max int
		want     string
	}{
		{0, -1, "*0.."}, // `*0..`（上限なし）: 以前 `*0..-1` を生成して IS6 が SyntaxError だった回帰ケース
		{1, -1, "*1.."}, // 下限あり・上限なし
		{-1, 3, "*..3"}, // 下限なし・上限あり
		{-1, -1, "*"},   // 無制限
		{2, 2, "*2"},    // 固定長
		{1, 3, "*1..3"}, // 範囲
	}
	for _, c := range cases {
		if got := VarLengthRange(c.min, c.max); got != c.want {
			t.Errorf("VarLengthRange(%d,%d) = %q, want %q", c.min, c.max, got, c.want)
		}
	}
}
