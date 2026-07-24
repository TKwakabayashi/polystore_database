package workload

import (
	"regexp"
	"sort"
	"testing"

	"polystore_database/src/go/migrator"
)

var paramRe = regexp.MustCompile(`\$(\w+)`)

// TestRegistryNamesSorted は AllWorkloadNames が Registry の全キーをソート済みで返すことを確認。
func TestRegistryNamesSorted(t *testing.T) {
	names := AllWorkloadNames()
	if len(names) != len(Registry) {
		t.Fatalf("AllWorkloadNames 件数 = %d, Registry = %d", len(names), len(Registry))
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("AllWorkloadNames がソートされていない: %v", names)
	}
	for _, n := range names {
		if _, ok := Registry[n]; !ok {
			t.Errorf("%q が Registry に無い", n)
		}
	}
}

// TestWorkloadParamsComplete は各ワークロードの cypher が参照する $param がすべて
// params に定義されていることを検証する（planner.replaceParameters の missing 相当を
// DB 無しで先取り検知する。ハードコード params のドリフト防止）。
func TestWorkloadParamsComplete(t *testing.T) {
	for _, name := range AllWorkloadNames() {
		name := name
		t.Run(name, func(t *testing.T) {
			def := Registry[name]
			cypher, params, _ := def(migrator.ModeGraphToRdb, true)
			if cypher == "" {
				t.Fatalf("cypher が空")
			}
			for _, m := range paramRe.FindAllStringSubmatch(cypher, -1) {
				key := m[1]
				if _, ok := params[key]; !ok {
					t.Errorf("$%s が params 未定義（params=%v）", key, params)
				}
			}
		})
	}
}

// TestWorkloadMigrationConfigs は isMigration=true で migration 設定が取得でき、
// 各設定に Entity 名があることを確認する（空 Entity の設定は移行時に不正）。
func TestWorkloadMigrationConfigs(t *testing.T) {
	for _, name := range AllWorkloadNames() {
		name := name
		t.Run(name, func(t *testing.T) {
			def := Registry[name]
			_, _, migs := def(migrator.ModeGraphToRdb, true)
			for i, m := range migs {
				if m.Entity == "" {
					t.Errorf("migs[%d] の Entity が空", i)
				}
			}
		})
	}
}
