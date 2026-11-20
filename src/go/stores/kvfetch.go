package stores

import "context"

// KVFetch is how a key-value store reads the values of one query. The executor sets it on the context
// (ExecOptions), so the stores do not depend on settings.
type KVFetch struct {
	// BatchKeys is the number of ids read per batch: their keys go in one MultiGet when the store has one
	// (RocksDB). 0 or 1 reads key by key.
	BatchKeys int
	// Parallelism is the number of batches read at once inside one query (1 = serial).
	Parallelism int
}

type kvFetchKey struct{}

// WithKVFetch returns ctx carrying f.
func WithKVFetch(ctx context.Context, f KVFetch) context.Context {
	return context.WithValue(ctx, kvFetchKey{}, f)
}

// KVFetchFrom returns the options on ctx (zero value: key by key, serial).
func KVFetchFrom(ctx context.Context) KVFetch {
	f, _ := ctx.Value(kvFetchKey{}).(KVFetch)
	return f
}
