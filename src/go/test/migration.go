package test

import (
	"context"
	"fmt"
	"time"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
)

// MigrationResult は1エンティティ分の移行結果。
type MigrationResult struct {
	Entity   string
	Mode     migrator.MigrationMode
	Count    int           // 移行件数
	Duration time.Duration // 移行時間
}

// RunMigration は migs を順に実行する。環境依存値は cfg から注入。
func RunMigration(ctx context.Context, cfg storage.Config, migs []migrator.MigrationConfig) ([]MigrationResult, error) {
	results := make([]MigrationResult, 0, len(migs))
	for _, mc := range migs {
		mc.MappingPath = cfg.MappingPath
		if cfg.Mongo != nil {
			mc.MongoDbName = cfg.Mongo.DBName
		}

		start := time.Now()
		count, err := migrator.MigrateData(mc, cfg)
		dur := time.Since(start)
		if err != nil {
			return results, fmt.Errorf("migration 失敗 (%s): %w", mc.Entity, err)
		}
		results = append(results, MigrationResult{
			Entity: mc.Entity, Mode: mc.Mode, Count: count, Duration: dur,
		})
	}
	return results, nil
}

// PrintMigration は移行結果（件数・時間）を出力する。
func PrintMigration(results []MigrationResult) {
	fmt.Println("=== Migration 結果 ===")
	for _, r := range results {
		fmt.Printf("  - %-14s mode=%-14s 件数=%-8d time=%v\n", r.Entity, r.Mode, r.Count, r.Duration)
	}
	fmt.Println()
}
