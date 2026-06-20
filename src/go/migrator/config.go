package migrator

import (
	"fmt"

	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
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
)

type MigrationConfig struct {
	ObjType     plan.ObjectType
	Entity      string
	Properties  []string
	Mode        MigrationMode
	MappingPath string
	MongoDbName string
}

type DataRowStream struct {
	UUID    id.UUID
	Payload map[string]interface{}
}

// modeStores は Mode を (src, dest) の StoreKind へ写す。
func modeStores(m MigrationMode) (src, dest storage.StoreKind, err error) {
	table := map[MigrationMode][2]storage.StoreKind{
		ModeGraphToKvs: {storage.Graph, storage.Kvs},
		ModeGraphToDoc: {storage.Graph, storage.Document},
		ModeGraphToCol: {storage.Graph, storage.Columnar},
		ModeGraphToRdb: {storage.Graph, storage.Relational},

		ModeKvsToGraph: {storage.Kvs, storage.Graph},
		ModeKvsToDoc:   {storage.Kvs, storage.Document},
		ModeKvsToCol:   {storage.Kvs, storage.Columnar},
		ModeKvsToRdb:   {storage.Kvs, storage.Relational},

		ModeDocToGraph: {storage.Document, storage.Graph},
		ModeDocToKvs:   {storage.Document, storage.Kvs},
		ModeDocToCol:   {storage.Document, storage.Columnar},
		ModeDocToRdb:   {storage.Document, storage.Relational},

		ModeColToGraph: {storage.Columnar, storage.Graph},
		ModeColToKvs:   {storage.Columnar, storage.Kvs},
		ModeColToDoc:   {storage.Columnar, storage.Document},
		ModeColToRdb:   {storage.Columnar, storage.Relational},

		ModeRdbToGraph: {storage.Relational, storage.Graph},
		ModeRdbToKvs:   {storage.Relational, storage.Kvs},
		ModeRdbToDoc:   {storage.Relational, storage.Document},
		ModeRdbToCol:   {storage.Relational, storage.Columnar},
	}
	pair, ok := table[m]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported mode: %s", m)
	}
	return pair[0], pair[1], nil
}
