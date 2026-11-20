// Package leveldb is the LevelDB key-value database. Compilation and execution are shared with
// RocksDB in stores/internal/kv; the connection is conn's LevelDB store.
package leveldb

import (
	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/stores/internal/kv"
)

// DB is LevelDB.
var DB = kv.DB{K: kind.LevelDB}
