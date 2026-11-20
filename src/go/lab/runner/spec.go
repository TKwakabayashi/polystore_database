// Package runner runs experiments from spec files (lab run) and single queries (cmd/query, lab workflow).
package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/settings"
)

// Spec is one experiment: workloads x query versions x parameter cases x placements x iteration models x
// tunings. Unset fields take defaults. Mode "opbench" hands the file to the operator benchmark (lab/opbench).
type Spec struct {
	Name string `json:"name"`
	// Mode is "run" (times only), "profile" (also spans per executed operator, with the store calls split
	// into first row / receive / decode) or "opbench".
	Mode      string   `json:"mode"`
	Workloads []string `json:"workloads"` // workload names (workload.Lookup), or ["all"] for every ldbc_snb workload
	// Queries are the versions of each workload's Cypher: "current" (the workload's, default) and
	// "official" (the LDBC original).
	Queries []string `json:"queries"`
	// Cases override the workload's parameters (default: one case "default" without overrides).
	Cases      []Case   `json:"cases"`
	Placements []string `json:"placements"` // database names (neo4j, mysql, ...); legacy short names also work
	Models     []string `json:"models"`
	// Vectors is a shorthand for one tuning per vector size (used when Tunings is empty).
	Vectors []int    `json:"vectors"`
	Tunings []Tuning `json:"tunings"`
	// Migrate are the properties copied to each non-Neo4j placement (default: the workload's own list).
	Migrate []Migration `json:"migrate"`
	Warmup  *int        `json:"warmup"`
	Trials  *int        `json:"trials"`
	Timeout string      `json:"timeout"` // per query run
	// Verify runs the Cypher on Neo4j as is (warmup + trials, recorded as placement neo4j, model cypher)
	// and compares every result with it. Neo4jCypher replaces the Cypher sent to Neo4j.
	Verify      bool   `json:"verify"`
	Neo4jCypher string `json:"neo4j_cypher"`
	// Consolidate brings every property back to Neo4j (except Exclude) before the runs.
	Consolidate bool     `json:"consolidate"`
	Exclude     []string `json:"exclude"`
	// StoreProfile adds one trial per condition in which every store reports how it ran each query (L4).
	// Its numbers go to spans.csv (level L4) and the stores' raw output to l4/*.json.
	StoreProfile bool `json:"store_profile"`
	// Pprof writes Go profiles (cpu, heap, allocs, block, mutex) of each condition's trials to pprof/.
	// Profiling slows the run, so do not compare its times with others.
	Pprof bool `json:"pprof"`

	timeout time.Duration
}

// Case is one set of parameter overrides.
type Case struct {
	Label  string            `json:"label"`
	Params map[string]string `json:"params"`
}

// Migration is one set of properties placed off Neo4j. Kind is "node" or "edge".
type Migration struct {
	Kind   string   `json:"kind"`
	Entity string   `json:"entity"`
	Props  []string `json:"props"`
}

// Tuning is one set of internal switches. Zero values keep the defaults from settings.
type Tuning struct {
	Label  string         `json:"label"`
	Vector int            `json:"vector"` // batch width of vectorized / stream / vecstream
	Chunk  int            `json:"chunk"`  // ids per store query (0 = the store's default)
	Chunks map[string]int `json:"chunks"` // Chunk per placement (as written in placements)
	Defer  *bool          `json:"defer"`  // defer non-anchor fetches and filters after the expansions
	Hand   bool           `json:"hand"`   // hand-built plan (handplan.go)
	Early  bool           `json:"early"`  // with Hand: filter before the traversal and pass uuids to it
	// PathPreds is how path element conditions on values outside Neo4j run: post / prefilter / stepwise ("" = settings).
	PathPreds string `json:"path_preds"`
}

const (
	queryCurrent  = "current"
	queryOfficial = "official"
	// defaultWarmup and defaultTrials apply when the spec does not set them.
	defaultWarmup = 3
	defaultTrials = 10
)

var allModels = []string{string(exec.ModelBulk), string(exec.ModelVolcano), string(exec.ModelVectorized),
	string(exec.ModelStream), string(exec.ModelVecStream)}

