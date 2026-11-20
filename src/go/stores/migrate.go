package stores

import (
	"context"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores/conn"
)

// Target is what a migration reads and writes: some properties of an entity or relationship type.
type Target struct {
	Obj     catalog.ObjectType
	Entity  string
	Props   []string
	Types   map[string]string // property -> type name (mapping)
	MongoDB string            // MongoDB database name
	// Stage is ClickHouse's staging table, set by Prepare when the destination table already has rows.
	Stage string
}

// Row is one migrated row: an id and its property values.
type Row struct {
	UUID   id.UUID
	Values map[string]any
}

// Scanner reads a target's rows for migrations (data/migrate).
type Scanner interface {
	// Scan sends the rows of uuid shard `shard` of `shards` to out (shards neither overlap nor miss rows).
	Scan(ctx context.Context, reg *conn.Registry, t Target, out chan<- Row, shard, shards int) error
	// ScanWorkers is the number of shards worth reading in parallel.
	ScanWorkers() int
}

// Writer writes migrated rows and removes them from a source.
type Writer interface {
	// Prepare makes the destination ready once before writing (tables, columns, indexes, the uuid type).
	Prepare(ctx context.Context, reg *conn.Registry, t *Target) error
	Upsert(ctx context.Context, reg *conn.Registry, t Target, rows []Row) error
	// Finish completes the writes (ClickHouse moves the staging table in).
	Finish(ctx context.Context, reg *conn.Registry, t Target) error
	// Delete removes the target's properties of rows (the source of a move).
	Delete(ctx context.Context, reg *conn.Registry, t Target, rows []Row) error
}

// BulkDeleter is a Writer that removes the whole target at once instead of row by row.
type BulkDeleter interface {
	DeleteAll(ctx context.Context, reg *conn.Registry, t Target) error
}

// Admin counts a target's rows (uuids), to verify migrations.
type Admin interface {
	Count(ctx context.Context, reg *conn.Registry, t Target) (int64, error)
}

// Mover is a database that migrations read from and write to.
type Mover interface {
	Scanner
	Writer
	Admin
}
