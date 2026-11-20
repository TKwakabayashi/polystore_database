// Package dbs lists the databases and dispatches compilation and execution to them by store kind.
package dbs

import (
	"context"
	"fmt"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/cassandra"
	"polystore_database/src/go/stores/clickhouse"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/internal/lower"
	"polystore_database/src/go/stores/leveldb"
	"polystore_database/src/go/stores/mongo"
	"polystore_database/src/go/stores/mysql"
	"polystore_database/src/go/stores/neo4j"
	"polystore_database/src/go/stores/rocksdb"
)

var all = map[kind.Kind]stores.DB{
	kind.Neo4j:      neo4j.DB{},
	kind.MySQL:      mysql.DB{},
	kind.Mongo:      mongo.DB{},
	kind.ClickHouse: clickhouse.DB{},
	kind.Cassandra:  cassandra.DB{},
	kind.LevelDB:    leveldb.DB,
	kind.RocksDB:    rocksdb.DB,
}

// Every database can take part in migrations.
var (
	_ stores.Mover       = neo4j.DB{}
	_ stores.Mover       = mysql.DB{}
	_ stores.Mover       = mongo.DB{}
	_ stores.Mover       = clickhouse.DB{}
	_ stores.Mover       = cassandra.DB{}
	_ stores.Mover       = leveldb.DB
	_ stores.BulkDeleter = clickhouse.DB{}
)

// MoverOf returns the migration side of the database of kind k.
func MoverOf(k kind.Kind) (stores.Mover, error) {
	m, ok := all[k].(stores.Mover)
	if !ok {
		return nil, fmt.Errorf("stores: %s cannot be migrated", k)
	}
	return m, nil
}

// Of returns the database of kind k (nil if none).
func Of(k kind.Kind) stores.DB { return all[k] }

// ShipEmulated reports whether shipped queries to k are emulated by the executor (the database
// cannot join a shipped table): it issues the query under the aggregation with the id boundary.
func ShipEmulated(k kind.Kind) bool {
	_, ok := all[k].(stores.Shipper)
	return !ok
}

// Compile turns a store query into the native query of its database.
func Compile(q *ir.StoreQuery, types stores.TypeFn) (*stores.Query, error) {
	db := all[q.Store]
	if db == nil {
		return nil, lower.Unsupported("store %s", q.Store)
	}
	if q.ShipMode == ir.ShipNone {
		out, err := db.Compile(q, types)
		if err != nil {
			return nil, err
		}
		return lower.Finish(q, out, types)
	}
	sh, ok := db.(stores.Shipper)
	if !ok {
		inner, _, ok := q.ShipInner()
		if !ok {
			return nil, lower.Unsupported("shipped %s query on %s", q.ShipMode, q.Store)
		}
		out, err := db.Compile(inner, types)
		if err != nil {
			return nil, err
		}
		out.ShipMode, out.ShipEmulated = q.ShipMode, true
		return lower.Finish(inner, out, types)
	}
	out, err := sh.CompileShipped(q, types)
	if err != nil {
		return nil, err
	}
	return lower.Finish(q, out, types)
}

// Session runs native queries on the connections of a conn.Registry.
type Session struct {
	reg *conn.Registry
}

// NewSession wraps connected stores.
func NewSession(reg *conn.Registry) *Session { return &Session{reg: reg} }

// Run collects every result row of q, decoded.
func (s *Session) Run(ctx context.Context, q *stores.Query, in []stores.Key) ([][]any, error) {
	return stores.Collect(func(emit func([]any) error) error { return s.Stream(ctx, q, in, emit) })
}

// Stream passes each result row of q, decoded, to emit.
func (s *Session) Stream(ctx context.Context, q *stores.Query, in []stores.Key, emit func(vals []any) error) error {
	db := all[q.Store]
	if db == nil {
		return fmt.Errorf("stores: no database for %s", q.Store)
	}
	return db.Stream(ctx, s.reg, q, in, func(raw []any) error {
		if err := Decode(q, raw); err != nil {
			return err
		}
		return emit(raw)
	})
}

// StreamReport is Stream that also measures the call (L3) and, with profile, asks the database for its
// own report (L4).
func (s *Session) StreamReport(ctx context.Context, q *stores.Query, in []stores.Key, emit func(vals []any) error, profile bool) (stores.Report, error) {
	var rep stores.Report
	db := all[q.Store]
	if db == nil {
		return rep, fmt.Errorf("stores: no database for %s", q.Store)
	}
	var inEmit time.Duration
	start := time.Now()
	err := db.Stream(ctx, s.reg, q, in, func(raw []any) error {
		t0 := time.Now()
		if rep.Rows == 0 {
			rep.FirstRow = t0.Sub(start)
		}
		rep.Rows++
		if err := Decode(q, raw); err != nil {
			return err
		}
		t1 := time.Now()
		rep.Decode += t1.Sub(t0)
		err := emit(raw)
		inEmit += time.Since(t1)
		return err
	})
	total := time.Since(start)
	if rep.Rows == 0 {
		rep.FirstRow = total
	} else {
		rep.Receive = max(total-rep.FirstRow-rep.Decode-inEmit, 0)
	}
	if err != nil {
		return rep, err
	}
	if p, ok := db.(stores.Profiler); ok && profile {
		if rep.Profile, err = p.Profile(ctx, s.reg, q, in); err != nil {
			return rep, fmt.Errorf("profile %s: %w", q.Store, err)
		}
	}
	return rep, nil
}

// RunShipped runs a shipped query with its input table and returns the decoded rows.
func (s *Session) RunShipped(ctx context.Context, q *stores.Query, sh stores.Shipment) ([][]any, error) {
	db, ok := all[q.Store].(stores.Shipper)
	if !ok {
		return nil, fmt.Errorf("stores: cannot ship rows to %s", q.Store)
	}
	rows, err := db.RunShipped(ctx, s.reg, q, sh)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if err := Decode(q, r); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

// Decode turns one raw result row into executor values in place: ids as id.UUID, segments as *eval.ListVal,
// other values normalized and parsed by their column type.
func Decode(q *stores.Query, raw []any) error {
	if len(raw) != len(q.Columns) {
		return fmt.Errorf("stores: %s returned %d columns, want %d", q.Store, len(raw), len(q.Columns))
	}
	for i, v := range raw {
		kind, typ := stores.KindValue, ""
		if i < len(q.ColKinds) {
			kind = q.ColKinds[i]
		}
		if i < len(q.ColTypes) {
			typ = q.ColTypes[i]
		}
		raw[i] = eval.Decode(v, kind, typ)
	}
	return nil
}

// Verbatim runs Cypher text on Neo4j as is.
func (s *Session) Verbatim(ctx context.Context, cypher string, params map[string]any) ([]map[string]any, error) {
	return neo4j.Verbatim(ctx, s.reg, cypher, params)
}
