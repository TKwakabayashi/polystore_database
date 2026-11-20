package exec

// limit.go: Limit counting weighted rows while rows flow.

import (
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
)

type limitState struct {
	skip, count int // count < 0 is unlimited
	seen        int
}

func limitOf(n physical.Node) *limitState {
	if c, ok := n.(*physical.Coord); ok {
		if l, ok := c.Op.(*ir.Limit); ok {
			return &limitState{skip: l.Skip, count: l.Count}
		}
	}
	return nil
}

// take advances the Limit state and returns the rows of this batch to keep (done: no more input is needed).
// Weighted rows count by weight; the boundary row is kept with a trimmed weight.
func (l *limitState) take(rows []Row) (out []Row, done bool) {
	full := func() bool { return l.count >= 0 && l.seen >= l.skip+l.count }
	for _, r := range rows {
		if full() {
			return out, true
		}
		w := eval.WeightOf(r)
		lo, hi := l.seen, l.seen+int(w) // the positions [lo, hi) this row occupies
		l.seen = hi
		from, to := lo, hi
		if from < l.skip {
			from = l.skip
		}
		if l.count >= 0 && to > l.skip+l.count {
			to = l.skip + l.count
			l.seen = to
		}
		if to <= from {
			continue
		}
		if keep := int64(to - from); keep != w {
			nr := r.Clone()
			eval.SetWeight(nr, keep)
			r = nr
		}
		out = append(out, r)
	}
	return out, full()
}
