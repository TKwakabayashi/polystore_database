package workload

import (
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/data/migrate"
)

// ldbc_snb: workloads on the schema of the LDBC SNB Neo4j dump (datastore/datasets/ldbc_snb_sf1), used to validate the engine
// and in the benchmarks. Queries do not depend on scale (the loaded data is the SF1 dump, dataset ldbc_snb_sf1). Not the
// datagen-standalone ldbc_snb_datagen (ldbc_snb_datagen.go; a different schema).

// ldbcSnbWorkloads are the ldbc_snb workloads (registered in Datasets as LDBC). They have no traits: their
// goldens are the query packages' (ast, logical, physical, native).
var ldbcSnbWorkloads = map[string]Workload{
	"IC2":  {Define: DefineWorkloadIC2},
	"IC8":  {Define: DefineWorkloadIC8},
	"IC9":  {Define: DefineWorkloadIC9},
	"IC11": {Define: DefineWorkloadIC11},
	"IS1":  {Define: DefineWorkloadIS1},
	"IS2":  {Define: DefineWorkloadIS2},
	"IS3":  {Define: DefineWorkloadIS3},
	"IS4":  {Define: DefineWorkloadIS4},
	"IS5":  {Define: DefineWorkloadIS5},
	"IS6":  {Define: DefineWorkloadIS6},
	"AGG1": {Define: DefineWorkloadAGG1},
	"AGG2": {Define: DefineWorkloadAGG2},
	"AGG3": {Define: DefineWorkloadAGG3},
	"AGG4": {Define: DefineWorkloadAGG4},
	"AGG5": {Define: DefineWorkloadAGG5},
	"AGG6": {Define: DefineWorkloadAGG6},
	"TP1":  {Define: DefineWorkloadTP1},
	// Variants of IS6 (custom): several starts and variable-length expansions (boundaries with several starts, weights, deduplication).
	"ISV1": {Define: DefineWorkloadISV1},
	"ISV2": {Define: DefineWorkloadISV2},
	"ISV3": {Define: DefineWorkloadISV3},
	"ISV4": {Define: DefineWorkloadISV4},
	// Experiments (custom): one variable-length traversal in Neo4j with the filters, aggregations and sorts before and after it moved around.
	"VP1": {Define: DefineWorkloadVP1},
	"VP2": {Define: DefineWorkloadVP2},
	"VP3": {Define: DefineWorkloadVP3},
	"VP4": {Define: DefineWorkloadVP4},
	// Workloads that validate the expression features.
	"INLIST":   {Define: DefineWorkloadINLIST},
	"ISNULL":   {Define: DefineWorkloadISNULL},
	"VARCMP":   {Define: DefineWorkloadVARCMP},
	"MULTIPAT": {Define: DefineWorkloadMULTIPAT},
	"PATHFILT": {Define: DefineWorkloadPATHFILT},
	// Official LDBC BI reads expressible with the basic grammar.
	"BI3": {Define: DefineWorkloadBI3},
	"BI9": {Define: DefineWorkloadBI9},
	// Official reads with WITH / OPTIONAL / CASE / collect / UNWIND / negation / cycles / shortestPath.
	"IS7":  {Define: DefineWorkloadIS7},
	"IC1":  {Define: DefineWorkloadIC1},
	"IC3":  {Define: DefineWorkloadIC3},
	"IC4":  {Define: DefineWorkloadIC4},
	"IC6":  {Define: DefineWorkloadIC6},
	"IC12": {Define: DefineWorkloadIC12},
	"BI5":  {Define: DefineWorkloadBI5},
	"BI6":  {Define: DefineWorkloadBI6},
	"BI7":  {Define: DefineWorkloadBI7},
	"BI8":  {Define: DefineWorkloadBI8},
	"BI11": {Define: DefineWorkloadBI11},
	"BI12": {Define: DefineWorkloadBI12},
	"BI14": {Define: DefineWorkloadBI14},
	"BI18": {Define: DefineWorkloadBI18},
	// The rest of the official LDBC queries (BI10 needs APOC, BI15/19/20 need GDS).
	"IC5":  {Define: DefineWorkloadIC5},
	"IC7":  {Define: DefineWorkloadIC7},
	"IC10": {Define: DefineWorkloadIC10},
	"IC13": {Define: DefineWorkloadIC13},
	"IC14": {Define: DefineWorkloadIC14},
	"BI1":  {Define: DefineWorkloadBI1},
	"BI2":  {Define: DefineWorkloadBI2},
	"BI4":  {Define: DefineWorkloadBI4},
	"BI10": {Define: DefineWorkloadBI10},
	"BI13": {Define: DefineWorkloadBI13},
	"BI15": {Define: DefineWorkloadBI15},
	"BI16": {Define: DefineWorkloadBI16},
	"BI17": {Define: DefineWorkloadBI17},
	"BI19": {Define: DefineWorkloadBI19},
	"BI20": {Define: DefineWorkloadBI20},
}

