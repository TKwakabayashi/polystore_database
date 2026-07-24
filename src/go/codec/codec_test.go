package codec

import (
	"bytes"
	"testing"
	"time"

	"polystore_database/src/go/id"
)

func TestMapToSQLType(t *testing.T) {
	cases := map[string]string{
		"string":   "TEXT",
		"int":      "INT",
		"long":     "BIGINT",
		"datetime": "DATETIME",
		"date":     "DATE",
		"json":     "JSON",
		"unknown":  "TEXT", // 未知型は安全側で TEXT
	}
	for in, want := range cases {
		if got := MapToSQLType(in); got != want {
			t.Errorf("MapToSQLType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapToCassandraType(t *testing.T) {
	cases := map[string]string{
		"int":       "int",
		"integer":   "int",
		"long":      "bigint",
		"float":     "float",
		"double":    "double",
		"string":    "text",
		"text":      "text",
		"datetime":  "timestamp",
		"timestamp": "timestamp",
		"date":      "date",
		"bool":      "boolean",
		"uuid":      "uuid",
		"json":      "text",
		"STRING":    "text", // 大文字も許容（ToLower）
	}
	for in, want := range cases {
		if got := MapToCassandraType(in); got != want {
			t.Errorf("MapToCassandraType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapToCassandraTypeUnsupportedPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("未対応型で panic するはずが panic しなかった")
		}
	}()
	MapToCassandraType("no-such-type")
}

func TestConvertToNativeType(t *testing.T) {
	// int/integer/long はすべて int64 に正規化される。
	for _, ty := range []string{"int", "integer", "long"} {
		v, err := ConvertToNativeType("100", ty)
		if err != nil {
			t.Fatalf("ConvertToNativeType(100,%q): %v", ty, err)
		}
		if v != int64(100) {
			t.Errorf("ConvertToNativeType(100,%q) = %#v, want int64(100)", ty, v)
		}
	}
	// float64 入力も int64 へ。
	if v, _ := ConvertToNativeType(float64(7), "long"); v != int64(7) {
		t.Errorf("ConvertToNativeType(7.0,long) = %#v, want int64(7)", v)
	}
	// datetime は RFC3339 → time.Time。
	v, err := ConvertToNativeType("2011-06-16T00:00:00Z", "datetime")
	if err != nil {
		t.Fatalf("datetime: %v", err)
	}
	if _, ok := v.(time.Time); !ok {
		t.Errorf("datetime = %#v, want time.Time", v)
	}
	// date は YYYY-MM-DD → time.Time。
	if v, err := ConvertToNativeType("2011-06-16", "date"); err != nil {
		t.Fatalf("date: %v", err)
	} else if _, ok := v.(time.Time); !ok {
		t.Errorf("date = %#v, want time.Time", v)
	}
	// string はそのまま文字列化。
	if v, _ := ConvertToNativeType(42, "string"); v != "42" {
		t.Errorf("string(42) = %#v, want \"42\"", v)
	}
	// 不正な整数文字列はエラー。
	if _, err := ConvertToNativeType("abc", "long"); err == nil {
		t.Errorf("不正な整数文字列でエラーになるはず")
	}
	// nil は nil のまま。
	if v, _ := ConvertToNativeType(nil, "long"); v != nil {
		t.Errorf("nil = %#v, want nil", v)
	}
}

func TestEncodeIntRoundTrip(t *testing.T) {
	// EncodeInt は long と同じ 8byte・符号ビット反転エンコード（DecodeValue で復元できる）。
	for _, n := range []int64{0, 1, -1, 100, -100, 1 << 40} {
		enc := EncodeInt(n)
		if len(enc) != 8 {
			t.Fatalf("EncodeInt(%d) len = %d, want 8", n, len(enc))
		}
		if got := DecodeValue(enc, "long"); got != n {
			t.Errorf("DecodeValue(EncodeInt(%d),long) = %#v, want %d", n, got, n)
		}
	}
}

func TestEncodeIntOrderPreserving(t *testing.T) {
	// バイト列辞書順が数値順と一致する（シャード境界・索引の前提）。
	a := EncodeInt(-5)
	b := EncodeInt(3)
	if bytes.Compare(a, b) >= 0 {
		t.Errorf("EncodeInt(-5) は EncodeInt(3) より小さいべき")
	}
}

func TestBuildEntityKey(t *testing.T) {
	u, _ := id.Parse("abc")
	got := BuildEntityKey("Person", u, "firstName")
	want := []byte("Person" + Sep + "abc" + Sep + "firstName")
	if !bytes.Equal(got, want) {
		t.Errorf("BuildEntityKey = %q, want %q", got, want)
	}
}

func TestBuildIndexKey(t *testing.T) {
	u, _ := id.Parse("u1")
	got := BuildIndexKey("Person", "firstName", []byte("Bob"), u)
	var want bytes.Buffer
	want.WriteString("index" + Sep + "Person" + Sep + "firstName" + Sep)
	want.Write([]byte("Bob"))
	want.WriteString(Sep)
	want.Write(u.Bytes())
	if !bytes.Equal(got, want.Bytes()) {
		t.Errorf("BuildIndexKey = %q, want %q", got, want.Bytes())
	}
}
