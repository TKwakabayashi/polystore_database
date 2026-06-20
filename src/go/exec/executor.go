package exec

import "polystore_database/src/go/id"

type Batch struct {
	Cols []id.UUID
	N    int
}

type Executor interface {
	Next() (*Batch, error)
}
