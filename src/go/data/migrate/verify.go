package migrate

import (
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"

	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
)

// VerifyMigration counts source and destination independently and requires destination >= source. Otherwise
// rows were missed: it returns an error and the caller skips the record update and the source deletion (the
// gate against data loss). On success it returns the number migrated (the source count).
func VerifyMigration(ctx context.Context, cfg MigrationConfig, srcKind, destKind kind.Kind, reg *conn.Registry) (int64, error) {
	fmt.Printf("verifying the migration (%s -> %s)...\n", srcKind, destKind)
	srcCount, err := CountEntity(ctx, cfg, srcKind, reg)
	if err != nil {
		return 0, fmt.Errorf("source count: %w", err)
	}
	destCount, err := CountEntity(ctx, cfg, destKind, reg)
	if err != nil {
		return 0, fmt.Errorf("dest count: %w", err)
	}
	if destCount < srcCount {
		return 0, fmt.Errorf("%s: the destination has fewer rows (source %d, destination %d)", destKind, srcCount, destCount)
	}
	fmt.Printf("   ok: source %d, destination %d\n", srcCount, destCount)
	return srcCount, nil
}

// CountEntity returns the number of cfg.Entity rows (uuids) in a store.
func CountEntity(ctx context.Context, cfg MigrationConfig, kind kind.Kind, reg *conn.Registry) (int64, error) {
	m, err := dbs.MoverOf(kind)
	if err != nil {
		return 0, err
	}
	return m.Count(ctx, reg, cfg.target(nil))
}
