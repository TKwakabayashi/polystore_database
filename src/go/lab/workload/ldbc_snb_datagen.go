package workload

import "polystore_database/src/go/data/migrate"

// ldbc_snb_datagen: LDBC SNB generated with datagen-standalone (Person, Forum, Post, Comment; sf1/sf3/sf10/sf30).
// The same queries at different scales show how cross-store costs grow with scale. The schema differs from ldbc_snb's
// (Post/Comment instead of Message, string keys and dates, snake_case). browser and gender values follow
// LDBC's generator (Firefox / Chrome / Internet Explorer / Safari / Opera, male / female).

var ldbcSnbDatagenWorkloads = map[string]Workload{
	"SNB1": {DefineWorkloadSNB1, traitsSNB1},
	"SNB2": {DefineWorkloadSNB2, traitsSNB2},
	"SNB3": {DefineWorkloadSNB3, traitsSNB3},
	"SNB4": {DefineWorkloadSNB4, traitsSNB4},
}

var traitsSNB1 = Traits{
	Purpose: "A mixed query with start, middle and end conditions and a text output, run at several scales (sf1 to sf30)",
	Anchor:  AnchorScan,
	Hops:    "2",
	TopK:    20,
	FanOut:  FanMid,
	Roles: map[Role][]string{
		RoleOutput:    {"f.person_id"},
		RoleAnchor:    {"p.first_name"},
		RoleEndpoint:  {"m.creation_date"},
		RoleInterior:  {"f.browser"},
		RoleGroupSort: {"m.creation_date", "m.post_id"},
		RolePayload:   {"m.content"},
	},
}

// SNB1: posts of acquaintances of people with a name, newest first, with texts output after the Top-k.
func DefineWorkloadSNB1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb_datagen", "SNB1")
	params := map[string]string{
		"firstName": "Jun", // to check: a name giving a few to tens of starts
		"browser":   "Firefox",
		"since":     "2012-01-01", // to check: creation_date format (string comparison)
	}
	return cypher, params, migs(mode, isMigration,
		node("Person", "first_name", "browser", "person_id"),
		node("Post", "post_id", "creation_date", "content"))
}

var traitsSNB2 = Traits{
	Purpose:   "A whole-path condition on reply chains: how the cost of extracting paths grows with scale",
	Anchor:    AnchorScan,
	Hops:      "1..4",
	VarLength: true,
	Aggregate: true,
	TopK:      20,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.creation_date"},
		RolePath:      {"x.length"},
		RoleGroupSort: {"root.comment_id"},
	},
}

// SNB2: reply chains of 1-4 hops where every comment is at least a length, counted per root comment.
func DefineWorkloadSNB2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb_datagen", "SNB2")
	params := map[string]string{
		"since":     "2012-12-01", // to check
		"minLength": "50",
	}
	return cypher, params, migs(mode, isMigration,
		node("Comment", "creation_date", "length", "comment_id"))
}

var traitsSNB3 = Traits{
	Purpose:   "Two hops plus an aggregation: where to place the end condition and the aggregated value, across scales",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"fo.creation_date"},
		RoleEndpoint:  {"c.length"},
		RoleGroupSort: {"fo.title", "m.length"},
	},
}

// SNB3: forum -> post -> reply aggregation (2 hops plus aggregation).
func DefineWorkloadSNB3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb_datagen", "SNB3")
	params := map[string]string{
		"since":     "2012-11-01", // to check
		"minLength": "100",
	}
	return cypher, params, migs(mode, isMigration,
		node("Forum", "creation_date", "title"),
		node("Post", "length"),
		node("Comment", "length"))
}

var traitsSNB4 = Traits{
	Purpose:   "Two hops through high-degree forum members: the trade-off of placing the middle condition, across scales",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleAnchor:    {"a.gender", "a.birthday"},
		RoleInterior:  {"fo.creation_date"},
		RoleGroupSort: {"mod.person_id", "a.person_id"},
	},
}

// SNB4: member -> forum <- moderator, with forums filtered by creation date and members counted per moderator.
func DefineWorkloadSNB4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb_datagen", "SNB4")
	params := map[string]string{
		"gender":    "female",
		"bornAfter": "1995-01-01", // to check
		"since":     "2012-06-01", // to check
	}
	return cypher, params, migs(mode, isMigration,
		node("Person", "gender", "birthday", "person_id"),
		node("Forum", "creation_date"))
}