// LoadSpec reads a spec and fills the defaults. Unknown fields are errors.
func LoadSpec(path string) (*Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var head struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(b, &head); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if head.Mode == "opbench" {
		return &Spec{Mode: head.Mode}, nil
	}
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	var s Spec
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := s.fill(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}

func (s *Spec) fill() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.Mode == "" {
		s.Mode = "run"
	}
	if s.Mode != "run" && s.Mode != "profile" {
		return errors.New("mode must be run, profile or opbench")
	}
	if len(s.Workloads) == 1 && s.Workloads[0] == "all" {
		s.Workloads = workload.Names(workload.LDBC)
	}
	if len(s.Workloads) == 0 {
		return errors.New("workloads is required")
	}
	if len(s.Queries) == 0 {
		s.Queries = []string{queryCurrent}
	}
	for _, w := range s.Workloads {
		if _, _, ok := workload.Lookup(w); !ok {
			return fmt.Errorf("unknown workload %q", w)
		}
		for _, q := range s.Queries {
			switch q {
			case queryCurrent:
			case queryOfficial:
				if _, ok := workload.OfficialQuery("ldbc_snb", w); !ok {
					return fmt.Errorf("%s has no official Cypher", w)
				}
			default:
				return fmt.Errorf("unknown query %q (%s / %s)", q, queryCurrent, queryOfficial)
			}
		}
	}
	if len(s.Cases) == 0 {
		s.Cases = []Case{{Label: "default"}}
	}
	seen := map[string]bool{}
	for _, c := range s.Cases {
		if c.Label == "" || seen[c.Label] {
			return fmt.Errorf("case label %q is empty or repeated", c.Label)
		}
		seen[c.Label] = true
	}
	if len(s.Placements) == 0 {
		s.Placements = []string{kind.Neo4j.String()}
	}
	for _, p := range s.Placements {
		if _, err := placementStore(p); err != nil {
			return err
		}
	}
	if len(s.Models) == 0 {
		s.Models = []string{string(exec.ModelBulk)}
	}
	for _, m := range s.Models {
		if !slices.Contains(allModels, m) {
			return fmt.Errorf("unknown model %q", m)
		}
	}
	if len(s.Tunings) == 0 {
		if len(s.Vectors) == 0 {
			s.Vectors = []int{settings.VectorSize}
		}
		for _, v := range s.Vectors {
			s.Tunings = append(s.Tunings, Tuning{Label: "v" + strconv.Itoa(v), Vector: v})
		}
	}
	for i := range s.Tunings {
		if s.Tunings[i].Label == "" {
			s.Tunings[i].Label = "t" + strconv.Itoa(i+1)
		}
		if p := s.Tunings[i].PathPreds; p != "" {
			if _, err := physical.ParsePathStrategy(p); err != nil {
				return fmt.Errorf("tuning %s: %w", s.Tunings[i].Label, err)
			}
		}
	}
	for _, m := range s.Migrate {
		if (m.Kind != "node" && m.Kind != "edge") || m.Entity == "" || len(m.Props) == 0 {
			return fmt.Errorf("bad migrate entry %+v", m)
		}
	}
	if s.Warmup == nil {
		w := defaultWarmup
		s.Warmup = &w
	}
	if s.Trials == nil {
		t := defaultTrials
		s.Trials = &t
	}
	if *s.Warmup < 0 || *s.Trials < 1 {
		return errors.New("warmup must be >= 0 and trials >= 1")
	}
	if s.Timeout == "" {
		s.Timeout = "90s"
	}
	d, err := time.ParseDuration(s.Timeout)
	if err != nil || d <= 0 {
		return fmt.Errorf("bad timeout %q", s.Timeout)
	}
	s.timeout = d
	return nil
}

func placementStore(p string) (kind.Kind, error) {
	if p == "graph" {
		return kind.Neo4j, nil
	}
	_, k, err := migrate.ModeStores(migrate.MigrationMode("graph_to_" + p))
	if err != nil {
		if k, perr := kind.ParseKind(p); perr == nil && k == kind.Neo4j {
			return k, nil
		}
		return 0, fmt.Errorf("unknown placement %q", p)
	}
	return k, nil
}

// options returns the planning and execution options of a tuning on a placement.
func (t Tuning) options(workloadName, place string, model string) (query.PlanOptions, exec.ExecOptions) {
	p := settings.Plan()
	e := settings.Exec(exec.Model(model))
	if t.Vector > 0 {
		e.VectorSize = t.Vector
	}
	if t.Chunk > 0 {
		e.MaterializeChunkSize = t.Chunk
	}
	if c, ok := t.Chunks[place]; ok {
		e.MaterializeChunkSize = c
	}
	if t.Defer != nil {
		p.Logical.DeferPredicates, p.Physical.DeferPredicates = *t.Defer, *t.Defer
	}
	if t.Hand {
		p.Rewrite = handPlan(workloadName, t.Early)
	}
	if t.PathPreds != "" {
		p.Physical.PathPredicates, _ = physical.ParsePathStrategy(t.PathPreds) // checked by LoadSpec
	}
	return p, e
}

// migrations are the spec's Migrate as migration configs (Mode is filled per placement).
func (s *Spec) migrations() []migrate.MigrationConfig {
	out := make([]migrate.MigrationConfig, len(s.Migrate))
	for i, m := range s.Migrate {
		obj := catalog.ObjEntity
		if m.Kind == "edge" {
			obj = catalog.ObjRelationship
		}
		out[i] = migrate.MigrationConfig{ObjType: obj, Entity: m.Entity, Properties: m.Props}
	}
	return out
}
