//go:build integration

// Package throwaway holds the tests that rewrite or delete data and run only on a throwaway environment
// given by its own config (build tag integration): the uuid storage form and migrations that delete the
// source. Each checks that the config's namespaces are dedicated to it before touching anything.
package throwaway

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
)

// TestMain moves the working directory to src/go once for the whole package (tests do not chdir). go test runs
// in the package directory (tests/db/ or tests/throwaway/), so this matches the relative paths the commands use
// (../../config/...). The working directory is process-wide: changing it per test would break later paths.
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic("chdir to src/go: " + err.Error())
	}
	os.Exit(m.Run())
}

var uuidE2EQueries = []string{
	"MATCH (p:Person {id: 1})-[:KNOWS]-(f:Person) RETURN f.firstName, f.lastName",
	"MATCH (p:Person)-[:IS_LOCATED_IN]->(c:Place) WHERE p.gender = 'female' RETURN c.name, count(*) AS n",
	"MATCH (p:Person {id: 1})-[:KNOWS*1..2]-(f:Person) RETURN f.id, f.gender, count(*) AS paths",
	"MATCH (a:Person {id: 1})-[:KNOWS]-(x:Person)-[:KNOWS]-(b:Person {id: 3}) RETURN x.lastName",
	"MATCH (p:Person) OPTIONAL MATCH (p)-[:IS_LOCATED_IN]->(c:Place) RETURN p.id, c.name",
	"MATCH p=(a:Person {id: 1})-[:KNOWS*1..2]-(c:Person) WHERE all(n IN nodes(p) WHERE n.gender = 'male') RETURN c.id",
	"MATCH (p:Person {id: 2})-[:KNOWS]-(f:Person) WITH collect(f) AS fs UNWIND fs AS g MATCH (g)-[:IS_LOCATED_IN]->(c:Place) RETURN g.id, c.name",
	"MATCH (p:Person) WHERE p.birthday >= date('1990-01-01') RETURN p.firstName, p.birthday ORDER BY p.birthday DESC LIMIT 5",
}

var uuidE2EMigs = []migrate.MigrationConfig{
	{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"firstName", "lastName", "gender", "birthday"}},
	{ObjType: catalog.ObjEntity, Entity: "Place", Properties: []string{"name"}},
}

// seedUUIDGraph empties Neo4j and loads a synthetic graph (60 Person, 5 Place, KNOWS, IS_LOCATED_IN).
func seedUUIDGraph(t *testing.T, cypher func(string, map[string]any) *neo4j.EagerResult) {
	t.Helper()
	cypher("MATCH (n) DETACH DELETE n", nil)
	first := []string{"Ann", "Bob", "Cid", "Dan", "Eve", "Fay", "Gus", "Hal"}
	last := []string{"Kim", "Lee", "Mori", "Sato"}
	var people []map[string]any
	for i := 1; i <= 60; i++ {
		g := "male"
		if i%3 == 0 {
			g = "female"
		}
		people = append(people, map[string]any{"uuid": id.Stored(id.New()), "id": int64(i),
			"firstName": first[i%len(first)], "lastName": last[i%len(last)], "gender": g,
			"birthday": fmt.Sprintf("19%02d-%02d-%02d", 70+i%30, 1+i%12, 1+i%28)})
	}
	cypher("UNWIND $rows AS r CREATE (:Person {uuid: r.uuid, id: r.id, firstName: r.firstName, lastName: r.lastName, gender: r.gender, birthday: date(r.birthday)})",
		map[string]any{"rows": people})
	var places []map[string]any
	for i := 1; i <= 5; i++ {
		places = append(places, map[string]any{"uuid": id.Stored(id.New()), "id": int64(i), "name": fmt.Sprintf("City%d", i)})
	}
	cypher("UNWIND $rows AS r CREATE (:Place {uuid: r.uuid, id: r.id, name: r.name, type: 'City'})", map[string]any{"rows": places})
	var knows, located []map[string]any
	for i := 1; i <= 60; i++ {
		for _, d := range []int{1, 7} {
			j := (i+d-1)%60 + 1
			knows = append(knows, map[string]any{"a": int64(i), "b": int64(j), "uuid": id.Stored(id.New())})
		}
		if i%6 != 0 { // multiples of 6 have no location (null in OPTIONAL)
			located = append(located, map[string]any{"a": int64(i), "c": int64(1 + i%5), "uuid": id.Stored(id.New())})
		}
	}
	cypher("UNWIND $rows AS r MATCH (a:Person {id: r.a}), (b:Person {id: r.b}) CREATE (a)-[:KNOWS {uuid: r.uuid, creationDate: datetime('2012-01-01T00:00:00Z')}]->(b)",
		map[string]any{"rows": knows})
	cypher("UNWIND $rows AS r MATCH (a:Person {id: r.a}), (c:Place {id: r.c}) CREATE (a)-[:IS_LOCATED_IN {uuid: r.uuid}]->(c)",
		map[string]any{"rows": located})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
