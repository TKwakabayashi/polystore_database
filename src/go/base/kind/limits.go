package kind

// limits.go holds the physical limits of each store: how many ids one query can carry. They are
// facts about the store, not settings. The tuning knobs (ExecOptions.MaterializeChunkSize and
// StoreFetchParallelism) are clamped by the executor to min(wanted, limit).

// Limits bound the number of ids in one query.
type Limits struct {
	// MaxInList is the hard limit on ids in an IN / $in (0 = no documented limit); exceeding it is an error.
	MaxInList int

	// DefaultChunk is the default chunk size (0 = no chunking): not a hard limit, the size above which a
	// query becomes slow or risky.
	DefaultChunk int
}

// Per store (hard limit vs default chunk):
// - MySQL: prepared statements allow at most 65535 placeholders (16-bit), and max_allowed_packet
// bounds the query length. Hard limit.
// - MongoDB: the query document must fit in 16MB of BSON; about 40B per uuid gives ~390k ids. Set
// conservatively. Hard limit.
// - ClickHouse: no IN limit, but max_query_size bounds the text (raised on connect in
// conn.openClickHouse); the default chunk keeps one query at a few hundred KB.
// - Cassandra: no protocol limit, but a large IN fans out on the coordinator. Default only.
// - Neo4j: no documented limit; large parameters cost memory and planning. Default only.
// - LevelDB / RocksDB: no IN; keys are read directly. RocksDB splits large inputs so that chunks are read in
// parallel (StoreFetchParallelism); each chunk is further read in MultiGet batches (ExecOptions.KVBatchKeys).
var limits = map[Kind]Limits{
	Neo4j:      {MaxInList: 0, DefaultChunk: 50000},
	MySQL:      {MaxInList: 65535, DefaultChunk: 1000},
	Mongo:      {MaxInList: 200000, DefaultChunk: 50000},
	ClickHouse: {MaxInList: 0, DefaultChunk: 5000},
	Cassandra:  {MaxInList: 0, DefaultChunk: 500},
	LevelDB:    {MaxInList: 0, DefaultChunk: 0},
	RocksDB:    {MaxInList: 0, DefaultChunk: 8192},
}

// LimitsOf returns the limits of store k (none if unregistered).
func LimitsOf(k Kind) Limits { return limits[k] }
