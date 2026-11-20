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

// uuidBatch is the number of values rewritten per write transaction.
const uuidBatch = 10000

// ConvertUUIDFormat rewrites every Neo4j node and relationship uuid to the storage form id.CurrentStorage()
// (text / 16 bytes). Values map one to one through id's conversions, so the uuid uniqueness constraints hold
// throughout (a string never equals bytes). Values already in the target form are skipped, so a rerun resumes.
// Legacy values (kinded "N-..." / 17 bytes) are converted with the kind dropped (one to one within a kind; node
// and relationship values collide only if old random v4s happened to be equal).
//
// The uuid columns of the other stores (MySQL / ClickHouse / Cassandra column types, MongoDB values, LevelDB /
// RocksDB keys) are not rewritten here: after switching, clear them with setup and migrate again (migrations make
// column types match the storage form).
func ConvertUUIDFormat(ctx context.Context, cfg conn.Config) error {
	if cfg.Neo4j == nil {
		return fmt.Errorf("neo4j setting was not found")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	defer driver.Close(ctx)

	target := "string"
	if id.Binary() {
		target = "16 byte"
	}
	log.Printf("rewriting the uuid storage form to %s", target)
	for _, k := range []struct {
		name, read, write string
	}{
		{"nodes", "MATCH (n) RETURN elementId(n) AS e, n.uuid AS u",
			"UNWIND $rows AS row MATCH (n) WHERE elementId(n) = row.e SET n.uuid = row.u"},
		{"relationships", "MATCH ()-[r]->() RETURN elementId(r) AS e, r.uuid AS u",
			"UNWIND $rows AS row MATCH ()-[r]->() WHERE elementId(r) = row.e SET r.uuid = row.u"},
	} {
		n, skipped, err := convertAll(ctx, driver, k.read, k.write)
		if err != nil {
			return fmt.Errorf("rewriting uuids of %s (%d done; a rerun continues): %w", k.name, n, err)
		}
		log.Printf("%s: %d rewritten, %d already in the target form", k.name, n, skipped)
	}
	return nil
}

// convertAll streams read's results (e: elementId, u: uuid) and rewrites values not in the target form in batches with write.
func convertAll(ctx context.Context, driver neo4j.DriverWithContext, read, write string) (converted, skipped int, err error) {
	rs := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer rs.Close(ctx)
	ws := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer ws.Close(ctx)

	res, err := rs.Run(ctx, read, nil)
	if err != nil {
		return 0, 0, err
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
		converted += len(rows)
		rows = rows[:0]
		return nil
	}
	start := time.Now()
	for res.Next(ctx) {
		rec := res.Record()
		e, _ := rec.Get("e")
		raw, _ := rec.Get("u")
		u := id.FromAny(raw)
		if u.Empty() {
			return converted, skipped, fmt.Errorf("element %v has a uuid that cannot be read (value %v)", e, raw)
		}
		if inTargetFormat(raw) {
			skipped++
			continue
		}
		rows = append(rows, map[string]any{"e": e, "u": id.Stored(u)})
		if len(rows) >= uuidBatch {
			if err := flush(); err != nil {
				return converted, skipped, err
			}
			if converted%(uuidBatch*50) == 0 {
				log.Printf("  … %d (%v)", converted, time.Since(start).Round(time.Second))
			}
		}
	}
	if err := res.Err(); err != nil {
		return converted, skipped, err
	}
	return converted, skipped, flush()
}

// inTargetFormat reports whether a uuid read from Neo4j is already in the id.CurrentStorage() form (legacy kinded values are not).
func inTargetFormat(v any) bool {
	switch x := v.(type) {
	case []byte:
		return id.CurrentStorage() == id.StorageBinary && len(x) == id.Size
	case string:
		return id.CurrentStorage() != id.StorageBinary && len(x) == id.TextSize
	}
	return false
}
