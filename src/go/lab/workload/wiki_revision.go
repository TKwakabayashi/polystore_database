package workload

import "polystore_database/src/go/data/migrate"

// wiki_revision: Wikipedia edit history (33k Page, 2.5M Revision, 310k Contributor). Revision chains (PREV) run
// thousands deep and full texts total 17.5GB: the large version of git_log's "deep chain plus text" (R4, R7, smaller
// Neo4j). The boolean column (Revision.minor) cannot be compared in the grammar, so it is not used in conditions.

var wikiRevisionWorkloads = map[string]Workload{
	"WIKI1": {DefineWorkloadWIKI1, traitsWIKI1},
	"WIKI2": {DefineWorkloadWIKI2, traitsWIKI2},
	"WIKI3": {DefineWorkloadWIKI3, traitsWIKI3},
	"WIKI4": {DefineWorkloadWIKI4, traitsWIKI4},
}

var traitsWIKI1 = Traits{
	Purpose:   "A whole-path condition over up to 10 hops of chains thousands long: a larger GIT2, the cost of extracting paths",
	Anchor:    AnchorScan,
	Hops:      "1..10",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"r.text_bytes"},
		RolePath:      {"x.text_bytes"},
		RoleGroupSort: {"o.page_id"},
	},
}

// WIKI1: from large revisions follow 1-10 previous revisions; chains where every revision is at least a size, counted per page.
func DefineWorkloadWIKI1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("wiki_revision", "WIKI1")
	params := map[string]string{
		"minBytes":      "500000", // to check: a value giving tens of starts
		"minChainBytes": "100000", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Revision", "text_bytes", "page_id"))
}

var traitsWIKI2 = Traits{
	Purpose: "Fetching full text (17.5GB in total) after Top-k, and the effect of a smaller graph, at their strongest",
	Anchor:  AnchorScan,
	Hops:    "2",
	TopK:    10,
	FanOut:  FanHigh,
	Roles: map[Role][]string{
		RoleOutput:    {"p.title", "c.name"},
		RoleAnchor:    {"p.namespace"},
		RoleInterior:  {"r.timestamp"},
		RoleGroupSort: {"r.text_bytes", "r.rev_id"},
		RolePayload:   {"r.text"},
	},
}

// WIKI2: an article's revisions in a period, largest first, with full texts output after the Top-k.
func DefineWorkloadWIKI2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("wiki_revision", "WIKI2")
	params := map[string]string{
		"namespace": "0",
		"since":     "2023-01-01", // to check: timestamp format (string comparison)
	}
	return cypher, params, migs(mode, isMigration,
		node("Page", "namespace", "title"),
		node("Revision", "timestamp", "text_bytes", "rev_id", "text"),
		node("Contributor", "name"))
}

var traitsWIKI3 = Traits{
	Purpose:   "Two hops narrowed only by a middle version: the cost when the middle condition is off graph and no start is fixed",
	Anchor:    AnchorInterior,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleInterior:  {"r.timestamp", "r.text_bytes"},
		RoleGroupSort: {"c.name", "p.page_id"},
	},
}

// WIKI3: contributor -> revision <- page, with interior revisions filtered by period and size, pages counted per contributor.
func DefineWorkloadWIKI3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("wiki_revision", "WIKI3")
	params := map[string]string{
		"since":    "2023-01-01", // to check
		"minBytes": "100000",
	}
	return cypher, params, migs(mode, isMigration,
		node("Revision", "timestamp", "text_bytes"),
		node("Contributor", "name"),
		node("Page", "page_id"))
}

var traitsWIKI4 = Traits{
	Purpose:   "A condition on the vertex after a variable-length chain (middle): how chain expansion and boundaries combine",
	Anchor:    AnchorScan,
	Hops:      "2..4",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"r.text_bytes"},
		RoleInterior:  {"o.timestamp"},
		RoleGroupSort: {"d.name"},
	},
}

// WIKI4: contributors of revisions 1-3 before large revisions (the later vertices of the chain filtered by period).
func DefineWorkloadWIKI4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("wiki_revision", "WIKI4")
	params := map[string]string{
		"minBytes": "500000",     // to check
		"since":    "2020-01-01", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Revision", "text_bytes", "timestamp"),
		node("Contributor", "name"))
}
