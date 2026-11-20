package exec

import (
	"context"
	"fmt"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// streamBackend is a Streamer that answers native queries mechanically (to test streaming, early stops and errors):
// - no input (a leaf): leafRows rows; id column tid("s<i>"), value column int64(i)
// - with input: one row per input uuid; id column that uuid, value column "v-<uuid prefix>"
//
// With gate, a leaf waits after gateAt rows until gate closes. With failAt > 0, a leaf fails at row failAt.
type streamBackend struct {
	leafRows int
	gateAt   int
	gate     chan struct{}
	failAt   int
	emitted  atomic.Int64 // rows emitted by leaves (over all queries)
	stopped  atomic.Int64 // leaves stopped early by ctx cancellation or an emit error
}

func (b *streamBackend) Stream(ctx context.Context, q *stores.Query, in []Key, emit func([]any) error) error {
	row := func(i int, u id.UUID) []any {
		vals := make([]any, len(q.Columns))
		for j := range q.Columns {
			switch {
			case j < len(q.ColKinds) && q.ColKinds[j] == stores.KindIdent:
				vals[j] = u
			case in != nil:
				vals[j] = "v-" + u.String()[2:8]
			default:
				vals[j] = int64(i)
			}
		}
		return vals
	}
	if in != nil {
		for i, k := range in {
			if err := emit(row(i, k[0])); err != nil {
				return err
			}
		}
		return nil
	}
	for i := 0; i < b.leafRows; i++ {
		if b.gate != nil && i == b.gateAt {
			<-b.gate
		}
		if b.failAt > 0 && i == b.failAt {
			return fmt.Errorf("injected stream failure at row %d", i)
		}
		if err := ctx.Err(); err != nil {
			b.stopped.Add(1)
			return err
		}
		if err := emit(row(i, tid("N", fmt.Sprintf("s%d", i)))); err != nil {
			b.stopped.Add(1)
			return err
		}
		b.emitted.Add(1)
	}
	return nil
}

func (b *streamBackend) Run(ctx context.Context, q *stores.Query, in []Key) ([][]any, error) {
	return stores.Collect(func(emit func([]any) error) error { return b.Stream(ctx, q, in, emit) })
}

func (b *streamBackend) Verbatim(context.Context, string, map[string]any) ([]map[string]any, error) {
	return nil, fmt.Errorf("streamBackend: verbatim not supported")
}

// runOnly shows streamBackend as a Run-only backend (to compare with the path without Streamer).
type runOnly struct{ b *streamBackend }

func (r runOnly) Run(ctx context.Context, q *stores.Query, in []Key) ([][]any, error) {
	return r.b.Run(ctx, q, in)
}

func (r runOnly) Verbatim(ctx context.Context, c string, p map[string]any) ([]map[string]any, error) {
	return r.b.Verbatim(ctx, c, p)
}

var allModels = []Model{ModelBulk, ModelVolcano, ModelVectorized,
	ModelStream, ModelVecStream}

// leafPlan is a plan with a single leaf in the Neo4j placement (MATCH (p:Person) RETURN p.id).
func leafPlan(t *testing.T) (physical.Node, *physical.Placement) {
	t.Helper()
	p, cat, err := Plan("MATCH (p:Person) RETURN p.id", nil, loadMapping(t), tPlan())
	if err != nil {
		t.Fatal(err)
	}
	if q, ok := p.(*physical.StoreQuery); !ok || !q.Leaf() {
		t.Fatalf("want a single leaf StoreQuery, got\n%s", RenderPlan(p, cat))
	}
	return p, cat
}

func withVectorSize(t *testing.T, n int) {
	prev := vectorSize
	vectorSize = n
	t.Cleanup(func() { vectorSize = prev })
}

// A streamable leaf passes the first batch downstream before the store returns its last row (transfer overlaps work).
func TestStreamLeafPipelines(t *testing.T) {
	t.Parallel()
	p, cat := leafPlan(t)
	be := &streamBackend{leafRows: 10, gateAt: 4, gate: make(chan struct{})}
	e := New(context.Background(), be, cat.TypeOf, tExec(ModelBulk))
	e.reset(p)
	it, err := e.iter(context.Background(), p, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer it.close()

	got := make(chan []Row, 1)
	go func() {
		b, _ := it.next()
		got <- b
	}()
	select {
	case b := <-got:
		if len(b) != 2 {
			t.Fatalf("first batch = %+v, want 2 rows", b)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first batch did not arrive while the store was still sending (results are buffered)")
	}
	close(be.gate)
	n := 2
	for {
		b, err := it.next()
		if err != nil {
			t.Fatal(err)
		}
		if b == nil {
			break
		}
		n += len(b)
	}
	if n != 10 {
		t.Errorf("rows = %d, want 10", n)
	}
	st := e.Steps()
	if len(st) != 1 || st[0].OutRows != 10 || st[0].RoundTrips != 1 {
		t.Errorf("steps = %+v", st)
	}
}

// A coordinator LIMIT stops streamed leaves once it has enough rows (all but bulk); every model returns LIMIT rows.
func TestLimitStopsStreamingLeaf(t *testing.T) {
	leaf, cat := leafPlan(t)
	withVectorSize(t, 4)
	const total = 200000
	p := &physical.Coord{Op: &ir.Limit{Count: 3}, Inputs: []physical.Node{leaf}}
	for _, m := range allModels {
		be := &streamBackend{leafRows: total}
		rows, err := New(context.Background(), be, cat.TypeOf, tExec(m)).ExecuteModel(p)
		if err != nil {
			t.Fatalf("%s: %v", m, err)
		}
		if len(rows) != 3 {
			t.Errorf("%s: rows = %d, want 3", m, len(rows))
		}
		if m == ModelBulk {
			continue // the model that materializes everything
		}
		if be.emitted.Load() >= total || be.stopped.Load() != 1 {
			t.Errorf("%s: leaf was not stopped (emitted %d of %d, stopped %d)", m, be.emitted.Load(), total, be.stopped.Load())
		}
	}
}

// An error in the middle of a stream is an error in every model (no partial results, no hang).
func TestStreamErrorPropagates(t *testing.T) {
	leaf, cat := leafPlan(t)
	withVectorSize(t, 2)
	for _, m := range allModels {
		done := make(chan error, 1)
		go func() {
			_, err := New(context.Background(), &streamBackend{leafRows: 100, failAt: 5}, cat.TypeOf, tExec(m)).ExecuteModel(leaf)
			done <- err
		}()
		select {
		case err := <-done:
			if err == nil || !strings.Contains(err.Error(), "injected") {
				t.Errorf("%s: want injected error, got %v", m, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: did not return after a stream failure", m)
		}
	}
}

// A caller's cancellation is an error even while streaming (no partial results).
func TestStreamCallerCancel(t *testing.T) {
	leaf, cat := leafPlan(t)
	withVectorSize(t, 2)
	for _, m := range allModels {
		ctx, cancel := context.WithCancel(context.Background())
		be := &streamBackend{leafRows: 100, gateAt: 3, gate: make(chan struct{})}
		done := make(chan error, 1)
		go func() {
			_, err := New(ctx, be, cat.TypeOf, tExec(m)).ExecuteModel(leaf)
			done <- err
		}()
		cancel()
		close(be.gate)
		select {
		case err := <-done:
			if err == nil {
				t.Errorf("%s: cancelled run returned no error", m)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: did not return after cancel", m)
		}
	}
}

// Stream and Run give the same results and round trips (a leaf plus a fetch from another store, through the join to input rows).
func TestStreamMatchesRun(t *testing.T) {
	withVectorSize(t, 3)
	md := placementsFor(t)["mixed"]
	p, cat, err := Plan("MATCH (p:Person) RETURN p.id, p.firstName", nil, md, tPlan())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(RenderPlan(p, cat), "mysql") {
		t.Fatalf("plan does not fetch from mysql:\n%s", RenderPlan(p, cat))
	}
	for _, m := range allModels {
		sb := &streamBackend{leafRows: 11}
		es := New(context.Background(), sb, cat.TypeOf, tExec(ModelBulk))
		got, err := es.model(m).ExecuteModel(p)
		if err != nil {
			t.Fatalf("%s stream: %v", m, err)
		}
		er := New(context.Background(), runOnly{&streamBackend{leafRows: 11}}, cat.TypeOf, tExec(ModelBulk))
		want, err := er.model(m).ExecuteModel(p)
		if err != nil {
			t.Fatalf("%s run: %v", m, err)
		}
		g, w := canon(OutputRows(p, got)), canon(OutputRows(p, want))
		if strings.Join(g, "\n") != strings.Join(w, "\n") || len(g) != 11 {
			t.Errorf("%s: stream and run differ\n--- run ---\n%s\n--- stream ---\n%s", m, strings.Join(w, "\n"), strings.Join(g, "\n"))
		}
		if es.RoundTrips() != er.RoundTrips() {
			t.Errorf("%s: round trips stream=%d run=%d", m, es.RoundTrips(), er.RoundTrips())
		}
		for _, r := range OutputRows(p, got) {
			if s, _ := r["p.firstName"].(string); !strings.HasPrefix(s, "v-") {
				t.Errorf("%s: firstName not joined: %v", m, r)
				break
			}
		}
	}
}

// Benchmarks of the executor's CPU and memory (no DB; streamBackend answers mechanically).
// go test ./query/exec/ -run '^$' -bench . -benchmem

const benchRows = 100000

// Boundary join: a Neo4j leaf (100k rows) -> a MySQL value fetch joined to the input rows.
func benchBoundary(b *testing.B, model Model) {
	md := placementsForBench(b)
	p, cat, err := Plan("MATCH (p:Person) RETURN p.id, p.firstName", nil, md, tPlan())
	if err != nil {
		b.Fatal(err)
	}
	prev := vectorSize
	vectorSize = 1024
	defer func() { vectorSize = prev }()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := New(context.Background(), &streamBackend{leafRows: benchRows}, cat.TypeOf, tExec(model)).ExecuteModel(p)
		if err != nil {
			b.Fatal(err)
		}
		if len(rows) != benchRows {
			b.Fatalf("rows = %d", len(rows))
		}
	}
}

func BenchmarkBoundaryBulk(b *testing.B)       { benchBoundary(b, ModelBulk) }
func BenchmarkBoundaryVectorized(b *testing.B) { benchBoundary(b, ModelVectorized) }
func BenchmarkBoundaryVecStream(b *testing.B)  { benchBoundary(b, ModelVecStream) }

// placementsForBench puts Person.firstName in MySQL (as the mixed placement of placementsFor).
func placementsForBench(b *testing.B) *catalog.MappingDictionary {
	b.Helper()
	md, err := catalog.LoadMappingDictionary("../../../../datastore/datasets/ldbc_snb_sf1/mapping.json")
	if err != nil {
		b.Fatal(err)
	}
	md.Entities["Person"].Properties["firstName"].Databases = []string{"relational"}
	return md
}

// Coordinator operators: a join on one uuid column, a distinct over all columns.
func benchRowsWithIDs(n, distinct int) []Row {
	rows := make([]Row, n)
	for i := range rows {
		rows[i] = Row{"a": tid("N", fmt.Sprintf("k%d", i%distinct)), "a.v": int64(i), "a.s": "x"}
	}
	return rows
}

func BenchmarkHashJoinUUID(b *testing.B) {
	left := benchRowsWithIDs(benchRows, benchRows)
	right := make([]Row, benchRows)
	for i := range right {
		right[i] = Row{"a": left[i]["a"], "b.v": int64(i)}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if out := hashJoin(left, right, []string{"a"}); len(out) != benchRows {
			b.Fatalf("rows = %d", len(out))
		}
	}
}

func BenchmarkDistinctAllCols(b *testing.B) {
	rows := benchRowsWithIDs(benchRows, benchRows/2)
	for i := range rows {
		rows[i]["a.v"] = int64(i % (benchRows / 2))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := runCoord(&ir.Distinct{}, [][]Row{rows})
		if err != nil || len(out) != benchRows/2 {
			b.Fatalf("rows = %d err = %v", len(out), err)
		}
	}
}

func BenchmarkDecodeDatetime(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, ok := eval.Decode("2012-02-22 14:05:58.785000", stores.KindValue, "datetime").(string); ok {
			b.Fatal("not parsed")
		}
	}
}
