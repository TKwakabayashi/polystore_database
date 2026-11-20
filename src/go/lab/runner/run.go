package runner

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/opbench"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
)

// RunSpec runs the experiment in specPath and writes, under outDir/<date>_<name>/: runs.csv (one row per
// trial, with the Neo4j baseline when verifying), migrations.csv, plans.txt, spans.csv (profile mode or
// store profiles), l4/*.json (store profiles), pprof/ (Pprof), and copies of the spec and options.
func RunSpec(ctx context.Context, cfg conn.Config, specPath, outDir string) (string, error) {
	s, err := LoadSpec(specPath)
	if err != nil {
		return "", err
	}
	name := s.Name
	if s.Mode == "opbench" {
		name = strings.TrimSuffix(filepath.Base(specPath), filepath.Ext(specPath))
	}
	dir := filepath.Join(outDir, time.Now().Format("20060102-150405")+"_"+name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if b, err := os.ReadFile(specPath); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "spec.json"), b, 0o644)
	}
	if s.Mode == "opbench" {
		return dir, opbench.Run(ctx, cfg, specPath, filepath.Join(dir, "opbench.csv"))
	}
	if b, err := json.MarshalIndent(struct {
		Plan any `json:"plan"`
		Exec any `json:"exec"`
	}{settings.Plan(), settings.Exec(exec.ModelBulk)}, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "options.json"), b, 0o644)
	}
	r := &specRun{s: s, cfg: cfg, dir: dir, runID: time.Now().Format("2006-01-02T15:04:05")}
	var closers []func()
	defer func() {
		for _, c := range closers {
			c()
		}
	}()
	open := func(name string, header []string) (*csv.Writer, error) {
		w, c, err := openCSV(filepath.Join(dir, name), header)
		if err == nil {
			closers = append(closers, c)
		}
		return w, err
	}
	if r.runs, err = open("runs.csv", runHeader); err != nil {
		return dir, err
	}
	if r.migs, err = open("migrations.csv", migrationHeader); err != nil {
		return dir, err
	}
	if s.Mode == "profile" || s.StoreProfile {
		if r.spans, err = open("spans.csv", spanHeader); err != nil {
			return dir, err
		}
	}
	plans, err := os.Create(filepath.Join(dir, "plans.txt"))
	if err != nil {
		return dir, err
	}
	defer plans.Close()
	r.plans = plans

	if s.Consolidate {
		if _, err := migrate.Consolidate(cfg, s.Exclude, false); err != nil {
			return dir, fmt.Errorf("consolidate: %w", err)
		}
	}
	if r.base, err = catalog.LoadRecord(cfg.MappingPath()); err != nil {
		return dir, err
	}
	reg, err := conn.NewRegistry(ctx, cfg)
	if err != nil {
		return dir, err
	}
	defer reg.Close(ctx)
	r.sess = dbs.NewSession(reg)
	for _, w := range s.Workloads {
		if err := r.workload(ctx, w); err != nil {
			return dir, err
		}
	}
	return dir, nil
}

// specRun is the state of one RunSpec.
type specRun struct {
	s                *Spec
	cfg              conn.Config
	dir, runID       string
	runs, migs       *csv.Writer
	spans            *csv.Writer
	plans            *os.File
	base             *catalog.MappingDictionary
	sess             *dbs.Session
	oracle           map[string]string // query/case -> canonical Neo4j rows
	wname, qv, label string            // the condition being run (for rows and spans)
	params           map[string]string
}

