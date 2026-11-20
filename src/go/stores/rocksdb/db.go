// Package rocksdb is the RocksDB key-value database (connections need a build with -tags rocksdb).
// Compilation and execution are shared with LevelDB in stores/internal/kv.
package rocksdb

import (
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/stores/internal/kv"
)

// DB is RocksDB.
var DB = kv.DB{K: kind.RocksDB}
