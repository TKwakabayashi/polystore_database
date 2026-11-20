package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// KVSDatetimeLayout is the fixed-width UTC layout (nanoseconds) of datetimes stored in KVS.
const KVSDatetimeLayout = "2006-01-02T15:04:05.000000000Z"

func EncodeForKVS(val interface{}, targetType string) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	tType := strings.ToLower(targetType)
	if _, ok := ListElem(tType); ok {
		tType = "json" // lists are stored as JSON too
	}

	switch tType {
	case "int", "integer":
		if v, ok := val.(int32); ok {
			buf := make([]byte, 4)
			// 4-byte encoding (offset so the byte order sorts)
			binary.BigEndian.PutUint32(buf, uint32(v)^(1<<31))
			return buf, nil
		}
		return nil, fmt.Errorf("KVS encode: expected int32 for %s, got %T", targetType, val)

	case "long":
		if v, ok := val.(int64); ok {
			buf := make([]byte, 8)
			// 8-byte encoding
			binary.BigEndian.PutUint64(buf, uint64(v)^(1<<63))
			return buf, nil
		}
		return nil, fmt.Errorf("KVS encode: expected int64 for %s, got %T", targetType, val)

	case "date":
		if v, ok := val.(time.Time); ok {
			// keep Neo4j's precision, stored in KVS as a searchable string
			return []byte(v.Format("2006-01-02")), nil
		}
	case "datetime":
		if v, ok := val.(time.Time); ok {
			// A fixed-width UTC layout keeps fractional seconds and makes the lexical order of index keys the time order
			// (RFC3339Nano drops trailing zeros, which would sort "...13Z" after "...13.234Z").
			return []byte(v.UTC().Format(KVSDatetimeLayout)), nil
		}

	case "string", "text":
		if isComplexValue(val) {
			return nil, fmt.Errorf("KVS encode: complex value for %q declared as %q; declare it as \"list<…>\" or \"json\"", fmt.Sprintf("%T", val), targetType)
		}
		return []byte(fmt.Sprint(val)), nil

	case "json":
		// complex values (arrays, structs) are stored as JSON (entity keys only; no reverse index)
		return json.Marshal(val)
	}

	return nil, fmt.Errorf("KVS encode unsupported type: %T for target %s", val, targetType)
}

// isComplexValue reports whether a value is an array, slice or map (a non-scalar complex value).
// []byte is raw bytes, a scalar.
func isComplexValue(val interface{}) bool {
	if val == nil {
		return false
	}
	if _, ok := val.([]byte); ok {
		return false
	}
	switch reflect.ValueOf(val).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	default:
		return false
	}
}

func DecodeValue(valBytes []byte, typeName string) interface{} {
	if len(valBytes) == 0 {
		return nil
	}
	typeName = strings.ToLower(typeName)
	if elem, ok := ListElem(typeName); ok {
		out, err := convertList(valBytes, elem, ConvertToNativeType)
		if err != nil {
			return string(valBytes) // return the raw string if malformed
		}
		return out
	}

	switch typeName {
	case "int", "integer":
		if len(valBytes) == 4 {
			u := binary.BigEndian.Uint32(valBytes)
			return int32(u ^ (1 << 31))
		}
		// string parsing kept for old values; the 4-byte binary is expected
		v, _ := strconv.ParseInt(string(valBytes), 10, 32)
		return int32(v)

	case "long":
		if len(valBytes) == 8 {
			u := binary.BigEndian.Uint64(valBytes)
			return int64(u ^ (1 << 63))
		}
		v, _ := strconv.ParseInt(string(valBytes), 10, 64)
		return v

	case "date", "datetime":
		// returned as a string; ParseToNative converts it to time.Time
		return string(valBytes)

	case "json":
		// restore complex values stored as JSON
		var out interface{}
		if err := json.Unmarshal(valBytes, &out); err != nil {
			return string(valBytes) // return the raw string if malformed
		}
		return out

	default:
		return string(valBytes)
	}
}

func ParseToNative(val interface{}, targetType string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	targetType = strings.ToLower(targetType)
	if elem, ok := ListElem(targetType); ok {
		// a list with elements of the element type (serialized per store on write)
		return convertList(val, elem, ParseToNative)
	}
	switch targetType {
	case "int", "integer":
		var i32 int32
		switch v := val.(type) {
		case int32:
			i32 = v
		case int64:
			// check values above 2.1 billion (overflow)
			if v > 2147483647 || v < -2147483648 {
				return nil, fmt.Errorf("value %d overflows int32", v)
			}
			i32 = int32(v)
		case int:
			i32 = int32(v)
		case float64:
			i32 = int32(v)
		case string:
			parsed, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return nil, err
			}
			i32 = int32(parsed)
		default:
			return nil, fmt.Errorf("invalid type for int32: %T", val)
		}
		return i32, nil

	case "long":
		var i64 int64
		switch v := val.(type) {
		case int64:
			i64 = v
		case int32:
			i64 = int64(v)
		case int:
			i64 = int64(v)
		case float64:
			i64 = int64(v)
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			i64 = parsed
		default:
			return nil, fmt.Errorf("invalid type for int64: %T", val)
		}
		return i64, nil

	case "date", "datetime":
		// time.Time (as from Neo4j) is the reference
		switch v := val.(type) {
		case time.Time:
			return v, nil
		case string:
			// from a string (RFC 3339 first)
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, nil
			}
			// MySQL DATETIME layout (mirrors PrepareForDB's relational write)
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t, nil
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t, nil
			}
			return nil, fmt.Errorf("unsupported date format: %s", v)
		default:
			// extract time.Time from Neo4j driver types and other stores' special types
			if t, ok := v.(interface{ Time() time.Time }); ok {
				return t.Time(), nil
			}
			return nil, fmt.Errorf("invalid type for time.Time: %T", val)
		}

	case "json":
		// complex values (arrays, structs) keep their structure; each store's writer serializes them.
		return val, nil

	case "string", "text":
		if s, ok := val.(string); ok {
			return s, nil
		}
		if isComplexValue(val) {
			// rounding a complex value to a string would mangle it irreversibly, so this is an error
			return nil, fmt.Errorf("complex value (%T) declared as %q; declare the property as \"list<…>\" or \"json\"", val, targetType)
		}
		return fmt.Sprint(val), nil

	case "bool", "boolean":
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		default:
			return nil, fmt.Errorf("invalid type for bool: %T", val)
		}

	default:
		// no type given: return as is
		return val, nil
	}
}

