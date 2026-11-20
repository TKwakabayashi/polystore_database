package migrate

import (
	"context"
	"fmt"
	"time"

	"polystore_database/src/go/stores/conn"
)

// Result is one entity's migration.
type Result struct {
	Entity   string
	Mode     MigrationMode
	Count    int // rows migrated
	Duration time.Duration
}

// Run runs migs in order with the connections and record path of cfg.
func Run(ctx context.Context, cfg conn.Config, migs []MigrationConfig) ([]Result, error) {
	results := make([]Result, 0, len(migs))
	for _, mc := range migs {
		mc.MappingPath = cfg.MappingPath()
		if cfg.Mongo != nil {
			mc.MongoDbName = cfg.Mongo.DBName
		}

		start := time.Now()
		count, err := MigrateData(mc, cfg)
		dur := time.Since(start)
		if err != nil {
			return results, fmt.Errorf("migrate %s: %w", mc.Entity, err)
		}
		results = append(results, Result{
			Entity: mc.Entity, Mode: mc.Mode, Count: count, Duration: dur,
		})
	}
	return results, nil
}
