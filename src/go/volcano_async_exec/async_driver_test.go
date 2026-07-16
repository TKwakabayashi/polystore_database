package volcano_async_executor

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

func goroutineCount() int { return runtime.NumGoroutine() }

// DB を必要とせず asyncDriver の並行性だけを検証する。
// 実 DB を用いた同期版との結果一致・往復回数の検証は cmd/volcanoasyncbench で行う。
//
// go test -race ./volcano_async_exec/... で data race も併せて検出する。

// fakeSource は scanIterator の代わり。batches 個のバッチを 1 行ずつ払い出す。
type fakeSource struct {
	batches int
	pos     int
	opened  int32
	closed  int32
}

func (f *fakeSource) Open(ctx context.Context) error  { atomic.AddInt32(&f.opened, 1); return nil }
func (f *fakeSource) Close(ctx context.Context) error { atomic.AddInt32(&f.closed, 1); return nil }
func (f *fakeSource) Next(ctx context.Context) (*Batch, error) {
	if f.pos >= f.batches {
		return nil, nil
	}
	b := newBatch(1, 1)
	b.appendRow([]string{fmt.Sprintf("id-%d", f.pos)})
	f.pos++
	return b, nil
}

// testProcessor は DB 接続を張らずに asyncDriver を動かすための最小 Processor。
func testProcessor(async AsyncMode, policy ExecPolicy) *Processor {
	p := &Processor{
		ctx:         context.Background(),
		mode:        ModeVolcano,
		vectorWidth: 1,
		asyncMode:   async,
		policy:      policy,
		metrics:     make(map[int]*Metrics),
	}
	if policy.Mode == ExecDynamic {
		p.sem = semaphore.NewWeighted(int64(policy.globalMax()))
	}
	return p
}

func drain(t *testing.T, it Iterator) ([]string, error) {
	t.Helper()
	var got []string
	for {
		b, err := it.Next(context.Background())
		if err != nil {
			return got, err
		}
		if b == nil {
			return got, nil
		}
		for i := 0; i < b.n; i++ {
			got = append(got, b.get(i, 0))
		}
	}
}

// Exchange / Prefetch のどちらでも、全行がちょうど 1 回ずつ通過すること。
func TestAsyncDriverPassesEveryRowOnce(t *testing.T) {
	const n = 200
	for _, tc := range []struct {
		name   string
		async  AsyncMode
		policy ExecPolicy
	}{
		{"exchange/dynamic", AsyncExchange, ExecPolicy{Mode: ExecDynamic, GlobalMaxConcurrency: 8}},
		{"exchange/fixed", AsyncExchange, ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 4}}},
		{"prefetch", AsyncPrefetch, ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 4}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := testProcessor(tc.async, tc.policy)
			src := &fakeSource{batches: n}
			d := newAsyncDriver(p, src, OpExpand, 1, func(in *Batch) (*Batch, error) {
				return in, nil // 素通し
			})
			if err := d.Open(context.Background()); err != nil {
				t.Fatalf("Open: %v", err)
			}
			got, err := drain(t, d)
			if err != nil {
				t.Fatalf("drain: %v", err)
			}
			if err := d.Close(context.Background()); err != nil {
				t.Fatalf("Close: %v", err)
			}

			if len(got) != n {
				t.Fatalf("行数 = %d, want %d", len(got), n)
			}
			seen := make(map[string]int, n)
			for _, v := range got {
				seen[v]++
			}
			for i := 0; i < n; i++ {
				k := fmt.Sprintf("id-%d", i)
				if seen[k] != 1 {
					t.Errorf("%s の出現回数 = %d, want 1", k, seen[k])
				}
			}
			if atomic.LoadInt32(&src.closed) != 1 {
				t.Errorf("child.Close 呼び出し回数 = %d, want 1", src.closed)
			}
		})
	}
}

