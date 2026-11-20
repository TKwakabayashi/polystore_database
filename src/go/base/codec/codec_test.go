package codec

import (
	"polystore_database/src/go/base/id"
	"reflect"
	"testing"
)

// Keys are read by position, not by separator, so a binary uuid containing \x00 does not break them.
func TestKVSKeysRoundTrip(t *testing.T) {
	prev := id.CurrentStorage()
	defer func() { id.SetStorage(prev) }()
	var raw [16]byte
	raw[3], raw[7] = 0x00, 0x00 // contains the separator byte
	raw[0] = 0x01
	for _, f := range []id.Storage{id.StorageText, id.StorageBinary} {
		id.SetStorage(f)
		for _, u := range []id.UUID{id.Make(raw), id.New()} {
			label, got, prop, ok := ParseEntityKey(BuildEntityKey("Person", u, "firstName"))
			if !ok || label != "Person" || got != u || prop != "firstName" {
				t.Errorf("format %d entity key: %q %v %q %v", f, label, got, prop, ok)
			}
			if got := IndexKeyUUID(BuildIndexKey("Person", "gender", []byte("male"), u)); got != u {
				t.Errorf("format %d index key: %v", f, got)
			}
		}
		if _, _, _, ok := ParseEntityKey([]byte("Person\x00short\x00p")); ok {
			t.Errorf("format %d: a key of the wrong length must give ok=false", f)
		}
	}
}

func TestListElem(t *testing.T) {
	for in, want := range map[string]string{"list<string>": "string", "LIST<long>": "long", "list": "", " list< int > ": "int"} {
		if got, ok := ListElem(in); !ok || got != want {
			t.Errorf("ListElem(%q) = %q %v, want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"string", "json", "listing", "list<string"} {
		if _, ok := ListElem(in); ok {
			t.Errorf("ListElem(%q) is not a list", in)
		}
	}
	if !IsComplexType("list<string>") || !IsComplexType("JSON") || IsComplexType("string") {
		t.Errorf("IsComplexType")
	}
}

// TestListRoundTrip checks that list<...> values come back from each store's written form as lists of the element type.
func TestListRoundTrip(t *testing.T) {
	want := []interface{}{"a@x.org", "b@y.org"}
	src := []interface{}{"a@x.org", "b@y.org"} // a Neo4j array

	native, err := ParseToNative(src, "list<string>")
	if err != nil || !reflect.DeepEqual(native, want) {
		t.Fatalf("ParseToNative: %v %v", native, err)
	}
	for _, dest := range []string{"relational", "columnar", "widecolumn"} {
		v, err := PrepareForDB(native, "list<string>", dest)
		if s, ok := v.(string); err != nil || !ok || s != `["a@x.org","b@y.org"]` {
			t.Errorf("PrepareForDB %s: %v %v", dest, v, err)
		}
		back, err := ConvertToNativeType(v, "list<string>") // reading back (JSON string)
		if err != nil || !reflect.DeepEqual(back, want) {
			t.Errorf("ConvertToNativeType %s: %v %v", dest, back, err)
		}
	}
	for _, dest := range []string{"document", "graph"} {
		if v, err := PrepareForDB(native, "list<string>", dest); err != nil || !reflect.DeepEqual(v, want) {
			t.Errorf("PrepareForDB %s keeps native values: %v %v", dest, v, err)
		}
	}
	b, err := EncodeForKVS(native, "list<string>")
	if err != nil {
		t.Fatal(err)
	}
	if got := DecodeValue(b, "list<string>"); !reflect.DeepEqual(got, want) {
		t.Errorf("KVS: %v", got)
	}
	// conversion to the element type (JSON numbers come back as float64)
	if got, err := ConvertToNativeType(`[1, 2]`, "list<long>"); err != nil || !reflect.DeepEqual(got, []interface{}{int64(1), int64(2)}) {
		t.Errorf("list<long>: %v %v", got, err)
	}
	if got, err := ParseToNative([]string{"x"}, "list"); err != nil || !reflect.DeepEqual(got, []interface{}{"x"}) {
		t.Errorf("list: %v %v", got, err)
	}
	if _, err := ParseToNative("x", "list<string>"); err == nil {
		t.Errorf("a scalar is not a list")
	}
	if _, err := ParseToNative(src, "string"); err == nil {
		t.Errorf("a list declared string is an error")
	}
}
