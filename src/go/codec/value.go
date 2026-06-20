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

func EncodeForKVS(val interface{}, targetType string) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	tType := strings.ToLower(targetType)

	switch tType {
	case "int", "integer":
		if v, ok := val.(int32); ok {
			buf := make([]byte, 4)
			// 4バイトでエンコード (ソート順序を考慮したオフセット)
			binary.BigEndian.PutUint32(buf, uint32(v)^(1<<31))
			return buf, nil
		}
		return nil, fmt.Errorf("KVS encode: expected int32 for %s, got %T", targetType, val)

	case "long":
		if v, ok := val.(int64); ok {
			buf := make([]byte, 8)
			// 8バイトでエンコード
			binary.BigEndian.PutUint64(buf, uint64(v)^(1<<63))
			return buf, nil
		}
		return nil, fmt.Errorf("KVS encode: expected int64 for %s, got %T", targetType, val)

	case "date":
		if v, ok := val.(time.Time); ok {
			// Neo4j基準の精度を維持しつつ、KVSでは検索性の高い文字列で保存
			return []byte(v.Format("2006-01-02")), nil
		}
	case "datetime":
		if v, ok := val.(time.Time); ok {
			return []byte(v.Format(time.RFC3339)), nil
		}

	case "string", "text":
		return []byte(fmt.Sprint(val)), nil
	}

	return nil, fmt.Errorf("KVS encode unsupported type: %T for target %s", val, targetType)
}

func DecodeValue(valBytes []byte, typeName string) interface{} {
	if len(valBytes) == 0 {
		return nil
	}
	typeName = strings.ToLower(typeName)

	switch typeName {
	case "int", "integer":
		if len(valBytes) == 4 {
			u := binary.BigEndian.Uint32(valBytes)
			return int32(u ^ (1 << 31))
		}
		// 下位互換性のため文字列パースも残すが、基本は4byteバイナリを想定
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
		// 文字列として返し、ParseToNative側で time.Time に変換させる
		return string(valBytes)

	default:
		return string(valBytes)
	}
}

func ParseToNative(val interface{}, targetType string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	targetType = strings.ToLower(targetType)
	switch targetType {
	case "int", "integer":
		var i32 int32
		switch v := val.(type) {
		case int32:
			i32 = v
		case int64:
			// 21億を超える値のチェック（オーバーフロー対策）
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
		// Neo4j方式（time.Time）を基準とする
		switch v := val.(type) {
		case time.Time:
			return v, nil
		case string:
			// 文字列からの復元（RFC3339優先）
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, nil
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t, nil
			}
			return nil, fmt.Errorf("unsupported date format: %s", v)
		default:
			// Neo4jドライバ独自の型や他DBの特殊な型から time.Time を抽出
			if t, ok := v.(interface{ Time() time.Time }); ok {
				return t.Time(), nil
			}
			return nil, fmt.Errorf("invalid type for time.Time: %T", val)
		}

	case "string", "text":
		if s, ok := val.(string); ok {
			return s, nil
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
		// 型指定がない場合はそのまま返す
		return val, nil
	}
}

func PrepareForDB(val interface{}, targetType string, destType string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}
	targetType = strings.ToLower(targetType)

	switch destType {
	case "document": // MongoDB
		// MongoDBドライバは time.Time を自動的に BSON Date に変換するので
		// 特別な変換は不要だが、精度の調整が必要ならここで行う
		return val, nil

	case "graph": // Neo4j
		if t, ok := val.(time.Time); ok {
			// "Local" タイムゾーンエラーを避けるため、UTCに変換した上でRFC3339文字列として渡す
			return t.UTC().Format(time.RFC3339), nil
		}

	case "relational": // MySQL等
		if t, ok := val.(time.Time); ok {
			if targetType == "date" {
				return t.Format("2006-01-02"), nil
			}
			return t.Format("2006-01-02 15:04:05"), nil
		}
	case "columnar": // Cassandra (重要)
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

// goで取り扱うための共通型に変換
func ConvertToNativeType(val interface{}, typeName string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	// JSONから復元された数値（json.Number）の処理
	if jn, ok := val.(json.Number); ok {
		switch typeName {
		case "int", "integer", "long":
			i, err := jn.Int64()
			if err != nil {
				return nil, fmt.Errorf("failed to parse integer '%v': %w", jn, err)
			}
			return i, nil // すべて int64 で一旦保持
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
				// フォールバック: 時刻付き形式から日付のみ抽出
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
