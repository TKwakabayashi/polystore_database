package codec

import (
	"testing"

	"polystore_database/src/go/id"
	"polystore_database/src/go/settings"
)

// キーは区切り文字ではなく位置で読むので、バイナリの uuid が \x00 を含んでも壊れない。
func TestKVSKeysRoundTrip(t *testing.T) {
	prev := settings.UUIDStorage
	defer func() { settings.UUIDStorage = prev }()
	var raw [16]byte
	raw[3], raw[7] = 0x00, 0x00 // 区切り文字と同じバイトを含む
	raw[0] = 0x01
	for _, f := range []settings.UUIDFormat{settings.UUIDText, settings.UUIDBinary} {
		settings.UUIDStorage = f
		for _, u := range []id.UUID{id.Make(id.KindNode, raw), id.New(id.KindNode)} {
			label, got, prop, ok := ParseEntityKey(BuildEntityKey("Person", u, "firstName"))
			if !ok || label != "Person" || got != u || prop != "firstName" {
				t.Errorf("format %d entity key: %q %v %q %v", f, label, got, prop, ok)
			}
			if got := IndexKeyUUID(BuildIndexKey("Person", "gender", []byte("male"), u)); got != u {
				t.Errorf("format %d index key: %v", f, got)
			}
		}
		if _, _, _, ok := ParseEntityKey([]byte("Person\x00short\x00p")); ok {
			t.Errorf("format %d: 長さの合わないキーは ok=false", f)
		}
	}
}
