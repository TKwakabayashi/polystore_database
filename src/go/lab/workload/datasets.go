package workload

import (
	"sort"

	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
)

// =====================================================================
// Every workload, per dataset. ldbc_snb (LDBC) validates the engine and runs the benchmarks; its goldens are the
// query packages'. The other datasets (mappings in datastore/datasets/<dataset>/mapping.json) are for placement
// experiments (is moving a property out of Neo4j a gain or a loss, per role); their goldens are this package's.
// Names are unique overall (the other datasets' carry a prefix).
//
// Each workload is a definition plus traits (Traits: purpose, anchor, hops, fan-out, property roles R0-R7;
// traits.go), written in traits<Name> right above the definition.
//
// Parameters:
// - ast turns values that read as numbers into number literals. Library keys are often digit-only strings
// (post_id, custkey, user_id, ...) that cannot be passed as string parameters, so those are literals in the Cypher.
// - the grammar compares strings and numbers only (no booleans), so boolean columns (minor, verified_purchase) are not used in conditions.
// - values were chosen before loading the data; values marked "to check" are fixed after counting on the data.
// Selectivity sweeps replace threshold parameters with quantiles.
// =====================================================================

// Datasets maps dataset name -> workload name -> definition and traits; dataset names are schema names
// (datastore/datasets folders without the scale).
// Each dataset's definitions and table are in <dataset>.go (add one line here per new dataset).
var Datasets = map[string]map[string]Workload{
	LDBC:                ldbcSnbWorkloads,
	"git_log":           gitLogWorkloads,
	"stack_exchange":    stackExchangeWorkloads,
	"stack_history":     stackHistoryWorkloads,
	"tpch":              tpchWorkloads,
	"movielens":         movielensWorkloads,
	"amazon_review2023": amazonReview2023Workloads,
	"ldbc_snb_datagen":  ldbcSnbDatagenWorkloads,
	"wiki_revision":     wikiRevisionWorkloads,
	"imdb":              imdbWorkloads,
	"openalex":          openalexWorkloads,
}

// LDBC is the dataset (schema) of the ldbc_snb workloads.
const LDBC = "ldbc_snb"

// DatasetNames returns the dataset names of Datasets, sorted.
func DatasetNames() []string {
	names := make([]string, 0, len(Datasets))
	for n := range Datasets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Names returns dataset ds's workload names, sorted (nil if unknown).
func Names(ds string) []string {
	defs := Datasets[ds]
	names := make([]string, 0, len(defs))
	for n := range defs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Lookup returns a workload's definition and dataset.
func Lookup(name string) (def workloadDef, dataset string, ok bool) {
	for ds, ws := range Datasets {
		if w, ok := ws[name]; ok {
			return w.Define, ds, true
		}
	}
	return nil, "", false
}

// mig builds migration configs: migs fills mode, and it returns nil when isMigration is false.
type mig struct {
	obj    catalog.ObjectType
	entity string
	props  []string
}

func node(entity string, props ...string) mig { return mig{catalog.ObjEntity, entity, props} }
func edge(rel string, props ...string) mig    { return mig{catalog.ObjRelationship, rel, props} }

func migs(mode migrate.MigrationMode, isMigration bool, ms ...mig) []migrate.MigrationConfig {
	if !isMigration {
		return nil
	}
	out := make([]migrate.MigrationConfig, len(ms))
	for i, m := range ms {
		out[i] = migrate.MigrationConfig{ObjType: m.obj, Entity: m.entity, Properties: m.props, Mode: mode}
	}
	return out
}
