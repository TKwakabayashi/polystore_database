package codec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// listtype.go handles the non-scalar mapping types:
//
// 	list<elem>  a list of elem (e.g. list<string>); plain "list" accepts any element type
// 	json        any complex value (arrays, structs)
//
// Neo4j (arrays) and MongoDB (arrays) store them natively; MySQL, ClickHouse, Cassandra and kvs store JSON
// strings. When read back, a list becomes []any with elements of its type.

// ListElem returns the element type and true for a list type ("" for plain "list").
func ListElem(typeName string) (elem string, ok bool) {
	t := strings.ToLower(strings.TrimSpace(typeName))
	if t == "list" {
		return "", true
	}
	if strings.HasPrefix(t, "list<") && strings.HasSuffix(t, ">") {
		return strings.TrimSpace(t[len("list<") : len(t)-1]), true
	}
	return "", false
}

// IsComplexType reports a non-scalar type (list / json): no reverse index, no predicates evaluated in stores.
func IsComplexType(typeName string) bool {
	if _, ok := ListElem(typeName); ok {
		return true
	}
	return strings.EqualFold(typeName, "json")
}

// toList turns a value into list elements (a slice or array, or a JSON array as string / bytes).
func toList(val interface{}) ([]interface{}, error) {
	switch v := val.(type) {
	case []interface{}:
		return v, nil
	case string:
		return unmarshalList([]byte(v))
	case []byte:
		return unmarshalList(v)
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("not a list: %T", val)
	}
	out := make([]interface{}, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, nil
}

func unmarshalList(b []byte) ([]interface{}, error) {
	var out []interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("not a JSON list: %w", err)
	}
	return out, nil
}

// convertList turns val into a list and converts the elements with conv (elements unchanged when elem is empty).
func convertList(val interface{}, elem string, conv func(interface{}, string) (interface{}, error)) (interface{}, error) {
	xs, err := toList(val)
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(xs))
	for i, x := range xs {
		if elem == "" || x == nil {
			out[i] = x
			continue
		}
		if out[i], err = conv(x, elem); err != nil {
			return nil, fmt.Errorf("list element %d: %w", i, err)
		}
	}
	return out, nil
}
