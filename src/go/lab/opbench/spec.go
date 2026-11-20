// Package opbench measures single logical operators under swept conditions.
//
// Experiments are JSON specs (mode "opbench") run with `go run ./cmd/lab run -spec <file>`. Each builds a logical plan of one
// target operator and the minimal earlier part that makes its input (`Argument` = N uuids sampled from the data ->
// earlier part -> target) and sweeps placement, input size, iteration model, chunk size and so on. The root
// physical step is the target: its time, the total time, round trips and rows go to a CSV, and the result is
// checked against the Neo4j placement. Data is only read (non-Neo4j placements use migrated columns).
//
// Specs: lab/opbench/specs/README.md, examples in lab/opbench/specs/*.json.
package opbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Spec is an experiment definition file.
type Spec struct {
	Name        string       `json:"name"`
	Mode        string       `json:"mode"`      // "opbench" (lab run dispatches on it)
	Warmup      int          `json:"warmup"`    // unmeasured runs before each condition
	Repeat      int          `json:"repeat"`    // measured runs per condition
	Timeout     string       `json:"timeout"`   // limit of one run (e.g. "60s"; default 60s)
	ListsFile   string       `json:"listsFile"` // input uuid lists ([]ListSpec as JSON, relative to the spec file)
	Lists       []ListSpec   `json:"-"`
	Experiments []Experiment `json:"experiments"`
}

// Experiment measures one target operator; its fields depend on Op (table in specs/README.md).
type Experiment struct {
	Name string `json:"name"`
	Op   string `json:"op"`   // EntityScan / GetProperty / Projection / Expand / VarLengthExpand / Filter / Aggregation / Sort / Limit / Distinct
	Role string `json:"role"` // the input's role (start / mid), copied to the CSV only

	Label   string   `json:"label"`  // label of the input vertices (sampled or scanned)
	Prop    string   `json:"prop"`   // the property (GetProperty / Filter / Sort / Distinct / Aggregation's group key)
	Columns []string `json:"props"`  // Projection: the projected columns (filtered first when Prop and Cmp are set)
	Sample  string   `json:"sample"` // condition selecting the input (a Cypher WHERE on n); empty: all

	Rel string `json:"rel"` // Expand / VarLengthExpand: relationship type
	Dir string `json:"dir"` // out / in / both (default out)
	To  string `json:"to"`  // destination label (empty: any)
	Min int    `json:"min"` // VarLengthExpand hops
	Max int    `json:"max"`

	Cmp   string `json:"cmp"`   // Filter / EntityScan: comparison (= <> < > <= >= IN)
	Value any    `json:"value"` // the compared value (an array for IN)

	Agg     string `json:"agg"`     // Aggregation: count / sum / avg / min / max / collect
	AggProp string `json:"aggProp"` // the aggregated property (empty: count(*))
	Desc    bool   `json:"desc"`    // Sort
	Count   int    `json:"count"`   // Limit

	Sweep Sweep `json:"sweep"`
}

// Sweep holds the swept axes; an empty axis is the single default point (in parentheses).
type Sweep struct {
	Placement   []string `json:"placement"`   // store holding the property (graph)
	Rows        []int    `json:"rows"`        // input rows (1000); ignored with input=scan
	Model       []string `json:"model"`       // iteration model (bulk)
	Chunk       []int    `json:"chunk"`       // ids per store query (current settings; 0 is the store's default)
	Parallelism []int    `json:"parallelism"` // concurrent chunks per store query (current settings)
	Vector      []int    `json:"vector"`      // batch width (current settings)
	At          []string `json:"at"`          // store / coordinator: run the target in the store or on the coordinator (store)
	Input       []string `json:"input"`       // ids (sampled uuids) / scan (all vertices of the label, to compare absorbable operators) (ids)
	List        []string `json:"list"`        // name of the uuid list for input=ids (empty: sample)
	Order       []string `json:"order"`       // raw (list order) / sorted (by id, duplicates kept) (raw)
}

// Condition is one point of the sweep.
type Condition struct {
	Placement   string
	Rows        int
	Model       string
	Chunk       int // -1 keeps the current setting
	Parallelism int // 0 keeps the current setting
	Vector      int // 0 keeps the current setting
	At          string
	Input       string
	List        string // input uuid list (sample if empty)
	Order       string // raw / sorted (empty for scans)
}

// Load reads and checks a spec file.
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Spec
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // do not silently ignore misspelled fields
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("opbench: %s: %w", path, err)
	}
	if s.ListsFile != "" {
		lists, err := os.ReadFile(filepath.Join(filepath.Dir(path), s.ListsFile))
		if err != nil {
			return nil, fmt.Errorf("opbench: %s: %w", path, err)
		}
		if err := json.Unmarshal(lists, &s.Lists); err != nil {
			return nil, fmt.Errorf("opbench: %s: %w", s.ListsFile, err)
		}
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("opbench: %s: %w", path, err)
	}
	return &s, nil
}

