package workload

import "polystore_database/src/go/data/migrate"

// stack_exchange: posts of a Q&A site (240k Post, 370k User): a shallow question -> answer tree (PARENT_OF) and links
// between posts (LINKS_TO). The control for stack_history (same posts plus edit revisions and their chain). stackQA1-4
// are shared with stack_history's SH1-4, so only the chain and store size differ. post_type 1 = question, 2 = answer
// (the Stack Exchange dump's convention).

var stackExchangeWorkloads = map[string]Workload{
	"SE1": {DefineWorkloadSE1, traitsStackQA1},
	"SE2": {DefineWorkloadSE2, traitsStackQA2},
	"SE3": {DefineWorkloadSE3, traitsStackQA3},
	"SE4": {DefineWorkloadSE4, traitsStackQA4},
}

var traitsStackQA1 = Traits{
	Purpose:   "Two hops with start, middle and end conditions: how placing the middle condition changes the boundaries",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"q.view_count"},
		RoleEndpoint:  {"u.reputation"},
		RoleInterior:  {"a.score"},
		RoleGroupSort: {"u.display_name"},
	},
}

// stackQA1: high-scoring answers to much-viewed questions, counted per high-reputation user.
func stackQA1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_exchange", "SE1")
	params := map[string]string{
		"minViews":      "10000", // to check
		"minScore":      "10",
		"minReputation": "1000",
	}
	return cypher, params, migs(mode, isMigration,
		node("Post", "view_count", "score"),
		node("User", "reputation", "display_name"))
}

var traitsStackQA2 = Traits{
	Purpose: "One hop question -> answer: fetching body after Top-k (stack_history has much more text)",
	Anchor:  AnchorScan,
	Hops:    "1",
	TopK:    10,
	FanOut:  FanLow,
	Roles: map[Role][]string{
		RoleOutput:    {"q.post_id"},
		RoleAnchor:    {"q.score"},
		RoleGroupSort: {"a.score", "a.post_id"},
		RolePayload:   {"a.body"},
	},
}

// stackQA2: answers to high-scoring questions by score, with bodies output after the Top-k.
func stackQA2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_exchange", "SE2")
	params := map[string]string{
		"minScore": "100", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Post", "post_id", "score", "body"))
}

var traitsStackQA3 = Traits{
	Purpose:   "Two hops over sparse links: placing the middle condition (the small fan-out control)",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"q.score"},
		RoleInterior:  {"m.view_count"},
		RoleGroupSort: {"r.post_id"},
	},
}

// stackQA3: two hops over links, with interior posts filtered by views.
func stackQA3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_exchange", "SE3")
	params := map[string]string{
		"minScore": "5",    // to check
		"minViews": "1000", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Post", "post_id", "score", "view_count"))
}

var traitsStackQA4 = Traits{
	Purpose:   "One hop plus an aggregation: whether it is absorbed when the aggregated value is off graph",
	Anchor:    AnchorScan,
	Hops:      "1",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"u.reputation"},
		RoleGroupSort: {"u.user_id", "p.score", "p.view_count"},
	},
}

// stackQA4: posts aggregated per user (one expansion plus aggregation).
func stackQA4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("stack_exchange", "SE4")
	params := map[string]string{
		"minReputation": "10000", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("User", "user_id", "reputation"),
		node("Post", "score", "view_count"))
}

func DefineWorkloadSE1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA1(mode, isMigration)
}

func DefineWorkloadSE2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA2(mode, isMigration)
}

func DefineWorkloadSE3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA3(mode, isMigration)
}

func DefineWorkloadSE4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	return stackQA4(mode, isMigration)
}
