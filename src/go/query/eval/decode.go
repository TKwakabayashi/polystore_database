package eval

import (
	"strconv"
	"time"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
)

func Decode(v any, kind stores.ColKind, typ string) any {
	switch kind {
	case stores.KindIdent:
		if v == nil {
			return nil
		}
		return id.FromAny(Normalize(v))
	case stores.KindList:
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		lv := &ListVal{}
		for _, u := range AsList(m["nodes"]) {
			lv.Nodes = append(lv.Nodes, id.FromAny(Normalize(u)))
		}
		for _, u := range AsList(m["rels"]) {
			lv.Rels = append(lv.Rels, id.FromAny(Normalize(u)))
		}
		return lv
	}
	v = Normalize(v)
	s, ok := v.(string)
	if typ == "" || !ok {
		return v
	}
	if typ == "date" || typ == "datetime" {
		// MySQL returns DATETIME as "2006-01-02 15:04:05[.ffffff]" (not RFC 3339). Pick the layout from the shape
		// and parse once (trying the other layouts if it fails).
		if t, ok := ParseDateString(s); ok {
			return t
		}
		return v
	}
	switch typ {
	case "double", "float":
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return v
	case "int", "integer", "long":
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f)
		}
		return v
	}
	if t, err := codec.ConvertToNativeType(s, typ); err == nil {
		return Normalize(t)
	}
	return v
}

var dateLayouts = []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999", "2006-01-02"}

// ParseDateString parses a date or datetime string to a UTC time.Time. The separator at index 10 picks the
// likely layout first, so rows do not repeat failing parses.
func ParseDateString(s string) (time.Time, bool) {
	first := -1
	switch {
	case len(s) == 10:
		first = 2
	case len(s) > 10 && s[10] == ' ':
		first = 1
	case len(s) > 10 && s[10] == 'T':
		first = 0
	}
	if first >= 0 {
		if t, err := time.Parse(dateLayouts[first], s); err == nil {
			return t.UTC(), true
		}
	}
	for i, layout := range dateLayouts {
		if i == first {
			continue
		}
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func AsList(v any) []any {
	switch x := v.(type) {
	case []any:
		return x
	case []string:
		out := make([]any, len(x))
		for i, s := range x {
			out[i] = s
		}
		return out
	}
	return nil
}
