// Command query runs one workload query once and prints its result. The iteration model and the other
// internal switches come from settings.
package main

import (
	"context"
	"flag"
	"log"
	"polystore_database/src/go/internal/paths"

	"polystore_database/src/go/lab/runner"
	"polystore_database/src/go/stores/conn"
)

func main() {
	var (
		workload   = flag.String("workload", "IC11", "workload name")
		configPath = flag.String("config", "../../config/config.json", "config file (JSON)")
		target     = flag.String("target", runner.TargetCustom, "custom (this system), neo4j (the Cypher as is) or both")
		format     = flag.String("format", runner.FormatTiming, "rows, timing or detail")
	)
	flag.Parse()

	cfg, err := conn.LoadConfig(paths.Resolve(*configPath))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := cfg.CheckRuntime(); err != nil {
		log.Printf("warning: %v", err)
	}
	runner.RunWorkloadByName(context.Background(), *workload, cfg, *target, *format)
}
