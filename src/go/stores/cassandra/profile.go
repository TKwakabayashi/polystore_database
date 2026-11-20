package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// traceID keeps the id of a traced query.
type traceID struct{ id []byte }

func (t *traceID) Trace(id []byte) { t.id = id }

// Profile runs q with tracing on and reads the trace (session duration and events) from system_traces.
func (DB) Profile(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key) (*result.StoreProfile, error) {
	sess, ok := reg.Cassandra()
	if !ok {
		return nil, fmt.Errorf("executor: widecolumn (cassandra) not connected")
	}
	t := q.TextBody()
	args := append([]any{}, t.Args...)
	if t.InArg >= 0 {
		args[t.InArg] = stores.StoredIDs(in)
	}
	tr := &traceID{}
	iter := sess.Query(t.Text, args...).WithContext(ctx).Trace(tr).Iter()
	var rows int64
	for iter.MapScan(map[string]any{}) {
		rows++
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("cql %q: %w", t.Text, err)
	}
	p := &result.StoreProfile{Store: "cassandra", Query: t.Text, Rows: rows}
	if tr.id == nil {
		return p, nil
	}
	// The trace is written asynchronously; wait briefly for the session row.
	var micros int
	for i := 0; i < 20; i++ {
		err := sess.Query(`SELECT duration FROM system_traces.sessions WHERE session_id = ?`, gocql.UUID(tr.id)).
			WithContext(ctx).Scan(&micros)
		if err == nil && micros > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	p.Time = time.Duration(micros) * time.Microsecond
	var events []map[string]any
	it := sess.Query(`SELECT activity, source_elapsed, thread FROM system_traces.events WHERE session_id = ?`,
		gocql.UUID(tr.id)).WithContext(ctx).Iter()
	for {
		m := map[string]any{}
		if !it.MapScan(m) {
			break
		}
		events = append(events, m)
	}
	_ = it.Close()
	p.Raw = map[string]any{"duration_us": micros, "events": events}
	return p, nil
}