// Prefetch(workers=1) は順序を保存する。Exchange は保証しない（ここでは検証しない）。
func TestPrefetchPreservesOrder(t *testing.T) {
	const n = 50
	p := testProcessor(AsyncPrefetch, ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 4}})
	src := &fakeSource{batches: n}
	d := newAsyncDriver(p, src, OpExpand, 1, func(in *Batch) (*Batch, error) { return in, nil })
	if err := d.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	got, err := drain(t, d)
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	d.Close(context.Background())

	for i, v := range got {
		if want := fmt.Sprintf("id-%d", i); v != want {
			t.Fatalf("順序が崩れた: got[%d] = %q, want %q", i, v, want)
		}
	}
}

// process のエラーが Next へ伝播し、パイプラインが畳まれること。
func TestAsyncDriverPropagatesProcessError(t *testing.T) {
	wantErr := errors.New("boom")
	p := testProcessor(AsyncExchange, ExecPolicy{Mode: ExecDynamic, GlobalMaxConcurrency: 4})
	src := &fakeSource{batches: 100}
	d := newAsyncDriver(p, src, OpExpand, 1, func(in *Batch) (*Batch, error) {
		return nil, wantErr
	})
	if err := d.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, err := drain(t, d)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if err := d.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// 空バッチ (n==0) は EOF と紛れないよう下流へ流さないこと。
func TestAsyncDriverSkipsEmptyBatches(t *testing.T) {
	const n = 60
	p := testProcessor(AsyncExchange, ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 4}})
	src := &fakeSource{batches: n}
	// 偶数番だけ通し、奇数番は空バッチにする（Filter が全滅したケース）。
	d := newAsyncDriver(p, src, OpFilter, 1, func(in *Batch) (*Batch, error) {
		var idx int
		fmt.Sscanf(in.get(0, 0), "id-%d", &idx)
		if idx%2 == 1 {
			return newBatch(1, 0), nil // 空
		}
		return in, nil
	})
	if err := d.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	got, err := drain(t, d)
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	d.Close(context.Background())

	if len(got) != n/2 {
		t.Fatalf("行数 = %d, want %d（空バッチで打ち切られていないか）", len(got), n/2)
	}
}

// ExecDynamic のセマフォがシステム全体の同時 DB 数を上限で抑えること。
func TestExecDynamicRespectsGlobalMax(t *testing.T) {
	const limit = 3
	p := testProcessor(AsyncExchange, ExecPolicy{Mode: ExecDynamic, GlobalMaxConcurrency: limit})
	src := &fakeSource{batches: 100}

	var mu sync.Mutex
	cur, peak := 0, 0
	d := newAsyncDriver(p, src, OpExpand, 1, func(in *Batch) (*Batch, error) {
		mu.Lock()
		cur++
		if cur > peak {
			peak = cur
		}
		mu.Unlock()

		time.Sleep(time.Millisecond) // DB 往復の代役

		mu.Lock()
		cur--
		mu.Unlock()
		return in, nil
	})
	if err := d.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := drain(t, d); err != nil {
		t.Fatalf("drain: %v", err)
	}
	d.Close(context.Background())

	mu.Lock()
	defer mu.Unlock()
	if peak > limit {
		t.Errorf("同時実行のピーク = %d, GlobalMaxConcurrency = %d を超えた", peak, limit)
	}
	if peak < 2 {
		t.Errorf("同時実行のピーク = %d, 並行化が効いていない", peak)
	}
}

// 途中で Close しても goroutine が送信待ちで残らない（drain して畳めている）こと。
func TestAsyncDriverCloseEarlyDoesNotLeak(t *testing.T) {
	before := goroutineCount()
	p := testProcessor(AsyncExchange, ExecPolicy{Mode: ExecFixed, Default: OpConcurrency{Workers: 4}})
	src := &fakeSource{batches: 10000}
	d := newAsyncDriver(p, src, OpExpand, 1, func(in *Batch) (*Batch, error) { return in, nil })
	if err := d.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	// 1 バッチだけ受け取って即 Close（下流が早期終了したケース）。
	if _, err := d.Next(context.Background()); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if err := d.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Close は wg.Wait() まで済ませているので、残るのは closer goroutine のみ。
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if goroutineCount() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine が残存: before=%d after=%d", before, goroutineCount())
}
