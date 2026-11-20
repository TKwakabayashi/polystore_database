package workload

import "polystore_database/src/go/data/migrate"

// amazon_review2023: a user -> item bipartite graph (23M users, 67M reviews); ratings and review texts are edge properties.
// Low degree, large scale; paired with movielens (same shape, high degree).

var amazonReview2023Workloads = map[string]Workload{
	"AMZ1": {DefineWorkloadAMZ1, traitsAMZ1},
	"AMZ2": {DefineWorkloadAMZ2, traitsAMZ2},
	"AMZ3": {DefineWorkloadAMZ3, traitsAMZ3},
}

var traitsAMZ1 = Traits{
	Purpose:   "The shape of ML3 at low degree and large scale, to separate the effects of degree and scale",
	Anchor:    AnchorScan,
	Hops:      "1",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"i.main_category"},
		RoleEdge:      {"r.rating"},
		RoleGroupSort: {"i.parent_asin", "r.rating"},
	},
}

// AMZ1: review count and average rating per item of a category (the shape of ML3).
func DefineWorkloadAMZ1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("amazon_review2023", "AMZ1")
	params := map[string]string{
		"category":  "All Beauty", // to check: the actual main_category value
		"minRating": "4",
	}
	return cypher, params, migs(mode, isMigration,
		node("Item", "main_category", "parent_asin"),
		edge("REVIEWED", "rating"))
}

var traitsAMZ2 = Traits{
	Purpose:   "The shape of ML2 at low degree and large scale: how the cost of an edge condition changes with degree",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"i.rating_number"},
		RoleEdge:      {"r1.helpful_vote", "r2.rating"},
		RoleGroupSort: {"j.parent_asin"},
	},
}

// AMZ2: items that users who rated a popular item with helpful reviews also rated highly (the shape of ML2, low degree).
func DefineWorkloadAMZ2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("amazon_review2023", "AMZ2")
	params := map[string]string{
		"minRatings": "50000", // to check: a value giving tens of start items
		"minHelpful": "10",
		"minRating":  "5",
	}
	return cypher, params, migs(mode, isMigration,
		node("Item", "rating_number", "parent_asin"),
		edge("REVIEWED", "helpful_vote", "rating"))
}

var traitsAMZ3 = Traits{
	Purpose: "Fetching review text on edges after Top-k (deferred materialization of edge properties)",
	Anchor:  AnchorScan,
	Hops:    "1",
	TopK:    10,
	FanOut:  FanMid,
	Roles: map[Role][]string{
		RoleOutput:    {"i.title"},
		RoleAnchor:    {"i.average_rating"},
		RoleEdge:      {"r.helpful_vote"},
		RoleGroupSort: {"r.helpful_vote", "i.title"},
		RolePayload:   {"r.text", "r.title"},
	},
}

// AMZ3: helpful reviews of low-rated items, with their texts output after the Top-k (edge text).
func DefineWorkloadAMZ3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("amazon_review2023", "AMZ3")
	params := map[string]string{
		"maxAvgRating": "2",
		"minHelpful":   "100", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Item", "average_rating", "title"),
		edge("REVIEWED", "helpful_vote", "title", "text"))
}
