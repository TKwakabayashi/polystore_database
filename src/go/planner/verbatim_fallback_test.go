package planner_test

import (
	"testing"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/planner"
	"polystore_database/src/go/workload"
)

// verbatim_fallback_test は Phase A の「現行文法で構造化できないクエリ → graph verbatim 委譲」
// フォールバックを DB 無しで検証する。ParseQuery は catalog/mapping.json だけで完結する。

const vfMappingPath = "../../../catalog/mapping.json"

func isGraphVerbatim(n plan.PlanNode) bool {
	f, ok := n.(*plan.StoreFragment)
	return ok && f.Verbatim != ""
}

// Phase A の graph-verbatim ワークロードは verbatim フラグメントへ落ちること。
func TestPhaseAWorkloadsRouteVerbatim(t *testing.T) {
	for _, name := range []string{
		"IS7", "IC1", "IC3", "IC4", "IC6", "IC12",
		"BI5", "BI6", "BI7", "BI8", "BI11", "BI12", "BI14", "BI18"} {
		name := name
		t.Run(name, func(t *testing.T) {
			def, ok := workload.Registry[name]
			if !ok {
				t.Fatalf("workload %s 未登録", name)
			}
			cypher, params, _ := def(migrator.ModeGraphToRdb, false)
			root, err := planner.ParseQuery(cypher, vfMappingPath, params)
			if err != nil {
				t.Fatalf("ParseQuery(%s) err: %v", name, err)
			}
			if !isGraphVerbatim(root) {
				t.Errorf("[%s] verbatim フォールバックされていない: %T", name, root)
			}
		})
	}
}

// 現行文法で構造化できるクエリ（既存/Phase0）は verbatim にならず通常プランを組むこと（回帰防止）。
func TestStructuredQueriesNotVerbatim(t *testing.T) {
	cases := map[string]string{
		"normal":  "MATCH (p:Person {id: $x})<-[:HAS_CREATOR]-(m:Message) RETURN m.id ORDER BY m.id ASC LIMIT 10",
		"in":      "MATCH (o:Organisation) WHERE o.type IN [\"Company\"] RETURN o.id",
		"varcmp":  "MATCH (a:Person {id: $x})-[:KNOWS]-(b:Person) WHERE a.id < b.id RETURN a.id, b.id",
		"pathflt": "MATCH p=(a:Person {id: $x})-[:KNOWS*1..2]-(b:Person) WHERE all(r IN relationships(p) WHERE r.creationDate > $d) RETURN b.id",
	}
	for name, q := range cases {
		root, err := planner.ParseQuery(q, vfMappingPath, map[string]string{"x": "1", "d": "2010-01-01T00:00:00.000Z"})
		if err != nil {
			t.Errorf("[%s] err: %v", name, err)
			continue
		}
		if isGraphVerbatim(root) {
			t.Errorf("[%s] 構造化可能なのに verbatim へフォールバックした", name)
		}
	}
}
