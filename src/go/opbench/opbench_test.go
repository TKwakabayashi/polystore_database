package opbench

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"polystore_database/src/go/executor"
	"polystore_database/src/go/id"
	"polystore_database/src/go/native"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/schema"
)

func baseMapping(t *testing.T) *schema.MappingDictionary {
	t.Helper()
	md, err := schema.LoadMappingDictionary("../../../catalog/mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func TestValidate(t *testing.T) {
	bad := []Experiment{
		{Op: "Nope", Label: "Person"},
		{Op: "GetProperty", Label: "Person"},                                                        // prop が無い
		{Op: "Filter", Label: "Person", Prop: "gender"},                                             // cmp が無い
		{Op: "Expand", Label: "Person", Rel: "KNOWS", Dir: "up"},                                    // dir
		{Op: "GetProperty", Label: "Person", Prop: "id", Sweep: Sweep{At: []string{"coordinator"}}}, // 常にストア
		{Op: "Limit", Label: "Person"},                                                              // count
		{Op: "EntityScan", Label: "Person", Sweep: Sweep{Input: []string{"ids"}}},
	}
	for _, e := range bad {
		s := &Spec{Name: "x", Experiments: []Experiment{e}}
		if err := s.Validate(); err == nil {
			t.Errorf("%+v は不正になるべき", e)
		}
	}
	s := &Spec{Name: "x", Experiments: []Experiment{{Op: "GetProperty", Label: "Person", Prop: "firstName"}}}
	if err := s.Validate(); err != nil || s.Repeat != 3 || s.Timeout != "60s" {
		t.Errorf("既定値: %+v %v", s, err)
	}
}

func TestConditions(t *testing.T) {
	e := Experiment{Op: "Filter", Label: "Person", Prop: "gender", Cmp: "=", Value: "male",
		Sweep: Sweep{Rows: []int{10, 100}, Placement: []string{"graph", "relational"}, At: []string{"store", "coordinator"},
			Input: []string{"ids", "scan"}}}
	cs := e.Conditions()
	// ids: 2 件数 × 2 配置 × 2 at = 8、scan: 件数を振らないので 2 × 2 = 4
	if len(cs) != 12 {
		t.Fatalf("%d 条件", len(cs))
	}
	if e.MaxRows() != 100 {
		t.Errorf("MaxRows %d", e.MaxRows())
	}
	scan := Experiment{Op: "EntityScan", Label: "Person"}
	if cs := scan.Conditions(); len(cs) != 1 || cs[0].Input != "scan" {
		t.Errorf("EntityScan: %+v", cs)
	}
}

// 根のステップが対象の演算子を含む形になり、at / input / placement が物理計画に効くこと。
func TestPlans(t *testing.T) {
	base := baseMapping(t)
	cases := []struct {
		name string
		e    Experiment
		c    Condition
		want []string // 物理計画のダンプに含まれるべき行（先頭の空白は無視）
		root string   // 根の物理ノード
	}{
		{"getprop_graph", Experiment{Op: "GetProperty", Label: "Person", Prop: "firstName"},
			Condition{Placement: "graph", Input: "ids", At: "store"}, []string{"Fetch[graph]", "ArgRows(a)"}, "*physical.StoreQuery"},
		{"getprop_relational", Experiment{Op: "GetProperty", Label: "Person", Prop: "firstName"},
			Condition{Placement: "relational", Input: "ids", At: "store"}, []string{"Fetch[relational]"}, "*physical.StoreQuery"},
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
			Condition{Placement: "kvs", Input: "scan", At: "store"}, []string{"StoreQuery[kvs]"}, "*physical.StoreQuery"},
		{"sort_coordinator", Experiment{Op: "Sort", Label: "Person", Prop: "birthday", Desc: true},
			Condition{Placement: "columnar", Input: "ids", At: "coordinator"}, []string{"Coord Sort[a.birthday DESC]"}, "*physical.Coord"},
		{"distinct_scan_document", Experiment{Op: "Distinct", Label: "Person", Prop: "lastName"},
			Condition{Placement: "document", Input: "scan", At: "store"}, []string{"Distinct"}, "*physical.StoreQuery"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			md, err := PlaceMapping(base, &c.e, c.c.Placement)
			if err != nil {
				t.Fatal(err)
			}
			r := &Runner{}
			pl, err := r.plan(&c.e, c.c, []id.UUID{id.New(id.KindNode)}, md)
			if err != nil {
				t.Fatal(err)
			}
			dump := physical.Dump(pl.p)
			for _, w := range c.want {
				if !strings.Contains(dump, w) {
					t.Errorf("計画に %q が無い:\n%s", w, dump)
				}
			}
			if got := fmt.Sprintf("%T", pl.p); got != c.root {
				t.Errorf("根 %s, want %s\n%s", got, c.root, dump)
			}
			if c.c.Input == "ids" && len(pl.args) != 1 {
				t.Errorf("入力の葉が %d 個", len(pl.args))
			}
			var lowerErr error
			var walk func(physical.Node)
			walk = func(n physical.Node) {
				if q, ok := n.(*physical.StoreQuery); ok && lowerErr == nil {
					_, lowerErr = native.Lower(q, pl.cat.TypeOf)
				}
				for _, ch := range n.Children() {
					walk(ch)
				}
			}
			walk(pl.p)
			if lowerErr != nil {
				t.Errorf("ネイティブクエリが作れない: %v", lowerErr)
			}
		})
	}
	// 配置は base を書き換えない
	if base.Entities["Person"].Properties["firstName"].Databases[0] != "graph" {
		t.Errorf("PlaceMapping が base を書き換えた")
	}
}

