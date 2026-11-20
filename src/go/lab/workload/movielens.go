package workload

import "polystore_database/src/go/data/migrate"

// movielens: a user -> movie bipartite graph (160k users, 25M ratings); ratings (rating, timestamp) are edge
// properties (the main R5 case). High degree per user; paired with amazon_review2023 (same shape, low degree, larger).
// user_id / movie_id are digit-only strings, so start keys are literals in the Cypher (movie_id '1' is Toy Story,
// user_id '1' the first user of ml-25m).

var movielensWorkloads = map[string]Workload{
	"ML1": {DefineWorkloadML1, traitsML1},
	"ML2": {DefineWorkloadML2, traitsML2},
	"ML3": {DefineWorkloadML3, traitsML3},
	"ML4": {DefineWorkloadML4, traitsML4},
}

var traitsML1 = Traits{
	Purpose:   "A high-degree bipartite graph: the cost of edge uuids crossing stores when the edge condition is off graph",
	Anchor:    AnchorUniqueKey,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleAnchor:    {"u.user_id"},
		RoleEdge:      {"r1.rating", "r2.rating"},
		RoleGroupSort: {"v.user_id"},
	},
}

// ML1: users who rated the same movies highly as a user (2 hops, both edges filtered by rating).
func DefineWorkloadML1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("movielens", "ML1")
	params := map[string]string{
		"minRating": "4.5",
	}
	return cypher, params, migs(mode, isMigration,
		node("User", "user_id"),
		edge("RATED", "rating"))
}

var traitsML2 = Traits{
	Purpose:   "Two hops with exploding fan-out: where placing the edge and end conditions starts to pay off",
	Anchor:    AnchorUniqueKey,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanExplosive,
	Roles: map[Role][]string{
		RoleAnchor:    {"m.movie_id"},
		RoleEndpoint:  {"o.year"},
		RoleEdge:      {"r.rating", "r.timestamp", "r2.rating"},
		RoleGroupSort: {"o.title"},
	},
}

// ML2: new movies rated highly by users who rated a movie highly (2 hops, a very large fan-out).
func DefineWorkloadML2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("movielens", "ML2")
	params := map[string]string{
		"minRating": "5",
		"since":     "1500000000", // 2017-07 (seconds). to check: controls the number of start users
		"minYear":   "2010",
	}
	return cypher, params, migs(mode, isMigration,
		node("Movie", "movie_id", "year", "title"),
		edge("RATED", "rating", "timestamp"))
}

var traitsML3 = Traits{
	Purpose:   "One hop aggregating an edge property: absorption when the aggregated value is on the edge",
	Anchor:    AnchorScan,
	Hops:      "1",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"m.year"},
		RoleGroupSort: {"m.title", "r.rating"},
	},
}

// ML3: rating count and average per movie of a year (aggregating edge properties).
func DefineWorkloadML3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("movielens", "ML3")
	params := map[string]string{
		"year": "1995",
	}
	return cypher, params, migs(mode, isMigration,
		node("Movie", "year", "title"),
		edge("RATED", "rating"))
}

var traitsML4 = Traits{
	Purpose:   "No vertex condition, only an edge condition: the cost of plans starting from an edge scan",
	Anchor:    AnchorEdge,
	Hops:      "1",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleEdge:      {"r.timestamp", "r.rating"},
		RoleGroupSort: {"m.movie_id"},
	},
}

// ML4: filtering on edges only (no condition on the start): edges chosen by period and rating, counted per movie.
func DefineWorkloadML4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("movielens", "ML4")
	params := map[string]string{
		"since":  "1560000000", // 2019-06 (seconds). to check
		"rating": "0.5",
	}
	return cypher, params, migs(mode, isMigration,
		node("Movie", "movie_id"),
		edge("RATED", "timestamp", "rating"))
}
