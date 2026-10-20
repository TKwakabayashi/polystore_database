package codec

import (
	"bytes"

	"polystore_database/src/go/id"
)

// KVS のキー:
//
//	エンティティ: <label> Sep <uuid> Sep <prop>
//	逆引き:       "index" Sep <label> Sep <prop> Sep <値> Sep <uuid>
//
// <uuid> は保存形式（settings.UUIDStorage）のバイト列で長さは id.StoredLen() に固定。バイナリには Sep（\x00）が
// 含まれ得るので、読むときは区切り文字で分けず位置で切り出す（ParseEntityKey / IndexKeyUUID）。
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

// ParseEntityKey はエンティティのキーを label・uuid・prop に分ける（形が違えば ok=false）。
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

// IndexKeyUUID は逆引きキーの末尾の uuid（形が違えば空）。
func IndexKeyUUID(key []byte) id.UUID {
	n := id.StoredLen()
	if len(key) < n+1 || key[len(key)-n-1] != Sep[0] {
		return id.UUID{}
	}
	u, _ := id.FromBytes(key[len(key)-n:])
	return u
}
