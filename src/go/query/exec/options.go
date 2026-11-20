package exec

import ()

// Model is an iteration model: how rows flow through the same physical plan.
type Model string

const (
	ModelBulk       Model = "bulk"       // materialize every operator's full output
	ModelVolcano    Model = "volcano"    // pull, one input row per store round trip
	ModelVectorized Model = "vectorized" // pull, VectorSize rows per round trip
	ModelStream     Model = "stream"     // push, a goroutine per operator
	ModelVecStream  Model = "vecstream"  // pull with an exchange of parallel workers
)

// ExecOptions are the execution switches for one run.
type ExecOptions struct {
	Model Model
	// VectorSize is the batch width of vectorized, stream and vecstream.
	VectorSize int
	// MaterializeChunkSize is the number of ids per store query (0 = the store's default).
	MaterializeChunkSize int
	// ShipChunkRows is the number of rows per shipped table (0 = one table).
	ShipChunkRows int
	// StoreFetchParallelism is the number of concurrent chunks per store query.
	StoreFetchParallelism int
	// GraphFetchParallelism replaces StoreFetchParallelism for Neo4j.
	GraphFetchParallelism int
	// KVBatchKeys is the number of ids a key-value store reads per batch (one MultiGet on RocksDB; 0 or 1 =
	// key by key).
	KVBatchKeys int
	// KVFetchParallelism is the number of batches a key-value store reads at once inside one query.
	KVFetchParallelism int
	// Breakdown splits store calls into first row / receive / decode (L3; StepMetric).
	Breakdown bool
	// StoreProfile asks the stores for their own report on every query (L4). It makes the stores do extra
	// work, so take it in a trial of its own.
	StoreProfile bool
}

// VectorWidth is the batch width used by the vectorized, stream and vecstream models.
func (o ExecOptions) VectorWidth() int {
	if o.VectorSize < 1 {
		return 1
	}
	return o.VectorSize
}
