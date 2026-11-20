package id

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"sort"
	"sync"
	"testing"
)

const (
	nodeText = "f57704b0-7e1b-4cb6-b7ba-a28ba57c2f51"
	edgeText = "b2c50fa4-82bf-446c-a1a8-a2220fdbba73"
)

func mustParse(t *testing.T, s string) UUID {
	t.Helper()
	u, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// TestRoundTrip checks that boundary conversions (text, binary, FromAny) round-trip.
func TestRoundTrip(t *testing.T) {
	for _, s := range []string{nodeText, edgeText} {
		u := mustParse(t, s)
		if u.String() != s {
			t.Errorf("Parse/String: got %q, want %q", u.String(), s)
		}
		if len(u.Bytes()) != Size {
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
		t.Errorf("an unreadable value is empty")
	}
	if u, err := Parse(""); err != nil || !u.Empty() || u.String() != "" || u.Bytes() != nil {
		t.Errorf("the empty string is the empty UUID")
	}
}

// TestReadLegacy checks that the legacy forms (kinded "N-..." / "E-..." and 17 bytes) are read with the kind dropped
// (so uuid-format can convert existing data).
func TestReadLegacy(t *testing.T) {
	u := mustParse(t, nodeText)
	for _, k := range []byte{'N', 'E'} {
		text := string(k) + "-" + nodeText
		bin := append([]byte{k}, u.Bytes()...)
		for _, v := range []any{text, []byte(text), bin, string(bin)} {
			if FromAny(v) != u {
				t.Errorf("legacy form %T %q: got %v", v, v, FromAny(v))
			}
		}
	}
	if _, err := FromBytes(append([]byte{'X'}, u.Bytes()...)); err == nil {
		t.Errorf("17 bytes whose first byte is not a kind are an error")
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	for _, s := range []string{
		"p1",
		"X-f57704b0-7e1b-4cb6-b7ba-a28ba57c2f51", // not a legacy kind prefix
		"f57704b0x7e1b-4cb6-b7ba-a28ba57c2f51",   // hyphen position
		"g57704b0-7e1b-4cb6-b7ba-a28ba57c2f51",   // not hex
		"f57704b0-7e1b-4cb6-b7ba-a28ba57c2f5",    // too short
	} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) must fail", s)
		}
	}
}

// TestCompareIsTextOrder pins "Compare (byte order of the internal form) = lexical order of the text form", the
// invariant that keeps store-side order and ranges the same in both storage forms.
func TestCompareIsTextOrder(t *testing.T) {
	var us []UUID
	a, _ := NewAllocator(3)
	for i := 0; i < 200; i++ {
		if i%3 == 0 {
			u, _ := a.Next()
			us = append(us, u)
		} else {
			us = append(us, New())
		}
	}
	us = append(us, UUID{})
	byBytes := append([]UUID{}, us...)
	sort.Slice(byBytes, func(i, j int) bool { return Less(byBytes[i], byBytes[j]) })
	byText := append([]UUID{}, us...)
	sort.Slice(byText, func(i, j int) bool { return byText[i].String() < byText[j].String() })
	for i := range byBytes {
		if byBytes[i] != byText[i] {
			t.Fatalf("order differs at %d: %s vs %s", i, byBytes[i], byText[i])
		}
	}
	if !byBytes[0].Empty() {
		t.Errorf("the empty UUID must be the smallest")
	}
}

