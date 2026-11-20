package migrate

import (
	"fmt"
	"polystore_database/src/go/base/kind"
	"strings"

	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores"
)

type MigrationMode string

const (
	ModeGraphToKvs MigrationMode = "graph_to_kvs"
	ModeGraphToDoc MigrationMode = "graph_to_doc"
	ModeGraphToCol MigrationMode = "graph_to_col"
	ModeGraphToRdb MigrationMode = "graph_to_rdb"

	ModeKvsToGraph MigrationMode = "kvs_to_graph"
	ModeKvsToDoc   MigrationMode = "kvs_to_doc"
	ModeKvsToCol   MigrationMode = "kvs_to_col"
	ModeKvsToRdb   MigrationMode = "kvs_to_rdb"

	ModeDocToGraph MigrationMode = "doc_to_graph"
	ModeDocToKvs   MigrationMode = "doc_to_kvs"
	ModeDocToCol   MigrationMode = "doc_to_col"
	ModeDocToRdb   MigrationMode = "doc_to_rdb"

	ModeColToGraph MigrationMode = "col_to_graph"
	ModeColToKvs   MigrationMode = "col_to_kvs"
	ModeColToDoc   MigrationMode = "col_to_doc"
	ModeColToRdb   MigrationMode = "col_to_rdb"

	ModeRdbToGraph MigrationMode = "rdb_to_graph"
	ModeRdbToKvs   MigrationMode = "rdb_to_kvs"
	ModeRdbToDoc   MigrationMode = "rdb_to_doc"
	ModeRdbToCol   MigrationMode = "rdb_to_col"

	// wide is Cassandra; col is ClickHouse.
	ModeGraphToWide MigrationMode = "graph_to_wide"
	ModeKvsToWide   MigrationMode = "kvs_to_wide"
	ModeDocToWide   MigrationMode = "doc_to_wide"
	ModeRdbToWide   MigrationMode = "rdb_to_wide"
	ModeColToWide   MigrationMode = "col_to_wide"
	ModeWideToGraph MigrationMode = "wide_to_graph"
	ModeWideToKvs   MigrationMode = "wide_to_kvs"
	ModeWideToDoc   MigrationMode = "wide_to_doc"
	ModeWideToRdb   MigrationMode = "wide_to_rdb"
	ModeWideToCol   MigrationMode = "wide_to_col"
)

type MigrationConfig struct {
	ObjType     catalog.ObjectType
	Entity      string
	Properties  []string
	Mode        MigrationMode
	MappingPath string
	MongoDbName string
	// DeleteSource deletes the source data after the verification and the record update succeed. By default
	// (false) nothing is deleted from the source, so no data is lost before copying and verification complete.
	DeleteSource bool
}

// target is the stores' view of a migration (types: property -> type name).
func (c MigrationConfig) target(types map[string]string) stores.Target {
	return stores.Target{Obj: c.ObjType, Entity: c.Entity, Props: c.Properties, Types: types, MongoDB: c.MongoDbName}
}

// shortNames are the short store names used in migration modes ("graph_to_rdb").
var shortNames = map[string]kind.Kind{
	"graph": kind.Neo4j, "rdb": kind.MySQL, "doc": kind.Mongo, "col": kind.ClickHouse,
	"wide": kind.Cassandra, "kvs": kind.LevelDB,
}

func parseSide(s string) (kind.Kind, error) {
	if k, ok := shortNames[s]; ok {
		return k, nil
	}
	return kind.ParseKind(s)
}

// ModeStores maps a mode "<from>_to_<to>" to its source and destination. Either side may be a
// short name (graph, rdb, doc, col, wide, kvs) or a database name (neo4j, mysql, ...), so any pair works.
func ModeStores(m MigrationMode) (src, dest kind.Kind, err error) {
	from, to, ok := strings.Cut(string(m), "_to_")
	if !ok {
		return 0, 0, fmt.Errorf("unsupported mode: %s", m)
	}
	if src, err = parseSide(from); err != nil {
		return 0, 0, fmt.Errorf("mode %s: %w", m, err)
	}
	if dest, err = parseSide(to); err != nil {
		return 0, 0, fmt.Errorf("mode %s: %w", m, err)
	}
	if src == dest {
		return 0, 0, fmt.Errorf("mode %s: source and destination are the same", m)
	}
	return src, dest, nil
}

// ModeOf returns the mode that moves data from one database to another.
func ModeOf(from, to kind.Kind) MigrationMode {
	return MigrationMode(from.String() + "_to_" + to.String())
}

func modeStores(m MigrationMode) (src, dest kind.Kind, err error) { return ModeStores(m) }
