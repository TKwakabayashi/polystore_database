package workload

import "polystore_database/src/go/data/migrate"

// tpch: TPC-H (about SF1) as a graph: Customer -> Order -> Lineitem -> Part / Supplier -> Nation with fixed hops. Each
// level has analytic numbers and dates, so relational strengths appear as traversals (the OLAP-on-graph control).
// Values follow TPC-H's generator (mktsegment, brand, returnflag, nation names; dates are 'YYYY-MM-DD' strings).
// Keys (custkey, ...) are digit-only strings, so start keys are literals in the Cypher.

var tpchWorkloads = map[string]Workload{
	"TPCH1": {DefineWorkloadTPCH1, traitsTPCH1},
	"TPCH2": {DefineWorkloadTPCH2, traitsTPCH2},
	"TPCH3": {DefineWorkloadTPCH3, traitsTPCH3},
	"TPCH4": {DefineWorkloadTPCH4, traitsTPCH4},
}

var traitsTPCH1 = Traits{
	Purpose:   "A fixed two-hop analysis with a condition per hop: how the number of boundaries changes with which hop is relational",
	Anchor:    AnchorScan,
	Hops:      "2",
	Aggregate: true,
	TopK:      10,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.mktsegment"},
		RoleEndpoint:  {"l.shipdate"},
		RoleInterior:  {"o.orderdate"},
		RoleGroupSort: {"o.orderkey", "o.orderdate", "l.extendedprice"},
	},
}

// TPCH1: TPC-H Q3 style (without the discount arithmetic): filters at each level, revenue per order.
func DefineWorkloadTPCH1(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("tpch", "TPCH1")
	params := map[string]string{
		"segment": "BUILDING",
		"date":    "1995-03-15",
	}
	return cypher, params, migs(mode, isMigration,
		node("Customer", "mktsegment"),
		node("Order", "orderkey", "orderdate"),
		node("Lineitem", "shipdate", "extendedprice"))
}

var traitsTPCH2 = Traits{
	Purpose:   "A three-hop analysis with middle conditions on two hops: the number of boundaries and aggregation absorption",
	Anchor:    AnchorScan,
	Hops:      "3",
	Aggregate: true,
	TopK:      10,
	FanOut:    FanMid,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.mktsegment"},
		RoleInterior:  {"o.orderdate", "l.discount"},
		RoleGroupSort: {"p.brand", "l.extendedprice"},
	},
}

// TPCH2: 4 hops (Customer -> Order -> Lineitem -> Part), filtered by period and discount, aggregated per brand.
func DefineWorkloadTPCH2(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("tpch", "TPCH2")
	params := map[string]string{
		"segment":     "AUTOMOBILE",
		"from":        "1994-01-01",
		"to":          "1995-01-01",
		"minDiscount": "0.05",
	}
	return cypher, params, migs(mode, isMigration,
		node("Customer", "mktsegment"),
		node("Order", "orderdate"),
		node("Lineitem", "discount", "extendedprice"),
		node("Part", "brand"))
}

var traitsTPCH3 = Traits{
	Purpose:   "Four hops from one unique-key start: with a small start, boundaries cost round trips",
	Anchor:    AnchorUniqueKey,
	Hops:      "4",
	Aggregate: true,
	FanOut:    FanLow,
	Roles: map[Role][]string{
		RoleAnchor:    {"c.custkey"},
		RoleInterior:  {"l.returnflag"},
		RoleGroupSort: {"n.name", "s.suppkey", "l.quantity"},
	},
}

// TPCH3: from one customer by unique key, 5 hops to the supplier's nation (a deep expansion from one start).
func DefineWorkloadTPCH3(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("tpch", "TPCH3")
	params := map[string]string{
		"flag": "N",
	}
	return cypher, params, migs(mode, isMigration,
		node("Customer", "custkey"),
		node("Lineitem", "returnflag", "quantity"),
		node("Supplier", "suppkey"),
		node("Nation", "name"))
}

var traitsTPCH4 = Traits{
	Purpose:   "A six-hop chain narrowed at both ends: middle conditions and whether to expand from both sides",
	Anchor:    AnchorBothEnds,
	Hops:      "6",
	Aggregate: true,
	FanOut:    FanHigh,
	Roles: map[Role][]string{
		RoleAnchor:    {"cn.name"},
		RoleEndpoint:  {"sn.name"},
		RoleInterior:  {"l.shipdate"},
		RoleGroupSort: {"l.extendedprice"},
	},
}

// TPCH4: TPC-H Q7 style: a 6-hop chain with both ends (customer and supplier nations) filtered by name and the period on the interior Lineitem.
func DefineWorkloadTPCH4(mode migrate.MigrationMode, isMigration bool) (string, map[string]string, []migrate.MigrationConfig) {
	cypher := query("tpch", "TPCH4")
	params := map[string]string{
		"custNation": "GERMANY",
		"suppNation": "FRANCE",
		"from":       "1995-01-01",
		"to":         "1996-12-31",
	}
	return cypher, params, migs(mode, isMigration,
		node("Nation", "name"),
		node("Lineitem", "shipdate", "extendedprice"))
}
