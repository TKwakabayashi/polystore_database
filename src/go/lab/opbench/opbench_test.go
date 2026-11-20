package opbench

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/dbs"
	"strings"
	"testing"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/physical"
)

func baseMapping(t *testing.T) *catalog.MappingDictionary {
	t.Helper()
	md, err := catalog.LoadMappingDictionary("../../../../datastore/datasets/ldbc_snb_sf1/mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func TestValidate(t *testing.T) {
	bad := []Experiment{
		{Op: "Nope", Label: "Person"},
		{Op: "GetProperty", Label: "Person"},                                                        // no prop
		{Op: "Filter", Label: "Person", Prop: "gender"},                                             // no cmp
		{Op: "Expand", Label: "Person", Rel: "KNOWS", Dir: "up"},                                    // dir
		{Op: "GetProperty", Label: "Person", Prop: "id", Sweep: Sweep{At: []string{"coordinator"}}}, // always the store
		{Op: "Limit", Label: "Person"},                                                              // count
		{Op: "EntityScan", Label: "Person", Sweep: Sweep{Input: []string{"ids"}}},
	}
	for _, e := range bad {
		s := &Spec{Name: "x", Experiments: []Experiment{e}}
		if err := s.Validate(); err == nil {
			t.Errorf("%+v must be invalid", e)
		}
	}
	s := &Spec{Name: "x", Experiments: []Experiment{{Op: "GetProperty", Label: "Person", Prop: "firstName"}}}
	if err := s.Validate(); err != nil || s.Repeat != 3 || s.Timeout != "60s" {
		t.Errorf("defaults: %+v %v", s, err)
	}
}

func TestConditions(t *testing.T) {
	e := Experiment{Op: "Filter", Label: "Person", Prop: "gender", Cmp: "=", Value: "male",
		Sweep: Sweep{Rows: []int{10, 100}, Placement: []string{"graph", "relational"}, At: []string{"store", "coordinator"},
			Input: []string{"ids", "scan"}}}
	cs := e.Conditions()
	// ids: 2 counts x 2 placements x 2 at = 8; scan: counts are not swept, 2 x 2 = 4
	if len(cs) != 12 {
		t.Fatalf("%d conditions", len(cs))
	}
	if e.MaxRows() != 100 {
		t.Errorf("MaxRows %d", e.MaxRows())
	}
	scan := Experiment{Op: "EntityScan", Label: "Person"}
	if cs := scan.Conditions(); len(cs) != 1 || cs[0].Input != "scan" {
		t.Errorf("EntityScan: %+v", cs)
	}
}

// The root step contains the target operator, and at / input / placement shape the physical plan.
func TestPlans(t *testing.T) {
	base := baseMapping(t)
	cases := []struct {
		name string
		e    Experiment
		c    Condition
		want []string // lines expected in the physical plan dump (leading spaces ignored)
		root string   // the root physical node
	}{
		{"getprop_graph", Experiment{Op: "GetProperty", Label: "Person", Prop: "firstName"},
			Condition{Placement: "graph", Input: "ids", At: "store"}, []string{"Fetch[neo4j]", "ArgRows(a)"}, "*physical.StoreQuery"},
		{"getprop_relational", Experiment{Op: "GetProperty", Label: "Person", Prop: "firstName"},
			Condition{Placement: "relational", Input: "ids", At: "store"}, []string{"Fetch[mysql]"}, "*physical.StoreQuery"},
		{"filter_in_store", Experiment{Op: "Filter", Label: "Person", Prop: "gender", Cmp: "=", Value: "male"},
			Condition{Placement: "document", Input: "ids", At: "store"}, []string{"Filter((a.gender = \"male\"))"}, "*physical.StoreQuery"},
		{"filter_at_coordinator", Experiment{Op: "Filter", Label: "Person", Prop: "gender", Cmp: "=", Value: "male"},
			Condition{Placement: "document", Input: "ids", At: "coordinator"}, []string{"Coord Filter", "Coord Unwind"}, "*physical.Coord"},
		{"aggregation_absorbed_scan", Experiment{Op: "Aggregation", Label: "Person", Prop: "gender", Agg: "count"},
			Condition{Placement: "relational", Input: "scan", At: "store"}, []string{"Aggregation[keys: a.gender AS key | count(*) AS value]"}, "*physical.StoreQuery"},
		{"aggregation_ids_coordinator", Experiment{Op: "Aggregation", Label: "Person", Prop: "gender", Agg: "count"},
			Condition{Placement: "relational", Input: "ids", At: "store"}, []string{"Coord Aggregation"}, "*physical.Coord"},
		{"expand", Experiment{Op: "Expand", Label: "Person", Rel: "KNOWS", Dir: "both", To: "Person"},
			Condition{Placement: "graph", Input: "ids", At: "store"}, []string{"Expand((a)-[r:KNOWS]-(b:Person))"}, "*physical.StoreQuery"},
		{"varlength", Experiment{Op: "VarLengthExpand", Label: "Person", Rel: "KNOWS", Dir: "both", Min: 1, Max: 2},
			Condition{Placement: "graph", Input: "ids", At: "store"}, []string{"VarLengthExpand((a)-[s:KNOWS*1..2]-(b))"}, "*physical.StoreQuery"},
		{"scan_filter_kvs", Experiment{Op: "EntityScan", Label: "Person", Prop: "gender", Cmp: "=", Value: "female"},
			Condition{Placement: "kvs", Input: "scan", At: "store"}, []string{"StoreQuery[leveldb]"}, "*physical.StoreQuery"},
		{"sort_coordinator", Experiment{Op: "Sort", Label: "Person", Prop: "birthday", Desc: true},
			Condition{Placement: "columnar", Input: "ids", At: "coordinator"}, []string{"Coord Sort[a.birthday DESC]"}, "*physical.Coord"},
		{"distinct_scan_document", Experiment{Op: "Distinct", Label: "Person", Prop: "lastName"},
			Condition{Placement: "document", Input: "scan", At: "store"}, []string{"Distinct"}, "*physical.StoreQuery"},
		{"projection_ids", Experiment{Op: "Projection", Label: "Person", Columns: []string{"firstName", "lastName"}},
			Condition{Placement: "relational", Input: "ids", At: "store"}, []string{"Fetch[mysql]"}, ""},
		{"projection_filter_scan", Experiment{Op: "Projection", Label: "Person", Prop: "gender", Cmp: "=", Value: "male",
			Columns: []string{"gender", "lastName"}},
			Condition{Placement: "relational", Input: "scan", At: "store"}, []string{"Filter((a.gender = \"male\"))"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			md, err := PlaceMapping(base, &c.e, c.c.Placement)
			if err != nil {
				t.Fatal(err)
			}
			r := &Runner{}
			pl, err := r.plan(&c.e, c.c, []id.UUID{id.New()}, md)
			if err != nil {
				t.Fatal(err)
			}
			dump := physical.Dump(pl.p)
			for _, w := range c.want {
				if !strings.Contains(dump, w) {
					t.Errorf("%q is not in the plan:\n%s", w, dump)
				}
			}
			if got := fmt.Sprintf("%T", pl.p); c.root != "" && got != c.root {
				t.Errorf("root %s, want %s\n%s", got, c.root, dump)
			}
			if c.c.Input == "ids" && len(pl.args) != 1 {
				t.Errorf("%d input leaves", len(pl.args))
			}
			var lowerErr error
			var walk func(physical.Node)
			walk = func(n physical.Node) {
				if q, ok := n.(*physical.StoreQuery); ok && lowerErr == nil {
					_, lowerErr = dbs.Compile(&q.StoreQuery, pl.cat.TypeOf)
				}
				for _, ch := range n.Children() {
					walk(ch)
				}
			}
			walk(pl.p)
			if lowerErr != nil {
				t.Errorf("cannot make the native query: %v", lowerErr)
			}
		})
	}
	// placement does not modify base
	if base.Entities["Person"].Properties["firstName"].Databases[0] != "neo4j" {
		t.Errorf("PlaceMapping changed base")
	}
}

// fakeBackend answers only Neo4j GetProperty (one row per input uuid).
type fakeBackend struct{ ids []id.UUID }

func (f *fakeBackend) Run(_ context.Context, q *stores.Query, in []exec.Key) ([][]any, error) {
	var out [][]any
	for _, k := range in {
		row := make([]any, len(q.Columns))
		for i, c := range q.Columns {
			if c == inAlias {
				row[i] = k[0]
			} else {
				row[i] = "v-" + k[0].String()[2:6]
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func (f *fakeBackend) Verbatim(_ context.Context, _ string, p map[string]any) ([]map[string]any, error) {
	n := int(p["n"].(int64))
	var out []map[string]any
	for _, u := range f.ids[:min(n, len(f.ids))] {
		out = append(out, map[string]any{"u": u.String()})
	}
	return out, nil
}

func TestListTake(t *testing.T) {
	a, b, c := id.New(), id.New(), id.New()
	l := &inputList{ids: []id.UUID{c, a, a, b}}
	if got := l.take(3, false); got[0] != c || got[1] != a || got[2] != a {
		t.Errorf("first 3: %v", got)
	}
	if got := l.take(3, true); !id.Less(got[0], got[2]) || distinctIDs(got) != 2 {
		t.Errorf("sorting is by id and keeps duplicates: %v", got)
	}
	if len(l.take(9, false)) != 4 {
		t.Errorf("the size is at most the list length")
	}
	r := &inputList{ids: []id.UUID{a, b, c}, rank: []int64{30, 20, 10}}
	if got := r.take(2, false); got[0] != b || got[1] != c {
		t.Errorf("the 2 smallest ranks in list order: %v", got)
	}
}

// listBackend returns each start uuid twice.
type listBackend struct{ fakeBackend }

func (b *listBackend) Verbatim(_ context.Context, q string, p map[string]any) ([]map[string]any, error) {
	var out []map[string]any
	if q == "starts" {
		for _, u := range b.ids {
			out = append(out, map[string]any{"u": u.String()})
		}
		return out, nil
	}
	for _, s := range p["s"].([]any) {
		out = append(out, map[string]any{"u": s}, map[string]any{"u": s})
	}
	return out, nil
}

func TestListFileRoundTrip(t *testing.T) {
	var ids []id.UUID
	for i := 0; i < 10; i++ {
		ids = append(ids, id.New())
	}
	var logs bytes.Buffer
	r := &Runner{Backend: &listBackend{fakeBackend{ids: ids}}, Log: &logs, InputDir: t.TempDir()}
	spec := &ListSpec{Name: "dup", Max: 7, Starts: "starts", Query: "expand", Batch: 2}
	built, err := r.loadList(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if built.len() != 7 || distinctIDs(built.ids) != 4 {
		t.Fatalf("%d rows, %d distinct", built.len(), distinctIDs(built.ids))
	}
	again, err := r.loadList(context.Background(), &ListSpec{Name: "dup", Max: 7, Query: "must not be remade"})
	if err != nil {
		t.Fatal(err)
	}
	for i := range built.ids {
		if built.ids[i] != again.ids[i] {
			t.Fatalf("differs from the saved list: %d", i)
		}
	}
}

func TestRunnerWritesCSV(t *testing.T) {
	var ids []id.UUID
	for i := 0; i < 50; i++ {
		ids = append(ids, id.New())
	}
	spec := &Spec{Name: "smoke", Warmup: 1, Repeat: 2, Experiments: []Experiment{{
		Op: "GetProperty", Label: "Person", Prop: "firstName",
		Sweep: Sweep{Rows: []int{10, 100}, Model: []string{"bulk", "stream"}},
	}}}
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	var buf, plans, logs bytes.Buffer
	r := &Runner{Backend: &fakeBackend{ids: ids}, Base: baseMapping(t), CSV: csv.NewWriter(&buf), Plans: &plans, Log: &logs}
	if err := r.Run(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	recs, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	// header + 2 counts x 2 models x 2 repeats
	if len(recs) != 1+8 {
		t.Fatalf("%d rows\n%s", len(recs), logs.String())
	}
	col := func(name string) int {
		for i, h := range Header {
			if h == name {
				return i
			}
		}
		t.Fatalf("no column %s", name)
		return -1
	}
	for _, rec := range recs[1:] {
		rows := rec[col("rows")]
		want := rows
		if rows == "100" {
			want = "50" // 50 were extracted
		}
		if rec[col("out_rows")] != want || rec[col("match_graph")] != "true" || rec[col("error")] != "" ||
			!strings.HasPrefix(rec[col("target_step")], "Fetch[neo4j]") {
			t.Errorf("row %v", rec)
		}
	}
	if !strings.Contains(plans.String(), "MATCH (a:Person) WHERE a.uuid IN $in") {
		t.Errorf("no native query in the plans file:\n%s", plans.String())
	}
	if !strings.Contains(logs.String(), "only 50 input rows") {
		t.Errorf("no notice about too few rows:\n%s", logs.String())
	}
}

// Every example spec loads and builds a plan.
func TestExampleSpecs(t *testing.T) {
	files, err := filepath.Glob("specs/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no specs/*.json: %v", err)
	}
	base := baseMapping(t)
	for _, f := range files {
		s, err := Load(f)
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		n := 0
		for i := range s.Experiments {
			e := &s.Experiments[i]
			for _, c := range e.Conditions() {
				n++
				md, err := PlaceMapping(base, e, c.Placement)
				if err != nil {
					t.Errorf("%s / %s: %v", f, e.Name, err)
					continue
				}
				if _, err := (&Runner{}).plan(e, c, []id.UUID{id.New()}, md); err != nil {
					t.Errorf("%s / %s %+v: %v", f, e.Name, c, err)
				}
			}
		}
		t.Logf("%s: %d experiments, %d conditions", f, len(s.Experiments), n)
	}
}
