//go:build integration

// Package integration は実 DB スタック（docker: Neo4j/Mongo/MySQL/Cassandra/LevelDB）を
// 必要とする結合テスト。通常のユニットテスト（DB 不要）とは分離し、build タグで隔離する。
//
//	実行: cwd=src/go で docker スタックを up した状態で
//	  go test -tags integration ./integration/ -run Roundtrip -v
//	設定は POLYSTORE_CONFIG（未設定なら ../../config/config.json）から読む。
//
// 旧 bench/verify_migrate.go（-mode verify-migrate）を go test 化したもの。
package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// TestMigrationRoundtrip は各ストアへの非破壊ラウンドトリップ（src→dest→src）で
// (1) データ消失なし (2) コピー整合 (3) 復元一致 を検証する。
func TestMigrationRoundtrip(t *testing.T) {
	// go test は package ディレクトリ（integration/）を cwd にするため、
	// アプリ（cwd=src/go）と同じ相対パス前提に合わせて 1 段上へ移動する。
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("chdir to src/go: %v", err)
	}
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := storage.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	cases := []struct {
		wl   string
		mode migrator.MigrationMode
	}{
		{"AGG5", "graph_to_rdb"},
		{"AGG5", "graph_to_kvs"},
		{"AGG5", "graph_to_doc"},
		{"AGG5", "graph_to_col"},
	}
	for _, c := range cases {
		t.Run(c.wl+"_"+string(c.mode), func(t *testing.T) {
			roundtrip(t, ctx, cfg, c.wl, c.mode)
		})
	}
}

// roundtrip は workload の各 migration 設定について fwd 方向の移行と逆方向の復元を行い、
// 件数ベースで検証する（DeleteSource=false のためソースは消えない）。
func roundtrip(t *testing.T, ctx context.Context, cfg storage.Config, name string, fwd migrator.MigrationMode) {
	def, ok := workload.Registry[name]
	if !ok {
		t.Fatalf("未知のワークロード %q", name)
	}
	_, _, migs := def(fwd, true)
	if len(migs) == 0 {
		t.Fatalf("ワークロード %s に migration 設定がありません", name)
	}
	back, err := reverseMode(fwd)
	if err != nil {
		t.Fatal(err)
	}
	srcKind, destKind, err := migrator.ModeStores(fwd)
	if err != nil {
		t.Fatalf("不正な migmode %q: %v", fwd, err)
	}

	for _, mig := range migs {
		mig.MappingPath = cfg.MappingPath
		if cfg.Mongo != nil {
			mig.MongoDbName = cfg.Mongo.DBName
		}
		mig.DeleteSource = false
		verifyOne(t, ctx, cfg, mig, srcKind, destKind, fwd, back)
	}
}

func verifyOne(t *testing.T, ctx context.Context, cfg storage.Config, mig migrator.MigrationConfig,
	srcKind, destKind storage.StoreKind, fwd, back migrator.MigrationMode) {

	srcBefore, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] 事前カウント失敗: %v", mig.Entity, err)
	}
	if srcBefore == 0 {
		t.Skipf("[%s] ソース(%s)が空。setup 済みか確認", mig.Entity, srcKind)
	}

	// --- 順方向: src → dest ---
	fmig := mig
	fmig.Mode = fwd
	if _, err := migrator.MigrateData(fmig, cfg); err != nil {
		t.Fatalf("[%s] 順方向 migrate 失敗: %v", mig.Entity, err)
	}
	srcAfter, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] 移行後 src カウント失敗: %v", mig.Entity, err)
	}
	destAfter, err := countVia(ctx, cfg, mig, destKind)
	if err != nil {
		t.Fatalf("[%s] 移行後 dest カウント失敗: %v", mig.Entity, err)
	}

	// 検証1: コピー成立（dest ≥ src移行前）
	if destAfter < srcBefore {
		t.Errorf("[%s] コピー不足: dest(%s)=%d < src移行前=%d", mig.Entity, destKind, destAfter, srcBefore)
	}
	// 検証2: ソース保全（Delete が走らない）
	if srcAfter != srcBefore {
		t.Errorf("[%s] ソースが変化: 移行前=%d → 移行後=%d（想定外の削除）", mig.Entity, srcBefore, srcAfter)
	}

	// --- 逆方向: dest → src（原状復帰） ---
	bmig := mig
	bmig.Mode = back
	if _, err := migrator.MigrateData(bmig, cfg); err != nil {
		t.Fatalf("[%s] 逆方向 migrate 失敗（mapping が %s のまま。setup で復旧可）: %v", mig.Entity, destKind, err)
	}
	srcRestored, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] 復元後 src カウント失敗: %v", mig.Entity, err)
	}
	// 検証3: ラウンドトリップ一致
	if srcRestored != srcBefore {
		t.Errorf("[%s] ラウンドトリップ不一致: %d → %d", mig.Entity, srcBefore, srcRestored)
	}
	t.Logf("[%s] %s→%s OK: src=%d dest=%d restored=%d", mig.Entity, srcKind, destKind, srcBefore, destAfter, srcRestored)
}

// countVia は kind ストアだけを開いて cfg.Entity を数え、すぐ閉じる（LevelDB のロック回避）。
func countVia(ctx context.Context, cfg storage.Config, mig migrator.MigrationConfig, kind storage.StoreKind) (int64, error) {
	reg, err := storage.NewRegistryFor(ctx, cfg, kind)
	if err != nil {
		return 0, err
	}
	defer reg.Close(ctx)
	return migrator.CountEntity(ctx, mig, kind, reg)
}

// reverseMode は "a_to_b" を "b_to_a" に反転する。
func reverseMode(m migrator.MigrationMode) (migrator.MigrationMode, error) {
	parts := strings.SplitN(string(m), "_to_", 2)
	if len(parts) != 2 {
		return "", nil
	}
	return migrator.MigrationMode(parts[1] + "_to_" + parts[0]), nil
}
