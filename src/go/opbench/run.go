package opbench

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/executor"
	"polystore_database/src/go/id"
	"polystore_database/src/go/native"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
)

// Run は定義ファイル specPath の実験をすべて実行し、計測を outPath（CSV）へ、条件ごとの物理計画と
// ネイティブクエリを outPath の隣の .plans.txt へ書く。outPath が空なら ../../results/opbench/<name>.csv。
func Run(ctx context.Context, cfg storage.Config, specPath, outPath string) error {
	spec, err := Load(specPath)
	if err != nil {
		return err
	}
	base, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		return err
	}
	reg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return err
	}
	defer reg.Close(ctx)
	if outPath == "" {
		outPath = filepath.Join("..", "..", "results", "opbench", spec.Name+".csv")
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

	r := &Runner{Backend: executor.NewStoreBackend(reg), Base: base, CSV: csv.NewWriter(csvFile), Plans: plansFile, Log: os.Stdout}
	if err := r.Run(ctx, spec); err != nil {
		return err
	}
	log.Printf("opbench: 計測を %s、計画を %s へ書いた", outPath, plansPath)
	return nil
}

// Runner は実験を実行する（Backend を差し替えればストア無しでも動かせる）。
type Runner struct {
	Backend executor.Backend
	Base    *schema.MappingDictionary
	CSV     *csv.Writer
	Plans   io.Writer
	Log     io.Writer

	refs map[string]string // 正しさの比較の基準（graph 配置・bulk・既定の設定）の結果
}

// Header は CSV の列。1 行が 1 回の実行。
var Header = []string{"spec", "experiment", "op", "input", "rows", "placement", "at", "model", "chunk", "parallelism",
	"vector", "rep", "exec_ms", "target_ms", "target_step", "round_trips", "out_rows", "out_rows_expanded",
	"match_graph", "steps", "error"}

