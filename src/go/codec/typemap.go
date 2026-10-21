package codec

import (
	"fmt"
	"strings"
)

// SQLTextCollation は文字列の列の照合順序。Cypher と同じくコードポイント順で、大文字小文字・アクセントを区別する
// （MySQL の既定 utf8mb4_0900_ai_ci だと DISTINCT / GROUP BY / = / ORDER BY で "Müller" と "Muller" が同じになる）。
const SQLTextCollation = "utf8mb4_bin"

func MapToSQLType(t string) string {
	switch t {
	case "string":
		return "TEXT COLLATE " + SQLTextCollation
	case "int":
		return "INT"
	case "long":
		return "BIGINT"
	case "datetime":
		// ISO 8601（時刻あり）。小数秒を落とさないようマイクロ秒精度で持つ（DATETIME は秒精度）。
		return "DATETIME(6)"
	case "date":
		// 時刻なしの日付のみに対応
		return "DATE"
	case "json":
		// 配列/構造体などの複雑値は JSON 文字列として保存
		return "JSON"
	default:
		// 未知の型は安全のために文字型として扱う
		return "TEXT COLLATE " + SQLTextCollation
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
	case "json":
		// 配列/構造体などの複雑値は JSON 文字列(text)として保存
		return "text"
	default:
		panic(fmt.Sprintf("unsupported cassandra type mapping for: %s", typeName))
	}
}
