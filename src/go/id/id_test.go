package id

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"
)

const (
	nodeText = "N-f57704b0-7e1b-4cb6-b7ba-a28ba57c2f51"
	edgeText = "E-b2c50fa4-82bf-446c-a1a8-a2220fdbba73"
)

func mustParse(t *testing.T, s string) UUID {
	t.Helper()
	u, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// TestRoundTrip は境界変換の往復（文字列・バイナリ・FromAny）が一致することを確認する。
func TestRoundTrip(t *testing.T) {
	for _, s := range []string{nodeText, edgeText} {
		u := mustParse(t, s)
		if u.String() != s {
			t.Errorf("Parse/String: got %q, want %q", u.String(), s)
		}
		if len(u.Bytes()) != Size || u.Bytes()[0] != s[0] {
			t.Errorf("Bytes: %x", u.Bytes())
		}
		if b, err := FromBytes(u.Bytes()); err != nil || b != u {
			t.Errorf("FromBytes(binary): %v %v", b, err)
		}
		if b, err := FromBytes([]byte(s)); err != nil || b != u {
			t.Errorf("FromBytes(text): %v %v", b, err)
		}
		if FromAny(s) != u || FromAny([]byte(s)) != u || FromAny(u.Bytes()) != u || FromAny(u) != u {
			t.Errorf("FromAny should accept text / text bytes / binary / UUID")
		}
	}
	if !FromAny(42).Empty() || !FromAny("p1").Empty() {
		t.Errorf("解釈できない値は空")
	}
	if u, err := Parse(""); err != nil || !u.Empty() || u.String() != "" || u.Bytes() != nil {
		t.Errorf("空文字列は空 UUID")
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	for _, s := range []string{
		"p1",
		"X-f57704b0-7e1b-4cb6-b7ba-a28ba57c2f51", // 種別
		"Nf57704b0-7e1b-4cb6-b7ba-a28ba57c2f51-", // 区切り
		"N-f57704b0x7e1b-4cb6-b7ba-a28ba57c2f51", // ハイフンの位置
		"N-g57704b0-7e1b-4cb6-b7ba-a28ba57c2f51", // 16 進でない
		"N-f57704b0-7e1b-4cb6-b7ba-a28ba57c2f5",  // 短い
	} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) はエラーになるべき", s)
		}
	}
}

// TestCompareIsTextOrder は「Compare（内部表現のバイト順）＝ 文字列表現の辞書順」を固定する。
// 保存形式（文字列 / バイナリ）を切り替えても、ストア側の並び・範囲が同じ順序になるための不変条件。
func TestCompareIsTextOrder(t *testing.T) {
	var us []UUID
	for i := 0; i < 200; i++ {
		k := KindNode
		if i%3 == 0 {
			k = KindEdge
		}
		us = append(us, New(k))
	}
	us = append(us, UUID{})
	byBytes := append([]UUID{}, us...)
	sort.Slice(byBytes, func(i, j int) bool { return Less(byBytes[i], byBytes[j]) })
	byText := append([]UUID{}, us...)
	sort.Slice(byText, func(i, j int) bool { return byText[i].String() < byText[j].String() })
	for i := range byBytes {
		if byBytes[i] != byText[i] {
			t.Fatalf("位置 %d で順序が違う: %s vs %s", i, byBytes[i], byText[i])
		}
	}
	if !byBytes[0].Empty() {
		t.Errorf("空 UUID が最小であるべき")
	}
}

func TestNewAndMake(t *testing.T) {
	a, b := New(KindNode), New(KindNode)
	if a == b || a.Kind() != KindNode || New(KindEdge).Kind() != KindEdge {
		t.Errorf("New: %s %s", a, b)
	}
	if a.String()[14+2] != '4' { // version 4（3 つ目のブロックの先頭）
		t.Errorf("version: %s", a)
	}
	var raw [16]byte
	raw[15] = 1
	if got := Make(KindEdge, raw).String(); got != "E-00000000-0000-0000-0000-000000000001" {
		t.Errorf("Make: %s", got)
	}
}

func TestJSON(t *testing.T) {
	u := mustParse(t, nodeText)
	b, err := json.Marshal(map[string]UUID{"u": u})
	if err != nil || string(b) != `{"u":"`+nodeText+`"}` {
		t.Fatalf("Marshal: %s %v", b, err)
	}
	var back map[string]UUID
	if err := json.Unmarshal(b, &back); err != nil || back["u"] != u {
		t.Errorf("Unmarshal: %v %v", back, err)
	}
}

// TestShardBoundsCover は ShardBounds が先頭バイト空間を過不足なく被覆することを確認する
// （隣接シャードの hi と次の lo が一致＝境界に穴も重なりもない）。
func TestShardBoundsCover(t *testing.T) {
	for _, total := range []int{1, 2, 3, 4, 8, 16} {
		if lo0, _ := ShardBounds(0, total); lo0 != nil {
			t.Errorf("total=%d: shard0 lo should be nil (下限なし), got %x", total, lo0)
		}
		if _, hiLast := ShardBounds(total-1, total); hiLast != nil {
			t.Errorf("total=%d: last shard hi should be nil (上限なし), got %x", total, hiLast)
		}
		for i := 0; i < total-1; i++ {
			_, hi := ShardBounds(i, total)
			loNext, _ := ShardBounds(i+1, total)
			if !bytes.Equal(hi, loNext) {
				t.Errorf("total=%d: shard %d hi=%x != shard %d lo=%x (境界不連続)", total, i, hi, i+1, loNext)
			}
		}
	}
}
