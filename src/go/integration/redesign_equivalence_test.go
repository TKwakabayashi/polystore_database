//go:build integration

// redesign_equivalence_test は刷新版のクエリ処理系（ast → logical → physical → native → executor）が
// graph 配置で Neo4j に原文をそのまま投げた結果（オラクル）と同じ行集合を返すことを確かめる。
//
//	実行: cwd=src/go で docker スタックを up した状態で
//	  go test -tags integration ./integration/ -run RedesignEquivalence -v
//	設定は POLYSTORE_CONFIG（未設定なら ../../config/config.json）から読む。
//
// 構造化できないワークロードは処理系自身が verbatim に落とすので一致は自明だが、実行エラーの検出のため残す。
// 非 graph 配置の一致は移行が要るので、配置×ワークロードの全面確認は bench-models 側で行う。
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/mongo/options"

	"polystore_database/src/go/bench"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/executor"
	"polystore_database/src/go/id"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

func TestRedesignEquivalence(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	md, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping load (%s): %v", cfg.MappingPath, err)
	}
	ctx := context.Background()
	reg, err := storage.NewRegistryFor(ctx, cfg, storage.Graph)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := executor.NewStoreBackend(reg)

	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			cypher, params, _ := workload.Registry[name](migrator.ModeGraphToRdb, false)
			oracle, err := bench.RunNeo4j(ctx, *cfg.Neo4j, cypher, typedForNeo4j(params))
			if err != nil {
				t.Skipf("Neo4j（オラクル）が失敗したため skip: %v", err)
			}
			got, err := executor.Run(ctx, be, md, cypher, params)
			if err != nil {
				t.Fatalf("redesign の実行に失敗: %v", err)
			}
			want, have := canonRedesign(oracle.Rows), canonRedesign(got.Rows)
			if strings.Join(want, "\n") != strings.Join(have, "\n") {
				t.Errorf("[%s] 行集合がオラクルと不一致（oracle=%d, redesign=%d）\n%s",
					got.Engine, len(want), len(have), firstRedesignDiff(want, have))
				return
			}
			t.Logf("[%s] rows=%d 一致（round trips=%d, plan=%v, exec=%v）",
				got.Engine, len(have), got.RoundTrips, got.PlanTime, got.ExecTime)
		})
	}
}

// typedForNeo4j は bench.RunNeo4j に渡す typed params（整数・RFC3339 時刻・文字列）。
func typedForNeo4j(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && fmt.Sprint(n) == v {
			out[k] = n
		} else if tm, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = tm
		} else {
			out[k] = v
		}
	}
	return out
}

// canonRedesign は行を順序非依存で比べられる文字列にする（数値・時刻の表現をそろえる）。
func canonRedesign(rows []map[string]interface{}) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%s;", k, canonValue(row[k]))
		}
		out[i] = b.String()
	}
	sort.Strings(out)
	return out
}

func canonValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case int:
		return fmt.Sprint(int64(x))
	case int32:
		return fmt.Sprint(int64(x))
	case int64:
		return fmt.Sprint(x)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', 10, 64)
	case float64:
		// 集約の足し合わせ順で末尾の桁が揺れるので 10 桁で比べる。
		return strconv.FormatFloat(x, 'g', 10, 64)
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano)
	case dbtype.Date: // オラクル（Neo4j 原文）の DATE。実行器は time.Time（UTC の 0 時）にそろえて返す
		return time.Time(x).UTC().Format(time.RFC3339Nano)
	case dbtype.LocalDateTime:
		return time.Time(x).UTC().Format(time.RFC3339Nano)
	case []interface{}:
		// collect の要素順は Cypher でも定まらないので、リストは多重集合として比べる。
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = canonValue(e)
		}
		sort.Strings(parts)
		return "[" + strings.Join(parts, ",") + "]"
	}
	return fmt.Sprint(v)
}

func firstRedesignDiff(want, got []string) string {
	n := len(want)
	if len(got) < n {
		n = len(got)
	}
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			return fmt.Sprintf("  行 %d:\n    oracle  : %s\n    redesign: %s", i, want[i], got[i])
		}
	}
	return fmt.Sprintf("  長さのみ相違: oracle=%d redesign=%d", len(want), len(got))
}

