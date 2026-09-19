package core

import (
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// scan + filter のみ（traversal 無し）→ MATCH (n) WHERE ... RETURN n.uuid。
func TestBuildGraphRecordCypherScan(t *testing.T) {
	sub := &plan.EntityScan{
		Alias: "n", Labels: []string{"Person"}, DataStore: store.Graph,
		Filter: []*plan.ConditionNode{
			{Type: plan.CondEq, Alias: "n", Property: "id", Value: "42", DataType: "int"},
		},
	}
	q, params := BuildGraphRecordCypher(sub, []string{"n"})
	want := "MATCH (n) WHERE (n:Person) AND n.id = $val0 RETURN n.uuid AS n"
	if q != want {
		t.Errorf("cypher:\n got %q\nwant %q", q, want)
	}
	if len(params) != 1 {
		t.Errorf("params = %v, want 1", params)
	}
}

// 1-hop 有向 traversal → MATCH (a)-[r:KNOWS]->(b) ... RETURN a.uuid, b.uuid。
func TestBuildGraphRecordCypherExpand(t *testing.T) {
	sub := &plan.Expand{
		Alias: "r", RelLabel: "KNOWS", Dir: plan.Outgoing,
		SourceEntity: "a", TargetEntity: "b", TargetLabels: []string{"Person"},
		Input: &plan.EntityScan{
			Alias: "a", Labels: []string{"Person"}, DataStore: store.Graph,
			Filter: []*plan.ConditionNode{{Type: plan.CondEq, Alias: "a", Property: "id", Value: "1", DataType: "int"}},
		},
	}
	q, _ := BuildGraphRecordCypher(sub, []string{"a", "b"})
	want := "MATCH (a)-[r:KNOWS]->(b) WHERE (a:Person) AND a.id = $val0 AND (b:Person) RETURN a.uuid AS a, b.uuid AS b"
	if q != want {
		t.Errorf("cypher:\n got %q\nwant %q", q, want)
	}
}

// 関係プロパティを projection する場合、関係 alias の uuid も RETURN できる（無向 KNOWS）。
func TestBuildGraphRecordCypherReturnsRelAlias(t *testing.T) {
	sub := &plan.Expand{
		Alias: "r", RelLabel: "KNOWS", Dir: plan.Bidirectional,
		SourceEntity: "p", TargetEntity: "friend", TargetLabels: []string{"Person"},
		Input: &plan.EntityScan{
			Alias: "p", Labels: []string{"Person"}, DataStore: store.Graph,
			Filter: []*plan.ConditionNode{{Type: plan.CondEq, Alias: "p", Property: "id", Value: "7", DataType: "int"}},
		},
	}
	q, _ := BuildGraphRecordCypher(sub, []string{"friend", "r"})
	want := "MATCH (p)-[r:KNOWS]-(friend) WHERE (p:Person) AND p.id = $val0 AND (friend:Person) RETURN friend.uuid AS friend, r.uuid AS r"
	if q != want {
		t.Errorf("cypher:\n got %q\nwant %q", q, want)
	}
}

// 可変長 traversal → *1..2 レンジ。
func TestBuildGraphRecordCypherVarLength(t *testing.T) {
	sub := &plan.VarLengthExpand{
		Alias: "r", RelLabel: "KNOWS", Dir: plan.Outgoing,
		SourceEntity: "a", TargetEntity: "b", MinHops: 1, MaxHops: 2,
		Input: &plan.EntityScan{Alias: "a", Labels: []string{"Person"}, DataStore: store.Graph},
	}
	q, _ := BuildGraphRecordCypher(sub, []string{"b"})
	want := "MATCH (a)-[r:KNOWS*1..2]->(b) WHERE (a:Person) RETURN b.uuid AS b"
	if q != want {
		t.Errorf("cypher:\n got %q\nwant %q", q, want)
	}
}

// 非 graph ノードが混ざる → 空（フォールバック）。
func TestBuildGraphRecordCypherNonGraphBails(t *testing.T) {
	sub := &plan.EntityScan{Alias: "n", Labels: []string{"Person"}, DataStore: store.Relational}
	if q, _ := BuildGraphRecordCypher(sub, []string{"n"}); q != "" {
		t.Errorf("non-graph scan should bail, got %q", q)
	}
}

// RETURN 対象 alias がパターンに無い → 空（フォールバック）。
func TestBuildGraphRecordCypherMissingAliasBails(t *testing.T) {
	sub := &plan.EntityScan{Alias: "n", Labels: []string{"Person"}, DataStore: store.Graph}
	if q, _ := BuildGraphRecordCypher(sub, []string{"missing"}); q != "" {
		t.Errorf("missing alias should bail, got %q", q)
	}
}