// =====================================================================
// Workload definitions return the Cypher, parameters (fixed) and migration configs (no connections);
// migrations run elsewhere with the returned configs.
// =====================================================================

// IC1 (official Interactive Complex 1): friends with a given name (1..3 hops, shortestPath distance) and their
// universities and companies: shortestPath, OPTIONAL, collect and CASE.
func DefineWorkloadIC1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC1")
	params := map[string]string{
		"personId":  "15393162799448",
		"firstName": "John", // TODO: tune to real values
	}
	return cypher, params, nil
}

func DefineWorkloadIC2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC2")
	params := map[string]string{
		"personId": "15393162799448",
		"maxDate":  "2011-06-16T00:00:00.000Z",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIC8(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC8")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Comment", Properties: []string{"creationDate", "id", "content"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIC9(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC9")
	params := map[string]string{
		// A person in the current dump: KNOWS*1..2 reaches about 760k HAS_CREATOR messages (a high-degree node); maxDate
		// narrows them to about 72k.
		"personId": "17592186049378",
		"maxDate":  "2011-06-16T00:00:00.000Z",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"creationDate", "id", "content", "imageFile"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIC11(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC11")
	params := map[string]string{
		"personId":     "15393162799448",
		"countryName":  "Germany",
		"workFromYear": "2008",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "browserUsed", "creationDate", "email", "gender", "locationIP", "speaks"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Organisation", Properties: []string{"name"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Place", Properties: []string{"name"}, Mode: mode},
			{ObjType: catalog.ObjRelationship, Entity: "WORK_AT", Properties: []string{"workFrom"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS1")
	params := map[string]string{
		"personId": "21990232558284",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "birthday", "gender", "locationIP", "browserUsed", "creationDate"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Place", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS2")
	params := map[string]string{
		"personId": "21990232558284", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS3")
	params := map[string]string{
		"personId": "21990232558284", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: catalog.ObjRelationship, Entity: "KNOWS", Properties: []string{"creationDate"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS4")
	params := map[string]string{
		"messageId": "1030792151051", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration { // authoritative
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"creationDate", "content", "imageFile"}},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS5(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS5")
	params := map[string]string{
		"messageId": "1030792151051", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG1: aggregation (count) with implicit GROUP BY, ORDER BY on the aggregate alias and LIMIT: replies per author of
// comments to a person's messages (like LDBC IC).
func DefineWorkloadAGG1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG1")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG2: grouped aggregation (count, avg, max) with ORDER BY on aggregate aliases and LIMIT: messages, and average and
// maximum length per friend. Writing :Message matches Comment and Post through multiple labels, and properties
// resolve on the Message entity (type is NULL, so no type condition).
func DefineWorkloadAGG2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG2")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG3: a global aggregation (no GROUP BY) covering count(*), count(DISTINCT ...), sum, avg, min and max.
func DefineWorkloadAGG3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG3")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG4: range filters >= / <= (no aggregation) with ORDER BY and LIMIT, on the numeric m.length (creationDate was
// stored as strings in Neo4j, which did not match the mapping's datetime type).
func DefineWorkloadAGG4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG4")
	params := map[string]string{
		"personId": "15393162799448",
		"minLen":   "50",
		"maxLen":   "200",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG5: a global aggregation over one entity without traversal (single-store absorption). With Organisation.id moved,
// MySQL, MongoDB, ClickHouse and Cassandra all aggregate natively.
func DefineWorkloadAGG5(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG5")
	params := map[string]string{
		"minId": "100",
		"maxId": "5000",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Organisation", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// TP1: replies per comment author collected by a Neo4j traversal, grouped by firstName / lastName. The start filter
// uses Person.id (kept in Neo4j) and only Person.{firstName,lastName} are read afterwards, so moving them to one store
// keeps the traversal in Neo4j and puts only the tail on that store (the tail can be measured A/B). AGG1 groups by
// author.id, so moving id would move the start filter too.
func DefineWorkloadTP1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "TP1")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG6: a grouped aggregation without traversal (GROUP BY o.type) with ORDER BY. MySQL, MongoDB and ClickHouse group
// natively; Cassandra (CQL) cannot, so the coordinator does.
func DefineWorkloadAGG6(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "AGG6")
	params := map[string]string{}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Organisation", Properties: []string{"type", "id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS6(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS6")
	params := map[string]string{
		"messageId": "1030792151051", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{ // derived: to check
			{ObjType: catalog.ObjEntity, Entity: "Forum", Properties: []string{"id", "title"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// =====================================================================
// Workloads for the expression features, each exercising one feature minimally.
// Values should match the dump's distribution (to avoid empty results).
// =====================================================================

// INLIST: WHERE ... IN [list] (a single-alias predicate), pushable to every store. Organisation.type has two values
// (Company / University), so the result is never empty.
func DefineWorkloadINLIST(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "INLIST")
	params := map[string]string{}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Organisation", Properties: []string{"id", "type"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// ISNULL: WHERE ... IS NOT NULL (a single-alias predicate). Cassandra cannot evaluate it, so the coordinator does.
// Posts have content NULL and comments have content, so IS NOT NULL keeps messages with a text.
func DefineWorkloadISNULL(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "ISNULL")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "content"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// VARCMP: a comparison between variables a.id < b.id, evaluated by the coordinator: friend pairs where the other id
// is larger (the usual way to drop symmetric duplicates).
func DefineWorkloadVARCMP(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "VARCMP")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// MULTIPAT: a MATCH with several comma-separated patterns sharing friend (acyclic); logically a linear chain, it tests
// parsing and joining comma-separated patterns.
func DefineWorkloadMULTIPAT(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "MULTIPAT")
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Place", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// PATHFILT: predicates on a whole variable-length path: the element predicate all(r IN relationships(p) WHERE ...)
// and the global length(p) <= 3.
func DefineWorkloadPATHFILT(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "PATHFILT")
	params := map[string]string{
		"personId": "17592186049378", // the high-degree Person of Q9 (KNOWS*1..2 reaches many)
		"minDate":  "2010-01-01T00:00:00.000Z",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// BI3: official LDBC BI Q3, popular topics (forums) of a tag class in a country: one linear path
// (country <- city <- Person <- Forum -> Post <- *Message -> Tag -> TagClass) and a single-level grouping with
// count(DISTINCT). Expressible with the basic grammar. City and Country are Place (with type) in the catalog, so
// :Place is used (cf. Q11). TODO(values): check that $country / $tagClass exist in the loaded dump.
func DefineWorkloadBI3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI3")
	params := map[string]string{
		"country":  "Germany",       // TODO: tune to real values
		"tagClass": "MusicalArtist", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Forum", Properties: []string{"id", "title", "creationDate"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// BI9: official LDBC BI Q9, the most active thread starters in a period: one linear path (Person <- Post <- *Message),
// a date range filter and two count(DISTINCT) in one aggregation. It compares creationDate ranges, so placements
// outside Neo4j depend on matching types (datetime vs string). Post / Comment are Message in the catalog.
// TODO(values): fit $startDate / $endDate to the loaded dump's distribution.
func DefineWorkloadBI9(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI9")
	params := map[string]string{
		"startDate": "2010-01-01T00:00:00.000Z", // TODO: tune to real values
		"endDate":   "2013-01-01T00:00:00.000Z", // TODO: tune to real values
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "creationDate"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// =====================================================================
// Official LDBC reads with WITH / OPTIONAL / CASE / collect / negation / cycles. Parameter values should exist in
// the loaded dump (to avoid empty results). Source: docs/LDBC_WORKLOADS.md.
// =====================================================================

// IS7 (official Interactive Short 7): replies to a message, with a CASE on whether the replier KNOWS the author.
// OPTIONAL MATCH and CASE.
func DefineWorkloadIS7(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IS7")
	params := map[string]string{
		"messageId": "1030792151051", // the existing message of IS6 (TODO: check for an id with replies)
	}
	return cypher, params, nil // Neo4j only (no other placement)
}

// IC3 (official Interactive Complex 3): friends (1..2 hops) who posted in two countries, counted per country.
// Multi-step WITH, collect, negation (NOT ... IN cities), CASE, sum, chained comparison.
func DefineWorkloadIC3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC3")
	params := map[string]string{
		"personId":     "15393162799448",
		"countryXName": "Germany", // TODO: tune to real values
		"countryYName": "China",   // TODO: tune to real values
		"startDate":    "2010-01-01T00:00:00.000Z",
		"endDate":      "2011-01-01T00:00:00.000Z",
	}
	return cypher, params, nil
}

// IC4 (official Interactive Complex 4): tags of friends' posts first used in a period, counted.
// Multi-step WITH, CASE, sum, chained date comparison.
func DefineWorkloadIC4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC4")
	params := map[string]string{
		"personId":  "15393162799448",
		"startDate": "2010-10-01T00:00:00.000Z", // TODO: tune to real values
		"endDate":   "2010-11-01T00:00:00.000Z", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI7 (official BI 7): "related tags" of comments replying to messages with a tag.
// Negated pattern NOT (comment)-[:HAS_TAG]->(tag) and several patterns.
func DefineWorkloadBI7(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI7")
	params := map[string]string{
		"tag": "Genghis_Khan", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI11 (official BI 11): triangles of three people in a country who all KNOW each other (cyclic pattern, comparison between variables).
func DefineWorkloadBI11(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI11")
	params := map[string]string{
		"country":   "Germany",                  // TODO: tune to real values
		"startDate": "2010-01-01T00:00:00.000Z", // TODO: tune to real values
		"endDate":   "2013-01-01T00:00:00.000Z", // TODO: tune to real values
	}
	return cypher, params, nil
}

// IC6 (official Interactive Complex 6): other tags co-occurring on friends' (1..2 hops) posts with a given tag.
// WITH, collect, UNWIND, several patterns, negation.
func DefineWorkloadIC6(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC6")
	params := map[string]string{
		"personId": "15393162799448",
		"tagName":  "Genghis_Khan", // TODO: tune to real values
	}
	return cypher, params, nil
}

// IC12 (official Interactive Complex 12): comments by friends replying to posts tagged under a tag class, counted.
// Variable length over alternated types [:HAS_TYPE|IS_SUBCLASS_OF*0..], collect, toInteger.
func DefineWorkloadIC12(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC12")
	params := map[string]string{
		"personId":     "15393162799448",
		"tagClassName": "Monarch", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI5 (official BI 5): the most active posters of a tag (weighted score of messages, replies and likes).
// OPTIONAL with LIKES / REPLY_OF and an arithmetic score; no temporals. Missing LIKES give likeCount 0 (no error).
func DefineWorkloadBI5(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI5")
	params := map[string]string{"tag": "Abbas_I_of_Persia"} // TODO: fit to real values
	return cypher, params, nil
}

// BI6 (official BI 6): the most authoritative users of a tag (chains of likes on replies). Two OPTIONAL with LIKES; no temporals.
func DefineWorkloadBI6(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI6")
	params := map[string]string{"tag": "Arnold_Schwarzenegger"} // TODO: fit to real values
	return cypher, params, nil
}

// BI8 (official BI 8): central people of a tag (interest and posts in a period, plus friends' scores).
// OPTIONAL, collect, UNWIND, size of a list comprehension, date string comparison. HAS_INTEREST exists.
func DefineWorkloadBI8(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI8")
	params := map[string]string{
		"tag":       "Che_Guevara",              // TODO: tune to real values
		"startDate": "2011-07-20T00:00:00.000Z", // TODO: tune to real values
		"endDate":   "2011-07-25T00:00:00.000Z", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI12 (official BI 12): distribution of people per message count (with text, length, period and language filters).
// OPTIONAL, IS NOT NULL, date string comparison, IN list (languages inline, no list parameter).
func DefineWorkloadBI12(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI12")
	params := map[string]string{
		"languages":       `["ar", "hu"]`, // lists are JSON arrays
		"lengthThreshold": "20",
		"startDate":       "2010-07-22T00:00:00.000Z", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI14 (official BI 14): the most interacting pairs between two countries (top per city).
// Four OPTIONAL, CASE, LIKES, indexing collected lists; no temporals. Schema adjusted: Country/City -> Place{type}.
func DefineWorkloadBI14(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI14")
	params := map[string]string{
		"country1": "Chile",     // TODO: tune to real values
		"country2": "Argentina", // TODO: tune to real values
	}
	return cypher, params, nil
}

// BI18 (official BI 18): unconnected pairs interested in a tag and their common friends (friend recommendation).
// Several patterns, a negated pattern, inequality between variables; no temporals.
func DefineWorkloadBI18(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI18")
	params := map[string]string{"tag": "Frank_Sinatra"} // TODO: fit to real values
	return cypher, params, nil
}

// =====================================================================
// The rest of the official LDBC queries (IC5/7/10/13/14, BI1/2/4/10/13/15/16/17/19/20). Sources:
// ldbc_snb_interactive_v1_impls/cypher/queries and ldbc_snb_bi/neo4j/queries (BI19/20 without precomputation).
// The dump holds datetime/date, so the originals' epoch-millisecond arithmetic is replaced by datetime arithmetic.
// Parameters are official examples that exist in the dump (person 14/16/66, ...).
// =====================================================================

// IC5 (official Interactive Complex 5): forums friends and friends of friends joined recently, and their post counts there.
// Adjusted: HAS_MEMBER's joinDate is creationDate in the dump.
func DefineWorkloadIC5(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC5")
	params := map[string]string{
		"personId": "66",
		"minDate":  "2012-01-01T00:00:00.000Z",
	}
	return cypher, params, nil
}

// IC7 (official Interactive Complex 7): recent likers of one's messages and the latency (minutes).
// Adjusted: latency via duration.inSeconds(...).minutes; a negated pattern expression becomes NOT EXISTS { }.
func DefineWorkloadIC7(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC7")
	params := map[string]string{"personId": "66"}
	return cypher, params, nil
}

// IC10 (official Interactive Complex 10): friends of friends with close birthdays, recommended by shared interests.
// Adjusted: birthday is a date, so the epoch-millisecond conversion is dropped.
func DefineWorkloadIC10(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC10")
	params := map[string]string{
		"personId": "66",
		"month":    "5",
	}
	return cypher, params, nil
}

// IC13 (official Interactive Complex 13): shortest KNOWS path length between two people (shortestPath).
func DefineWorkloadIC13(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC13")
	params := map[string]string{
		"person1Id": "14",
		"person2Id": "66",
	}
	return cypher, params, nil
}

// IC14 (official Interactive Complex 14): all shortest paths and their weights from reply interactions (allShortestPaths).
// Adjusted: the original weight scans every comment with an unbound pattern comprehension; it is rewritten with both edge ends bound (same values).
func DefineWorkloadIC14(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "IC14")
	params := map[string]string{
		"person1Id": "14",
		"person2Id": "66",
	}
	return cypher, params, nil
}

// BI1 (official BI 1): message counts per year, kind and length class (full scan with label tests).
func DefineWorkloadBI1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI1")
	params := map[string]string{"datetime": "2011-12-01T00:00:00.000Z"}
	return cypher, params, nil
}

// BI2 (official BI 2): per tag of a tag class, the difference in message counts between two consecutive 100-day windows (duration).
func DefineWorkloadBI2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI2")
	params := map[string]string{
		"date":     "2012-06-01T00:00:00.000Z",
		"tagClass": "MusicalArtist",
	}
	return cypher, params, nil
}

// BI4 (official BI 4): post counts of members of the top 100 forums by members per country (CALL subqueries, UNION ALL, REPLY_OF*0..).
func DefineWorkloadBI4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI4")
	params := map[string]string{"date": "2010-01-29T00:00:00.000Z"}
	return cypher, params, nil
}

// BI10 (official BI 10): tag counts of people 3-4 hops away living in a country who posted on a tag class (APOC subgraphNodes).
func DefineWorkloadBI10(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI10")
	params := map[string]string{
		"personId":        "14",
		"country":         "China",
		"tagClass":        "MusicalArtist",
		"minPathDistance": "3",
		"maxPathDistance": "4",
	}
	return cypher, params, nil
}

// BI13 (official BI 13): "zombies" of a country (fewer than one post per month) and the share of likes from zombies.
func DefineWorkloadBI13(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI13")
	params := map[string]string{
		"country": "France",
		"endDate": "2013-01-01T00:00:00.000Z",
	}
	return cypher, params, nil
}

// BI15 (official BI 15): shortest paths over KNOWS weighted by replies in forums of a period (needs GDS; the oracle fails without it).
func DefineWorkloadBI15(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI15")
	params := map[string]string{
		"person1Id": "14",
		"person2Id": "16",
		"startDate": "2010-11-01T00:00:00.000Z",
		"endDate":   "2010-12-01T00:00:00.000Z",
	}
	return cypher, params, nil
}

// BI16 (official BI 16): people who posted on two (tag, date) pairs with few friends doing the same (CALL subqueries, UNWIND of maps).
func DefineWorkloadBI16(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI16")
	params := map[string]string{
		"tagA":          "Meryl_Streep",
		"dateA":         "2012-09-15T00:00:00.000Z", // the official example date gives 0 rows on this dump; a day with posters of both tags
		"tagB":          "Hank_Williams",
		"dateB":         "2012-11-26T00:00:00.000Z",
		"maxKnowsLimit": "4",
	}
	return cypher, params, nil
}

// BI17 (official BI 17): information of a tag reaching another forum later (two REPLY_OF*0.., a negated pattern, duration).
func DefineWorkloadBI17(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI17")
	params := map[string]string{
		"tag":   "Never_Gonna_Give_You_Up", // the official example Slavoj_Žižek (4,412 messages) takes minutes on Neo4j
		"delta": "4",
	}
	return cypher, params, nil
}

// BI19 (official BI 19, without precomputation): shortest paths between residents of two cities weighted by replies (needs GDS).
func DefineWorkloadBI19(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI19")
	params := map[string]string{
		"city1Id": "669",
		"city2Id": "648",
	}
	return cypher, params, nil
}

// BI20 (official BI 20, without precomputation): shortest paths from a company's employees to a person weighted by alumni year gaps (needs GDS).
func DefineWorkloadBI20(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "BI20")
	params := map[string]string{
		"company":   "Falcon_Air",
		"person2Id": "66",
	}
	return cypher, params, nil
}

// ISV1 (a variant of IS6): from messages in a time window follow reply chains to the root post and return the forum and moderator.
// The window width sets the number of starts; each start gives exactly one row (a post is its own root at 0 hops).
// The label is not Comment: store tables are per Message, so Comment would make Neo4j scan every comment and pass the uuids.
func DefineWorkloadISV1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "ISV1")
	params := map[string]string{
		"start": "2012-11-01T00:00:00.000Z",
		"end":   "2012-11-02T00:00:00.000Z",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"creationDate"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Forum", Properties: []string{"id", "title"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// ISV2 (a variant of IS6): from everyone with a given name follow KNOWS twice and return the distinct home cities.
// The starts are not one row, so WITH DISTINCT cannot be absorbed into the Neo4j query (when the start is outside Neo4j).
// Jun gives 133 starts, 586,484 paths, 9,430 distinct people.
func DefineWorkloadISV2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "ISV2")
	params := map[string]string{"name": "Jun"}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: catalog.ObjEntity, Entity: "Place", Properties: []string{"type", "name"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// ISV3 (a variant of ISV2): from everyone with a given name follow KNOWS twice, filter the ends by birthday and project.
// No DISTINCT or aggregation: the traversal can stay in Neo4j and the filter and projection go to the store with the columns.
func DefineWorkloadISV3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "ISV3")
	params := map[string]string{"name": "Deepak", "start": "1985-01-01", "end": "1986-01-01"}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "birthday"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// ISV4 (a variant of ISV1): from messages in a time window follow reply chains to the root post, filter by the post's
// language and project. Like ISV3, no DISTINCT or aggregation.
func DefineWorkloadISV4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "ISV4")
	params := map[string]string{
		"start":    "2012-11-01T00:00:00.000Z",
		"end":      "2012-11-02T00:00:00.000Z",
		"language": "zh",
	}
	var migs []migrate.MigrationConfig
	if isMigration {
		migs = []migrate.MigrationConfig{
			{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"id", "length", "creationDate", "language"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// VP1-VP4: starts filtered by id IN, one KNOWS*1..2 traversal in Neo4j, and the later filters, aggregations and sorts
// moved around. The default starts are 11 people every 1000th by id (34,636 paths).
var vpIDs = "[14,2199023262647,6597069770240,10995116278068,13194139540794,17592186048307,21990232556371,24189255818910,28587302327203,32985348834784,35184372097588]"

// VP1: σ → T → σ → π。
func DefineWorkloadVP1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "VP1")
	params := map[string]string{"ids": vpIDs, "start": "1985-01-01", "end": "1986-01-01"}
	return cypher, params, vpMigrations(mode, isMigration, "id", "firstName", "lastName", "birthday")
}

// VP2: σ -> T -> σ -> γ -> π; the traversal goes up to HAS_CREATOR.
func DefineWorkloadVP2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "VP2")
	params := map[string]string{"ids": vpIDs, "start": "2012-10-01T00:00:00.000Z", "end": "2012-11-01T00:00:00.000Z"}
	migs := vpMigrations(mode, isMigration, "id", "lastName")
	if isMigration {
		migs = append(migs, migrate.MigrationConfig{ObjType: catalog.ObjEntity, Entity: "Message", Properties: []string{"length", "creationDate"}, Mode: mode})
	}
	return cypher, params, migs
}

// VP3: σ → T → σ → Sort/Limit → π。
func DefineWorkloadVP3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "VP3")
	params := map[string]string{"ids": vpIDs, "start": "1985-01-01", "end": "1986-01-01"}
	return cypher, params, vpMigrations(mode, isMigration, "id", "firstName", "lastName", "birthday")
}

// VP4: σ -> T -> Distinct -> σ -> π. Deduplication in Neo4j, then a filter that barely filters (a birthday range) in the store with the column.
func DefineWorkloadVP4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("ldbc_snb", "VP4")
	params := map[string]string{"ids": vpIDs, "start": "1981-01-01", "end": "1990-01-01"}
	return cypher, params, vpMigrations(mode, isMigration, "id", "firstName", "lastName", "birthday")
}

func vpMigrations(mode migrate.MigrationMode, isMigration bool, props ...string) []migrate.MigrationConfig {
	if !isMigration {
		return nil
	}
	return []migrate.MigrationConfig{{ObjType: catalog.ObjEntity, Entity: "Person", Properties: props, Mode: mode}}
}
