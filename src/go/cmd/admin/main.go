// Command admin maintains a store instance: setup, migration, statistics and uuid format.
//
//	admin setup [-dataset NAME]   rebuild datastore/runtime/ for the running dataset
//	admin migrate -workload NAME -migmode a_to_b
//	admin consolidate [-exclude Label.prop,TYPE.prop] [-delete]
//	admin stats
//	admin uuid-format -format text|binary
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/internal/paths"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/data/setup"
	"polystore_database/src/go/lab/runner"
	workloads "polystore_database/src/go/lab/workload"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: admin <setup|migrate|consolidate|stats|uuid-format> [flags]")
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	configPath := fs.String("config", "../../config/config.json", "config file (JSON)")
	dataset := fs.String("dataset", "", "setup: dataset folder in datastore/datasets (default: the config's dataset)")
	workload := fs.String("workload", "IC11", "migrate: workload whose migration configs are used")
	migMode := fs.String("migmode", "graph_to_rdb", "migrate: direction a_to_b")
	format := fs.String("format", "", "uuid-format: target uuid storage, text or binary")
	exclude := fs.String("exclude", "", "consolidate: properties that keep their placement (Label.prop or TYPE.prop, comma separated)")
	deleteSource := fs.Bool("delete", false, "consolidate: delete the values from the other databases after copying them back")
	_ = fs.Parse(os.Args[2:])

	cfg, err := conn.LoadConfig(paths.Resolve(*configPath))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	layout := cfg.Layout()
	ctx := context.Background()

	switch cmd {
	case "setup":
		// Rebuild runtime/ for the running dataset: the record from the dataset's template, uuids, the uuid
		// indexes and empty other stores.
		ds := *dataset
		if ds == "" {
			ds = cfg.Dataset
		}
		running, rerr := setup.RunningDataset(ctx, layout)
		switch {
		case ds == "" && rerr != nil:
			log.Fatalf("setup: no dataset in the config and %v", rerr)
		case ds == "":
			ds = running
		case rerr == nil && running != ds:
			log.Fatalf("setup: the dataset is %s but the running containers are %s", ds, running)
		}
		log.Printf("setting up %s", ds)
		if err := setup.PrepareRuntime(ctx, layout, cfg, ds); err != nil {
			log.Fatalf("runtime: %v", err)
		}
		if _, err := catalog.LoadRecord(cfg.MappingPath()); err != nil {
			log.Fatalf("record mapping: %v", err)
		}
		if err := setup.AssignUUIDs(ctx, cfg, layout.LeaseLedger(), ds); err != nil {
			log.Fatalf("assign uuids: %v", err)
		}
		if err := setup.RunDataset(ctx, cfg, ds); err != nil {
			log.Fatalf("setup: %v", err)
		}
	case "migrate":
		def, _, ok := workloads.Lookup(*workload)
		if !ok {
			log.Fatalf("unknown workload %q (available: %s)", *workload, workloads.Available())
		}
		mode := migrate.MigrationMode(*migMode)
		_, _, migs := def(mode, true)
		for i := range migs {
			migs[i].Mode = mode
			migs[i].DeleteSource = settings.MigrationDeleteSource
		}
		res, err := migrate.Run(ctx, cfg, migs)
		if err != nil {
			log.Fatalf("migrate: %v", err)
		}
		runner.PrintMigration(res)
	case "consolidate":
		// Bring every property back to Neo4j except the excluded ones, and record it.
		var ex []string
		for _, e := range strings.Split(*exclude, ",") {
			if e = strings.TrimSpace(e); e != "" {
				ex = append(ex, e)
			}
		}
		plan, err := migrate.Consolidate(cfg, ex, *deleteSource)
		for _, m := range plan.Moves {
			log.Printf("moved %s %v (%s)", m.Entity, m.Properties, m.Mode)
		}
		if len(plan.Replicas) > 0 {
			log.Printf("record set to neo4j for replicas: %v", plan.Replicas)
		}
		if err != nil {
			log.Fatalf("consolidate: %v", err)
		}
		if len(plan.Moves) == 0 && len(plan.Replicas) == 0 {
			log.Printf("everything is already in neo4j")
		}
	case "stats":
		// Planner statistics from the Neo4j count store, saved next to mapping.json.
		driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
		if err != nil {
			log.Fatalf("connect neo4j: %v", err)
		}
		defer driver.Close(ctx)
		st, err := setup.CollectStats(ctx, driver)
		if err != nil {
			log.Fatalf("collect stats: %v", err)
		}
		path := stats.PathFor(cfg.MappingPath())
		if err := st.Save(path); err != nil {
			log.Fatalf("save stats: %v", err)
		}
		log.Printf("stats saved to %s (%d nodes, %d labels, %d relationship types)", path, st.Nodes, len(st.Labels), len(st.RelTypes))
	case "uuid-format":
		// Rewrites every Neo4j uuid to the target storage and records it. The other stores must be
		// set up and migrated again.
		target, err := id.ParseStorage(*format)
		if err != nil || *format == "" {
			log.Fatalf("uuid-format: -format text|binary is required")
		}
		md, err := catalog.LoadRecord(cfg.MappingPath())
		if err != nil {
			log.Fatalf("record mapping: %v", err)
		}
		id.SetStorage(target)
		if err := setup.ConvertUUIDFormat(ctx, cfg); err != nil {
			log.Fatalf("uuid-format: %v", err)
		}
		md.UUIDStorage = target.String()
		if err := md.SaveMappingDictionary(cfg.MappingPath()); err != nil {
			log.Fatalf("record mapping: %v", err)
		}
	default:
		usage()
	}
}
