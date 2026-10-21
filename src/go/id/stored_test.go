package id

import (
	"bytes"
	"testing"

	"polystore_database/src/go/settings"
)

func withFormat(t *testing.T, f settings.UUIDFormat) {
	t.Helper()
	prev := settings.UUIDStorage
	settings.UUIDStorage = f
	t.Cleanup(func() { settings.UUIDStorage = prev })
}

func TestStoredFormats(t *testing.T) {
	u := New(KindNode)
	withFormat(t, settings.UUIDText)
	if s, ok := Stored(u).(string); !ok || s != u.String() || StoredLen() != TextSize || string(StoredKey(u)) != u.String() {
		t.Errorf("text: %v", Stored(u))
	}
	if l, ok := StoredSlice([]UUID{u}).([]string); !ok || l[0] != u.String() {
		t.Errorf("text slice: %v", StoredSlice([]UUID{u}))
	}
	withFormat(t, settings.UUIDBinary)
	if b, ok := Stored(u).([]byte); !ok || !bytes.Equal(b, u.Bytes()) || StoredLen() != Size || !bytes.Equal(StoredKey(u), u.Bytes()) {
		t.Errorf("binary: %v", Stored(u))
	}
	if l, ok := StoredSlice([]UUID{u}).([][]byte); !ok || !bytes.Equal(l[0], u.Bytes()) {
		t.Errorf("binary slice: %v", StoredSlice([]UUID{u}))
	}
	if Stored(UUID{}) != nil {
		t.Errorf("空は nil")
	}
	if b, ok := StoredBound([]byte{0x80}).([]byte); !ok || len(b) != Size || b[0] != 0x80 {
		t.Errorf("binary bound: %v", StoredBound([]byte{0x80}))
	}
}

// どちらの保存形式で書かれた値も、設定に関係なく読める（形式の切り替え途中・移行元と移行先の形式違いに耐える）。
func TestReadEitherFormat(t *testing.T) {
	u := New(KindEdge)
	for _, f := range []settings.UUIDFormat{settings.UUIDText, settings.UUIDBinary} {
		withFormat(t, f)
		for _, v := range []any{u.String(), []byte(u.String()), u.Bytes(), string(u.Bytes())} {
			if FromAny(v) != u {
				t.Errorf("format %d: FromAny(%T %v) = %v", f, v, v, FromAny(v))
			}
		}
	}
}
