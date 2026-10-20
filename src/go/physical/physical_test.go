package physical_test

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/workload"
)

var update = flag.Bool("update", false, "golden ファイルを再生成する")

const mappingPath = "../../../catalog/mapping.json"

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s が無い（-update で生成）: %v", path, err)
	}
	if string(want) != got {
		t.Errorf("golden %s と不一致\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func baseMapping(t *testing.T) *schema.MappingDictionary {
	t.Helper()
	md, err := schema.LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func clone(t *testing.T, md *schema.MappingDictionary) *schema.MappingDictionary {
	t.Helper()
	b, err := json.Marshal(md)
	if err != nil {
		t.Fatal(err)
	}
	var out schema.MappingDictionary
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return &out
}

// place は列をストアへ移した配置を作る（migs の内容を mapping に反映するのと同じ）。
func place(t *testing.T, base *schema.MappingDictionary, target string, migs []migrator.MigrationConfig) *schema.MappingDictionary {
	md := clone(t, base)
	for _, m := range migs {
		for _, p := range m.Properties {
			if m.ObjType == schema.ObjRelationship {
				if r, ok := md.Relationships[m.Entity]; ok {
					if pm, ok := r.Properties[p]; ok {
						pm.Databases = []string{target}
					}
				}
				continue
			}
			if e, ok := md.Entities[m.Entity]; ok {
				if pm, ok := e.Properties[p]; ok {
					pm.Databases = []string{target}
				}
			}
		}
	}
	return md
}

func plan(t *testing.T, q string, params map[string]string, md *schema.MappingDictionary) (string, error) {
	t.Helper()
	a, err := ast.Parse(q, params)
	if err != nil {
		return "", err
	}
	l, err := logical.Translate(a)
	if err != nil {
		return "", err
	}
	p, err := physical.Build(logical.Rewrite(l), physical.NewCatalog(md))
	if err != nil {
		return "", err
	}
	return physical.Dump(p), nil
}

var placements = []struct {
	name   string
	target string
	mode   migrator.MigrationMode
}{
	{"graph", "", ""},
	{"relational", "relational", migrator.ModeGraphToRdb},
	{"document", "document", migrator.ModeGraphToDoc},
	{"columnar", "columnar", migrator.ModeGraphToCol},
	{"kvs", "kvs", migrator.ModeGraphToKvs},
}

// 全ワークロード × 5 配置の物理計画を golden で固定する。配置は各ワークロードの移行設定に従う。
func TestPhysicalWorkloads(t *testing.T) {
	base := baseMapping(t)
	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		for _, pc := range placements {
			mode := pc.mode
			if mode == "" {
				mode = migrator.ModeGraphToRdb
			}
			cypher, params, migs := workload.Registry[n](mode, pc.target != "")
			md := base
			if pc.target != "" {
				md = place(t, base, pc.target, migs)
			}
			got, err := plan(t, cypher, params, md)
			if errors.Is(err, ast.ErrUnsupported) {
				if pc.name == "graph" {
					b.WriteString("### " + n + "\nUNSUPPORTED: " + err.Error() + "\n\n")
				}
				break
			}
			if err != nil {
				t.Fatalf("%s/%s: %v", n, pc.name, err)
			}
			b.WriteString("### " + n + " @" + pc.name + "\n" + got + "\n")
		}
	}
	golden(t, "workloads.golden", b.String())
}

// 計画書 §6 の例と、統合・吸収・パス化の代表ケース（説明用の混在配置）。
func TestPhysicalCases(t *testing.T) {
	base := baseMapping(t)
	mixed := clone(t, base)
	person := mixed.Entities["Person"].Properties
	person["firstName"].Databases = []string{"relational"}
	person["lastName"].Databases = []string{"relational"}
	person["gender"].Databases = []string{"document"}
	person["birthday"].Databases = []string{"columnar"}

	cases := []struct {
		name string
		md   *schema.MappingDictionary
		q    string
	}{
		{"all_graph_single_query", base, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"three_stores_chain", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.gender = 'male' RETURN f.firstName, f.id"},
		{"sort_limit_then_fetch", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.firstName ORDER BY f.id ASC LIMIT 10"},
		{"scan_in_document", mixed, "MATCH (p:Person) WHERE p.gender = 'male' MATCH (p)-[:KNOWS]->(f:Person) RETURN f.id"},
		{"aggregate_single_store", mixed, "MATCH (p:Person) WHERE p.gender = 'male' RETURN p.gender, count(*) AS cnt"},
		{"aggregate_after_traversal", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) RETURN f.gender, count(*) AS cnt ORDER BY cnt DESC LIMIT 5"},
		{"columnar_is_null_at_coordinator", mixed, "MATCH (p:Person {id: $x})-[:KNOWS]->(f:Person) WHERE f.birthday IS NULL RETURN f.id"},
		{"cross_store_predicate", mixed, "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person) WHERE a.firstName = b.gender RETURN b.id"},
		{"cycle_all_graph", base, "MATCH (a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS]->(c:Person)-[:KNOWS]->(a) RETURN b.id, c.id"},
		{"cartesian_graph", base, "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.id, t.id"},
		{"cartesian_mixed", mixed, "MATCH (a:Person {id: $x}), (t:Tag {name: 'x'}) RETURN a.firstName, t.id"},
		{"path_preds_graph", base, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_preds_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS]->(b:Person)-[:KNOWS*1..3]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') AND length(p) <= 3 RETURN c.id"},
		{"path_any_document", mixed, "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(c:Person) WHERE any(n IN nodes(p) WHERE n.gender = 'female') RETURN c.id"},
		{"with_stage_mixed", mixed, "MATCH (p:Person)-[:KNOWS]->(f:Person) WITH p, count(f) AS deg WHERE deg > 10 RETURN p.firstName, deg ORDER BY deg DESC LIMIT 3"},
	}
	var b strings.Builder
	for _, c := range cases {
		got, err := plan(t, c.q, map[string]string{"x": "1"}, c.md)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		b.WriteString("### " + c.name + "\n" + c.q + "\n" + got + "\n")
	}
	golden(t, "cases.golden", b.String())
}
