// Package eval is the value semantics of the query engine: rows, comparison and ordering,
// expression evaluation and decoding of stored values. The executor and the key-value store use it.
package eval

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"math"
	"polystore_database/src/go/query/ir"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/base/id"
)

// Row is one row. Column keys follow the ir conventions (id column "a" holds an id.UUID, value column
// "a.p", segment and path columns hold *ListVal, list value columns []any).
type Row map[string]any

// ListVal is the value of a segment or path column (node and relationship ids). A segment's Nodes are the inner nodes only.
type ListVal struct {
	Nodes []id.UUID
	Rels  []id.UUID
}

func (r Row) Clone() Row {
	out := make(Row, len(r))
	for k, v := range r {
		out[k] = v
	}
	return out
}

// Normalize brings driver values to comparable common types (integers -> int64, floats -> float64,
// []byte -> string, times -> UTC).
func Normalize(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return float64(x)
	case []byte:
		return string(x)
	case time.Time:
		return x.UTC()
	case dbtype.Date: // Neo4j DATE; aligned with other stores' dates (UTC midnight)
		return time.Time(x).UTC()
	case dbtype.LocalDateTime:
		return time.Time(x).UTC()
	case []any: // list values (collect, ...)
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = Normalize(e)
		}
		return out
	}
	return v
}

// ---- comparison (Cypher semantics) ----

// typeRank orders values of different types (for ORDER BY). null is always last.
func typeRank(v any) int {
	switch v.(type) {
	case bool:
		return 1
	case int64, float64:
		return 2
	case string:
		return 3
	case time.Time:
		return 4
	case id.UUID:
		return 5
	case nil:
		return 9
	}
	return 6
}

// compare compares a and b. ok=false means incomparable (different types or null): null in three-valued logic.
func compare(a, b any) (int, bool) {
	if a == nil || b == nil {
		return 0, false
	}
	switch x := a.(type) {
	case int64:
		switch y := b.(type) {
		case int64:
			return cmpOrdered(x, y), true
		case float64:
			return cmpOrdered(float64(x), y), true
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return cmpOrdered(x, float64(y)), true
		case float64:
			return cmpOrdered(x, y), true
		}
	case string:
		if y, ok := b.(string); ok {
			return strings.Compare(x, y), true
		}
	case bool:
		if y, ok := b.(bool); ok {
			switch {
			case x == y:
				return 0, true
			case !x:
				return -1, true
			}
			return 1, true
		}
	case time.Time:
		if y, ok := b.(time.Time); ok {
			return x.Compare(y), true
		}
	case id.UUID:
		if y, ok := b.(id.UUID); ok {
			return id.Compare(x, y), true
		}
	}
	return 0, false
}

func cmpOrdered[T int64 | float64](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// OrderCompare is the ORDER BY comparison (a total order: null last, different types by type).
func OrderCompare(a, b any) int {
	if c, ok := compare(a, b); ok {
		return c
	}
	ra, rb := typeRank(a), typeRank(b)
	if ra != rb {
		return ra - rb
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

// equalValues is equality for DISTINCT and grouping (null equals null).
func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	c, ok := compare(a, b)
	return ok && c == 0
}

// KeyString turns a tuple of values into a key for grouping, deduplication and joins.
func KeyString(vals ...any) string {
	return string(appendKeys(nil, vals...))
}

// appendKeys appends a tuple's key to b (KeyString without allocation; a map lookup with m[string(b)] does not copy).
// Each value is a type mark, its content and a \x00 separator.
func appendKeys(b []byte, vals ...any) []byte {
	for _, v := range vals {
		b = AppendKey(b, v)
		b = append(b, 0)
	}
	return b
}

func AppendKey(b []byte, v any) []byte {
	switch x := v.(type) {
	case nil:
		return append(b, 'n')
	case int64:
		return strconv.AppendFloat(append(b, 'f'), float64(x), 'g', -1, 64) // 1 and 1.0 get the same key
	case float64:
		if math.IsNaN(x) {
			return append(b, "nan"...)
		}
		return strconv.AppendFloat(append(b, 'f'), x, 'g', -1, 64)
	case string:
		return append(append(b, 's'), x...)
	case id.UUID:
		return append(append(b, 'u'), x[:]...) // fixed-length raw bytes, not hex
	case time.Time:
		return x.UTC().AppendFormat(append(b, 't'), time.RFC3339Nano)
	case *ListVal:
		b = strconv.AppendInt(append(b, 'L'), int64(len(x.Nodes)), 10) // raw bytes may contain the separator, so the length comes first
		b = append(b, ':')
		for _, n := range x.Nodes {
			b = append(b, n[:]...)
		}
		b = strconv.AppendInt(append(b, '|'), int64(len(x.Rels)), 10)
		b = append(b, ':')
		for _, r := range x.Rels {
			b = append(b, r[:]...)
		}
		return b
	case []any:
		b = append(b, '[')
		for _, e := range x {
			b = append(AppendKey(b, e), ',')
		}
		return append(b, ']')
	default:
		return fmt.Appendf(b, "?%v", x)
	}
}

// UUIDKey appends the raw bytes of id columns cols to b when all are non-empty id.UUIDs (false otherwise).
// A cheaper key than KeyString for boundary deduplication and result lookup in StoreQuery.
func UUIDKey(b []byte, r Row, cols []string) ([]byte, bool) {
	b = b[:0]
	for _, c := range cols {
		u, ok := r[c].(id.UUID)
		if !ok || u.Empty() {
			return b, false
		}
		b = append(b, u[:]...)
	}
	return b, true
}

// ---- weights (ir.StoreQuery.Weighted) ----

// WeightCol holds a row's weight (how many original rows it stands for); 1 when absent.
const WeightCol = ir.WeightCol

func WeightOf(r Row) int64 {
	if w, ok := r[WeightCol].(int64); ok && w > 0 {
		return w
	}
	return 1
}

// SetWeight resets the weight (removes the column for 1).
func SetWeight(r Row, w int64) {
	if w == 1 {
		delete(r, WeightCol)
		return
	}
	r[WeightCol] = w
}
