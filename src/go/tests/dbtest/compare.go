// Package dbtest holds helpers shared by the database tests (tests/db, tests/throwaway): canonical rows
// and readable differences.
package dbtest

import (
	"encoding/json"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"sort"
	"strconv"
	"strings"
	"time"
)

// JSONNumbers reads a JSON array with integers as int64 and other numbers as float64 (json.Unmarshal makes every
// number a float64, so an id IN list would reach Neo4j as floats and come back as 1.099511628e+13).
func JSONNumbers(s string, out *[]any) error {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return err
	}
	for i, e := range *out {
		if n, ok := e.(json.Number); ok {
			if iv, err := n.Int64(); err == nil {
				(*out)[i] = iv
			} else if fv, err := n.Float64(); err == nil {
				(*out)[i] = fv
			}
		}
	}
	return nil
}

// Canon renders rows as order-independent strings (with uniform number and time forms).
func Canon(rows []map[string]interface{}) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%s;", k, canonValue(row[k]))
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return out
}

func canonValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case int:
		return fmt.Sprint(int64(x))
	case int32:
		return fmt.Sprint(int64(x))
	case int64:
		return fmt.Sprint(x)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', 10, 64)
	case float64:
		// The summation order of aggregates moves the last digits, so compare 10 digits.
		return strconv.FormatFloat(x, 'g', 10, 64)
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano)
	case dbtype.Date: // the oracle's (Neo4j) DATE; the executor returns time.Time at UTC midnight
		return time.Time(x).UTC().Format(time.RFC3339Nano)
	case dbtype.LocalDateTime:
		return time.Time(x).UTC().Format(time.RFC3339Nano)
	case []interface{}:
		// Cypher does not fix collect's element order either, so lists compare as multisets.
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = canonValue(e)
		}
		sort.Strings(parts)
		return "[" + strings.Join(parts, ",") + "]"
	}
	return fmt.Sprint(v)
}

func FirstDiff(want, got []string) string {
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			return fmt.Sprintf("  row %d:\n    oracle: %s\n    got:    %s", i, want[i], got[i])
		}
	}
	return fmt.Sprintf("  only the lengths differ: oracle=%d got=%d", len(want), len(got))
}
