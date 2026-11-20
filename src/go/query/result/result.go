// Package result holds the result and the measurements of one query execution.
package result

import "time"

// StepMetric measures one physical operator (all iteration models; summed over batches).
type StepMetric struct {
	Step            int           // execution order (leaf to root)
	Op              string        // operator kind (StoreQuery[neo4j], Sort, ...)
	Desc            string        // the plan's Describe, to match steps with the plan
	Store           string        // the store queried (empty for coordinator operators)
	Duration        time.Duration // time spent in the operator
	InRows          int           // input rows
	OutRows         int           // output (intermediate) rows
	OutRowsExpanded int64         // output rows with weighted rows expanded (the real count, with variable-length duplicates)
	RoundTrips      int64         // queries the operator sent to stores

	// L3 breakdown of the operator's store calls (ExecOptions.Breakdown): sending to the first row,
	// receiving the rest, and decoding values. The remainder of Duration is the executor's own work.
	FirstRow, Receive, Decode time.Duration
	// Profiles are the stores' own reports (L4, ExecOptions.StoreProfile), one per store call.
	Profiles []StoreProfile
}

// StoreProfile is a store's own report on one query (L4): Neo4j PROFILE, MySQL EXPLAIN ANALYZE,
// MongoDB explain. Rows, DBHits and Time are the summable numbers (0 when the store does not give them);
// Raw is the store's output as is (saved as JSON).
type StoreProfile struct {
	Store  string        `json:"store"`
	Query  string        `json:"query"`
	Rows   int64         `json:"rows"`
	DBHits int64         `json:"db_hits"`
	Time   time.Duration `json:"time_ns"`
	Raw    any           `json:"raw"`
}

// Result is the result and the measurements of one query execution. PlanTime is parsing and
// planning, ExecTime is execution; TotalLatency is their sum. VectorWidth is set only for the
// vectorized, stream and vecstream models.
type Result struct {
	Rows        []map[string]interface{}
	PlanTime    time.Duration
	ExecTime    time.Duration
	Steps       []StepMetric
	RoundTrips  int64            // queries sent to stores
	StoreTrips  map[string]int64 // RoundTrips per store (store.Kind name -> count)
	VectorWidth int              // vectorized, stream and vecstream only (0 otherwise)
	Engine      string           // engine name
	ExecKind    string           // how it ran (ExecStructured / ExecVerbatim)
}

// How a query ran. Verbatim sent the original Cypher to Neo4j without a plan: it runs the same in every placement and has no Steps.
const (
	ExecStructured = "structured"
	ExecVerbatim   = "verbatim"
)

// RowCount is the number of result rows.
func (r Result) RowCount() int { return len(r.Rows) }

// TotalLatency is planning plus execution.
func (r Result) TotalLatency() time.Duration { return r.PlanTime + r.ExecTime }

// SumStepTime is the total operator time.
func (r Result) SumStepTime() time.Duration {
	var s time.Duration
	for _, st := range r.Steps {
		s += st.Duration
	}
	return s
}
