package codec

import (
	"bytes"

	"polystore_database/src/go/base/id"
)

// KVS keys:
//
//	entity:  <label> Sep <uuid> Sep <prop>
//	index:   "index" Sep <label> Sep <prop> Sep <value> Sep <uuid>
//
// <uuid> is the stored form (id.CurrentStorage()), always id.StoredLen() bytes. Binary may contain Sep (\x00),
// so keys are read by position, not split on the separator (ParseEntityKey / IndexKeyUUID).
const Sep = "\x00"

func BuildEntityKey(entityLabel string, uuid id.UUID, property string) []byte {
	var buf bytes.Buffer
	buf.WriteString(entityLabel)
	buf.WriteString(Sep)
	buf.Write(id.StoredKey(uuid))
	buf.WriteString(Sep)
	buf.WriteString(property)
	return buf.Bytes()
}

func BuildIndexKey(entityLabel, property string, valBytes []byte, uuid id.UUID) []byte {
	var buf bytes.Buffer
	buf.WriteString("index")
	buf.WriteString(Sep)
	buf.WriteString(entityLabel)
	buf.WriteString(Sep)
	buf.WriteString(property)
	buf.WriteString(Sep)
	buf.Write(valBytes)
	buf.WriteString(Sep)
	buf.Write(id.StoredKey(uuid))
	return buf.Bytes()
}

// ParseEntityKey splits an entity key into label, uuid and prop (ok=false on a malformed key).
func ParseEntityKey(key []byte) (label string, u id.UUID, prop string, ok bool) {
	i := bytes.IndexByte(key, Sep[0])
	n := id.StoredLen()
	if i < 0 || len(key) < i+1+n+1 || key[i+1+n] != Sep[0] {
		return "", id.UUID{}, "", false
	}
	u, err := id.FromBytes(key[i+1 : i+1+n])
	if err != nil {
		return "", id.UUID{}, "", false
	}
	return string(key[:i]), u, string(key[i+1+n+1:]), true
}

// IndexKeyUUID returns the uuid at the end of an index key (empty if malformed).
func IndexKeyUUID(key []byte) id.UUID {
	n := id.StoredLen()
	if len(key) < n+1 || key[len(key)-n-1] != Sep[0] {
		return id.UUID{}
	}
	u, _ := id.FromBytes(key[len(key)-n:])
	return u
}
