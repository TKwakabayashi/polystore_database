package neo4j

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func fetchGraph(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var label, query string
	if cfg.Obj == catalog.ObjRelationship {
		label = "r"
		// the directed -> reads each relationship once (undirected - would fetch it twice)
		query = fmt.Sprintf("MATCH ()-[%s:%s]->() ", label, cfg.Entity)
	} else {
		label = "n"
		query = fmt.Sprintf("MATCH (%s:%s) ", label, cfg.Entity)
	}

	propList := []string{"uuid: " + label + ".uuid"}
	for _, p := range cfg.Props {
		if p != "uuid" {
			propList = append(propList, fmt.Sprintf("%s: %s.%s", p, label, p))
		}
	}
	query += fmt.Sprintf("RETURN { %s } AS data", strings.Join(propList, ", "))

	res, err := session.Run(ctx, query, nil)
	if err != nil {
		return err
	}

	for res.Next(ctx) {
		record := res.Record()
		dataInterface, _ := record.Get("data")
		dataMap, ok := dataInterface.(map[string]interface{})
		if !ok {
			// data object error
			continue
		}

		uuid := id.FromAny(dataMap["uuid"])
		if uuid.Empty() {
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Props))
		for _, p := range cfg.Props {
			// ParseToNative turns Neo4j's float64 into int32/int64
			nativeVal, err := codec.ParseToNative(dataMap[p], typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [graph] UUID %s, Prop %s: %w", uuid.String(), p, err)
			}
			payload[p] = nativeVal
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: uuid, Values: payload}:
		}
	}

	return nil
}

func upsertGraph(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	unwindParams := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		cleanPayload := make(map[string]interface{})
		for _, p := range cfg.Props {
			if v, ok := row.Values[p]; ok {
				cleanPayload[p] = graphValue(v, typeMap[p])
			}
		}
		// boundary conversion: id.UUID -> stored value (string or []byte; neo4j packstream cannot take named types)
		unwindParams[i] = map[string]interface{}{"uuid": id.Stored(row.UUID), "payload": cleanPayload}
	}

	var query string
	if cfg.Obj == catalog.ObjRelationship {
		query = fmt.Sprintf("UNWIND $batch AS row MATCH ()-[r:%s {uuid: row.uuid}]-() SET r += row.payload", cfg.Entity)
	} else {
		query = fmt.Sprintf("UNWIND $batch AS row MERGE (n:%s {uuid: row.uuid}) SET n += row.payload", cfg.Entity)
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, map[string]interface{}{"batch": unwindParams})
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	if err != nil {
		return fmt.Errorf("neo4j bulk upsert failed: %w", err)
	}

	return nil
}

func deleteGraph(ctx context.Context, cfg stores.Target, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	uuids := make([]interface{}, len(rows))
	for i, r := range rows {
		uuids[i] = id.Stored(r.UUID)
	}

	var query string
	if cfg.Obj == catalog.ObjRelationship {
		// delete relationship properties (with APOC)
		query = fmt.Sprintf(`
					MATCH ()-[t:%s]-() WHERE t.uuid IN $uuids 
					CALL apoc.create.setRelProperties(t, $props, [x IN $props | null]) 
					YIELD rel RETURN count(*)`, cfg.Entity)
	} else {
		// delete node properties
		query = fmt.Sprintf(`
					MATCH (t:%s) WHERE t.uuid IN $uuids 
					CALL apoc.create.removeProperties(t, $props) 
					YIELD node RETURN count(*)`, cfg.Entity)
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, map[string]interface{}{"uuids": uuids, "props": cfg.Props})
	})
	if err != nil {
		return fmt.Errorf("neo4j properties removal failed: %w", err)
	}

	return nil
}

func countGraph(ctx context.Context, cfg stores.Target, reg *conn.Registry) (int64, error) {
	drv, ok := reg.Neo4j()
	if !ok {
		return 0, fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	if cfg.Obj == catalog.ObjRelationship {
		// the directed -> counts each relationship once (undirected - would count it twice)
		query = fmt.Sprintf("MATCH ()-[r:%s]->() RETURN count(r) AS c", cfg.Entity)
	} else {
		query = fmt.Sprintf("MATCH (n:%s) RETURN count(n) AS c", cfg.Entity)
	}

	res, err := session.Run(ctx, query, nil)
	if err != nil || !res.Next(ctx) {
		return 0, fmt.Errorf("neo4j: count %s: %v", cfg.Entity, err)
	}
	c, _ := res.Record().Get("c")
	n, ok := c.(int64)
	if !ok {
		return 0, fmt.Errorf("neo4j: the count is not an integer")
	}
	return n, nil
}

// Prepare creates the uuid uniqueness constraint that MERGE / MATCH by uuid need (idempotent). A schema
// conflict only warns: without the constraint the migration is slower but proceeds.
func (DB) Prepare(ctx context.Context, reg *conn.Registry, t *stores.Target) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("neo4j not connected")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	name := strings.ToLower(t.Entity)
	var stmts []string
	if t.Obj == catalog.ObjRelationship {
		// one constraint per relationship type; an old _uuid_idx cannot coexist, so it is dropped first
		stmts = []string{
			fmt.Sprintf("DROP INDEX %s_uuid_idx IF EXISTS", name),
			fmt.Sprintf("CREATE CONSTRAINT %s_uuid_unique IF NOT EXISTS FOR ()-[r:%s]-() REQUIRE r.uuid IS UNIQUE", name, t.Entity),
		}
	} else {
		stmts = []string{fmt.Sprintf("CREATE CONSTRAINT %s_uuid_unique IF NOT EXISTS FOR (n:%s) REQUIRE n.uuid IS UNIQUE", name, t.Entity)}
	}
	for _, stmt := range stmts {
		if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			res, err := tx.Run(ctx, stmt, nil)
			if err != nil {
				return nil, err
			}
			return res.Consume(ctx)
		}); err != nil {
			fmt.Printf("neo4j: uuid constraint for %s skipped: %v\n", t.Entity, err)
			break
		}
	}
	return nil
}

// ScanWorkers is 1: reads are dominated by round trips and the uuid index.
func (DB) ScanWorkers() int { return 1 }
func (DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, _, _ int) error {
	return fetchGraph(ctx, t, reg, t.Types, out)
}
func (DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertGraph(ctx, t, reg, t.Types, rows)
}
func (DB) Finish(context.Context, *conn.Registry, stores.Target) error { return nil }
func (DB) Delete(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return deleteGraph(ctx, t, reg, t.Types, rows)
}
func (DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countGraph(ctx, t, reg)
}

// graphValue is a migrated value as Neo4j stores it: dates as Date and datetimes as DateTime (UTC), so the
// property keeps its type (a string would no longer compare with date() / datetime()).
func graphValue(v any, typ string) any {
	if t, ok := v.(time.Time); ok {
		if strings.EqualFold(typ, "date") {
			return neo4j.DateOf(t.UTC())
		}
		return t.UTC()
	}
	out, _ := codec.PrepareForDB(v, typ, "graph")
	return out
}
