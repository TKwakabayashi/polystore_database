// Command lab runs experiments.
//
//	lab run -spec FILE [-out DIR]   an experiment from a spec (results/<date>_<name>/); lab/specs has
//	                                examples, opbench specs (mode "opbench") run the operator benchmark
//	lab workflow -workload NAME -migmode a_to_b [-format rows|timing|detail]
//	                                Neo4j baseline, migration, then this system
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"polystore_database/src/go/internal/paths"

	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/runner"
	workloads "polystore_database/src/go/lab/workload"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: lab <run|workflow> [flags]")
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	configPath := fs.String("config", "../../config/config.json", "config file (JSON)")
	workload := fs.String("workload", "IC11", "workflow: workload name")
	migMode := fs.String("migmode", "graph_to_rdb", "workflow: migration direction a_to_b")
	outPath := fs.String("out", "../../results", "run: directory for the result folders")
	specPath := fs.String("spec", "", "run: experiment definition (JSON)")
	format := fs.String("format", runner.FormatTiming, "workflow: rows, timing or detail")
	_ = fs.Parse(os.Args[2:])

	cfg, err := conn.LoadConfig(paths.Resolve(*configPath))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := cfg.CheckRuntime(); err != nil {
		log.Printf("warning: %v", err)
	}
	ctx := context.Background()

	switch cmd {
	case "run":
		if *specPath == "" {
			log.Fatal("-spec is required")
		}
		var outDir string
		outDir, err = runner.RunSpec(ctx, cfg, *specPath, paths.Resolve(*outPath))
		if outDir != "" {
			log.Printf("results in %s", outDir)
		}
	case "workflow":
		def, _, ok := workloads.Lookup(*workload)
		if !ok {
			log.Fatalf("unknown workload %q (available: %s)", *workload, workloads.Available())
		}
		mode := migrate.MigrationMode(*migMode)
		cypher, params, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode
			migs[i].DeleteSource = settings.MigrationDeleteSource
		}
		runner.RunWorkflow(ctx, *workload, cfg, cypher, params, migs, *format)
	default:
		usage()
	}
	if err != nil {
		log.Fatalf("%s: %v", cmd, err)
	}
}
