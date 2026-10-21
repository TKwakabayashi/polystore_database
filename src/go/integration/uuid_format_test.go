//go:build integration

// uuid_format_test は uuid の保存形式（settings.UUIDStorage: 文字列 / 17 バイト）を切り替えても、
// 全ストアで同じ結果になることを、使い捨ての環境で端から端まで確かめる:
//
//	合成グラフを Neo4j に投入（文字列）→ 各ストアへ移行して問い合わせ → `uuid-format` 相当でバイナリへ書き換え →
//	非 graph を空にして移行し直し、問い合わせ → 文字列へ書き戻して問い合わせ
//
// 問い合わせはどれも Neo4j に原文を投げた結果（オラクル）と比べる。
//
// 実データを壊さないよう、専用の設定（POLYSTORE_UUID_E2E_CONFIG）でだけ動き、Mongo の DB 名・MySQL の DSN・
// Cassandra の keyspace・LevelDB のパスがすべて "uuidtest" を含み、Neo4j のノードが 10 万件未満のときに限る。
//
// 注意: 使い捨ての Neo4j を本番データの docker スタックと同じ VM で動かすとメモリを取り合う
// （2026-10-20 に ldbc_sf1 の Neo4j が OOM で落ちた）。小さいヒープで動かし、終わったら止めること。
//
//	実行例: POLYSTORE_UUID_E2E_CONFIG=/path/to/uuidtest.json \
//	  go test -tags integration ./integration/ -run UUIDFormatE2E -v
package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/executor"
	"polystore_database/src/go/id"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/setup"
	"polystore_database/src/go/storage"
)

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

var uuidE2EMigs = []migrator.MigrationConfig{
	{ObjType: schema.ObjEntity, Entity: "Person", Properties: []string{"firstName", "lastName", "gender", "birthday"}},
	{ObjType: schema.ObjEntity, Entity: "Place", Properties: []string{"name"}},
}

