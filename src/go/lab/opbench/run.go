package opbench

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"polystore_database/src/go/internal/paths"
	"polystore_database/src/go/query"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/dbs"
	"sort"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores/conn"
)

// Run runs every experiment of specPath, writing measurements to outPath (CSV) and each condition's physical plan
// and native queries to the .plans.txt next to it. An empty outPath means results/opbench/<name>.csv.
func Run(ctx context.Context, cfg conn.Config, specPath, outPath string) error {
	spec, err := Load(specPath)
	if err != nil {
		return err
	}
	base, err := catalog.LoadRecord(cfg.MappingPath())
	if err != nil {
		return err
	}
	reg, err := conn.NewRegistry(ctx, cfg)
	if err != nil {
		return err
	}
	defer reg.Close(ctx)
	if outPath == "" {
		outPath = paths.Resolve(filepath.Join("..", "..", "results", "opbench", spec.Name+".csv"))
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	csvFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	plansPath := strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".plans.txt"
	plansFile, err := os.Create(plansPath)
	if err != nil {
		return err
	}
	defer plansFile.Close()

	r := &Runner{Backend: dbs.NewSession(reg), Base: base, CSV: csv.NewWriter(csvFile), Plans: plansFile, Log: os.Stdout,
		InputDir: filepath.Join(filepath.Dir(outPath), "inputs")}
	if err := r.Run(ctx, spec); err != nil {
		return err
	}
	log.Printf("opbench: measurements in %s, plans in %s", outPath, plansPath)
	return nil
}

// Runner runs experiments (with another Backend it runs without stores).
type Runner struct {
	Backend exec.Backend
	Base    *catalog.MappingDictionary
	CSV     *csv.Writer
	Plans   io.Writer
	Log     io.Writer

	InputDir string // directory of the uuid list files

	refs  map[string]string     // result hashes of the reference (Neo4j placement, bulk, default settings)
	lists map[string]*inputList // name -> uuid list
}

// Header is the CSV header; one row per run.
var Header = []string{"spec", "experiment", "role", "op", "prop", "predicate", "input", "list", "order", "rows", "in_distinct",
	"placement", "at", "model", "chunk", "parallelism", "vector", "rep", "exec_ms", "target_ms", "target_step", "round_trips",
	"out_rows", "out_rows_expanded", "match_graph", "steps", "error"}

// Run runs every experiment and condition of spec.
func (r *Runner) Run(ctx context.Context, spec *Spec) error {
	timeout, _ := time.ParseDuration(spec.Timeout)
	r.refs = map[string]string{}
	r.lists = map[string]*inputList{}
	if err := r.CSV.Write(Header); err != nil {
		return err
	}
	for i := range spec.Lists {
		l, err := r.loadList(ctx, &spec.Lists[i])
		if err != nil {
			return err
		}
		r.lists[spec.Lists[i].Name] = l
	}
	for i := range spec.Experiments {
		e := &spec.Experiments[i]
		var ids []id.UUID
		if n := e.MaxRows(); n > 0 {
			var err error
			if ids, err = r.sample(ctx, e, n); err != nil {
				return fmt.Errorf("experiment %q: sampling the input: %w", e.Name, err)
			}
			if len(ids) < n {
				fmt.Fprintf(r.Log, "[%s] only %d input rows (%d asked); larger sizes run with %d\n", e.Name, len(ids), n, len(ids))
			}
		}
		for _, c := range e.Conditions() {
			in := ids
			switch {
			case c.List != "":
				l := r.lists[c.List]
				if c.Rows > l.len() {
					continue // the list is too short for the count (counts are not compared on different inputs)
				}
				in = l.take(c.Rows, c.Order == "sorted")
			case c.Order == "sorted":
				in = append([]id.UUID{}, ids...)
				sortIDs(in)
			}
			if err := r.condition(ctx, spec, e, c, in, timeout); err != nil {
				return fmt.Errorf("experiment %q: %w", e.Name, err)
			}
			r.CSV.Flush()
			if err := r.CSV.Error(); err != nil {
				return err
			}
		}
	}
	return nil
}

// sample takes n uuids of the input vertex (in uuid order: deterministic; uuids are random, so unbiased).
func (r *Runner) sample(ctx context.Context, e *Experiment, n int) ([]id.UUID, error) {
	q := fmt.Sprintf("MATCH (n:`%s`)", e.Label)
	if e.Sample != "" {
		q += " WHERE " + e.Sample
	}
	q += " RETURN n.uuid AS u ORDER BY n.uuid LIMIT $n"
	rows, err := r.Backend.Verbatim(ctx, q, map[string]any{"n": int64(n)})
	if err != nil {
		return nil, err
	}
	out := make([]id.UUID, 0, len(rows))
	for _, row := range rows {
		if u := id.FromAny(row["u"]); !u.Empty() {
			out = append(out, u)
		}
	}
	return out, nil
}

// planned is the plan of one condition.
type planned struct {
	p    physical.Node
	cat  *physical.Placement
	args map[*physical.ArgRows][]exec.Row
}

func (r *Runner) plan(e *Experiment, c Condition, ids []id.UUID, md *catalog.MappingDictionary) (*planned, error) {
	lp, err := BuildLogical(e, c)
	if err != nil {
		return nil, err
	}
	cat := physical.NewPlacement(md)
	popt := settings.Plan().Physical
	popt.UniqueKeys = md.UniqueKeys()
	p, err := physical.Build(lp, cat, popt)
	if err != nil {
		return nil, err
	}
	pl := &planned{p: p, cat: cat, args: map[*physical.ArgRows][]exec.Row{}}
	n := c.Rows
	if n > len(ids) {
		n = len(ids)
	}
	var walk func(physical.Node)
	walk = func(x physical.Node) {
		if a, ok := x.(*physical.ArgRows); ok {
			rows := make([]exec.Row, n)
			for i, u := range ids[:n] {
				rows[i] = exec.Row{inAlias: u}
			}
			pl.args[a] = rows
		}
		for _, ch := range x.Children() {
			walk(ch)
		}
	}
	walk(p)
	return pl, nil
}

func (r *Runner) condition(ctx context.Context, spec *Spec, e *Experiment, c Condition, ids []id.UUID, timeout time.Duration) error {
	label := fmt.Sprintf("[%s] input=%s list=%s order=%s rows=%d placement=%s at=%s model=%s chunk=%s par=%s vec=%s", e.Name, c.Input,
		c.List, c.Order, c.Rows, c.Placement, c.At, c.Model, knob(c.Chunk, -1), knob(c.Parallelism, 0), knob(c.Vector, 0))
	predicate := ""
	if e.Cmp != "" {
		predicate = fmt.Sprintf("%s %s %v", e.Prop, e.Cmp, e.Value)
	}
	distinct := ""
	if c.Input == "ids" {
		distinct = strconv.Itoa(distinctIDs(ids[:min(c.Rows, len(ids))]))
	}
	base := []string{spec.Name, e.Name, e.Role, e.Op, strings.Join(e.Props(), ";"), predicate, c.Input, c.List, c.Order,
		strconv.Itoa(c.Rows), distinct, c.Placement, c.At, c.Model, knob(c.Chunk, -1), knob(c.Parallelism, 0), knob(c.Vector, 0)}
	fail := func(rep int, err error) error {
		fmt.Fprintf(r.Log, "%s: %v\n", label, err)
		return r.CSV.Write(append(append([]string{}, base...), strconv.Itoa(rep), "", "", "", "", "", "", "", "", err.Error()))
	}

	md, err := PlaceMapping(r.Base, e, c.Placement)
	if err != nil {
		return fail(0, err)
	}
	pl, err := r.plan(e, c, ids, md)
	if err != nil {
		return fail(0, err)
	}
	r.writePlan(label, pl)

	match := r.check(ctx, e, c, ids, pl, timeout)

	restore := apply(c)
	defer restore()
	var execs, targets []time.Duration
	for rep := -spec.Warmup; rep < spec.Repeat; rep++ {
		tctx, cancel := context.WithTimeout(ctx, timeout)
		ex := exec.New(tctx, r.Backend, pl.cat.TypeOf, settings.Exec(exec.Model(c.Model)))
		start := time.Now()
		rows, err := ex.ExecuteWithArgs(pl.p, pl.args)
		dur := time.Since(start)
		cancel()
		if err != nil {
			return fail(rep, err) // timeouts and failures repeat, so stop
		}
		if rep < 0 {
			continue // warmup runs are not recorded
		}
		steps := ex.Steps()
		var target time.Duration
		targetStep := ""
		if len(steps) > 0 {
			last := steps[len(steps)-1] // the root (target) step is last
			target, targetStep = last.Duration, last.Op
		}
		execs, targets = append(execs, dur), append(targets, target)
		parts := make([]string, len(steps))
		for i, s := range steps {
			parts[i] = fmt.Sprintf("%s:%.3fms:%d", s.Op, ms(s.Duration), s.OutRows)
		}
		row := append(append([]string{}, base...), strconv.Itoa(rep+1), fmt.Sprintf("%.3f", ms(dur)),
			fmt.Sprintf("%.3f", ms(target)), targetStep, strconv.FormatInt(ex.RoundTrips(), 10), strconv.Itoa(len(rows)),
			strconv.Itoa(len(exec.OutputRows(pl.p, rows))), match, strings.Join(parts, " | "), "")
		if err := r.CSV.Write(row); err != nil {
			return err
		}
	}
	fmt.Fprintf(r.Log, "%s: exec %.1fms target %.1fms (median of %d) match=%s\n", label, ms(median(execs)), ms(median(targets)), len(execs), match)
	return nil
}

// check reports whether the result equals the Neo4j placement's (bulk, default settings) as a multiset (true /
// false / n/a). A Limit without a sort may pick other rows per placement, so it is not compared.
func (r *Runner) check(ctx context.Context, e *Experiment, c Condition, ids []id.UUID, pl *planned, timeout time.Duration) string {
	if e.Op == "Limit" {
		return "n/a"
	}
	key := fmt.Sprintf("%s|%s|%s|%d|%s", e.Name, c.Input, c.List, c.Rows, c.At) // sorting does not change the multiset
	run := func(p *planned, model exec.Model) (string, error) {
		tctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		rows, err := exec.New(tctx, r.Backend, p.cat.TypeOf, settings.Exec(model)).ExecuteWithArgs(p.p, p.args)
		if err != nil {
			return "", err
		}
		return canon(exec.OutputRows(p.p, rows)), nil
	}
	ref, ok := r.refs[key]
	if !ok {
		gc := c
		gc.Placement, gc.Model = "graph", "bulk"
		gp, err := r.plan(e, gc, ids, r.Base)
		if err != nil {
			return "n/a"
		}
		if ref, err = run(gp, exec.ModelBulk); err != nil {
			return "n/a"
		}
		r.refs[key] = ref
	}
	got, err := run(pl, exec.Model(c.Model))
	if err != nil {
		return "n/a"
	}
	return strconv.FormatBool(got == ref)
}

// writePlan writes the condition's physical plan and the native query sent to each store.
func (r *Runner) writePlan(label string, pl *planned) {
	fmt.Fprintf(r.Plans, "### %s\n%s\n", label, query.RenderPlan(pl.p, pl.cat))
}

// apply sets the condition's settings (chunk size, parallelism, vector size) and returns a function restoring them.
func apply(c Condition) func() {
	chunk, par, vec := settings.MaterializeChunkSize, settings.StoreFetchParallelism, settings.VectorSize
	if c.Chunk >= 0 {
		settings.MaterializeChunkSize = c.Chunk
	}
	if c.Parallelism > 0 {
		settings.StoreFetchParallelism = c.Parallelism
	}
	if c.Vector > 0 {
		settings.VectorSize = c.Vector
	}
	return func() {
		settings.MaterializeChunkSize, settings.StoreFetchParallelism, settings.VectorSize = chunk, par, vec
	}
}

func knob(v, unset int) string {
	if v == unset {
		return "default"
	}
	return strconv.Itoa(v)
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func median(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	s := append([]time.Duration{}, ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s[len(s)/2]
}

// canon reduces a multiset of rows to a short comparable hash (so large results are not kept as references).
func canon(rows []map[string]any) string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			if k != barrierVar {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%v;", k, row[k])
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	sum := sha256.Sum256([]byte(strings.Join(out, "\n")))
	return hex.EncodeToString(sum[:])
}
