package conn

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The kv implementations (kind.LevelDB / kind.RocksDB) behave the same. kind.RocksDB only with `-tags rocksdb`.
func TestKVImplementations(t *testing.T) {
	ctx := context.Background()
	impls := map[string]func(dir string) (KV, func(), error){
		"leveldb": func(dir string) (KV, func(), error) {
			s, err := openLevelDB(ctx, dir)
			if err != nil {
				return nil, nil, err
			}
			return s, func() { s.Close(ctx) }, nil
		},
	}
	if RocksDBAvailable {
		impls["rocksdb"] = func(dir string) (KV, func(), error) {
			s, err := openRocksDB(ctx, dir)
			if err != nil {
				return nil, nil, err
			}
			kv, _ := any(s).(KV)
			return kv, func() { s.Close(ctx) }, nil
		}
	}
	for name, open := range impls {
		t.Run(name, func(t *testing.T) {
			kv, done, err := open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer done()
			b := &KVBatch{}
			for _, k := range []string{"a\x00\x01", "a\x00\x02", "a\x01", "b\x00x", "a\x00\x03"} {
				b.Put([]byte(k), []byte("v"+k))
			}
			b.Put([]byte("empty"), []byte{})
			if err := kv.Write(b); err != nil {
				t.Fatal(err)
			}
			if v, ok, err := kv.Get([]byte("a\x00\x02")); err != nil || !ok || string(v) != "va\x00\x02" {
				t.Fatalf("Get = %q %v %v", v, ok, err)
			}
			if v, ok, err := kv.Get([]byte("empty")); err != nil || !ok || len(v) != 0 {
				t.Fatalf("Get(empty) = %q %v %v", v, ok, err)
			}
			if _, ok, err := kv.Get([]byte("zzz")); err != nil || ok {
				t.Fatalf("Get(missing) = %v %v", ok, err)
			}
			if _, multi := kv.(MultiGetter); multi != (name == "rocksdb") {
				t.Fatalf("MultiGetter = %v", multi)
			}
			for _, sorted := range []bool{true, false} {
				in := [][]byte{[]byte("a\x00\x01"), []byte("a\x00\x02"), []byte("a\x00\x09"), []byte("empty"), []byte("zzz")}
				if !sorted {
					in[0], in[4] = in[4], in[0]
				}
				vals, found, err := GetMany(kv, in, sorted)
				if err != nil {
					t.Fatal(err)
				}
				for i, k := range in {
					v, ok, _ := kv.Get(k)
					if found[i] != ok || string(vals[i]) != string(v) || (ok && vals[i] == nil) {
						t.Fatalf("GetMany(sorted=%v)[%q] = %q %v, Get = %q %v", sorted, k, vals[i], found[i], v, ok)
					}
				}
			}
			var keys []string
			if err := ScanPrefix(kv, []byte("a\x00"), func(k, _ []byte) error {
				keys = append(keys, string(k))
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if len(keys) != 3 || keys[0] != "a\x00\x01" || keys[2] != "a\x00\x03" {
				t.Fatalf("ScanPrefix = %q", keys)
			}
			stop := errors.New("stop")
			n := 0
			if err := kv.Scan(nil, nil, func(_, _ []byte) error { n++; return stop }); err != stop || n != 1 {
				t.Fatalf("Scan stop = %v n=%d", err, n)
			}
			d := &KVBatch{}
			d.Delete([]byte("a\x00\x01"))
			if err := kv.Write(d); err != nil {
				t.Fatal(err)
			}
			if _, ok, _ := kv.Get([]byte("a\x00\x01")); ok {
				t.Fatal("delete did not take effect")
			}
		})
	}
}

func writeLedger(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "uuid_leases.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A missing ledger is an error, not a start at 0.
func TestUUIDLeaseRequiresLedger(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "uuid_leases.json")
	if _, err := LeaseUUIDAllocator(missing, "a"); err == nil || !strings.Contains(err.Error(), UUIDLeaseExampleFile) {
		t.Errorf("a missing ledger must give an error pointing to the template: %v", err)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Errorf("the ledger must not be created")
	}
}

// Lease numbers are one sequence across datasets and every lease is recorded with its dataset.
func TestUUIDLeaseSharedSequence(t *testing.T) {
	path := writeLedger(t, `{"_comment": "keep", "next": 1, "leases": [{"lease": 0, "dataset": "a"}]}`)
	for _, c := range []struct {
		ds   string
		want uint64
	}{{"b", 1}, {"a", 2}, {"b", 3}} {
		a, err := LeaseUUIDAllocator(path, c.ds)
		if err != nil || a.Lease() != c.want {
			t.Fatalf("%s: lease %v %v, want %d", c.ds, a, err, c.want)
		}
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), `"keep"`) || !strings.Contains(string(got), `"next": 4`) {
		t.Errorf("ledger:\n%s", got)
	}
	if ls, err := DatasetLeases(path, "b"); err != nil || len(ls) != 2 || ls[0] != 1 || ls[1] != 3 {
		t.Errorf("leases of b: %v %v", ls, err)
	}
}

// Bulk assignment works only for a dataset without leases.
func TestUUIDLeaseBulk(t *testing.T) {
	path := writeLedger(t, `{"next": 5, "leases": []}`)
	a, err := LeaseUUIDAllocatorForBulk(path, "a")
	if err != nil || a.Lease() != 5 {
		t.Fatalf("a dataset without leases must pass: %v %v", a, err)
	}
	if _, err := LeaseUUIDAllocatorForBulk(path, "a"); err == nil || !strings.Contains(err.Error(), "already took lease 5") {
		t.Errorf("a dataset with a lease must be an error: %v", err)
	}
	if a, err := LeaseUUIDAllocator(path, "a"); err != nil || a.Lease() != 6 {
		t.Errorf("an update takes the next number: %v %v", a, err)
	}
}

// Concurrent takers get distinct, gapless numbers (the file lock).
func TestUUIDLeaseConcurrent(t *testing.T) {
	path := writeLedger(t, `{"next": 0, "leases": []}`)
	const n = 40
	var (
		mu  sync.Mutex
		got []uint64
		wg  sync.WaitGroup
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := LeaseUUIDAllocator(path, "a")
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			got = append(got, a.Lease())
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	for i, v := range got {
		if v != uint64(i) {
			t.Fatalf("numbers overlap or are missing: %v", got)
		}
	}
}
