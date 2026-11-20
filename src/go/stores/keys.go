package stores

import (
	"fmt"

	"polystore_database/src/go/base/id"
)

// Collect turns a row-by-row function into one returning every row.
func Collect(stream func(emit func([]any) error) error) ([][]any, error) {
	var out [][]any
	err := stream(func(vals []any) error {
		out = append(out, vals)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ColNames returns the result column names c0, c1, ... (the order of Query.Columns); built once per query, not per row.
func ColNames(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("c%d", i)
	}
	return out
}

// StoredIDs returns the first id column of the input as a typed slice of stored values (id.StoredSlice).
func StoredIDs(in []Key) any {
	us := make([]id.UUID, len(in))
	for i, k := range in {
		us[i] = k[0]
	}
	return id.StoredSlice(us)
}

// StoredList returns the stored values of the first id column one by one (for SQL placeholders).
func StoredList(in []Key) []any {
	out := make([]any, len(in))
	for i, k := range in {
		out[i] = id.Stored(k[0])
	}
	return out
}

// ShipValue converts one shipped value for the driver (ids as stored values; as text inside JSON).
func ShipValue(v any, inJSON bool) any {
	u, ok := v.(id.UUID)
	if !ok {
		return v // row number, position, weight (int64) or null
	}
	if inJSON {
		return id.StoredText(u)
	}
	return id.Stored(u)
}

// Key is one input row's id columns, in the order of Query.InKeys.
type Key []id.UUID

// Shipment is the table passed to a shipped query; each row follows Query.ShipCols.
type Shipment struct {
	Rows [][]any
}