// fakeBackend は graph の GetProperty だけを答える（入力の uuid ごとに 1 行）。
type fakeBackend struct{ ids []id.UUID }

func (f *fakeBackend) Run(_ context.Context, q *native.Query, in []executor.Key) ([][]any, error) {
	var out [][]any
	for _, k := range in {
		row := make([]any, len(q.Columns))
		for i, c := range q.Columns {
			if c == inAlias {
				row[i] = k[0].String()
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

func TestRunnerWritesCSV(t *testing.T) {
	var ids []id.UUID
	for i := 0; i < 50; i++ {
		ids = append(ids, id.New(id.KindNode))
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
	// ヘッダ + 2 件数 × 2 モデル × 2 回
	if len(recs) != 1+8 {
		t.Fatalf("%d 行\n%s", len(recs), logs.String())
	}
	col := func(name string) int {
		for i, h := range Header {
			if h == name {
				return i
			}
		}
		t.Fatalf("列 %s が無い", name)
		return -1
	}
	for _, rec := range recs[1:] {
		rows := rec[col("rows")]
		want := rows
		if rows == "100" {
			want = "50" // 抜き出せたのは 50 件
		}
		if rec[col("out_rows")] != want || rec[col("match_graph")] != "true" || rec[col("error")] != "" ||
			!strings.HasPrefix(rec[col("target_step")], "Fetch[graph]") {
			t.Errorf("行 %v", rec)
		}
	}
	if !strings.Contains(plans.String(), "MATCH (a:Entity) WHERE a.uuid IN $in") {
		t.Errorf("計画ファイルにネイティブクエリが無い:\n%s", plans.String())
	}
	if !strings.Contains(logs.String(), "入力が 50 件しかない") {
		t.Errorf("件数不足の知らせが無い:\n%s", logs.String())
	}
}

// 例の定義ファイルはどれも読めて、組める計画になる。
func TestExampleSpecs(t *testing.T) {
	files, err := filepath.Glob("specs/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("specs/*.json が無い: %v", err)
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
				if _, err := (&Runner{}).plan(e, c, []id.UUID{id.New(id.KindNode)}, md); err != nil {
					t.Errorf("%s / %s %+v: %v", f, e.Name, c, err)
				}
			}
		}
		t.Logf("%s: %d 実験・%d 条件", f, len(s.Experiments), n)
	}
}
