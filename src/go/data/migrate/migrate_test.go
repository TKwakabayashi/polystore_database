package migrate

import (
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"testing"
)

func TestModeStores(t *testing.T) {
	cases := map[MigrationMode][2]kind.Kind{
		ModeGraphToRdb:                   {kind.Neo4j, kind.MySQL},
		ModeWideToCol:                    {kind.Cassandra, kind.ClickHouse},
		"neo4j_to_mongo":                 {kind.Neo4j, kind.Mongo},
		ModeOf(kind.LevelDB, kind.Neo4j): {kind.LevelDB, kind.Neo4j},
	}
	for m, want := range cases {
		src, dst, err := ModeStores(m)
		if err != nil || src != want[0] || dst != want[1] {
			t.Errorf("%s: got %v %v %v", m, src, dst, err)
		}
	}
	for _, bad := range []MigrationMode{"graph", "graph_to_graph", "x_to_rdb"} {
		if _, _, err := ModeStores(bad); err == nil {
			t.Errorf("%s: want an error", bad)
		}
	}
}

func TestPlanConsolidation(t *testing.T) {
	prop := func(dbs ...string) *catalog.PropertyMapping {
		return &catalog.PropertyMapping{Type: "string", Databases: dbs}
	}
	md := &catalog.MappingDictionary{
		Entities: map[string]*catalog.EntityMapping{"Person": {Properties: map[string]*catalog.PropertyMapping{
			"id": prop("neo4j"), "firstName": prop("mysql"), "lastName": prop("mysql"),
			"gender": prop("neo4j", "mongo"), "email": prop("kvs"), "speaks": prop("relational"),
		}}},
		Relationships: map[string]*catalog.RelationshipMapping{"KNOWS": {Properties: map[string]*catalog.PropertyMapping{
			"creationDate": prop("cassandra"),
		}}},
	}
	plan := PlanConsolidation(md, []string{"Person.email"})
	if len(plan.Replicas) != 1 || plan.Replicas[0] != "Person.gender" {
		t.Errorf("replicas = %v", plan.Replicas)
	}
	got := map[string][]string{}
	for _, m := range plan.Moves {
		got[m.Entity+" "+string(m.Mode)] = m.Properties
	}
	want := map[string][]string{
		"KNOWS cassandra_to_neo4j": {"creationDate"},
		"Person mysql_to_neo4j":    {"firstName", "lastName", "speaks"},
	}
	if len(got) != len(want) {
		t.Fatalf("moves = %v", got)
	}
	for k, v := range want {
		if g := got[k]; len(g) != len(v) || g[0] != v[0] {
			t.Errorf("%s: %v, want %v", k, g, v)
		}
	}
}
