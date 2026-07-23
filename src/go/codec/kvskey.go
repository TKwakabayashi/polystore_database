package codec

import (
	"bytes"

	"polystore_database/src/go/id"
)

const Sep = "\x00"

func BuildEntityKey(entityLabel string, uuid id.UUID, property string) []byte {
	return []byte(entityLabel + Sep + uuid.String() + Sep + property)
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
	buf.Write(uuid.Bytes())
	return buf.Bytes()
}