func PrepareForDB(val interface{}, targetType string, destType string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}
	targetType = strings.ToLower(targetType)

	// Complex values (list / json) go as JSON strings to row and column stores, natively to MongoDB and Neo4j
	// (MongoDB stores arrays and objects; Neo4j stores arrays of primitives).
	if IsComplexType(targetType) {
		switch destType {
		case "relational", "columnar", "widecolumn":
			b, err := json.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("json marshal for %s failed: %w", destType, err)
			}
			return string(b), nil
		default: // document / graph
			return val, nil
		}
	}

	switch destType {
	case "document": // MongoDB
		// The MongoDB driver converts time.Time to a BSON Date itself; adjust precision here if ever needed.
		return val, nil

	case "graph": // Neo4j
		if t, ok := val.(time.Time); ok {
			// convert to UTC and pass as an RFC 3339 string with fractional seconds (avoids the "Local" time zone error)
			return t.UTC().Format(time.RFC3339Nano), nil
		}

	case "relational": // MySQL
		if t, ok := val.(time.Time); ok {
			if targetType == "date" {
				return t.Format("2006-01-02"), nil
			}
			return t.UTC().Format("2006-01-02 15:04:05.000000"), nil // DATETIME(6)
		}
	case "columnar": // ClickHouse (Go types matching the column types)
		switch targetType {
		case "int", "integer":
			if v, ok := toInt64(val); ok {
				return int32(v), nil
			}
		case "long":
			if v, ok := toInt64(val); ok {
				return v, nil
			}
		case "float":
			if v, ok := toFloat64(val); ok {
				return float32(v), nil
			}
		case "double":
			if v, ok := toFloat64(val); ok {
				return v, nil
			}
		case "datetime", "timestamp", "date":
			if t, ok := val.(time.Time); ok {
				return t.UTC(), nil
			}
		case "bool", "boolean":
			return val, nil
		default:
			if s, ok := val.(string); ok {
				return s, nil
			}
			return fmt.Sprint(val), nil
		}
	case "widecolumn": // Cassandra
		switch targetType {
		case "int", "integer":
			if v, ok := val.(int32); ok {
				return v, nil
			}
		case "long":
			if v, ok := val.(int64); ok {
				return v, nil
			}
		}
	}
	return val, nil
}

func UnmarshalWithNumber(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

func EncodeInt(n int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n)^(1<<63))
	return buf
}

// convert to the common Go types
func ConvertToNativeType(val interface{}, typeName string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	if elem, ok := ListElem(typeName); ok {
		return convertList(val, elem, ConvertToNativeType)
	}

	// numbers restored from JSON (json.Number)
	if jn, ok := val.(json.Number); ok {
		switch typeName {
		case "int", "integer", "long":
			i, err := jn.Int64()
			if err != nil {
				return nil, fmt.Errorf("failed to parse integer '%v': %w", jn, err)
			}
			return i, nil // keep everything as int64 first
		case "float", "double":
			f, err := jn.Float64()
			if err != nil {
				return nil, fmt.Errorf("failed to parse float '%v': %w", jn, err)
			}
			return f, nil
		default:
			return jn.String(), nil
		}
	}

	switch typeName {
	case "int", "integer", "long":
		switch v := val.(type) {
		case int, int32, int64:
			return reflect.ValueOf(v).Int(), nil
		case float64:
			return int64(v), nil
		case string:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer string '%s': %w", v, err)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("unexpected type for integer: %T", val)
		}

	case "datetime":
		if s, ok := val.(string); ok {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, fmt.Errorf("invalid datetime format '%s' (expected RFC3339): %w", s, err)
			}
			return t, nil
		}
		if t, ok := val.(time.Time); ok {
			return t, nil
		}
		return nil, fmt.Errorf("unexpected type for datetime: %T", val)

	case "date":
		if s, ok := val.(string); ok {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				// fallback: take the date from a datetime layout
				t2, err2 := time.Parse(time.RFC3339, s)
				if err2 != nil {
					return nil, fmt.Errorf("invalid date format '%s' (expected YYYY-MM-DD): %w", s, err2)
				}
				return time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC), nil
			}
			return t, nil
		}
		if t, ok := val.(time.Time); ok {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
		return nil, fmt.Errorf("unexpected type for date: %T", val)

	case "string":
		return fmt.Sprintf("%v", val), nil
	}

	return val, nil
}

func toInt64(v interface{}) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case float64:
		return int64(x), true
	}
	return 0, false
}

func toFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}
