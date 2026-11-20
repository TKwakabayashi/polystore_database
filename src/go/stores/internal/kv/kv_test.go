package kv

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync/atomic"
	"testing"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// mapKV is an in-memory conn.KV (Scan is not needed here).
type mapKV struct {
	m    map[string][]byte
	gets atomic.Int64
}

func (s *mapKV) Get(key []byte) ([]byte, bool, error) {
	s.gets.Add(1)
	v, ok := s.m[string(key)]
	return v, ok, nil
}
func (s *mapKV) Write(*conn.KVBatch) error                             { return nil }
func (s *mapKV) Scan(_, _ []byte, _ func(key, val []byte) error) error { return nil }

// multiKV adds MultiGet and counts its calls.
type multiKV struct {
	*mapKV
	multi, sorted atomic.Int64
}

func (s *multiKV) MultiGet(keys [][]byte, sorted bool) ([][]byte, []bool, error) {
	s.multi.Add(1)
	if sorted {
		if !slices.IsSortedFunc(keys, bytes.Compare) {
			return nil, nil, fmt.Errorf("keys passed as sorted are not")
		}
		s.sorted.Add(1)
	}
	vals, found := make([][]byte, len(keys)), make([]bool, len(keys))
	for i, k := range keys {
		vals[i], found[i] = s.m[string(k)]
	}
	return vals, found, nil
}

func TestFetchBatchesAgree(t *testing.T) {
	const n = 1000
	base := &mapKV{m: map[string][]byte{}}
	ids := make([]id.UUID, n)
	for i := range ids {
		var raw [16]byte
		raw[14], raw[15] = byte(i>>8), byte(i)
		ids[i] = id.Make(raw)
		enc, err := codec.EncodeForKVS(int64(i), "long")
		if err != nil {
			t.Fatal(err)
		}
		base.m[string(codec.BuildEntityKey("Person", ids[i], "age"))] = enc
		if i%3 != 0 { // some entities lack the name
			base.m[string(codec.BuildEntityKey("Person", ids[i], "name"))] = []byte("p")
		}
	}
	// reverse half of the ids so that some batches are not sorted
	for i, j := 0, n/2-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	k := &stores.KVSQuery{
		Label: "Person", Alias: "p", Props: []string{"name", "age"},
		Types:   map[string]string{"name": "string", "age": "long"},
		Filters: []ir.Expr{&ir.Bin{Op: ">=", L: &ir.Column{Key: "p.age"}, R: &ir.Lit{Val: int64(10)}}},
	}
	q := &stores.Query{Columns: []string{"p", "p.name", "p.age"}}
	run := func(kv conn.KV, f stores.KVFetch) [][]any {
		var out [][]any
		err := fetchKVS(context.Background(), kv, q, k, ids, f, func(v []any) error {
			out = append(out, v)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	want := run(base, stores.KVFetch{})
	if len(want) != n-10 {
		t.Fatalf("key by key: %d rows, want %d", len(want), n-10)
	}
	for _, f := range []stores.KVFetch{{BatchKeys: 1}, {BatchKeys: 64}, {BatchKeys: 64, Parallelism: 4}, {BatchKeys: 5000, Parallelism: 3}, {Parallelism: 8}} {
		m := &multiKV{mapKV: &mapKV{m: base.m}}
		if got := run(m, f); !reflect.DeepEqual(got, want) {
			t.Errorf("%+v: rows differ from key by key", f)
		}
		if f.BatchKeys > 0 {
			if wantCalls := int64((n + f.BatchKeys - 1) / f.BatchKeys); m.multi.Load() != wantCalls || m.gets.Load() != 0 {
				t.Errorf("%+v: %d MultiGet / %d Get, want %d / 0", f, m.multi.Load(), m.gets.Load(), wantCalls)
			}
			if f.BatchKeys == 64 && m.sorted.Load() == 0 { // the ids of the second half ascend
				t.Errorf("%+v: no batch passed as sorted", f)
			}
		} else if m.multi.Load() != 0 {
			t.Errorf("%+v: MultiGet used with BatchKeys 0", f)
		}
	}
}
