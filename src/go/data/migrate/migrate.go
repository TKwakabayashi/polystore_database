package migrate

import (
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"

	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"

	"golang.org/x/sync/errgroup"
)

func MigrateData(config MigrationConfig, sconfig conn.Config) (int, error) {
	mappingDictionary, err := catalog.LoadRecord(config.MappingPath)
	if err != nil {
		return 0, err
	}

	srcKind, destKind, err := modeStores(config.Mode)
	if err != nil {
		return 0, err
	}
	count, err := ExecuteMigrationStream(config, sconfig, srcKind, destKind, mappingDictionary)
	if err != nil {
		return 0, err
	}
	return count, mappingDictionary.SaveMappingDictionary(config.MappingPath)
}

// ExecuteMigrationStream runs a migration and returns the number of migrated rows on success.
func ExecuteMigrationStream(config MigrationConfig, scfg conn.Config, srcKind, destKind kind.Kind, md *catalog.MappingDictionary) (int, error) {
	if !md.CheckDatastore(config.ObjType, config.Entity, config.Properties, srcKind.String()) {
		return 0, fmt.Errorf("data not found")
	}

	typeMap := make(map[string]string)
	for _, prop := range config.Properties {
		typeMap[prop], _ = md.GetPropertyDataType(config.ObjType, config.Entity, prop)
	}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	reg, err := conn.NewRegistryFor(mainCtx, scfg, srcKind, destKind)
	if err != nil {
		return 0, fmt.Errorf("registry init: %w", err)
	}
	defer reg.Close(mainCtx)

	src, err := dbs.MoverOf(srcKind)
	if err != nil {
		return 0, err
	}
	dest, err := dbs.MoverOf(destKind)
	if err != nil {
		return 0, err
	}
	target := config.target(typeMap)
	// Prepare the destination once, single-threaded (no concurrent ALTER).
	if err := dest.Prepare(mainCtx, reg, &target); err != nil {
		return 0, fmt.Errorf("prepare dest schema: %w", err)
	}

	// ===== Phase 1: copy (fetch -> upsert only) =====
	// Deleting is not part of this phase: the source is deleted only after verification and the record update
	// (phase 3), and only with DeleteSource. When any stage returns an error ctx.Done() closes and every stage stops.
	eg, ctx := errgroup.WithContext(mainCtx)

	fetchWorkers := src.ScanWorkers()
	upsertWorkers := 4

	// buffers of about 10k rows (tens of MB even at 1KB per row)
	fetchToUpsertCh := make(chan stores.Row, 10000)

	eg.Go(func() error {
		defer close(fetchToUpsertCh)
		fEg, fCtx := errgroup.WithContext(ctx)
		for i := 0; i < fetchWorkers; i++ {
			i := i
			fEg.Go(func() error {
				// each worker fetches only uuid shard (i, fetchWorkers), so the work is split without overlap
				return src.Scan(fCtx, reg, target, fetchToUpsertCh, i, fetchWorkers)
			})
		}
		return fEg.Wait()
	})

	eg.Go(func() error {
		uEg, uCtx := errgroup.WithContext(ctx)
		for i := 0; i < upsertWorkers; i++ {
			uEg.Go(func() error {
				return inBatches(uCtx, fetchToUpsertCh, func(rows []stores.Row) error { return dest.Upsert(uCtx, reg, target, rows) })
			})
		}
		return uEg.Wait()
	})

	if err := eg.Wait(); err != nil {
		return 0, err
	}
	if err := dest.Finish(mainCtx, reg, target); err != nil {
		return 0, err
	}

	// ===== Phase 2: verify (the gate for a successful copy) =====
	// Count source and destination independently and require destination >= source. Otherwise (rows missed by the
	// fetch or lost writes) stop without updating the record or deleting; the source is intact, so it can be rerun.
	migrated, err := VerifyMigration(mainCtx, config, srcKind, destKind, reg)
	if err != nil {
		return 0, fmt.Errorf("verify failed, mapping switch and source delete skipped: %w", err)
	}

	// ===== Record update (only after a successful verification) =====
	// Reads now go to the destination; deleting the source later does not affect queries.
	// A copy keeps the source, so the record lists both databases (a replica); a move replaces it.
	if config.DeleteSource {
		md.UpdateDatastore(config.ObjType, config.Entity, config.Properties, destKind.String())
	} else if err := md.AddDatastore(config.ObjType, config.Entity, config.Properties, destKind.String()); err != nil {
		return 0, err
	}

	// ===== Phase 3: delete (only with DeleteSource) =====
	if config.DeleteSource {
		if err := runDeletePhase(mainCtx, src, reg, target); err != nil {
			return 0, fmt.Errorf("delete phase failed (data already copied and mapping switched): %w", err)
		}
	}

	fmt.Println("migration done")
	return int(migrated), nil
}

// runDeletePhase deletes the migrated data from the source after copying, verification and the record update
// (only with DeleteSource): it scans the source again and deletes in batches, or all at once (BulkDeleter).
func runDeletePhase(ctx context.Context, src stores.Mover, reg *conn.Registry, t stores.Target) error {
	if b, ok := src.(stores.BulkDeleter); ok {
		return b.DeleteAll(ctx, reg, t)
	}
	const deleteWorkers = 2
	deleteCh := make(chan stores.Row, 10000)
	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(deleteCh)
		return src.Scan(gctx, reg, t, deleteCh, 0, 1)
	})
	for i := 0; i < deleteWorkers; i++ {
		eg.Go(func() error {
			return inBatches(gctx, deleteCh, func(rows []stores.Row) error { return src.Delete(gctx, reg, t, rows) })
		})
	}
	return eg.Wait()
}

// inBatches passes the rows of in to fn 2000 at a time until in closes.
func inBatches(ctx context.Context, in <-chan stores.Row, fn func([]stores.Row) error) error {
	const batchSize = 2000
	batch := make([]stores.Row, 0, batchSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case row, ok := <-in:
			if !ok {
				if len(batch) == 0 {
					return nil
				}
				return fn(batch)
			}
			batch = append(batch, row)
			if len(batch) >= batchSize {
				if err := fn(batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}
	}
}
