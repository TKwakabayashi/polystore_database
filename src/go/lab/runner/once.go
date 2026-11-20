package runner

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
	"polystore_database/src/go/lab/workload"
	"polystore_database/src/go/query"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/result"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/stores/conn"
	"polystore_database/src/go/stores/dbs"
)

// Targets of single query runs.
const (
	TargetCustom = "custom" // this system
	TargetNeo4j  = "neo4j"  // Neo4j running the Cypher as is
	TargetBoth   = "both"
)

// RunWorkloadByName runs a workload's query once on target and prints it in format. It does not set up or
// migrate data (that would clear the other stores).
func RunWorkloadByName(ctx context.Context, name string, cfg conn.Config, target, format string) {
	def, _, ok := workload.Lookup(name)
	if !ok {
		log.Fatalf("unknown workload %q (available: %s)", name, workload.Available())
	}
	cypher, params, _ := def(migrate.ModeKvsToGraph, false)
	fmt.Printf("=== workload %s ===\n", name)
	if target == TargetCustom || target == TargetBoth {
		r, err := RunEngine(ctx, cfg, nil, settings.Plan(), settings.Engine, cypher, params)
		if err != nil {
			log.Fatalf("%v", err)
		}
		Print("this system", format, r)
	}
	if target == TargetNeo4j || target == TargetBoth {
		if cfg.Neo4j == nil {
			log.Printf("no Neo4j in the config, skipped")
		} else if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("neo4j: %v", err)
		} else {
			Print("neo4j (baseline)", format, r)
		}
	}
}

// RunWorkflow measures the Neo4j baseline, migrates, then runs this system on the new placement. The
// baseline runs first because a moving migration deletes the source properties.
func RunWorkflow(ctx context.Context, name string, cfg conn.Config, cypher string, params map[string]string, migs []migrate.MigrationConfig, format string) {
	fmt.Printf("===== workload %s =====\n", name)
	if cfg.Neo4j != nil {
		if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("neo4j: %v", err)
		} else {
			Print("neo4j (baseline)", format, r)
		}
	}
	if len(migs) > 0 {
		res, err := migrate.Run(ctx, cfg, migs)
		if err != nil {
			log.Printf("migration: %v", err)
		}
		PrintMigration(res)
	}
	if r, err := RunEngine(ctx, cfg, nil, settings.Plan(), settings.Engine, cypher, params); err != nil {
		log.Printf("this system: %v", err)
	} else {
		Print("this system", format, r)
	}
}

// RunEngine runs cypher once on this system with the iteration model kind. md is the runtime placement
// (nil: the record as it is), popt the planning options.
func RunEngine(ctx context.Context, cfg conn.Config, md *catalog.MappingDictionary, popt query.PlanOptions, kind exec.Model, cypher string, params map[string]string) (result.Result, error) {
	if md == nil {
		var err error
		if md, err = catalog.LoadRecord(cfg.MappingPath()); err != nil {
			return result.Result{}, fmt.Errorf("load mapping: %w", err)
		}
	}
	reg, err := conn.NewRegistry(ctx, cfg)
	if err != nil {
		return result.Result{}, fmt.Errorf("connect stores (are all databases up?): %w", err)
	}
	defer reg.Close(ctx)
	return query.Run(ctx, dbs.NewSession(reg), md, cypher, params, popt, settings.Exec(kind))
}

// RunNeo4j runs cypher once on Neo4j as is (the baseline).
func RunNeo4j(ctx context.Context, cfg conn.Neo4jConfig, cypher string, params map[string]interface{}) (result.Result, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return result.Result{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	one := func() (result.Result, error) {
		start := time.Now()
		// An auto-commit session.Run, as the executor sends to Neo4j.
		res, err := session.Run(ctx, cypher, params)
		if err != nil {
			return result.Result{}, err
		}
		var rows []map[string]interface{}
		for res.Next(ctx) {
			rec := res.Record()
			row := make(map[string]interface{}, len(rec.Keys))
			for i, key := range rec.Keys {
				row[key] = rec.Values[i]
			}
			rows = append(rows, row)
		}
		if err := res.Err(); err != nil {
			return result.Result{}, err
		}
		return result.Result{Rows: rows, ExecTime: time.Since(start), Engine: "neo4j", ExecKind: "direct"}, nil
	}
	return one()
}

// toValuedParams types string parameters for Neo4j.
func toValuedParams(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		if n, err := strconv.Atoi(v); err == nil {
			out[k] = n
		} else if list, ok := query.ParseListParam(v); ok {
			out[k] = list
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = t
		} else if t, err := time.Parse(time.DateOnly, v); err == nil {
			out[k] = neo4j.DateOf(t) // a time.Time is a DateTime, which never equals a Date
		} else {
			out[k] = v
		}
	}
	return out
}
