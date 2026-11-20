package codec

import (
	"fmt"
	"strings"
)

// SQLTextCollation is the collation of string columns: code point order, case- and accent-sensitive, like Cypher
// (MySQL's default utf8mb4_0900_ai_ci makes "Müller" equal "Muller" in DISTINCT / GROUP BY / = / ORDER BY).
const SQLTextCollation = "utf8mb4_bin"

func MapToSQLType(t string) string {
	if _, ok := ListElem(t); ok {
		return "JSON" // lists are stored as JSON arrays
	}
	switch t {
	case "string":
		return "TEXT COLLATE " + SQLTextCollation
	case "int":
		return "INT"
	case "long":
		return "BIGINT"
	case "datetime":
		// ISO 8601 with time, at microsecond precision so fractional seconds are kept (DATETIME has seconds only).
		return "DATETIME(6)"
	case "date":
		// dates without time only
		return "DATE"
	case "json":
		// complex values (arrays, structs) are stored as JSON strings
		return "JSON"
	default:
		// unknown types are stored as text to be safe
		return "TEXT COLLATE " + SQLTextCollation
	}
}

func MapToCassandraType(typeName string) string {
	if _, ok := ListElem(typeName); ok {
		return "text" // lists are stored as JSON array strings
	}
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
		// complex values (arrays, structs) are stored as JSON text
		return "text"
	default:
		panic(fmt.Sprintf("unsupported cassandra type mapping for: %s", typeName))
	}
}

// MapToClickHouseType maps a mapping type name to a ClickHouse column type (every column but uuid is Nullable).
// Strings compare bytewise, which for UTF-8 is code point order, matching Cypher.
func MapToClickHouseType(typeName string) string {
	var t string
	switch strings.ToLower(typeName) {
	case "int", "integer":
		t = "Int32"
	case "long":
		t = "Int64"
	case "float":
		t = "Float32"
	case "double":
		t = "Float64"
	case "datetime", "timestamp":
		t = "DateTime64(6, 'UTC')" // keep fractional seconds
	case "date":
		t = "Date32" // Date only covers 1970-2149
	case "boolean", "bool":
		t = "Bool"
	default: // string / text / list and json (JSON strings) / unknown types
		t = "String"
	}
	return "Nullable(" + t + ")"
}