func TestUUIDFormatE2E(t *testing.T) {
	path := os.Getenv("POLYSTORE_UUID_E2E_CONFIG")
	if path == "" {
		t.Skip("POLYSTORE_UUID_E2E_CONFIG（使い捨て環境の設定）が無い")
	}
	cfg, err := storage.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	for name, v := range map[string]string{"mongo.dbname": cfg.Mongo.DBName, "mysql.dsn": cfg.MySQL.DSN,
		"cassandra.keyspace": cfg.Cassandra.Keyspace, "leveldb.path": cfg.LevelDB.Path} {
		if !strings.Contains(v, "uuidtest") {
			t.Fatalf("%s=%q は uuidtest を含まない（実データを消さないため中止）", name, v)
		}
	}
	ctx := context.Background()
	prev := settings.UUIDStorage
	defer func() { settings.UUIDStorage = prev }()

	// mapping は移行が書き換えるので、毎回元の写しから始める。
	original := cfg.MappingPath
	cfg.MappingPath = filepath.Join(t.TempDir(), "mapping.json")
	resetMapping := func() {
		t.Helper()
		if err := copyFile(original, cfg.MappingPath); err != nil {
			t.Fatal(err)
		}
	}
	resetMapping()

	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close(ctx)
	cypher := func(q string, p map[string]any) *neo4j.EagerResult {
		t.Helper()
		r, err := neo4j.ExecuteQuery(ctx, driver, q, p, neo4j.EagerResultTransformer)
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		return r
	}
	if n := cypher("MATCH (n) RETURN count(n) AS c", nil).Records[0].Values[0].(int64); n >= 100000 {
		t.Fatalf("Neo4j に %d ノードある（使い捨て環境ではない）", n)
	}

	settings.UUIDStorage = settings.UUIDText
	seedUUIDGraph(t, cypher)

	for _, phase := range []struct {
		name   string
		format settings.UUIDFormat
	}{{"text", settings.UUIDText}, {"binary", settings.UUIDBinary}, {"text-again", settings.UUIDText}} {
		t.Run(phase.name, func(t *testing.T) {
			settings.UUIDStorage = phase.format
			if err := setup.ConvertUUIDFormat(ctx, cfg); err != nil {
				t.Fatal(err)
			}
			checkNeo4jFormat(t, cypher, phase.format)
			if err := setup.Run(ctx, cfg); err != nil { // 非 graph の 4 ストアを空にする（列型を作り直すため）
				t.Fatal(err)
			}
			for _, pl := range []struct {
				name string
				mode migrator.MigrationMode
			}{{"graph", ""}, {"relational", migrator.ModeGraphToRdb}, {"columnar", migrator.ModeGraphToCol},
				{"kvs", migrator.ModeGraphToKvs}, {"document", migrator.ModeGraphToDoc}} {
				resetMapping()
				for _, m := range uuidE2EMigs {
					if pl.mode == "" {
						break
					}
					m.Mode, m.MappingPath, m.MongoDbName = pl.mode, cfg.MappingPath, cfg.Mongo.DBName
					if _, err := migrator.MigrateData(m, cfg); err != nil {
						t.Fatalf("%s: migrate %s: %v", pl.name, m.Entity, err)
					}
				}
				md, err := schema.LoadMappingDictionary(cfg.MappingPath)
				if err != nil {
					t.Fatal(err)
				}
				reg, err := storage.NewRegistry(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				be := executor.NewStoreBackend(reg)
				for qi, q := range uuidE2EQueries {
					oracle, err := bench.RunNeo4j(ctx, *cfg.Neo4j, q, map[string]any{})
					if err != nil {
						t.Fatalf("oracle %d: %v", qi, err)
					}
					want := canonRedesign(oracle.Rows)
					for _, model := range []settings.EngineKind{settings.EngineBulk, settings.EngineStream, settings.EngineVectorized} {
						got, err := executor.RunModel(ctx, be, md, q, nil, model)
						if err != nil {
							t.Errorf("%s/%s q%d: %v", pl.name, model, qi, err)
							continue
						}
						if have := canonRedesign(got.Rows); strings.Join(have, "\n") != strings.Join(want, "\n") {
							t.Errorf("%s/%s q%d %q: オラクルと不一致\n%s", pl.name, model, qi, q, firstRedesignDiff(want, have))
						}
					}
				}
				checkStoredFormat(t, ctx, reg, pl.name, phase.format)
				reg.Close(ctx)
				t.Logf("%s: %d クエリ × 3 モデル一致", pl.name, len(uuidE2EQueries))
			}
		})
	}
}

// seedUUIDGraph は Neo4j を空にして合成グラフ（Person 60・Place 5・KNOWS・IS_LOCATED_IN）を入れる。
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
		people = append(people, map[string]any{"uuid": id.Stored(id.New(id.KindNode)), "id": int64(i),
			"firstName": first[i%len(first)], "lastName": last[i%len(last)], "gender": g,
			"birthday": fmt.Sprintf("19%02d-%02d-%02d", 70+i%30, 1+i%12, 1+i%28)})
	}
	cypher("UNWIND $rows AS r CREATE (:Entity:Person {uuid: r.uuid, id: r.id, firstName: r.firstName, lastName: r.lastName, gender: r.gender, birthday: date(r.birthday)})",
		map[string]any{"rows": people})
	var places []map[string]any
	for i := 1; i <= 5; i++ {
		places = append(places, map[string]any{"uuid": id.Stored(id.New(id.KindNode)), "id": int64(i), "name": fmt.Sprintf("City%d", i)})
	}
	cypher("UNWIND $rows AS r CREATE (:Entity:Place {uuid: r.uuid, id: r.id, name: r.name, type: 'City'})", map[string]any{"rows": places})
	var knows, located []map[string]any
	for i := 1; i <= 60; i++ {
		for _, d := range []int{1, 7} {
			j := (i+d-1)%60 + 1
			knows = append(knows, map[string]any{"a": int64(i), "b": int64(j), "uuid": id.Stored(id.New(id.KindEdge))})
		}
		if i%6 != 0 { // 6 の倍数は所在地なし（OPTIONAL の null）
			located = append(located, map[string]any{"a": int64(i), "c": int64(1 + i%5), "uuid": id.Stored(id.New(id.KindEdge))})
		}
	}
	cypher("UNWIND $rows AS r MATCH (a:Person {id: r.a}), (b:Person {id: r.b}) CREATE (a)-[:KNOWS {uuid: r.uuid, creationDate: datetime('2012-01-01T00:00:00Z')}]->(b)",
		map[string]any{"rows": knows})
	cypher("UNWIND $rows AS r MATCH (a:Person {id: r.a}), (c:Place {id: r.c}) CREATE (a)-[:IS_LOCATED_IN {uuid: r.uuid}]->(c)",
		map[string]any{"rows": located})
}

