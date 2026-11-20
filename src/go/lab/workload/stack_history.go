package workload

import "polystore_database/src/go/data/migrate"

// stack_history: stack_exchange's posts plus edit revisions (650k Rev with texts) and their chain (PREV_EDIT).
// SH1-4 are stack_exchange.go's stackQA1-4; SH5 and SH6 use Rev.

var stackHistoryWorkloads = map[string]Workload{
	"SH1": {DefineWorkloadSH1, traitsStackQA1},
	"SH2": {DefineWorkloadSH2, traitsStackQA2},
	"SH3": {DefineWorkloadSH3, traitsStackQA3},
	"SH4": {DefineWorkloadSH4, traitsStackQA4},
	"SH5": {DefineWorkloadSH5, traitsSH5},
	"SH6": {DefineWorkloadSH6, traitsSH6},
}

func DefineWorkloadSH1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA1(mode, isMigration)
}

func DefineWorkloadSH2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA2(mode, isMigration)
}

func DefineWorkloadSH3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA3(mode, isMigration)
}

func DefineWorkloadSH4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA4(mode, isMigration)
}

var traitsSH5 = Traits{
	Purpose:   "A whole-path condition (none) on the linear chain of versions: the cost of extracting paths",
	Anchor:    AnchorScan,
	Hops:      "1..5",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"r.creation_date"},
		RolePath:      {"x.history_type"},
		RoleGroupSort: {"o.post_id"},
	},
}

// SH5: revision chains (newer -> previous) of 1-5 hops without a rollback (history_type 8), counted.
func DefineWorkloadSH5(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_history", "SH5")
	params := map[string]string{
		"since": "2020-01-01", // to check: creation_date format (string comparison)
	}
	return cypher, params, migs(mode, isMigration,
		node("Rev", "creation_date", "history_type", "post_id"))
}

var traitsSH6 = Traits{
	Purpose: "Fetching version text (large in total) after Top-k, and the effect of a smaller graph",
	Anchor:  AnchorScan,
	Hops:    "2",
	TopK:    10,
	FanOut:  FanLow,
	Roles: map[Role][]string{
		RoleOutput:    {"p.post_id"},
		RoleAnchor:    {"u.reputation"},
		RoleEndpoint:  {"p.score"},
		RoleGroupSort: {"r.creation_date", "r.rev_id"},
		RolePayload:   {"r.text"},
	},
}

// SH6: revisions of high-scoring posts edited by high-reputation users, newest first, with texts output after the Top-k.
func DefineWorkloadSH6(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_history", "SH6")
	params := map[string]string{
		"minReputation": "10000", // to check
		"minScore":      "50",
	}
	return cypher, params, migs(mode, isMigration,
		node("User", "reputation"),
		node("Post", "post_id", "score"),
		node("Rev", "rev_id", "creation_date", "text"))
}
