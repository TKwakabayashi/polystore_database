package setup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores/conn"
)

// AssignUUIDs gives uuids to Neo4j nodes and relationships without one, from dataset's allocator (id.Allocator),
// in bulk. Values are decided in Go and written in batches with their elementId (not by APOC); the storage form
// follows id.Stored.
//
// With nothing to assign it does nothing (and takes no lease). Otherwise it takes a bulk lease, which fails when
// the dataset's counter was already used (a rerun of a crashed assignment fails too; only after reloading data
// without uuids reset next to 0 in leasePath).
func AssignUUIDs(ctx context.Context, cfg conn.Config, leasePath, dataset string) error {
	if cfg.Neo4j == nil {
		return fmt.Errorf("neo4j setting was not found")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	defer driver.Close(ctx)

	kinds := []struct {
		name        string
		probe, read string
		write       string
	}{
		{"nodes",
			"MATCH (n) WHERE n.uuid IS NULL RETURN 1 LIMIT 1",
			"MATCH (n) WHERE n.uuid IS NULL RETURN elementId(n) AS e",
			"UNWIND $rows AS row MATCH (n) WHERE elementId(n) = row.e SET n.uuid = row.u"},
		{"relationships",
			"MATCH ()-[r]->() WHERE r.uuid IS NULL RETURN 1 LIMIT 1",
			"MATCH ()-[r]->() WHERE r.uuid IS NULL RETURN elementId(r) AS e",
			"UNWIND $rows AS row MATCH ()-[r]->() WHERE elementId(r) = row.e SET r.uuid = row.u"},
	}
	var missing []string
	for _, k := range kinds {
		found, err := anyRow(ctx, driver, k.probe)
		if err != nil {
			return fmt.Errorf("checking %s without uuid: %w", k.name, err)
		}
		if found {
			missing = append(missing, k.name)
		}
	}
	if len(missing) == 0 {
		log.Printf("every node and relationship has a uuid (nothing to assign)")
		return nil
	}
	if dataset == "" {
		return fmt.Errorf("%v without uuid exist; name the dataset to assign with -dataset", missing)
	}
	alloc, err := conn.LeaseUUIDAllocatorForBulk(leasePath, dataset)
	if err != nil {
		return err
	}
	log.Printf("assigning uuids of dataset %s from lease %d", dataset, alloc.Lease())
	for _, k := range kinds {
		n, err := assignAll(ctx, driver, alloc, k.read, k.write)
		if err != nil {
			return fmt.Errorf("assigning uuids to %s (%d done): %w", k.name, n, err)
		}
		log.Printf("%s: %d assigned", k.name, n)
	}
	return nil
}

// anyRow reports whether query returns at least one row.
func anyRow(ctx context.Context, driver neo4j.DriverWithContext, query string) (bool, error) {
	s := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer s.Close(ctx)
	res, err := s.Run(ctx, query, nil)
	if err != nil {
		return false, err
	}
	found := res.Next(ctx)
	if err := res.Err(); err != nil {
		return false, err
	}
	return found, nil
}

// assignAll streams read's results (e: elementId) and writes uuids from alloc in batches with write.
func assignAll(ctx context.Context, driver neo4j.DriverWithContext, alloc *id.Allocator, read, write string) (assigned int, err error) {
	rs := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer rs.Close(ctx)
	ws := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer ws.Close(ctx)

	res, err := rs.Run(ctx, read, nil)
	if err != nil {
		return 0, err
	}
	var rows []map[string]any
	flush := func() error {
		if len(rows) == 0 {
			return nil
		}
		if _, err := ws.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			r, err := tx.Run(ctx, write, map[string]any{"rows": rows})
			if err != nil {
				return nil, err
			}
			return r.Consume(ctx)
		}); err != nil {
			return err
		}
		assigned += len(rows)
		rows = rows[:0]
		return nil
	}
	start := time.Now()
	for res.Next(ctx) {
		e, _ := res.Record().Get("e")
		u, err := alloc.Next()
		if err != nil {
			return assigned, err
		}
		rows = append(rows, map[string]any{"e": e, "u": id.Stored(u)})
		if len(rows) >= uuidBatch {
			if err := flush(); err != nil {
				return assigned, err
			}
			if assigned%(uuidBatch*50) == 0 {
				log.Printf("  … %d (%v)", assigned, time.Since(start).Round(time.Second))
			}
		}
	}
	if err := res.Err(); err != nil {
		return assigned, err
	}
	return assigned, flush()
}
