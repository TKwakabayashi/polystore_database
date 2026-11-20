package conn

import (
	"bytes"
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// kv.go is the minimal read/write interface that lets the kvs implementation be swapped. Key formats
// (codec.BuildEntityKey / BuildIndexKey) and fetch plans are shared by kind.LevelDB and kind.RocksDB.

// KV is an ordered key-value store.
type KV interface {
	// Get returns key's value (found=false if absent). The caller may keep the value.
	Get(key []byte) (val []byte, found bool, err error)
	// Write persists a batch (synchronously).
	Write(b *KVBatch) error
	// Scan calls fn for the keys in [start, limit) in ascending order (nil start: from the first key; nil limit: to the end).
	// key and val are valid only during the call. When fn returns an error, Scan stops and returns it.
	Scan(start, limit []byte, fn func(key, val []byte) error) error
}

// MultiGetter is a KV that reads many keys in one call.
type MultiGetter interface {
	// MultiGet returns the value of each key (found[i]=false if absent). sorted tells that keys are in
	// ascending order. The caller may keep the values.
	MultiGet(keys [][]byte, sorted bool) (vals [][]byte, found []bool, err error)
}

// GetMany reads keys with one MultiGet when kv supports it, otherwise with a Get per key.
func GetMany(kv KV, keys [][]byte, sorted bool) ([][]byte, []bool, error) {
	if m, ok := kv.(MultiGetter); ok {
		return m.MultiGet(keys, sorted)
	}
	vals, found := make([][]byte, len(keys)), make([]bool, len(keys))
	for i, k := range keys {
		v, ok, err := kv.Get(k)
		if err != nil {
			return nil, nil, err
		}
		vals[i], found[i] = v, ok
	}
	return vals, found, nil
}

// KVBatch collects Puts and Deletes.
type KVBatch struct{ ops []kvOp }

type kvOp struct {
	key, val []byte
	del      bool
}

// Put writes val at key.
func (b *KVBatch) Put(key, val []byte) { b.ops = append(b.ops, kvOp{key: key, val: val}) }

// Delete removes key.
func (b *KVBatch) Delete(key []byte) { b.ops = append(b.ops, kvOp{key: key, del: true}) }

// Len is the number of operations in the batch.
func (b *KVBatch) Len() int { return len(b.ops) }

// PrefixRange returns the range [start, limit) of keys starting with prefix.
func PrefixRange(prefix []byte) (start, limit []byte) {
	r := util.BytesPrefix(prefix)
	return r.Start, r.Limit
}

// ScanPrefix walks the keys starting with prefix.
func ScanPrefix(kv KV, prefix []byte, fn func(key, val []byte) error) error {
	start, limit := PrefixRange(prefix)
	return kv.Scan(start, limit, fn)
}

// KV returns the connection of key-value store k (kind.LevelDB or kind.RocksDB).
func (rg *Registry) KV(k kind.Kind) (KV, bool) {
	kv, ok := rg.stores[k].(KV)
	return kv, ok
}

func openKV(ctx context.Context, cfg Config, k kind.Kind) (Store, error) {
	switch k {
	case kind.LevelDB:
		return openLevelDB(ctx, cfg.KVPath(k))
	case kind.RocksDB:
		return openRocksDB(ctx, cfg.KVPath(k))
	}
	return nil, fmt.Errorf("%s is not a key-value store", k)
}

// kvConfigured reports whether key-value store k is available: LevelDB always (embedded), RocksDB in a
// build with -tags rocksdb.
func (cfg Config) kvConfigured(k kind.Kind) bool {
	switch k {
	case kind.LevelDB:
		return true
	case kind.RocksDB:
		return RocksDBAvailable
	}
	return false
}

// KVPath returns the data directory of key-value store k (datastore/runtime/leveldb or rocksdb).
func (cfg Config) KVPath(k kind.Kind) string { return cfg.Layout().KVDir(k) }

// ---- LevelDB ----

type LevelDBStore struct{ db *leveldb.DB }

func openLevelDB(_ context.Context, path string) (*LevelDBStore, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("leveldb open: %w", err)
	}
	return &LevelDBStore{db: db}, nil
}
func (s *LevelDBStore) Kind() kind.Kind               { return kind.LevelDB }
func (s *LevelDBStore) DB() *leveldb.DB               { return s.db }
func (s *LevelDBStore) Close(_ context.Context) error { return s.db.Close() }

func (s *LevelDBStore) Get(key []byte) ([]byte, bool, error) {
	v, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func (s *LevelDBStore) Write(b *KVBatch) error {
	lb := new(leveldb.Batch)
	for _, op := range b.ops {
		if op.del {
			lb.Delete(op.key)
		} else {
			lb.Put(op.key, op.val)
		}
	}
	return s.db.Write(lb, &opt.WriteOptions{Sync: true})
}

func (s *LevelDBStore) Scan(start, limit []byte, fn func(key, val []byte) error) error {
	iter := s.db.NewIterator(&util.Range{Start: start, Limit: limit}, nil)
	defer iter.Release()
	for iter.Next() {
		if limit != nil && bytes.Compare(iter.Key(), limit) >= 0 {
			break
		}
		if err := fn(iter.Key(), iter.Value()); err != nil {
			return err
		}
	}
	return iter.Error()
}
