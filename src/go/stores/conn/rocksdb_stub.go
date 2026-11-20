//go:build !rocksdb

package conn

import (
	"context"
	"fmt"
)

// RocksDBAvailable reports whether the RocksDB implementation is built in (true only with `-tags rocksdb`).
const RocksDBAvailable = false

// openRocksDB fails in builds without RocksDB (it needs cgo and librocksdb, so the default build leaves it out).
func openRocksDB(_ context.Context, _ string) (Store, error) {
	return nil, fmt.Errorf("rocksdb: this binary was built without -tags rocksdb (see stores/conn/rocksdb.go)")
}