func TestNewAndMake(t *testing.T) {
	a, b := New(), New()
	if a == b {
		t.Errorf("New: %s %s", a, b)
	}
	if a.String()[14] != '4' { // version 4 (start of the third group)
		t.Errorf("version: %s", a)
	}
	var raw [16]byte
	raw[15] = 1
	if got := Make(raw).String(); got != "00000000-0000-0000-0000-000000000001" {
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

// TestShardBoundsCover checks that ShardBounds covers the leading-byte space exactly (each shard's hi is the
// next shard's lo: no holes, no overlaps).
func TestShardBoundsCover(t *testing.T) {
	for _, total := range []int{1, 2, 3, 4, 8, 16} {
		if lo0, _ := ShardBounds(0, total); lo0 != nil {
			t.Errorf("total=%d: shard0 lo should be nil (no lower bound), got %x", total, lo0)
		}
		if _, hiLast := ShardBounds(total-1, total); hiLast != nil {
			t.Errorf("total=%d: last shard hi should be nil (no upper bound), got %x", total, hiLast)
		}
		for i := 0; i < total-1; i++ {
			_, hi := ShardBounds(i, total)
			loNext, _ := ShardBounds(i+1, total)
			if !bytes.Equal(hi, loNext) {
				t.Errorf("total=%d: shard %d hi=%x != shard %d lo=%x (bounds not contiguous)", total, i, hi, i+1, loNext)
			}
		}
	}
}

func withFormat(t *testing.T, f Storage) {
	t.Helper()
	prev := CurrentStorage()
	SetStorage(f)
	t.Cleanup(func() { SetStorage(prev) })
}

func TestStoredFormats(t *testing.T) {
	u := New()
	withFormat(t, StorageText)
	if s, ok := Stored(u).(string); !ok || s != u.String() || StoredLen() != TextSize || string(StoredKey(u)) != u.String() {
		t.Errorf("text: %v", Stored(u))
	}
	if l, ok := StoredSlice([]UUID{u}).([]string); !ok || l[0] != u.String() {
		t.Errorf("text slice: %v", StoredSlice([]UUID{u}))
	}
	withFormat(t, StorageBinary)
	if b, ok := Stored(u).([]byte); !ok || !bytes.Equal(b, u.Bytes()) || StoredLen() != Size || !bytes.Equal(StoredKey(u), u.Bytes()) {
		t.Errorf("binary: %v", Stored(u))
	}
	if l, ok := StoredSlice([]UUID{u}).([][]byte); !ok || !bytes.Equal(l[0], u.Bytes()) {
		t.Errorf("binary slice: %v", StoredSlice([]UUID{u}))
	}
	if Stored(UUID{}) != nil {
		t.Errorf("empty is nil")
	}
	if b, ok := StoredBound([]byte{0x80}).([]byte); !ok || len(b) != Size || b[0] != 0x80 {
		t.Errorf("binary bound: %v", StoredBound([]byte{0x80}))
	}
}

// Values written in either form are read whatever the current storage is (during a conversion, or between stores in different forms).
func TestReadEitherFormat(t *testing.T) {
	u := New()
	for _, f := range []Storage{StorageText, StorageBinary} {
		withFormat(t, f)
		for _, v := range []any{u.String(), []byte(u.String()), u.Bytes(), string(u.Bytes())} {
			if FromAny(v) != u {
				t.Errorf("format %d: FromAny(%T %v) = %v", f, v, v, FromAny(v))
			}
		}
	}
}

// unpack inverts pack.
func unpack(b UUID) uint64 {
	var v uint64
	for i := 0; i < 6; i++ {
		v |= uint64(b[i]) << (56 - 8*i)
	}
	v |= uint64(b[6]&0x0f) << 12
	v |= uint64(b[7]) << 4
	v |= uint64(b[8] & 0x0f)
	return v
}

// unmix inverts mix (right xorshifts by iteration, odd multiplications by the inverse modulo 2^64).
func unmix(z uint64) uint64 {
	unshift := func(x uint64, s uint) uint64 {
		y := x
		for i := s; i < 64; i += s {
			y = x ^ y>>s
		}
		return y
	}
	inv := func(c uint64) uint64 {
		x := c // Newton's method: for odd c, starting from x = c each step doubles the correct bits
		for i := 0; i < 6; i++ {
			x *= 2 - c*x
		}
		return x
	}
	z = unshift(z, 31)
	z *= inv(0x94d049bb133111eb)
	z = unshift(z, 27)
	z *= inv(0xbf58476d1ce4e5b9)
	z = unshift(z, 30)
	return z
}

// TestAllocBijective checks that mix and pack have inverses (injective: different pairs give different values).
func TestAllocBijective(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for _, v := range append([]uint64{0, 1, 1<<64 - 1, 1 << SeqBits}, func() []uint64 {
		var vs []uint64
		for i := 0; i < 10000; i++ {
			vs = append(vs, r.Uint64())
		}
		return vs
	}()...) {
		if got := unmix(mix(v)); got != v {
			t.Fatalf("unmix(mix(%x)) = %x", v, got)
		}
		b := pack(v)
		if got := unpack(UUID(b)); got != v {
			t.Fatalf("unpack(pack(%x)) = %x", v, got)
		}
		if b[6]>>4 != 8 || b[8]>>6 != 2 {
			t.Fatalf("pack(%x) = %x: not version 8 / variant 10", v, b)
		}
	}
}

// TestAllocUnique checks that concurrent allocation and different leases never collide and no sequence number is skipped.
func TestAllocUnique(t *testing.T) {
	const perLease, workers = 20000, 8
	seen := map[UUID]struct{}{}
	var mu sync.Mutex
	for _, lease := range []uint64{0, 1, 1<<LeaseBits - 1} {
		a, err := NewAllocator(lease)
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < perLease/workers; i++ {
					u, err := a.Next()
					if err != nil {
						t.Error(err)
						return
					}
					mu.Lock()
					if _, dup := seen[u]; dup {
						t.Errorf("duplicate: %v", u)
					}
					seen[u] = struct{}{}
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
	}
	// Decoded, each lease shows 0..perLease-1 exactly once.
	got := map[uint64][]uint64{}
	for u := range seen {
		v := unmix(unpack(u))
		got[v>>SeqBits] = append(got[v>>SeqBits], v&(1<<SeqBits-1))
	}
	for lease, seqs := range got {
		sort.Slice(seqs, func(i, j int) bool { return seqs[i] < seqs[j] })
		if len(seqs) != perLease || seqs[0] != 0 || seqs[len(seqs)-1] != perLease-1 {
			t.Errorf("lease %d: %d values, range %d..%d", lease, len(seqs), seqs[0], seqs[len(seqs)-1])
		}
	}
}

// TestAllocSpread checks that sequential numbers still spread the leading byte (ShardBounds splits evenly).
func TestAllocSpread(t *testing.T) {
	a, _ := NewAllocator(0)
	var counts [256]int
	const n = 256 * 400
	for i := 0; i < n; i++ {
		u, _ := a.Next()
		counts[u[0]]++
	}
	for b, c := range counts {
		if c < n/256/2 || c > n/256*2 {
			t.Errorf("first byte %02x: %d (about %d expected)", b, c, n/256)
		}
	}
}

func TestAllocLimits(t *testing.T) {
	if _, err := NewAllocator(1 << LeaseBits); err == nil {
		t.Errorf("past the lease limit must be an error")
	}
	a, _ := NewAllocator(5)
	a.seq.Store(1<<SeqBits - 1)
	if _, err := a.Next(); err != nil {
		t.Errorf("the last sequence number can be handed out: %v", err)
	}
	if _, err := a.Next(); err == nil {
		t.Errorf("running out of sequence numbers must be an error")
	}
	b, _ := NewAllocator(5)
	u, err := b.Next()
	if err != nil || u.String()[14] != '8' {
		t.Errorf("next must be version 8: %v %v", u, err)
	}
	if back, err := Parse(u.String()); err != nil || back != u {
		t.Errorf("no round trip through the string form: %v %v", back, err)
	}
}
