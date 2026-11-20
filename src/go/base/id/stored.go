package id

import (
	"encoding/hex"
	"fmt"
	"sync/atomic"
)

// stored.go converts at the store boundary following the storage form (Storage; uuid_storage in the instance's
// record mapping). Values written to stores, values to match, and KVS key parts always go through here (display
// and logs use String). Reading (FromAny / FromBytes) accepts both forms.

// Storage is how uuids are stored in every database of an instance.
type Storage int

const (
	// StorageText stores the 36-character string form.
	StorageText Storage = iota
	// StorageBinary stores the 16 bytes.
	StorageBinary
)

func (s Storage) String() string {
	if s == StorageBinary {
		return "binary"
	}
	return "text"
}

// ParseStorage reads "text" (or "") and "binary".
func ParseStorage(name string) (Storage, error) {
	switch name {
	case "", "text":
		return StorageText, nil
	case "binary":
		return StorageBinary, nil
	}
	return StorageText, fmt.Errorf("id: unknown uuid storage %q", name)
}

// The storage of the running instance. It is a property of the data, set once from the record
// mapping when the instance is opened (schema.LoadRecord).
var binary atomic.Bool

// SetStorage sets the storage of the running instance.
func SetStorage(s Storage) { binary.Store(s == StorageBinary) }

// CurrentStorage returns the storage of the running instance.
func CurrentStorage() Storage {
	if binary.Load() {
		return StorageBinary
	}
	return StorageText
}

// Binary reports whether uuids are stored as bytes.
func Binary() bool { return binary.Load() }

// Stored returns u's stored value (string for text, 16-byte []byte for binary); nil when empty.
func Stored(u UUID) any {
	if u.Empty() {
		return nil
	}
	if Binary() {
		return u.Bytes()
	}
	return u.String()
}

// StoredText returns u's stored value as text (for tables passed inside JSON): the text form, or for binary the
// 16 bytes in hex (32 characters; the receiver UNHEXes them); nil when empty.
func StoredText(u UUID) any {
	if u.Empty() {
		return nil
	}
	if Binary() {
		return hex.EncodeToString(u.Bytes())
	}
	return u.String()
}

// StoredSlice returns the stored values of us as a typed slice ([]string or [][]byte) for IN lists.
func StoredSlice(us []UUID) any {
	if Binary() {
		out := make([][]byte, len(us))
		for i, u := range us {
			out[i] = u.Bytes()
		}
		return out
	}
	out := make([]string, len(us))
	for i, u := range us {
		out[i] = u.String()
	}
	return out
}

// StoredKey returns the bytes embedded in KVS keys (always StoredLen long; read by position, not by separator).
func StoredKey(u UUID) []byte {
	if Binary() {
		return u.Bytes()
	}
	return []byte(u.String())
}

// StoredLen is the byte length of a stored value (for reading keys by position).
func StoredLen() int {
	if Binary() {
		return Size
	}
	return TextSize
}

// StoredBound turns a ShardBounds leading-byte bound into a value comparable with stored values (MongoDB range
// conditions). Binary compares lengths too (BinData), so it is padded to 16 bytes with zeros; nil stays nil.
func StoredBound(prefix []byte) any {
	if prefix == nil {
		return nil
	}
	if Binary() {
		b := make([]byte, Size)
		copy(b, prefix)
		return b
	}
	return string(prefix)
}
