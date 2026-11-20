// Package store names the databases a value can be placed in (Kind). It is a leaf package so
// that the plan, the stores and the connections share one vocabulary.
package kind

import "fmt"

// Kind is a database.
type Kind int

const (
	Neo4j Kind = iota
	ClickHouse
	MySQL
	Mongo
	LevelDB
	Cassandra
	RocksDB
)

var kindNames = map[Kind]string{
	Neo4j:      "neo4j",
	Mongo:      "mongo",
	LevelDB:    "leveldb",
	RocksDB:    "rocksdb",
	MySQL:      "mysql",
	ClickHouse: "clickhouse",
	Cassandra:  "cassandra",
}

// legacyNames are the data-model names used before placements were named after databases.
// Mapping files and experiment specs that still use them keep working.
var legacyNames = map[string]Kind{
	"graph":      Neo4j,
	"relational": MySQL,
	"document":   Mongo,
	"columnar":   ClickHouse,
	"widecolumn": Cassandra,
	"kvs":        LevelDB, // the key-value placement before LevelDB and RocksDB were separate
}

func (k Kind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// ParseKind converts a database name ("neo4j", ...) or a legacy data-model name ("graph", ...) to a Kind.
func ParseKind(s string) (Kind, error) {
	for k, name := range kindNames {
		if name == s {
			return k, nil
		}
	}
	if k, ok := legacyNames[s]; ok {
		return k, nil
	}
	return 0, fmt.Errorf("unknown store kind: %q", s)
}

// IsKV reports whether k is a key-value store (LevelDB or RocksDB).
func (k Kind) IsKV() bool { return k == LevelDB || k == RocksDB }