func (r *specRun) workload(ctx context.Context, w string) error {
	s := r.s
	def := workload.Def(w)
	r.wname = w
	cypherOf := func(q, current string) string {
		if q == queryOfficial {
			c, _ := workload.OfficialQuery("ldbc_snb", w)
			return c
		}
		return current
	}
	// The baseline runs before any copy, on the data as it is.
	r.oracle = map[string]string{}
	if s.Verify && r.cfg.Neo4j != nil {
		current, params0, _ := def(migrate.ModeGraphToRdb, false)
		for _, q := range s.Queries {
			for _, c := range s.Cases {
				r.qv, r.label, r.params = q, c.Label, mergeParams(params0, c.Params)
				cy := cypherOf(q, current)
				if s.Neo4jCypher != "" {
					cy = s.Neo4jCypher
				}
				if err := r.baseline(ctx, cy); err != nil {
					return err
				}
			}
		}
	}
	for _, place := range s.Placements {
		target, _ := placementStore(place)
		mode := migrate.ModeOf(kind.Neo4j, target)
		current, params0, wmigs := def(mode, target != kind.Neo4j)
		if len(s.Migrate) > 0 {
			wmigs = s.migrations()
		}
		md := r.base
		if target != kind.Neo4j {
			if len(wmigs) == 0 {
				fmt.Printf("[%s] no properties to place on %s, skipped\n", w, target)
				continue
			}
			for i := range wmigs {
				wmigs[i].Mode, wmigs[i].DeleteSource = mode, false
			}
			res, err := migrate.Run(ctx, r.cfg, wmigs)
			if err != nil {
				return fmt.Errorf("%s: copy to %s: %w", w, target, err)
			}
			for _, m := range res {
				_ = r.migs.Write([]string{r.runID, s.Name, w, target.String(), m.Entity, fmt.Sprint(wmigs[0].Properties),
					strconv.Itoa(m.Count), ms(m.Duration)})
			}
			r.migs.Flush()
			if md, err = placedFor(r.base, wmigs); err != nil {
				return err
			}
		}
		for _, q := range s.Queries {
			for _, c := range s.Cases {
				r.qv, r.label, r.params = q, c.Label, mergeParams(params0, c.Params)
				cypher := cypherOf(q, current)
				popt, _ := s.Tunings[0].options(w, place, s.Models[0])
				writePlan(r.plans, md, popt, r.runID, w+"/"+q+"/"+c.Label, target.String(), cypher, r.params)
				for _, model := range s.Models {
					for _, t := range s.Tunings {
						if err := r.condition(ctx, md, place, target, model, t, cypher); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

// baseline runs cypher on Neo4j as is (warmup + trials) and keeps the canonical rows for matching.
func (r *specRun) baseline(ctx context.Context, cypher string) error {
	s := r.s
	key := r.qv + "/" + r.label
	for trial := -*s.Warmup + 1; trial <= *s.Trials; trial++ {
		tctx, cancel := context.WithTimeout(ctx, s.timeout)
		res, err := RunNeo4j(tctx, *r.cfg.Neo4j, cypher, toValuedParams(r.params))
		cancel()
		st := runStatus(err)
		if trial > 0 || err != nil {
			r.row("neo4j", "cypher", "-", 0, 0, trial, res, "", st)
		}
		if err != nil {
			fmt.Printf("[%s/%s] neo4j: %v\n", r.wname, key, err)
			if !r.neo4jAlive(ctx) {
				return fmt.Errorf("neo4j does not answer: %w", err)
			}
			return nil
		}
		r.oracle[key] = canonRows(res.Rows)
	}
	r.runs.Flush()
	return nil
}

// condition runs one model x tuning on a placement: warmup + trials, then the store-profile trial.
// A failed run is recorded with its status and ends the condition; a Neo4j that stopped answering ends
// the experiment.
func (r *specRun) condition(ctx context.Context, md *catalog.MappingDictionary, place string, target kind.Kind, model string, t Tuning, cypher string) error {
	s := r.s
	popt, eopt := t.options(r.wname, place, model)
	eopt.Breakdown = s.Mode == "profile"
	var prof *profileSession
	for trial := -*s.Warmup + 1; trial <= *s.Trials; trial++ {
		if trial == 1 && s.Pprof {
			prof = startProfile(filepath.Join(r.dir, "pprof"),
				strings.Join([]string{r.wname, r.qv, r.label, target.String(), model, t.Label}, "_"))
		}
		tctx, cancel := context.WithTimeout(ctx, s.timeout)
		res, err := query.Run(tctx, r.sess, md, cypher, r.params, popt, eopt)
		cancel()
		st := runStatus(err)
		match := ""
		if want, ok := r.oracle[r.qv+"/"+r.label]; ok && err == nil {
			match = strconv.FormatBool(canonRows(res.Rows) == want)
		}
		if trial > 0 || err != nil {
			r.row(target.String(), model, t.Label, eopt.VectorSize, eopt.MaterializeChunkSize, trial, res, match, st)
			if r.spans != nil && err == nil {
				r.writeSpans(target.String(), model, t.Label, strconv.Itoa(trial), res)
			}
		}
		if err != nil {
			prof.Stop()
			r.runs.Flush()
			fmt.Printf("[%s/%s/%s @%s %s %s] %s: %v\n", r.wname, r.qv, r.label, target, model, t.Label, st, err)
			if !r.neo4jAlive(ctx) {
				return fmt.Errorf("neo4j does not answer: %w", err)
			}
			return nil
		}
	}
	prof.Stop()
	r.runs.Flush()
	if s.StoreProfile {
		popt2 := eopt
		popt2.Breakdown, popt2.StoreProfile = false, true
		tctx, cancel := context.WithTimeout(ctx, s.timeout)
		res, err := query.Run(tctx, r.sess, md, cypher, r.params, popt, popt2)
		cancel()
		if err != nil {
			fmt.Printf("[%s @%s %s] store profile: %v\n", r.wname, target, model, err)
			return nil
		}
		return r.writeStoreProfiles(target.String(), model, t.Label, res)
	}
	return nil
}

// neo4jAlive reports whether Neo4j still answers.
func (r *specRun) neo4jAlive(ctx context.Context) bool {
	if r.cfg.Neo4j == nil {
		return true
	}
	tctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	_, err := RunNeo4j(tctx, *r.cfg.Neo4j, "RETURN 1", nil)
	return err == nil
}

// writePlan writes the physical plan and native queries of a condition.
func writePlan(w *os.File, md *catalog.MappingDictionary, popt query.PlanOptions, runID, name, place, cypher string, params map[string]string) {
	fmt.Fprintf(w, "### %s %s @%s\n", runID, name, place)
	p, cat, err := query.Plan(cypher, params, md, popt)
	if err != nil {
		fmt.Fprintf(w, "(plan: %v)\n\n", err)
		return
	}
	fmt.Fprintf(w, "%s\n", query.RenderPlan(p, cat))
}

// runStatus is a run's status column.
func runStatus(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(err.Error(), "context deadline exceeded"):
		return "timeout"
	}
	return "error"
}

func mergeParams(base, over map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(over))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		out[k] = v
	}
	return out
}

// placedFor returns the runtime placement of a placement run: base with the migrated properties
// read from the migrations' destination only. The record file is not touched.
func placedFor(base *catalog.MappingDictionary, migs []migrate.MigrationConfig) (*catalog.MappingDictionary, error) {
	md := base
	for _, m := range migs {
		_, dest, err := migrate.ModeStores(m.Mode)
		if err != nil {
			return nil, err
		}
		if md, err = md.Placed(m.ObjType, m.Entity, m.Properties, dest.String()); err != nil {
			return nil, err
		}
	}
	return md, nil
}
