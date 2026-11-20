//go:build integration

// Package db holds the tests that need the docker stack (build tag integration), compared with Neo4j running
// the original Cypher (the oracle). Run from src/go with the stack up:
//
//	go test -tags integration ./tests/db/ -run 'TestSmoke$' -v
//
// The config comes from POLYSTORE_CONFIG (default ../../config/config.json).
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/tests/dbtest"
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

// loadStackConfig loads the docker stack's config (its datastore holds the running dataset's runtime/).
func loadStackConfig(path string) (conn.Config, error) {
	cfg, err := conn.LoadConfig(path)
	if err != nil {
		return cfg, err
	}
	if err := cfg.CheckRuntime(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// typedForNeo4j types parameters for runner.RunNeo4j (integers, RFC 3339 times, strings).
func typedForNeo4j(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		var n int
		var list []any
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && fmt.Sprint(n) == v {
			out[k] = n
		} else if strings.HasPrefix(v, "[") && dbtest.JSONNumbers(v, &list) == nil {
			out[k] = list
		} else if tm, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = tm
		} else if d, err := time.Parse("2006-01-02", v); err == nil {
			// date-only strings go as Date (as strings, comparing with Date columns is null and the oracle returns nothing)
			out[k] = neo4j.DateOf(d)
		} else {
			out[k] = v
		}
	}
	return out
}

// reachableRegistry opens a Registry with the configured stores that respond (Neo4j is required); callers skip
// placements on stopped stores (the docker VM may not fit every store in memory).
func reachableRegistry(t *testing.T, ctx context.Context, cfg conn.Config) (*conn.Registry, map[kind.Kind]bool) {
	t.Helper()
	up := map[kind.Kind]bool{}
	var kinds []kind.Kind
	for _, k := range []kind.Kind{kind.Neo4j, kind.MySQL, kind.Mongo, kind.ClickHouse, kind.Cassandra, kind.LevelDB, kind.RocksDB} {
		pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		r, err := conn.NewRegistryFor(pctx, cfg, k)
		cancel()
		if err != nil {
			if k == kind.Neo4j {
				t.Fatalf("registry: %v", err)
			}
			t.Logf("%s is not reachable, its placement is skipped: %v", k, err)
			continue
		}
		_ = r.Close(ctx)
		up[k] = true
		kinds = append(kinds, k)
	}
	reg, err := conn.NewRegistryFor(ctx, cfg, kinds...)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	return reg, up
}

// placeMapping returns a mapping with only the migration columns that have data in the target store moved
// (tables from earlier migrations may lack some of a workload's columns).
func placeMapping(t *testing.T, base *catalog.MappingDictionary, target string, migs []migrate.MigrationConfig,
	ok func(kind catalog.ObjectType, label, prop string) bool) (*catalog.MappingDictionary, int) {
	t.Helper()
	b, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var md catalog.MappingDictionary
	if err := json.Unmarshal(b, &md); err != nil {
		t.Fatal(err)
	}
	md.Stats = base.Stats
	moved := 0
	for _, m := range migs {
		for _, p := range m.Properties {
			if !ok(m.ObjType, m.Entity, p) {
				continue
			}
			if m.ObjType == catalog.ObjRelationship {
				if r, has := md.Relationships[m.Entity]; has {
					if pm, has := r.Properties[p]; has {
						pm.Databases = []string{target}
						moved++
					}
				}
				continue
			}
			if e, has := md.Entities[m.Entity]; has {
				if pm, has := e.Properties[p]; has {
					pm.Databases = []string{target}
					moved++
				}
			}
		}
	}
	return &md, moved
}

var populatedCache = map[string]bool{}

// populated reports whether the target store has values for label/prop (false for missing or all-null columns).
func populated(ctx context.Context, reg *conn.Registry, target string, obj catalog.ObjectType, label, prop string) bool {
	key := target + "/" + label + "/" + prop
	if v, ok := populatedCache[key]; ok {
		return v
	}
	v := false
	switch target {
	case "relational":
		if db, ok := reg.MySQL(); ok {
			var one int
			err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT 1 FROM `%s` WHERE `%s` IS NOT NULL LIMIT 1", label, prop)).Scan(&one)
			v = err == nil
		}
	case "columnar":
		if db, ok := reg.ClickHouse(); ok {
			var one uint8
			err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT 1 FROM `%s` WHERE `%s` IS NOT NULL LIMIT 1", label, prop)).Scan(&one)
			v = err == nil
		}
	case "widecolumn":
		if sess, ok := reg.Cassandra(); ok {
			// MapScan turns null into the zero value (0 for bigint); scanning into a **T keeps null as nil.
			iter := sess.Query(fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 200`, prop, label)).WithContext(ctx).Iter()
			if cols := iter.Columns(); len(cols) == 1 {
				if typed, err := cols[0].TypeInfo.NewWithError(); err == nil {
					dst := reflect.New(reflect.TypeOf(typed))
					for iter.Scan(dst.Interface()) {
						if !dst.Elem().IsNil() {
							v = true
							break
						}
					}
				}
			}
			_ = iter.Close()
		}
	case "kvs", "rocksdb":
		k, _ := kind.ParseKind(target)
		v = kvsPopulated(ctx, reg, k, obj, label, prop)
	case "document":
		if db, ok := reg.Mongo(); ok {
			n, err := db.Collection(label).CountDocuments(ctx, map[string]interface{}{prop: map[string]interface{}{"$ne": nil}},
				options.Count().SetLimit(1))
			v = err == nil && n > 0
		}
	}
	populatedCache[key] = v
	return v
}

func kvsPopulated(ctx context.Context, reg *conn.Registry, k kind.Kind, obj catalog.ObjectType, label, prop string) bool {
	kv, ok := reg.KV(k)
	if !ok {
		return false
	}
	d, ok := reg.Neo4j()
	if !ok {
		return false
	}
	q := fmt.Sprintf("MATCH (n:%s) RETURN n.uuid AS u LIMIT 20", label)
	if obj == catalog.ObjRelationship {
		q = fmt.Sprintf("MATCH ()-[n:%s]->() RETURN n.uuid AS u LIMIT 20", label)
	}
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(ctx)
	res, err := sess.Run(ctx, q, nil)
	if err != nil {
		return false
	}
	for res.Next(ctx) {
		u, _ := res.Record().Get("u")
		if s, ok := u.(string); ok {
			if _, found, err := kv.Get(codec.BuildEntityKey(label, id.FromAny(s), prop)); err == nil && found {
				return true
			}
		}
	}
	return false
}

// allProperties returns migration configs for every property of every entity and relationship (placeMapping
// moves only the columns with values in the target).
func allProperties(md *catalog.MappingDictionary, mode migrate.MigrationMode) []migrate.MigrationConfig {
	var out []migrate.MigrationConfig
	add := func(kind catalog.ObjectType, name string, props map[string]*catalog.PropertyMapping) {
		var ps []string
		for p := range props {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		out = append(out, migrate.MigrationConfig{ObjType: kind, Entity: name, Properties: ps, Mode: mode})
	}
	for name, e := range md.Entities {
		add(catalog.ObjEntity, name, e.Properties)
	}
	for name, r := range md.Relationships {
		add(catalog.ObjRelationship, name, r.Properties)
	}
	return out
}
