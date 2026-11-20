package runner

import (
	"fmt"
	"sort"
	"time"

	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/query/result"
)

// TrialResult is the summary of one workload trial (scaffolding for detailed reports).
type TrialResult struct {
	WorkloadName string
	Mode         string
	TotalTime    time.Duration
	Steps        []result.StepMetric
}

// Output formats of single query runs (cmd/query, lab workflow).
const (
	FormatRows   = "rows"   // result rows
	FormatTiming = "timing" // total time and row count
	FormatDetail = "detail" // time and output rows per operator
)

// Print writes r in format.
func Print(title, format string, r result.Result) {
	switch format {
	case FormatRows:
		PrintRows(r)
	case FormatDetail:
		PrintDetail(title, r)
	default:
		PrintTiming(title, r)
	}
}

// PrintRows writes the result rows one per line.
func PrintRows(r result.Result) {
	fmt.Printf("\n--- %d rows ---\n", r.RowCount())
	for i, row := range r.Rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Printf("[%d] ", i)
		for j, k := range keys {
			if j > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s=%v", k, row[k])
		}
		fmt.Println()
	}
}

// PrintTiming writes the total time and the row count.
func PrintTiming(title string, r result.Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - latency: %v\n", r.TotalLatency())
	fmt.Printf("  - rows: %d\n", r.RowCount())
	fmt.Println()
}

// PrintDetail writes the time and output rows of every operator.
func PrintDetail(title string, r result.Result) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - latency: %v\n", r.TotalLatency())
	fmt.Printf("  - operator time (sum): %v\n", r.SumStepTime())
	fmt.Printf("  - rows: %d\n", r.RowCount())
	if len(r.Steps) > 0 {
		fmt.Println("  - operators:")
		for _, st := range r.Steps {
			fmt.Printf("      %-16s time=%-12v rows=%d\n", st.Op, st.Duration, st.OutRows)
		}
	}
	fmt.Println()
}

// PrintMigration writes the migrated rows and times.
func PrintMigration(results []migrate.Result) {
	fmt.Println("=== migration ===")
	for _, r := range results {
		fmt.Printf("  - %-14s mode=%-14s rows=%-8d time=%v\n", r.Entity, r.Mode, r.Count, r.Duration)
	}
	fmt.Println()
}
