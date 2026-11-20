//go:build integration

// roundtrip_test copies a workload's properties to each store and back (the source is kept) and checks the
// counts.
//
//	go test -tags integration ./tests/db/ -run TestRoundtrip -v
package db

import (
	"context"
	"os"
	"strings"
	"testing"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/stores/conn"
)

// TestRoundtrip runs a non-destructive round trip (src -> dest -> src) per store and checks
// (1) no data loss, (2) a consistent copy, (3) an identical restore.
func TestRoundtrip(t *testing.T) {
	cfgPath := os.Getenv("POLYSTORE_CONFIG")
	if cfgPath == "" {
		cfgPath = "../../config/config.json"
	}
	cfg, err := loadStackConfig(cfgPath)
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath, err)
	}
	ctx := context.Background()

	cases := []struct {
		wl   string
		mode migrate.MigrationMode
	}{
		{"AGG5", "graph_to_rdb"},
		{"AGG5", "graph_to_kvs"},
		{"AGG5", "graph_to_doc"},
		{"AGG5", "graph_to_col"},
		{"AGG5", "graph_to_wide"},
	}
	for _, c := range cases {
		t.Run(c.wl+"_"+string(c.mode), func(t *testing.T) {
			roundtrip(t, ctx, cfg, c.wl, c.mode)
		})
	}
}

// roundtrip migrates each migration config of the workload forward and back and checks counts (DeleteSource is
// false, so the source stays).
func roundtrip(t *testing.T, ctx context.Context, cfg conn.Config, name string, fwd migrate.MigrationMode) {
	def, _, ok := workload.Lookup(name)
	if !ok {
		t.Fatalf("unknown workload %q", name)
	}
	_, _, migs := def(fwd, true)
	if len(migs) == 0 {
		t.Fatalf("workload %s has no migration settings", name)
	}
	back, err := reverseMode(fwd)
	if err != nil {
		t.Fatal(err)
	}
	srcKind, destKind, err := migrate.ModeStores(fwd)
	if err != nil {
		t.Fatalf("bad migmode %q: %v", fwd, err)
	}

	for _, mig := range migs {
		mig.MappingPath = cfg.MappingPath()
		if cfg.Mongo != nil {
			mig.MongoDbName = cfg.Mongo.DBName
		}
		mig.DeleteSource = false
		verifyOne(t, ctx, cfg, mig, srcKind, destKind, fwd, back)
	}
}

func verifyOne(t *testing.T, ctx context.Context, cfg conn.Config, mig migrate.MigrationConfig,
	srcKind, destKind kind.Kind, fwd, back migrate.MigrationMode) {

	srcBefore, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] count before: %v", mig.Entity, err)
	}
	if srcBefore == 0 {
		t.Skipf("[%s] the source (%s) is empty; was setup run?", mig.Entity, srcKind)
	}

	// --- forward: src -> dest ---
	fmig := mig
	fmig.Mode = fwd
	if _, err := migrate.MigrateData(fmig, cfg); err != nil {
		t.Fatalf("[%s] forward migration: %v", mig.Entity, err)
	}
	srcAfter, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] source count after: %v", mig.Entity, err)
	}
	destAfter, err := countVia(ctx, cfg, mig, destKind)
	if err != nil {
		t.Fatalf("[%s] destination count after: %v", mig.Entity, err)
	}

	// check 1: the copy holds (dest >= src before the migration)
	if destAfter < srcBefore {
		t.Errorf("[%s] incomplete copy: dest(%s)=%d < source before=%d", mig.Entity, destKind, destAfter, srcBefore)
	}
	// check 2: the source is kept (no delete)
	if srcAfter != srcBefore {
		t.Errorf("[%s] the source changed: %d before, %d after (unexpected delete)", mig.Entity, srcBefore, srcAfter)
	}

	// --- back: dest -> src (restore) ---
	bmig := mig
	bmig.Mode = back
	if _, err := migrate.MigrateData(bmig, cfg); err != nil {
		t.Fatalf("[%s] backward migration (the mapping still says %s; setup restores it): %v", mig.Entity, destKind, err)
	}
	srcRestored, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		t.Fatalf("[%s] source count after restoring: %v", mig.Entity, err)
	}
	// check 3: the round trip matches
	if srcRestored != srcBefore {
		t.Errorf("[%s] round trip differs: %d -> %d", mig.Entity, srcBefore, srcRestored)
	}
	t.Logf("[%s] %s→%s OK: src=%d dest=%d restored=%d", mig.Entity, srcKind, destKind, srcBefore, destAfter, srcRestored)
}

// countVia opens only the kind store, counts cfg.Entity and closes it (avoids the LevelDB lock).
func countVia(ctx context.Context, cfg conn.Config, mig migrate.MigrationConfig, kind kind.Kind) (int64, error) {
	reg, err := conn.NewRegistryFor(ctx, cfg, kind)
	if err != nil {
		return 0, err
	}
	defer reg.Close(ctx)
	return migrate.CountEntity(ctx, mig, kind, reg)
}

// reverseMode turns "a_to_b" into "b_to_a".
func reverseMode(m migrate.MigrationMode) (migrate.MigrationMode, error) {
	parts := strings.SplitN(string(m), "_to_", 2)
	if len(parts) != 2 {
		return "", nil
	}
	return migrate.MigrationMode(parts[1] + "_to_" + parts[0]), nil
}
