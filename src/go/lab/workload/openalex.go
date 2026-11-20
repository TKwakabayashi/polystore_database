package workload

import "polystore_database/src/go/data/migrate"

// openalex: scholarly works (21.91M Work, 9.63M Author) with citation chains (160M CITES) and co-authorship
// (33.55M AUTHORED). Like imdb, 2 hops with a large fan-out; citation chains also exercise whole-path conditions (R4).

var openalexWorkloads = map[string]Workload{
	"OA1": {DefineWorkloadOA1, traitsOA1},
	"OA2": {DefineWorkloadOA2, traitsOA2},
	"OA3": {DefineWorkloadOA3, traitsOA3},
}

var traitsOA1 = Traits{
	Purpose:   "Two hops of citations with a large fan-out: placing the middle condition at the largest scale",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanExplosive,
	Roles: map[Role][]string{
		RoleAnchor:    {"w.year", "w.cited_by_count"},
		RoleInterior:  {"m.year"},
		RoleGroupSort: {"o.title"},
	},
}

// OA1: citations of citations of highly cited works (2 hops), with interior works filtered by year.
func DefineWorkloadOA1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("openalex", "OA1")
	params := map[string]string{
		"year":     "2020",
		"minCited": "1000", // to check: a value giving tens of starts
		"minYear":  "2015",
	}
	return cypher, params, migs(mode, isMigration,
		node("Work", "year", "cited_by_count", "title"))
}

var traitsOA2 = Traits{
	Purpose:   "A whole-path condition on citation chains (many branches): the cost of extracting many paths",
	Anchor:    AnchorScan,
	Hops:      "1..3",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanExplosive,
	Roles: map[Role][]string{
		RoleAnchor:    {"w.cited_by_count"},
		RolePath:      {"x.year"},
		RoleGroupSort: {"o.venue"},
	},
}

// OA2: citation chains of 1-3 hops where every work is from a year on, counted per destination venue.
func DefineWorkloadOA2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("openalex", "OA2")
	params := map[string]string{
		"minCited": "20000", // to check
		"minYear":  "2010",
	}
	return cypher, params, migs(mode, isMigration,
		node("Work", "cited_by_count", "year", "venue"))
}

var traitsOA3 = Traits{
	Purpose:   "Two hops of co-authorship narrowed only by a middle paper, at the largest scale",
	Anchor:    AnchorInterior,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleInterior:  {"w.year", "w.cited_by_count"},
		RoleGroupSort: {"b.author_id"},
	},
}

// OA3: co-authorship (author -> work <- author), with interior works filtered by year and citation count.
func DefineWorkloadOA3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("openalex", "OA3")
	params := map[string]string{
		"year":     "2020",
		"minCited": "500", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Work", "year", "cited_by_count"),
		node("Author", "author_id"))
}
