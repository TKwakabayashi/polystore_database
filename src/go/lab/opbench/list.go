package opbench

import (
	"context"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"polystore_database/src/go/base/id"
)

// ListSpec says how to build the uuid list that is actually passed in: run the real query's earlier part on Neo4j and
// save the results in order (duplicates kept).
//
// - Starts only: Starts' results (column u) shuffled with a fixed seed.
// - Starts and Query: Starts' results shuffled and passed Batch at a time as Query's $s; column u collected up to Max.
// - Query only: run Query once. When it also returns an integer column k, a list of N takes the N smallest k in
// result order (like picking N with a range predicate).
//
// Query also receives the remaining count as $limit.
type ListSpec struct {
	Name   string `json:"name"`
	Max    int    `json:"max"`    // the most uuids to collect
	Starts string `json:"starts"` // Cypher returning the starts (column u; need not be uuids)
	Query  string `json:"query"`  // Cypher returning column u (uuids)
	Batch  int    `json:"batch"`  // starts passed to Query at a time (default 1)
}

func (l *ListSpec) validate() error {
	if l.Name == "" || l.Max <= 0 {
		return fmt.Errorf("list: name and max (positive) are required: %+v", l)
	}
	if l.Starts == "" && l.Query == "" {
		return fmt.Errorf("list %q: starts or query is required", l.Name)
	}
	if l.Batch < 0 {
		return fmt.Errorf("list %q: batch must be >= 0", l.Name)
	}
	return nil
}

// inputList is a saved uuid list; rank holds column k (nil if absent).
type inputList struct {
	ids  []id.UUID
	rank []int64
}

func (l *inputList) len() int { return len(l.ids) }

// take returns n inputs: the first n without rank, otherwise the n smallest ranks in their original order. sorted
// orders them by id (duplicates kept).
func (l *inputList) take(n int, sorted bool) []id.UUID {
	if n > len(l.ids) {
		n = len(l.ids)
	}
	var out []id.UUID
	if l.rank == nil {
		out = append(out, l.ids[:n]...)
	} else {
		idx := make([]int, len(l.ids))
		for i := range idx {
			idx[i] = i
		}
		sort.SliceStable(idx, func(a, b int) bool { return l.rank[idx[a]] < l.rank[idx[b]] })
		idx = idx[:n]
		sort.Ints(idx)
		out = make([]id.UUID, n)
		for i, x := range idx {
			out[i] = l.ids[x]
		}
	}
	if sorted {
		sortIDs(out)
	}
	return out
}

func sortIDs(ids []id.UUID) {
	sort.Slice(ids, func(i, j int) bool { return id.Less(ids[i], ids[j]) })
}

func distinctIDs(ids []id.UUID) int {
	seen := make(map[id.UUID]struct{}, len(ids))
	for _, u := range ids {
		seen[u] = struct{}{}
	}
	return len(seen)
}

// loadList reads dir/<name>.csv if present, else builds and saves it. The same list replays for every condition, so
// a saved list is never rebuilt (delete the file to rebuild).
func (r *Runner) loadList(ctx context.Context, spec *ListSpec) (*inputList, error) {
	path := filepath.Join(r.InputDir, spec.Name+".csv")
	l, err := readList(path)
	if err == nil {
		fmt.Fprintf(r.Log, "[list %s] read %[3]d from %[2]s\n", spec.Name, path, l.len())
		return l, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("list %q: %w", spec.Name, err)
	}
	if l, err = r.buildList(ctx, spec); err != nil {
		return nil, fmt.Errorf("list %q: %w", spec.Name, err)
	}
	if err := writeList(path, l); err != nil {
		return nil, fmt.Errorf("list %q: %w", spec.Name, err)
	}
	fmt.Fprintf(r.Log, "[list %s] saved %d (%d distinct uuids) to %s\n", spec.Name, l.len(), distinctIDs(l.ids), path)
	return l, nil
}

func (r *Runner) buildList(ctx context.Context, spec *ListSpec) (*inputList, error) {
	l := &inputList{}
	add := func(rows []map[string]any, room int) error {
		if len(rows) > room {
			rows = rows[:room]
		}
		for _, row := range rows {
			u := id.FromAny(row["u"])
			if u.Empty() {
				return fmt.Errorf("not a uuid: %v", row["u"])
			}
			l.ids = append(l.ids, u)
			if k, ok := row["k"].(int64); ok {
				l.rank = append(l.rank, k)
			}
		}
		return nil
	}
	switch {
	case spec.Starts == "":
		rows, err := r.Backend.Verbatim(ctx, spec.Query, map[string]any{"limit": int64(spec.Max)})
		if err != nil {
			return nil, err
		}
		if err := add(rows, spec.Max); err != nil {
			return nil, err
		}
	default:
		rows, err := r.Backend.Verbatim(ctx, spec.Starts, map[string]any{})
		if err != nil {
			return nil, err
		}
		starts := shuffled(rows, spec.Name)
		if spec.Query == "" {
			picked := make([]map[string]any, min(len(starts), spec.Max))
			for i := range picked {
				picked[i] = map[string]any{"u": starts[i]}
			}
			return l, add(picked, spec.Max)
		}
		batch := max(spec.Batch, 1)
		for i := 0; i < len(starts) && l.len() < spec.Max; i += batch {
			room := spec.Max - l.len()
			params := map[string]any{"s": starts[i:min(i+batch, len(starts))], "limit": int64(room)}
			rows, err := r.Backend.Verbatim(ctx, spec.Query, params)
			if err != nil {
				return nil, err
			}
			if err := add(rows, room); err != nil {
				return nil, err
			}
		}
	}
	if l.rank != nil && len(l.rank) != len(l.ids) {
		return nil, fmt.Errorf("a row whose column k is not an integer")
	}
	return l, nil
}

// shuffled sorts the result column u by value and shuffles it with a seed from the list name (the same data gives the same order).
func shuffled(rows []map[string]any, name string) []any {
	keys := make([]string, len(rows))
	byKey := make(map[string]any, len(rows))
	for i, row := range rows {
		keys[i] = fmt.Sprint(row["u"])
		byKey[keys[i]] = row["u"]
	}
	sort.Strings(keys)
	h := fnv.New64a()
	h.Write([]byte(name))
	rng := rand.New(rand.NewSource(int64(h.Sum64())))
	rng.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	out := make([]any, len(keys))
	for i, k := range keys {
		out[i] = byKey[k]
	}
	return out
}

func writeList(path string, l *inputList) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	header := []string{"u"}
	if l.rank != nil {
		header = append(header, "k")
	}
	if err := w.Write(header); err != nil {
		return err
	}
	for i, u := range l.ids {
		rec := []string{u.String()}
		if l.rank != nil {
			rec = append(rec, strconv.FormatInt(l.rank[i], 10))
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readList(path string) (*inputList, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cr := csv.NewReader(f)
	header, err := cr.Read()
	if err != nil {
		return nil, err
	}
	l := &inputList{}
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			return l, nil
		}
		if err != nil {
			return nil, err
		}
		u, err := id.Parse(rec[0])
		if err != nil {
			return nil, err
		}
		l.ids = append(l.ids, u)
		if len(header) > 1 {
			k, err := strconv.ParseInt(rec[1], 10, 64)
			if err != nil {
				return nil, err
			}
			l.rank = append(l.rank, k)
		}
	}
}
