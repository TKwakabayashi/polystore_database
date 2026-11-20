package runner

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/query/result"
)

var (
	runHeader = []string{"run", "spec", "workload", "query", "case", "params", "placement", "model", "tuning", "vector",
		"chunk", "trial", "rows", "plan_ms", "exec_ms", "round_trips", "trips_by_store", "exec_kind", "match", "status"}
	migrationHeader = []string{"run", "spec", "workload", "placement", "entity", "properties", "rows", "ms"}
	spanHeader      = []string{"run", "workload", "query", "case", "placement", "model", "tuning", "trial", "span", "parent",
		"level", "kind", "store", "rows_in", "rows_out", "round_trips", "ns", "db_hits"}
)

func (r *specRun) row(place, model, tuning string, vec, chunk, trial int, res result.Result, match, status string) {
	b, _ := json.Marshal(r.params) // keys sorted
	_ = r.runs.Write([]string{r.runID, r.s.Name, r.wname, r.qv, r.label, string(b), place, model, tuning,
		strconv.Itoa(vec), strconv.Itoa(chunk), strconv.Itoa(trial), strconv.Itoa(res.RowCount()), ms(res.PlanTime),
		ms(res.ExecTime), strconv.FormatInt(res.RoundTrips, 10), formatTrips(res.StoreTrips), res.ExecKind, match, status})
}

// writeSpans writes one L3 span per executed operator (leaf to root; the query is span 0) and, under each
// operator that queried a store, its store calls' first row / receive / decode times.
func (r *specRun) writeSpans(place, model, tuning, trial string, res result.Result) {
	head := []string{r.runID, r.wname, r.qv, r.label, place, model, tuning, trial}
	row := func(cols ...string) { _ = r.spans.Write(append(append([]string{}, head...), cols...)) }
	row("0", "", "L2", "query", "", "", strconv.Itoa(res.RowCount()), strconv.FormatInt(res.RoundTrips, 10), ns(res.ExecTime), "")
	for _, st := range res.Steps {
		sid := strconv.Itoa(st.Step)
		row(sid, "0", "L3", st.Op, st.Store, strconv.Itoa(st.InRows), strconv.Itoa(st.OutRows),
			strconv.FormatInt(st.RoundTrips, 10), ns(st.Duration), "")
		if st.FirstRow+st.Receive+st.Decode == 0 {
			continue
		}
		for i, part := range []struct {
			kind string
			d    time.Duration
		}{{"first_row", st.FirstRow}, {"receive", st.Receive}, {"decode", st.Decode}} {
			row(sid+"."+strconv.Itoa(i+1), sid, "L3", part.kind, st.Store, "", "", "", ns(part.d), "")
		}
	}
	r.spans.Flush()
}

// writeStoreProfiles writes the stores' reports (L4) of a profiled trial: one span per store query under its
// operator (trial "l4"), and the raw output per operator to l4/<condition>.json.
func (r *specRun) writeStoreProfiles(place, model, tuning string, res result.Result) error {
	type stepOut struct {
		Step     int                   `json:"step"`
		Op       string                `json:"op"`
		Desc     string                `json:"desc"`
		Profiles []result.StoreProfile `json:"profiles"`
	}
	var out []stepOut
	head := []string{r.runID, r.wname, r.qv, r.label, place, model, tuning, "l4"}
	for _, st := range res.Steps {
		if len(st.Profiles) == 0 {
			continue
		}
		sid := strconv.Itoa(st.Step)
		for i, p := range st.Profiles {
			_ = r.spans.Write(append(append([]string{}, head...), sid+".p"+strconv.Itoa(i+1), sid, "L4", "store_profile",
				p.Store, "", strconv.FormatInt(p.Rows, 10), "", ns(p.Time), strconv.FormatInt(p.DBHits, 10)))
		}
		out = append(out, stepOut{st.Step, st.Op, st.Desc, st.Profiles})
	}
	r.spans.Flush()
	if len(out) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(r.dir, "l4"), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	name := strings.Join([]string{r.wname, r.qv, r.label, place, model, tuning}, "_") + ".json"
	return os.WriteFile(filepath.Join(r.dir, "l4", name), b, 0o644)
}

func formatTrips(byStore map[string]int64) string {
	keys := make([]string, 0, len(byStore))
	for k := range byStore {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + strconv.FormatInt(byStore[k], 10)
	}
	return strings.Join(parts, ";")
}

func openCSV(path string, header []string) (*csv.Writer, func(), error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	w := csv.NewWriter(f)
	_ = w.Write(header)
	return w, func() { w.Flush(); f.Close() }, nil
}

func ms(d time.Duration) string {
	return strconv.FormatFloat(float64(d.Microseconds())/1000, 'f', 3, 64)
}

func ns(d time.Duration) string { return strconv.FormatInt(d.Nanoseconds(), 10) }
