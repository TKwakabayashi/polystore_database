package runner

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// canonRows makes a row multiset comparable as a string (floats to 10 significant digits).
func canonRows(rows []map[string]any) string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			v := row[k]
			switch x := v.(type) {
			case float64:
				v = fmt.Sprintf("%.10g", x) // the order of an avg changes the last digits
			case neo4j.Date:
				v = x.Time().UTC().Format(time.RFC3339Nano) // the executor holds dates as UTC midnight
			case time.Time:
				v = x.UTC().Format(time.RFC3339Nano) // Neo4j and the executor write zones differently
			}
			fmt.Fprintf(&b, "%s=%v;", k, v)
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return strings.Join(out, "\n")
}
