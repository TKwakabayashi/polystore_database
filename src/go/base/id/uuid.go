// Package id is the only authority on UUIDs: generation, boundary conversion, order and the property name.
//
// The internal form is the 16 bytes of an RFC 9562 UUID in a fixed array, comparable (usable as a map key).
// The text form is "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" in lowercase hex. Nodes and relationships are not
// distinguished (Allocator gives values unique across both). Other code does not depend on the form: it uses
// String, Bytes and Compare only.
//
// Reading also accepts the legacy forms (a kind letter N/E before the text, "N-xxxxxxxx-...", and a kind byte
// plus 16 bytes) and drops the kind, so uuid-format can convert existing data. Writing and matching always use
// the current form, so unconverted data does not match.
package id

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Size is the byte length of the internal form.
const Size = 16

// TextSize is the length of the text form.
const TextSize = 36

// legacySize / legacyTextSize are the lengths of the legacy (kinded) forms, accepted on read only.
const (
	legacySize     = 1 + Size
	legacyTextSize = 2 + TextSize
)

// UUID is the internal form. The zero value is empty (no id).
type UUID [Size]byte

// PropName is the only definition of the uuid property / column name (a schema name, not a representation).
const PropName = "uuid"

// New makes a random UUID (v4) for tests and throwaway ids (unique only probabilistically; data ids come from Allocator).
func New() UUID {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	raw[6] = (raw[6] & 0x0f) | 0x40 // version 4
	raw[8] = (raw[8] & 0x3f) | 0x80 // variant 10
	return Make(raw)
}

// Make makes a UUID from 16 bytes (for deterministic test ids and conversions).
func Make(raw [16]byte) UUID { return UUID(raw) }

// legacyKind reports whether a byte is a legacy kind letter.
func legacyKind(b byte) bool { return b == 'N' || b == 'E' }

// Parse converts the text boundary (Cypher values, text stored values, display) to a UUID; "" is the empty UUID.
// The legacy "N-..." / "E-..." are accepted (the kind is dropped).
func Parse(s string) (UUID, error) {
	if s == "" {
		return UUID{}, nil
	}
	body := s
	if len(s) == legacyTextSize && legacyKind(s[0]) && s[1] == '-' {
		body = s[2:]
	}
	if len(body) != TextSize || body[8] != '-' || body[13] != '-' || body[18] != '-' || body[23] != '-' {
		return UUID{}, fmt.Errorf("id: malformed uuid %q", s)
	}
	var hexs [32]byte
	n := 0
	for i := 0; i < len(body); i++ {
		if body[i] == '-' {
			continue
		}
		hexs[n] = body[i]
		n++
	}
	var u UUID
	if _, err := hex.Decode(u[:], hexs[:]); err != nil {
		return UUID{}, fmt.Errorf("id: malformed uuid %q: %w", s, err)
	}
	return u, nil
}

// FromBytes converts the byte boundary (binary stored values, KVS key parts) to a UUID: the 16-byte internal
// form or the bytes of the text form (legacy 17 bytes and "N-..." bytes too).
func FromBytes(b []byte) (UUID, error) {
	switch len(b) {
	case 0:
		return UUID{}, nil
	case Size:
		return UUID(b), nil
	case legacySize:
		if !legacyKind(b[0]) {
			return UUID{}, fmt.Errorf("id: malformed binary uuid %x", b)
		}
		return UUID(b[1:]), nil
	}
	return Parse(string(b))
}

// FromAny converts a driver's native value (interface{} from SQL / CQL / Neo4j) to a UUID: strings, bytes (either
// storage form, legacy too) and UUIDs; anything else gives the empty UUID. Driver-specific types (bson Binary, ...)
// are turned into []byte at each driver boundary first.
func FromAny(v interface{}) UUID {
	var (
		u   UUID
		err error
	)
	switch x := v.(type) {
	case UUID:
		return x
	case string:
		if len(x) == Size || len(x) == legacySize { // a binary stored value received as a string (an SQL driver's []byte -> string, ...)
			u, err = FromBytes([]byte(x))
		} else {
			u, err = Parse(x)
		}
	case []byte:
		u, err = FromBytes(x)
	default:
		return UUID{}
	}
	if err != nil {
		return UUID{}
	}
	return u
}

// String converts a UUID to its text form (display, text stored values); "" when empty.
func (u UUID) String() string {
	if u.Empty() {
		return ""
	}
	var buf [TextSize]byte
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], u[10:16])
	return string(buf[:])
}

// Bytes returns the 16-byte internal form (binary stored values); nil when empty.
func (u UUID) Bytes() []byte {
	if u.Empty() {
		return nil
	}
	b := make([]byte, Size)
	copy(b, u[:])
	return b
}

// Empty reports whether the UUID is empty.
func (u UUID) Empty() bool { return u == UUID{} }

// MarshalText / UnmarshalText use the text form in text encodings such as JSON.
func (u UUID) MarshalText() ([]byte, error) { return []byte(u.String()), nil }

func (u *UUID) UnmarshalText(b []byte) error {
	v, err := Parse(string(b))
	if err != nil {
		return err
	}
	*u = v
	return nil
}

// Compare is the only order (byte order of the internal form; equal to lexical order of the text form, since lowercase hex sorts numerically).
func Compare(a, b UUID) int { return bytes.Compare(a[:], b[:]) }

// Less is Compare(a, b) < 0.
func Less(a, b UUID) bool { return Compare(a, b) < 0 }
