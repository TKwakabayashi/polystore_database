package vecstream

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
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
)

// DB を必要とせず exchangeIterator の並列性・session 再利用・再チャンクを検証する。
// go test -race で data race も併せて検出する。

// fakeSource は scanIterator の代わり。batches 個のバッチを 1 行ずつ払い出す Iterator。
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
	b.appendRow([]uid.UUID{uid.UUID(fmt.Sprintf("id-%d", f.pos))})
	f.pos++
	return b, nil
}

func testProcessor(workers, globalMax, vecWidth int, withSem bool) *Processor {
	p := &Processor{
		ctx: context.Background(),
		exec: ExecPolicy{
			Default:              OpConcurrency{Workers: workers},
			GlobalMaxConcurrency: globalMax,
			VectorWidth:          vecWidth,
		},
		instr: core.NewInstr(),
	}
	if withSem {
		p.sem = semaphore.NewWeighted(int64(globalMax))
	}
	return p
}

// driveInts は int リソース版 exchange を回し、列スロット0の値・エラー・newRes呼び出し回数を返す。
func driveInts(p *Processor, child Iterator, process func(*int, *Batch) (*Batch, error)) ([]uid.UUID, *int32, error) {
	var newResCalls int32
	newRes := func() *int { atomic.AddInt32(&newResCalls, 1); v := 0; return &v }
	ex := newExchange(p, child, OpExpand, "Expand", 1, newRes, func(*int) {}, process)
	if err := ex.Open(context.Background()); err != nil {
		return nil, &newResCalls, err
	}
	var got []uid.UUID
	var rerr error
	for {
		b, err := ex.Next(context.Background())
		if err != nil {
			rerr = err
			break
		}
		if b == nil {
			break
		}
		for i := 0; i < b.n; i++ {
			got = append(got, b.get(i, 0))
		}
	}
	ex.Close(context.Background())
	return got, &newResCalls, rerr
}

func TestExchangeEveryRowOnce(t *testing.T) {
	const n = 200
	p := testProcessor(4, 8, 1024, true)
	src := &fakeSource{batches: n}
	got, _, err := driveInts(p, src, func(_ *int, b *Batch) (*Batch, error) { return b, nil })
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != n {
		t.Fatalf("行数 = %d, want %d", len(got), n)
	}
	seen := make(map[uid.UUID]int, n)
	for _, v := range got {
		seen[v]++
	}
	for i := 0; i < n; i++ {
		if k := uid.UUID(fmt.Sprintf("id-%d", i)); seen[k] != 1 {
			t.Errorf("%s 出現 = %d, want 1", k, seen[k])
		}
	}
	if atomic.LoadInt32(&src.closed) != 1 {
		t.Errorf("child.Close 呼び出し = %d, want 1", src.closed)
	}
}

// session 再利用: newRes（＝session 生成）はワーカー数ちょうど（バッチ数ではない）。
func TestExchangeReusesResourcePerWorker(t *testing.T) {
	const workers = 4
	p := testProcessor(workers, 8, 1024, false)
	src := &fakeSource{batches: 500}
	_, calls, err := driveInts(p, src, func(_ *int, b *Batch) (*Batch, error) { return b, nil })
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got := atomic.LoadInt32(calls); got != workers {
		t.Errorf("newRes 呼び出し = %d, want %d（ワーカーにつき1本の session 再利用）", got, workers)
	}
}

// process が幅広バッチを返しても、出力は VectorWidth 幅へ再チャンクされる。
func TestExchangeReChunksToVectorWidth(t *testing.T) {
	p := testProcessor(1, 8, 4, false)
	src := &fakeSource{batches: 1}
	var sizes []int
	var mu sync.Mutex
	ex := newExchange(p, src, OpExpand, "Expand", 1,
		noRes, noResClose,
		func(_ struct{}, _ *Batch) (*Batch, error) {
			wide := newBatch(1, 10)
			for i := 0; i < 10; i++ {
				wide.appendRow([]uid.UUID{uid.UUID(fmt.Sprintf("r%d", i))})
			}
			return wide, nil
		})
	if err := ex.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	total := 0
	for {
		b, err := ex.Next(context.Background())
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if b == nil {
			break
		}
		mu.Lock()
		sizes = append(sizes, b.n)
		mu.Unlock()
		total += b.n
		if b.n > p.exec.vectorWidth() {
			t.Errorf("出力幅 %d が VectorWidth %d 超過", b.n, p.exec.vectorWidth())
		}
	}
	ex.Close(context.Background())
	if total != 10 {
		t.Fatalf("総行数 = %d, want 10", total)
	}
	if len(sizes) != 3 { // 4,4,2
		t.Errorf("バッチ数 = %d (%v), want 3", len(sizes), sizes)
	}
}

func TestExchangePropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	p := testProcessor(4, 8, 1024, true)
	src := &fakeSource{batches: 100}
	_, _, err := driveInts(p, src, func(_ *int, _ *Batch) (*Batch, error) { return nil, wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

// 空バッチ (n==0) は下流へ流さない。
func TestExchangeSkipsEmpty(t *testing.T) {
	const n = 60
	p := testProcessor(4, 8, 1024, false)
	src := &fakeSource{batches: n}
	got, _, err := driveInts(p, src, func(_ *int, b *Batch) (*Batch, error) {
		var idx int
		fmt.Sscanf(b.get(0, 0).String(), "id-%d", &idx)
		if idx%2 == 1 {
			return newBatch(1, 0), nil // 空
		}
		return b, nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != n/2 {
		t.Fatalf("行数 = %d, want %d", len(got), n/2)
	}
}

// ExecDynamic の共有セマフォがシステム全体の同時実行を上限で抑える（かつ2以上＝並列が効く）。
func TestExchangeRespectsGlobalMax(t *testing.T) {
	const limit = 3
	p := testProcessor(8, limit, 1024, true) // workers=8 だが sem(3) が抑える
	src := &fakeSource{batches: 100}
	var mu sync.Mutex
	cur, peak := 0, 0
	_, _, err := driveInts(p, src, func(_ *int, b *Batch) (*Batch, error) {
		mu.Lock()
		cur++
		if cur > peak {
			peak = cur
		}
		mu.Unlock()
		time.Sleep(time.Millisecond)
		mu.Lock()
		cur--
		mu.Unlock()
		return b, nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if peak > limit {
		t.Errorf("同時実行ピーク = %d, globalMax = %d 超過", peak, limit)
	}
	if peak < 2 {
		t.Errorf("同時実行ピーク = %d, 並列が効いていない", peak)
	}
}

// 途中で Close してもワーカー goroutine が残らない（drain して畳めている）。
func TestExchangeCloseEarlyNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	p := testProcessor(4, 8, 1024, false)
	src := &fakeSource{batches: 100000}
	ex := newExchange(p, src, OpExpand, "Expand", 1, noRes, noResClose,
		func(_ struct{}, b *Batch) (*Batch, error) { return b, nil })
	if err := ex.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := ex.Next(context.Background()); err != nil { // 1 バッチだけ受けて即 Close
		t.Fatalf("Next: %v", err)
	}
	ex.Close(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine 残存: before=%d after=%d", before, runtime.NumGoroutine())
}