// Run は spec の全実験・全条件を実行する。
func (r *Runner) Run(ctx context.Context, spec *Spec) error {
	timeout, _ := time.ParseDuration(spec.Timeout)
	r.refs = map[string]string{}
	if err := r.CSV.Write(Header); err != nil {
		return err
	}
	for i := range spec.Experiments {
		e := &spec.Experiments[i]
		var ids []id.UUID
		if n := e.MaxRows(); n > 0 {
			var err error
			if ids, err = r.sample(ctx, e, n); err != nil {
				return fmt.Errorf("experiment %q: 入力の抜き出し: %w", e.Name, err)
			}
			if len(ids) < n {
				fmt.Fprintf(r.Log, "[%s] 入力が %d 件しかない（%d 件を要求）。それより多い件数は %d 件で実行する\n", e.Name, len(ids), n, len(ids))
			}
		}
		for _, c := range e.Conditions() {
			if err := r.condition(ctx, spec, e, c, ids, timeout); err != nil {
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

// sample は入力の頂点の uuid を n 件抜き出す（uuid 順なので決定的。uuid は乱数なので標本は偏らない）。
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

// planned は条件 1 つ分の計画。
type planned struct {
	p    physical.Node
	cat  *physical.Catalog
	args map[*physical.ArgRows][]executor.Row
}

func (r *Runner) plan(e *Experiment, c Condition, ids []id.UUID, md *schema.MappingDictionary) (*planned, error) {
	lp, err := BuildLogical(e, c)
	if err != nil {
		return nil, err
	}
	cat := physical.NewCatalog(md)
	p, err := physical.Build(lp, cat)
	if err != nil {
		return nil, err
	}
	pl := &planned{p: p, cat: cat, args: map[*physical.ArgRows][]executor.Row{}}
	n := c.Rows
	if n > len(ids) {
		n = len(ids)
	}
	var walk func(physical.Node)
	walk = func(x physical.Node) {
		if a, ok := x.(*physical.ArgRows); ok {
			rows := make([]executor.Row, n)
			for i, u := range ids[:n] {
				rows[i] = executor.Row{inAlias: u}
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
	label := fmt.Sprintf("[%s] input=%s rows=%d placement=%s at=%s model=%s chunk=%s par=%s vec=%s", e.Name, c.Input, c.Rows,
		c.Placement, c.At, c.Model, knob(c.Chunk, -1), knob(c.Parallelism, 0), knob(c.Vector, 0))
	base := []string{spec.Name, e.Name, e.Op, c.Input, strconv.Itoa(c.Rows), c.Placement, c.At, c.Model,
		knob(c.Chunk, -1), knob(c.Parallelism, 0), knob(c.Vector, 0)}
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
		ex := executor.New(tctx, r.Backend, pl.cat.TypeOf)
		start := time.Now()
		rows, err := ex.ExecuteWithArgs(pl.p, settings.EngineKind(c.Model), pl.args)
		dur := time.Since(start)
		cancel()
		if err != nil {
			return fail(rep, err) // 打ち切り・失敗は以降の繰り返しも同じなので止める
		}
		if rep < 0 {
			continue // 事前実行は記録しない
		}
		steps := ex.Steps()
		var target time.Duration
		targetStep := ""
		if len(steps) > 0 {
			last := steps[len(steps)-1] // 根（対象）のステップは最後
			target, targetStep = last.Duration, last.Op
		}
		execs, targets = append(execs, dur), append(targets, target)
		parts := make([]string, len(steps))
		for i, s := range steps {
			parts[i] = fmt.Sprintf("%s:%.3fms:%d", s.Op, ms(s.Duration), s.OutRows)
		}
		row := append(append([]string{}, base...), strconv.Itoa(rep+1), fmt.Sprintf("%.3f", ms(dur)),
			fmt.Sprintf("%.3f", ms(target)), targetStep, strconv.FormatInt(ex.RoundTrips(), 10), strconv.Itoa(len(rows)),
			strconv.Itoa(len(executor.OutputRows(pl.p, rows))), match, strings.Join(parts, " | "), "")
		if err := r.CSV.Write(row); err != nil {
			return err
		}
	}
	fmt.Fprintf(r.Log, "%s: exec %.1fms target %.1fms (median of %d) match=%s\n", label, ms(median(execs)), ms(median(targets)), len(execs), match)
	return nil
}

// check は結果が graph 配置（bulk・既定の設定）と同じ行の多重集合かを返す（true / false / n/a）。
// 並べ替えの無い Limit は配置で選ばれる行が変わり得るので比べない。
func (r *Runner) check(ctx context.Context, e *Experiment, c Condition, ids []id.UUID, pl *planned, timeout time.Duration) string {
	if e.Op == "Limit" {
		return "n/a"
	}
	key := fmt.Sprintf("%s|%s|%d|%s", e.Name, c.Input, c.Rows, c.At)
	run := func(p *planned, model settings.EngineKind) (string, error) {
		tctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		rows, err := executor.New(tctx, r.Backend, p.cat.TypeOf).ExecuteWithArgs(p.p, model, p.args)
		if err != nil {
			return "", err
		}
		return canon(executor.OutputRows(p.p, rows)), nil
	}
	ref, ok := r.refs[key]
	if !ok {
		gc := c
		gc.Placement, gc.Model = "graph", "bulk"
		gp, err := r.plan(e, gc, ids, r.Base)
		if err != nil {
			return "n/a"
		}
		if ref, err = run(gp, settings.EngineBulk); err != nil {
			return "n/a"
		}
		r.refs[key] = ref
	}
	got, err := run(pl, settings.EngineKind(c.Model))
	if err != nil {
		return "n/a"
	}
	return strconv.FormatBool(got == ref)
}

// writePlan は条件の物理計画と、各ストアへ投げるネイティブクエリを書く。
func (r *Runner) writePlan(label string, pl *planned) {
	fmt.Fprintf(r.Plans, "### %s\n%s", label, physical.Dump(pl.p))
	var walk func(physical.Node)
	walk = func(n physical.Node) {
		for _, ch := range n.Children() {
			walk(ch)
		}
		if q, ok := n.(*physical.StoreQuery); ok {
			if nq, err := native.Lower(q, pl.cat.TypeOf); err == nil {
				fmt.Fprintf(r.Plans, "%s\n", nq.Render())
			} else {
				fmt.Fprintf(r.Plans, "(lower: %v)\n", err)
			}
		}
	}
	walk(pl.p)
	fmt.Fprintln(r.Plans)
}

// apply は条件の設定（チャンク幅・並列度・ベクトル長）を反映し、元へ戻す関数を返す。
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

// canon は行の多重集合を比べられる文字列にする。
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
	return strings.Join(out, "\n")
}