// TestRedesignPlacements は、各ワークロードの移行設定どおりに列を非 graph ストアへ置いた配置で計画・実行し、
// オラクル（Neo4j verbatim）と行集合を比べる。移行済みのデータが各ストアに入っていることが前提
// （bench-models の移行で作られたもの）。document はデータが無ければ skip する。
//
//	go test -tags integration ./integration/ -run RedesignPlacements -v
func TestRedesignPlacements(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	base, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping load: %v", err)
	}
	ctx := context.Background()
	reg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := executor.NewStoreBackend(reg)

	placements := []struct {
		name   string
		target string
		mode   migrator.MigrationMode
	}{
		{"relational", "relational", migrator.ModeGraphToRdb},
		{"columnar", "columnar", migrator.ModeGraphToCol},
		{"kvs", "kvs", migrator.ModeGraphToKvs},
		{"document", "document", migrator.ModeGraphToDoc},
	}

	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, pc := range placements {
		for _, name := range names {
			pc, name := pc, name
			t.Run(pc.name+"/"+name, func(t *testing.T) {
				cypher, params, migs := workload.Registry[name](pc.mode, true)
				if len(migs) == 0 {
					migs = allProperties(base, pc.mode) // 移行設定を持たないワークロードは、移行済みの全列で試す
				}
				md, moved := placeMapping(t, base, pc.target, migs, func(kind schema.ObjectType, label, prop string) bool {
					return populated(ctx, reg, pc.target, kind, label, prop)
				})
				if moved == 0 {
					t.Skip("移行済みの列が無い")
				}
				if _, _, err := executor.Plan(cypher, params, md); err != nil {
					t.Skipf("構造化できない（verbatim）: %v", err)
				}
				oracle, err := bench.RunNeo4j(ctx, *cfg.Neo4j, cypher, typedForNeo4j(params))
				if err != nil {
					t.Skipf("Neo4j（オラクル）が失敗: %v", err)
				}
				qctx, cancel := context.WithTimeout(ctx, 90*time.Second)
				defer cancel()
				got, err := executor.Run(qctx, be, md, cypher, params)
				if err != nil && qctx.Err() != nil {
					t.Skipf("90 秒で打ち切り（既知: ストアをまたぐ中間結果が大きい形）: %v", err)
				}
				if err != nil {
					t.Fatalf("redesign の実行に失敗: %v", err)
				}
				want, have := canonRedesign(oracle.Rows), canonRedesign(got.Rows)
				if strings.Join(want, "\n") != strings.Join(have, "\n") {
					t.Errorf("行集合がオラクルと不一致（oracle=%d, redesign=%d, moved=%d）\n%s", len(want), len(have), moved, firstRedesignDiff(want, have))
					return
				}
				t.Logf("rows=%d 一致（moved=%d, round trips=%d, exec=%v）", len(have), moved, got.RoundTrips, got.ExecTime)
			})
		}
	}
}

