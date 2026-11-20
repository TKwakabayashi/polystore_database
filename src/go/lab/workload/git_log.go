package workload

import "polystore_database/src/go/data/migrate"

// git_log: the kubernetes commit history (160k Commit; PARENT is a child -> parent DAG; TOUCHED carries line counts).
// Deep chains and texts (message) at small scale; paired with wiki_revision (same shape, larger). Start shas cannot
// be chosen before seeing the data, so starts are "large commits" (thresholds on files_changed / insertions);
// the threshold controls the number of starts (the selectivity).

var gitLogWorkloads = map[string]Workload{
	"GIT1": {DefineWorkloadGIT1, traitsGIT1},
	"GIT2": {DefineWorkloadGIT2, traitsGIT2},
	"GIT3": {DefineWorkloadGIT3, traitsGIT3},
	"GIT4": {DefineWorkloadGIT4, traitsGIT4},
}

var traitsGIT1 = Traits{
	Purpose:   "Short walks on deep chains: the cost when the sort key and output-only columns are off graph",
	Anchor:    AnchorScan,
	Hops:      "1..3",
	VarLength: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleOutput:    {"c.sha", "a.subject"},
		RoleAnchor:    {"c.files_changed"},
		RoleGroupSort: {"a.committed_date", "a.sha"},
	},
}

// GIT1: from large commits follow 1-3 ancestors and output subjects, newest first.
func DefineWorkloadGIT1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("git_log", "GIT1")
	params := map[string]string{
		"minFiles": "200", // to check: a value giving tens of starts
	}
	return cypher, params, migs(mode, isMigration,
		node("Commit", "sha", "files_changed", "committed_date", "subject"))
}

var traitsGIT2 = Traits{
	Purpose:   "The cost of extracting paths (paths x length) when a whole-path condition is off graph",
	Anchor:    AnchorScan,
	Hops:      "1..5",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.files_changed"},
		RolePath:      {"x.deletions"},
		RoleGroupSort: {"a.sha"},
	},
}

// GIT2: a whole-path condition: 1-5 hop ancestor paths where every commit deletes fewer lines than a threshold.
func DefineWorkloadGIT2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("git_log", "GIT2")
	params := map[string]string{
		"minFiles":     "200",  // to check
		"maxDeletions": "1000", // to check: controls the share of paths kept
	}
	return cypher, params, migs(mode, isMigration,
		node("Commit", "sha", "files_changed", "deletions"))
}

var traitsGIT3 = Traits{
	Purpose:   "The cost of splitting the graph query on middle vertex and edge conditions, and aggregation absorption",
	Anchor:    AnchorInterior,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleInterior:  {"c.committed_date"},
		RoleEdge:      {"t.insertions"},
		RoleGroupSort: {"au.name", "f.path"},
	},
}

// GIT3: filter interior vertices (Commit) and edges (TOUCHED), then count touched files per author.
func DefineWorkloadGIT3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("git_log", "GIT3")
	params := map[string]string{
		"since":         "2024-01-01", // to check: committed_date format (string comparison)
		"minInsertions": "100",
	}
	return cypher, params, migs(mode, isMigration,
		node("Commit", "committed_date"),
		edge("TOUCHED", "insertions"),
		node("Author", "name"),
		node("File", "path"))
}

var traitsGIT4 = Traits{
	Purpose: "Deferred materialization (fetching only message after Top-k) at a small scale",
	Anchor:  AnchorScan,
	Hops:    "1",
	TopK:    10,
	FanOut:  FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.insertions"},
		RoleGroupSort: {"c.insertions", "c.sha"},
		RolePayload:   {"p.message"},
	},
}

// GIT4: output parent commits' texts (message) after the Top-k.
func DefineWorkloadGIT4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("git_log", "GIT4")
	params := map[string]string{
		"minInsertions": "5000", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Commit", "sha", "insertions", "message"))
}
