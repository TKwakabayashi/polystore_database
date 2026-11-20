package sqlq

import (
	"database/sql"
	"strings"
)

// ExpandIn widens the InArg-th `?` to one `?` per input id and expands the uuids into the arguments.
func ExpandIn(text string, args []any, inArg int, ids []any) (string, []any) {
	if inArg < 0 {
		return text, args
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '?' {
			if n == inArg {
				b.WriteString(strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", "))
				n++
				continue
			}
			n++
		}
		b.WriteByte(text[i])
	}
	out := make([]any, 0, len(args)+len(ids))
	for i, a := range args {
		if i == inArg {
			for _, s := range ids {
				out = append(out, s)
			}
			continue
		}
		out = append(out, a)
	}
	return b.String(), out
}

// Scan passes the result rows to emit as they arrive ([]byte becomes string).
func Scan(rows *sql.Rows, n int, emit func([]any) error) error {
	for rows.Next() {
		vals := make([]any, n)
		ptrs := make([]any, n)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		for i, v := range vals {
			if bs, ok := v.([]byte); ok {
				vals[i] = string(bs)
			}
		}
		if err := emit(vals); err != nil {
			return err
		}
	}
	return rows.Err()
}
