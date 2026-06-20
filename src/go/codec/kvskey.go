package codec

import (
	"bytes"
)

const Sep = "\x00"

func BuildEntityKey(entityLabel, uuid, property string) []byte {
	return []byte(entityLabel + Sep + uuid + Sep + property)
}

func BuildIndexKey(entityLabel, property string, valBytes []byte, uuid string) []byte {
	var buf bytes.Buffer
	buf.WriteString("index")
	buf.WriteString(Sep)
	buf.WriteString(entityLabel)
	buf.WriteString(Sep)
	buf.WriteString(property)
	buf.WriteString(Sep)
	buf.Write(valBytes)
	buf.WriteString(Sep)
	buf.WriteString(uuid)
	return buf.Bytes()
}
