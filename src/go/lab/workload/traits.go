package workload

import (
	"fmt"
	"sort"
	"strings"
)

// Role is a property's role in a query (the main axis of placement experiments; the hypothesis is that it decides the gain or loss of moving the property out of Neo4j).
type Role int

const (
	RoleOutput    Role = iota // R0 output only (deferred materialization fetches only surviving rows)
	RoleAnchor                // R1 start filter (the first query; the boundary passes the start uuids once)
	RoleEndpoint              // R2 end filter (after the expansion; the whole expansion result is passed)
	RoleInterior              // R3 interior vertex filter (splits the Neo4j query; more boundaries)
	RolePath                  // R4 whole-path condition all/none (outside Neo4j the path is extracted and evaluated by the coordinator)
	RoleEdge                  // R5 edge property (edge uuids cross stores)
	RoleGroupSort             // R6 aggregation and sort keys and aggregate arguments (can the store absorb them)
	RolePayload               // R7 heavy texts output after the Top-k (deferred materialization and a smaller Neo4j)
)

var roleNames = [...]string{"R0 output", "R1 start", "R2 end", "R3 middle", "R4 path", "R5 edge", "R6 aggregate / sort", "R7 text"}

func (r Role) String() string {
	if int(r) < len(roleNames) {
		return roleNames[r]
	}
	return fmt.Sprintf("Role(%d)", int(r))
}

// Code is "R0"-"R7" (CSV columns and aggregation keys).
func (r Role) Code() string { return fmt.Sprintf("R%d", int(r)) }

// Anchor is how the pattern's start is chosen (it sets the number of starts, i.e. uuids passed at the first boundary).
type Anchor int

const (
	AnchorUniqueKey Anchor = iota // one by unique-key equality
	AnchorScan                    // a filtered full scan (the threshold sets the number of starts)
	AnchorBothEnds                // both ends filtered (may expand from both sides)
	AnchorInterior                // no condition on the ends; filtered on interior vertices
	AnchorEdge                    // no condition on vertices; filtered on edges
)

var anchorNames = [...]string{"unique key", "filtered scan", "both ends", "narrowed in the middle", "narrowed by edges"}

func (a Anchor) String() string {
	if int(a) < len(anchorNames) {
		return anchorNames[a]
	}
	return fmt.Sprintf("Anchor(%d)", int(a))
}

// FanOut estimates the rows an expansion makes per start (estimated before loading; corrected by measurement).
type FanOut int

const (
	FanLow       FanOut = iota // up to 10 rows (chains, sparse trees)
	FanMid                     // up to 1,000 rows
	FanHigh                    // up to 100k rows
	FanExplosive               // over 100k rows (2+ hops through high-degree vertices)
)

var fanOutNames = [...]string{"low", "medium", "high", "explosive"}

func (f FanOut) String() string {
	if int(f) < len(fanOutNames) {
		return fanOutNames[f]
	}
	return fmt.Sprintf("FanOut(%d)", int(f))
}

// Traits describe a workload, for aggregating experiment results (gain or loss per role, shape and scale) and as
// plan golden headings. VarLength, Aggregate, TopK and the Roles columns are checked against the Cypher by tests.
type Traits struct {
	Purpose   string            // what the query is for
	Anchor    Anchor            // how the start is chosen
	Hops      string            // hops ("2", or min..max with variable length such as "1..5")
	VarLength bool              // has a variable-length expansion
	Aggregate bool              // has an aggregation
	TopK      int               // the LIMIT count (0 if none)
	FanOut    FanOut            // estimated expansion rows per start
	Roles     map[Role][]string // role -> columns ("var.prop"; path conditions use the quantifier variable)
}

// RoleList returns the roles in Roles in order.
func (t Traits) RoleList() []Role {
	rs := make([]Role, 0, len(t.Roles))
	for r := range t.Roles {
		rs = append(rs, r)
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i] < rs[j] })
	return rs
}

// String is a one-line summary (golden headings, logs).
func (t Traits) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "start=%s hops=%s fan-out=%s", t.Anchor, t.Hops, t.FanOut)
	if t.VarLength {
		b.WriteString(" variable-length")
	}
	if t.Aggregate {
		b.WriteString(" aggregate")
	}
	if t.TopK > 0 {
		fmt.Fprintf(&b, " Top-%d", t.TopK)
	}
	for _, r := range t.RoleList() {
		fmt.Fprintf(&b, " | %s: %s", r, strings.Join(t.Roles[r], ", "))
	}
	return b.String()
}

// Workload is one workload of a dataset (definition and traits).
type Workload struct {
	Define workloadDef
	Traits Traits
}