// Validate checks the spec for contradictions and fills defaults.
func (s *Spec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is empty")
	}
	if s.Repeat <= 0 {
		s.Repeat = 3
	}
	if s.Timeout == "" {
		s.Timeout = "60s"
	}
	if _, err := time.ParseDuration(s.Timeout); err != nil {
		return fmt.Errorf("timeout: %w", err)
	}
	if len(s.Experiments) == 0 {
		return fmt.Errorf("experiments is empty")
	}
	lists := map[string]bool{}
	for i := range s.Lists {
		if err := s.Lists[i].validate(); err != nil {
			return err
		}
		lists[s.Lists[i].Name] = true
	}
	for i := range s.Experiments {
		e := &s.Experiments[i]
		if e.Name == "" {
			e.Name = fmt.Sprintf("%s#%d", e.Op, i+1)
		}
		if err := e.validate(); err != nil {
			return fmt.Errorf("experiment %q: %w", e.Name, err)
		}
		for _, l := range e.Sweep.List {
			if !lists[l] {
				return fmt.Errorf("experiment %q: list %q is not defined", e.Name, l)
			}
		}
	}
	return nil
}

func (e *Experiment) validate() error {
	need := func(field, v string) error {
		if v == "" {
			return fmt.Errorf("%s needs %s", e.Op, field)
		}
		return nil
	}
	if err := need("label", e.Label); err != nil {
		return err
	}
	switch e.Op {
	case "EntityScan":
		if e.Cmp != "" {
			if err := need("prop", e.Prop); err != nil {
				return err
			}
		}
	case "GetProperty", "Sort", "Distinct":
		if err := need("prop", e.Prop); err != nil {
			return err
		}
	case "Projection":
		if len(e.Columns) == 0 {
			return fmt.Errorf("projection needs props")
		}
		if e.Cmp != "" {
			if err := need("prop", e.Prop); err != nil {
				return err
			}
		}
	case "Filter":
		if err := need("prop", e.Prop); err != nil {
			return err
		}
		if err := need("cmp", e.Cmp); err != nil {
			return err
		}
	case "Expand", "VarLengthExpand":
		if err := need("rel", e.Rel); err != nil {
			return err
		}
		switch e.Dir {
		case "", "out", "in", "both":
		default:
			return fmt.Errorf("dir is out / in / both: %q", e.Dir)
		}
		if e.Op == "VarLengthExpand" && (e.Min < 0 || (e.Max >= 0 && e.Max < e.Min)) {
			return fmt.Errorf("bad hops min=%d max=%d", e.Min, e.Max)
		}
	case "Aggregation":
		if err := need("agg", e.Agg); err != nil {
			return err
		}
	case "Limit":
		if e.Count <= 0 {
			return fmt.Errorf("limit needs count (positive)")
		}
	default:
		return fmt.Errorf("unsupported op %q", e.Op)
	}
	for _, at := range e.Sweep.At {
		if at != "store" && at != "coordinator" {
			return fmt.Errorf("at is store / coordinator: %q", at)
		}
		if at == "coordinator" && !barrierable(e.Op) {
			return fmt.Errorf("%s cannot take at=coordinator (it always runs in the store)", e.Op)
		}
	}
	for _, in := range e.Sweep.Input {
		if in != "ids" && in != "scan" {
			return fmt.Errorf("input is ids / scan: %q", in)
		}
	}
	for _, o := range e.Sweep.Order {
		if o != "raw" && o != "sorted" {
			return fmt.Errorf("order is raw / sorted: %q", o)
		}
	}
	if e.Op == "EntityScan" && len(e.Sweep.Input) > 0 {
		return fmt.Errorf("entityScan always scans (do not set input)")
	}
	for _, r := range e.Sweep.Rows {
		if r <= 0 {
			return fmt.Errorf("rows must be positive: %d", r)
		}
	}
	return nil
}

// barrierable reports whether the operator can be forced to the coordinator (at=coordinator).
func barrierable(op string) bool {
	switch op {
	case "Filter", "Aggregation", "Sort", "Limit", "Distinct":
		return true
	}
	return false
}

// Conditions returns every point of the sweep (the product of the axes) in a fixed order.
func (e *Experiment) Conditions() []Condition {
	strs := func(xs []string, def string) []string {
		if len(xs) == 0 {
			return []string{def}
		}
		return xs
	}
	ints := func(xs []int, def int) []int {
		if len(xs) == 0 {
			return []int{def}
		}
		return xs
	}
	inputs := strs(e.Sweep.Input, "ids")
	if e.Op == "EntityScan" {
		inputs = []string{"scan"}
	}
	var out []Condition
	for _, in := range inputs {
		rows := ints(e.Sweep.Rows, 1000)
		lists, orders := strs(e.Sweep.List, ""), strs(e.Sweep.Order, "raw")
		if in == "scan" {
			rows, lists, orders = []int{0}, []string{""}, []string{""} // scans sweep neither counts nor input lists
		}
		for _, r := range rows {
			for _, l := range lists {
				for _, o := range orders {
					for _, pl := range strs(e.Sweep.Placement, "graph") {
						for _, at := range strs(e.Sweep.At, "store") {
							for _, m := range strs(e.Sweep.Model, "bulk") {
								for _, c := range ints(e.Sweep.Chunk, -1) {
									for _, p := range ints(e.Sweep.Parallelism, 0) {
										for _, v := range ints(e.Sweep.Vector, 0) {
											out = append(out, Condition{Placement: pl, Rows: r, Model: m, Chunk: c,
												Parallelism: p, Vector: v, At: at, Input: in, List: l, Order: o})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return out
}

// MaxRows is the number of inputs to sample (the sweep's maximum; smaller counts use a prefix, so inputs nest).
func (e *Experiment) MaxRows() int {
	if len(e.Sweep.List) > 0 {
		return 0 // inputs come from the defined uuid list
	}
	max := 0
	for _, c := range e.Conditions() {
		if c.Rows > max {
			max = c.Rows
		}
	}
	return max
}