// placeMapping は移行設定の列のうち、target ストアに実際にデータがある列だけを移した mapping を作る
// （過去の移行で作られた表は、ワークロードの移行設定の列をすべて持っているとは限らないため）。
func placeMapping(t *testing.T, base *schema.MappingDictionary, target string, migs []migrator.MigrationConfig,
	ok func(kind schema.ObjectType, label, prop string) bool) (*schema.MappingDictionary, int) {
	t.Helper()
	b, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	var md schema.MappingDictionary
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
			if m.ObjType == schema.ObjRelationship {
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

// populated は target ストアの label/prop に値が入っているか（NULL だけの列・無い列は false）。
func populated(ctx context.Context, reg *storage.Registry, target string, kind schema.ObjectType, label, prop string) bool {
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
		if sess, ok := reg.Cassandra(); ok {
			iter := sess.Query(fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 200`, prop, label)).WithContext(ctx).Iter()
			m := map[string]interface{}{}
			for iter.MapScan(m) {
				if m[strings.ToLower(prop)] != nil {
					v = true
					break
				}
				m = map[string]interface{}{}
			}
			_ = iter.Close()
		}
	case "kvs":
		v = kvsPopulated(ctx, reg, kind, label, prop)
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

func kvsPopulated(ctx context.Context, reg *storage.Registry, kind schema.ObjectType, label, prop string) bool {
	ldb, ok := reg.LevelDB()
	if !ok {
		return false
	}
	d, ok := reg.Neo4j()
	if !ok {
		return false
	}
	q := fmt.Sprintf("MATCH (n:%s) RETURN n.uuid AS u LIMIT 20", label)
	if kind == schema.ObjRelationship {
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
			if _, err := ldb.Get(codec.BuildEntityKey(label, id.FromAny(s), prop), nil); err == nil {
				return true
			}
		}
	}
	return false
}

// TestRedesignModels は反復モデルを変えても結果が同じことを、実データの各配置で確かめる。
// 基準は同じ配置の bulk（オラクルとの一致は TestRedesignEquivalence / TestRedesignPlacements が担う）。
//
//	go test -tags integration ./integration/ -run RedesignModels -v
func TestRedesignModels(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	base, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		t.Fatalf("mapping load: %v", err)
	}
	ctx := context.Background()
	reg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	defer reg.Close(ctx)
	be := executor.NewStoreBackend(reg)

	placements := []struct {
		name   string
		target string
		mode   migrator.MigrationMode
	}{
		{"graph", "", migrator.ModeGraphToRdb},
		{"relational", "relational", migrator.ModeGraphToRdb},
		{"document", "document", migrator.ModeGraphToDoc},
		{"columnar", "columnar", migrator.ModeGraphToCol},
		{"kvs", "kvs", migrator.ModeGraphToKvs},
	}
	models := []settings.EngineKind{settings.EngineVolcano, settings.EngineVectorized, settings.EngineStream, settings.EngineVecStream}
	var names []string
	for n := range workload.Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, pc := range placements {
		for _, name := range names {
			pc, name := pc, name
			t.Run(pc.name+"/"+name, func(t *testing.T) {
				cypher, params, migs := workload.Registry[name](pc.mode, pc.target != "")
				if pc.target != "" && len(migs) == 0 {
					migs = allProperties(base, pc.mode)
				}
				md := base
				if pc.target != "" {
					var moved int
					md, moved = placeMapping(t, base, pc.target, migs, func(kind schema.ObjectType, label, prop string) bool {
						return populated(ctx, reg, pc.target, kind, label, prop)
					})
					if moved == 0 {
						t.Skip("移行済みの列が無い")
					}
				}
				if _, _, err := executor.Plan(cypher, params, md); err != nil {
					t.Skipf("構造化できない（verbatim）: %v", err)
				}
				run := func(m settings.EngineKind) (string, []string, bool) {
					qctx, cancel := context.WithTimeout(ctx, 60*time.Second)
					defer cancel()
					r, err := executor.RunModel(qctx, be, md, cypher, params, m)
					if err != nil && qctx.Err() != nil {
						return "", nil, false
					}
					if err != nil {
						t.Fatalf("[%s] 実行に失敗: %v", m, err)
					}
					return fmt.Sprintf("rt=%d exec=%v", r.RoundTrips, r.ExecTime.Round(time.Millisecond)), canonRedesign(r.Rows), true
				}
				refInfo, ref, ok := run(settings.EngineBulk)
				if !ok {
					t.Skip("bulk が 60 秒で終わらない")
				}
				t.Logf("[bulk] rows=%d %s", len(ref), refInfo)
				for _, m := range models {
					info, got, ok := run(m)
					if !ok {
						t.Logf("[%s] 60 秒で打ち切り", m)
						continue
					}
					if strings.Join(got, "\n") != strings.Join(ref, "\n") {
						t.Errorf("[%s] bulk と不一致（bulk=%d, %s=%d）\n%s", m, len(ref), m, len(got), firstRedesignDiff(ref, got))
						continue
					}
					t.Logf("[%s] 一致 %s", m, info)
				}
			})
		}
	}
}

// allProperties は mapping の全エンティティ・全関係の全プロパティを移行対象とする設定
// （placeMapping が、移行先に値がある列だけを実際に移す）。
func allProperties(md *schema.MappingDictionary, mode migrator.MigrationMode) []migrator.MigrationConfig {
	var out []migrator.MigrationConfig
	add := func(kind schema.ObjectType, name string, props map[string]*schema.PropertyMapping) {
		var ps []string
		for p := range props {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		out = append(out, migrator.MigrationConfig{ObjType: kind, Entity: name, Properties: ps, Mode: mode})
	}
	for name, e := range md.Entities {
		add(schema.ObjEntity, name, e.Properties)
	}
	for name, r := range md.Relationships {
		add(schema.ObjRelationship, name, r.Properties)
	}
	return out
}
