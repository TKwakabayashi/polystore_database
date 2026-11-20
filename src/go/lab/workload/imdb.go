package workload

import "polystore_database/src/go/data/migrate"

// imdb: 7.14M Person -> 12.66M Title (91M WORKED_ON). Two hops explode the fan-out (the break-even of R3).
// nconst / tconst are strings with letters, so they can be parameters.

var imdbWorkloads = map[string]Workload{
	"IMDB1": {DefineWorkloadIMDB1, traitsIMDB1},
	"IMDB2": {DefineWorkloadIMDB2, traitsIMDB2},
	"IMDB3": {DefineWorkloadIMDB3, traitsIMDB3},
}

var traitsIMDB1 = Traits{
	Purpose:   "Two hops from a unique-key start through high-degree titles: the trade-off of placing the middle condition",
	Anchor:    AnchorUniqueKey,
	Hops:      "2",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleAnchor:    {"p.nconst"},
		RoleInterior:  {"t.title_type", "t.start_year", "t.average_rating"},
		RoleGroupSort: {"q.primary_name"},
	},
}

// IMDB1: a person's co-workers (2 hops), with the titles in between filtered by type, year and rating. nconst 'nm0000138' (Leonardo DiCaprio).
func DefineWorkloadIMDB1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("imdb", "IMDB1")
	params := map[string]string{
		"person":    "nm0000138",
		"type":      "movie",
		"fromYear":  "2000",
		"minRating": "7",
	}
	return cypher, params, migs(mode, isMigration,
		node("Person", "nconst", "primary_name"),
		node("Title", "title_type", "start_year", "average_rating"))
}

var traitsIMDB2 = Traits{
	Purpose:   "Three hops with exploding fan-out: the extreme case where a middle condition off graph sends every intermediate row across stores",
	Anchor:    AnchorUniqueKey,
	Hops:      "3",
	Aggregate: true,
	TopK:      20,
	FanOut:    FanExplosive,
	Roles: map[Role][]string{
		RoleAnchor:    {"p.nconst"},
		RoleEndpoint:  {"u.num_votes"},
		RoleInterior:  {"t.start_year", "q.birth_year"},
		RoleGroupSort: {"u.primary_title"},
	},
}

// IMDB2: titles of co-workers (3 hops; the fan-out explodes), with interior titles and people filtered.
func DefineWorkloadIMDB2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("imdb", "IMDB2")
	params := map[string]string{
		"person":    "nm0000138",
		"fromYear":  "2010",
		"bornAfter": "1980",
		"minVotes":  "100000",
	}
	return cypher, params, migs(mode, isMigration,
		node("Person", "nconst", "birth_year"),
		node("Title", "start_year", "num_votes", "primary_title"))
}

var traitsIMDB3 = Traits{
	Purpose:   "Two hops from a unique-key start plus a per-year aggregation: the end condition and aggregation absorption",
	Anchor:    AnchorUniqueKey,
	Hops:      "2",
	Aggregate: true,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleAnchor:    {"t.tconst"},
		RoleEndpoint:  {"o.title_type"},
		RoleGroupSort: {"o.start_year", "o.average_rating"},
	},
}

// IMDB3: other titles of the people of a title, aggregated per year. tconst 'tt0111161' (The Shawshank Redemption).
func DefineWorkloadIMDB3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("imdb", "IMDB3")
	params := map[string]string{
		"title": "tt0111161",
		"type":  "movie",
	}
	return cypher, params, migs(mode, isMigration,
		node("Title", "tconst", "title_type", "start_year", "average_rating"))
}
