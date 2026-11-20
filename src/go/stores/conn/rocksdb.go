//go:build rocksdb

package conn

// rocksdb.go is the kind.RocksDB kvs implementation (grocksdb, cgo), built only with `-tags rocksdb`. For example
// (Homebrew's librocksdb is a shared library, so -lrocksdb is enough):
//
// 	SDKROOT=/Library/Developer/CommandLineTools/SDKs/MacOSX26.5.sdk \
// 	CGO_ENABLED=1 CGO_CFLAGS="-I/opt/homebrew/include" CGO_LDFLAGS="-L/opt/homebrew/lib -lrocksdb" \
// 	go build -tags rocksdb ./...
//
// SDKROOT works around the default MacOSX27.2.sdk, whose Security.tbd ld cannot read ("unknown architecture" for arm64e.x1).

import (
	"bytes"
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"

	"github.com/linxGnu/grocksdb"
)

// RocksDBAvailable reports whether the kind.RocksDB implementation is built in.
const RocksDBAvailable = true

type RocksDBStore struct {
	db   *grocksdb.DB
	opts *grocksdb.Options
	ro   *grocksdb.ReadOptions
	wo   *grocksdb.WriteOptions
}

func openRocksDB(_ context.Context, path string) (*RocksDBStore, error) {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(256 << 20))
	bbto.SetFilterPolicy(grocksdb.NewBloomFilter(10))
	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	db, err := grocksdb.OpenDb(opts, path)
	if err != nil {
		opts.Destroy()
		return nil, fmt.Errorf("rocksdb open: %w", err)
	}
	wo := grocksdb.NewDefaultWriteOptions()
	wo.SetSync(true) // same as the kind.LevelDB implementation (Sync: true)
	return &RocksDBStore{db: db, opts: opts, ro: grocksdb.NewDefaultReadOptions(), wo: wo}, nil
}

func (s *RocksDBStore) Kind() kind.Kind { return kind.RocksDB }
func (s *RocksDBStore) Close(_ context.Context) error {
	s.db.Close()
	s.ro.Destroy()
	s.wo.Destroy()
	s.opts.Destroy()
	return nil
}

func (s *RocksDBStore) Get(key []byte) ([]byte, bool, error) {
	v, err := s.db.Get(s.ro, key)
	if err != nil {
		return nil, false, err
	}
	defer v.Free()
	if !v.Exists() {
		return nil, false, nil
	}
	return append([]byte{}, v.Data()...), true, nil
}

// MultiGet reads keys with BatchedMultiGetCF on the default column family (it batches the filter and block
// reads of keys in the same block).
func (s *RocksDBStore) MultiGet(keys [][]byte, sorted bool) ([][]byte, []bool, error) {
	if len(keys) == 0 {
		return nil, nil, nil
	}
	ps, err := s.db.BatchedMultiGetCF(s.ro, s.db.GetDefaultColumnFamily(), sorted, keys...)
	if err != nil {
		return nil, nil, err
	}
	defer ps.Destroy()
	vals, found := make([][]byte, len(keys)), make([]bool, len(keys))
	for i, p := range ps {
		if p.Exists() {
			vals[i], found[i] = append([]byte{}, p.Data()...), true
		}
	}
	return vals, found, nil
}

func (s *RocksDBStore) Write(b *KVBatch) error {
	wb := grocksdb.NewWriteBatch()
	defer wb.Destroy()
	for _, op := range b.ops {
		if op.del {
			wb.Delete(op.key)
		} else {
			wb.Put(op.key, op.val)
		}
	}
	return s.db.Write(s.wo, wb)
}

func (s *RocksDBStore) Scan(start, limit []byte, fn func(key, val []byte) error) error {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	if limit != nil {
		ro.SetIterateUpperBound(limit)
	}
	it := s.db.NewIterator(ro)
	defer it.Close()
	if start != nil {
		it.Seek(start)
	} else {
		it.SeekToFirst()
	}
	for ; it.Valid(); it.Next() {
		k, v := it.Key(), it.Value()
		// a Slice points to C memory, so pass a Go copy valid during the call
		key, val := append([]byte{}, k.Data()...), append([]byte{}, v.Data()...)
		k.Free()
		v.Free()
		if limit != nil && bytes.Compare(key, limit) >= 0 {
			break
		}
		if err := fn(key, val); err != nil {
			return err
		}
	}
	return it.Err()
}