// checkNeo4jFormat は Neo4j の uuid がすべて format の型（文字列 / バイト列）で、どれも解釈できることを確かめる。
func checkNeo4jFormat(t *testing.T, cypher func(string, map[string]any) *neo4j.EagerResult, format settings.UUIDFormat) {
	t.Helper()
	for _, q := range []string{"MATCH (n) RETURN n.uuid AS u", "MATCH ()-[r]->() RETURN r.uuid AS u"} {
		for _, rec := range cypher(q, nil).Records {
			v := rec.Values[0]
			_, isBytes := v.([]byte)
			if isBytes != (format == settings.UUIDBinary) || id.FromAny(v).Empty() {
				t.Fatalf("%s: 形式が違う値 %T %v", q, v, v)
			}
		}
	}
}

// checkStoredFormat は移行先ストアに書かれた Person の uuid が format の物理表現であることを確かめる。
func checkStoredFormat(t *testing.T, ctx context.Context, reg *storage.Registry, store string, format settings.UUIDFormat) {
	t.Helper()
	binary := format == settings.UUIDBinary
	switch store {
	case "relational":
		db, _ := reg.MySQL()
		var typ string
		if err := db.QueryRowContext(ctx, "SELECT DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'Person' AND COLUMN_NAME = 'uuid'").Scan(&typ); err != nil {
			t.Fatal(err)
		}
		if (typ == "binary") != binary {
			t.Errorf("relational: uuid 列の型 %s", typ)
		}
	case "columnar":
		sess, _ := reg.Cassandra()
		iter := sess.Query(`SELECT uuid FROM "Person" LIMIT 1`).WithContext(ctx).Iter()
		cols := iter.Columns()
		_ = iter.Close()
		if len(cols) == 0 || (cols[0].TypeInfo.Type() == gocql.TypeBlob) != binary {
			t.Errorf("columnar: uuid 列の型 %v", cols)
		}
	case "document":
		db, _ := reg.Mongo()
		var doc bson.M
		if err := db.Collection("Person").FindOne(ctx, bson.M{}).Decode(&doc); err != nil {
			t.Fatal(err)
		}
		if _, isBin := doc["uuid"].(primitive.Binary); isBin != binary {
			t.Errorf("document: uuid の値 %T", doc["uuid"])
		}
	case "kvs":
		ldb, _ := reg.LevelDB()
		iter := ldb.NewIterator(util.BytesPrefix([]byte("Person\x00")), nil)
		defer iter.Release()
		if !iter.Next() {
			t.Fatal("kvs: Person のキーが無い")
		}
		// ParseEntityKey は uuid の位置を保存形式の長さで決めるので、形式が違うキーは ok=false になる。
		if _, u, _, ok := codec.ParseEntityKey(iter.Key()); !ok || u.Empty() {
			t.Errorf("kvs: キー %q が保存形式と違う", iter.Key())
		}
	}
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
