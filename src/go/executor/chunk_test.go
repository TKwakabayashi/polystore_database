package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"polystore_database/src/go/id"
	"polystore_database/src/go/native"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// echoBackend は入力 uuid をそのまま 1 行ずつ返す（順序と往復数の検証用）。fail を含むチャンクは失敗する。
type echoBackend struct {
	delay time.Duration
	fail  id.UUID
}

func (b *echoBackend) Run(_ context.Context, _ *native.Query, in []Key) ([][]any, error) {
	time.Sleep(b.delay)
	var out [][]any
	for _, k := range in {
		if k[0] == b.fail {
			return nil, fmt.Errorf("chunk with %s failed", k[0])
		}
		out = append(out, []any{k[0].String()})
	}
	return out, nil
}

func (b *echoBackend) Verbatim(context.Context, string, map[string]any) ([]map[string]any, error) {
	return nil, nil
}

func TestRunChunksInParallelAndInOrder(t *testing.T) {
	prevChunk, prevPar := settings.MaterializeChunkSize, settings.StoreFetchParallelism
	settings.MaterializeChunkSize, settings.StoreFetchParallelism = 3, 4
	defer func() { settings.MaterializeChunkSize, settings.StoreFetchParallelism = prevChunk, prevPar }()

	var keys []Key
	for i := 0; i < 10; i++ {
		keys = append(keys, Key{tid(id.KindNode, fmt.Sprintf("u%02d", i))})
	}
	q := &native.Query{Store: store.Columnar, Columns: []string{"a"}}
	e := New(context.Background(), &echoBackend{delay: 50 * time.Millisecond}, nil)
	e.reset(nil)
	start := time.Now()
	rows, err := e.run(q, keys)
	if err != nil {
		t.Fatal(err)
	}
	if took := time.Since(start); took > 150*time.Millisecond {
		t.Errorf("4 chunks with parallelism 4 took %v (not parallel?)", took)
	}
	if e.RoundTrips() != 4 {
		t.Errorf("round trips = %d, want 4", e.RoundTrips())
	}
	for i, r := range rows {
		if r[0] != tid(id.KindNode, fmt.Sprintf("u%02d", i)).String() {
			t.Fatalf("row %d = %v (order lost)", i, r[0])
		}
	}

	e = New(context.Background(), &echoBackend{fail: tid(id.KindNode, "u07")}, nil)
	e.reset(nil)
	if _, err := e.run(q, keys); err == nil {
		t.Error("a failing chunk must fail the whole fetch")
	}
}
