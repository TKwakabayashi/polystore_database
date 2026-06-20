package codec

import (
	"fmt"
	"strings"
)

func MapToSQLType(t string) string {
	switch t {
	case "string":
		// 255文字を超える可能性がある場合は TEXT 型を検討してください
		return "TEXT"
	case "int":
		return "INT"
	case "long":
		return "BIGINT"
	case "datetime":
		// ISO 8601（時刻あり）に対応
		return "DATETIME"
	case "date":
		// 時刻なしの日付のみに対応
		return "DATE"
	default:
		// 未知の型は安全のために文字型として扱う
		return "TEXT"
	}
}

func MapToCassandraType(typeName string) string {
	switch strings.ToLower(typeName) {
	case "int", "integer":
		return "int" // 32-bit signed integer
	case "long":
		return "bigint" // 64-bit signed integer
	case "float":
		return "float" // 32-bit IEEE-754
	case "double":
		return "double" // 64-bit IEEE-754
	case "string", "text":
		return "text"
	case "datetime", "timestamp":
		return "timestamp"
	case "date":
		return "date"
	case "boolean", "bool":
		return "boolean"
	case "uuid":
		return "uuid"
	default:
		panic(fmt.Sprintf("unsupported cassandra type mapping for: %s", typeName))
	}
}
