package core

import (
	"testing"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// 純投影フラグメント（scan + filter + projection）→ GROUP BY なしの SELECT。
func TestLowerFragmentProjectionOnly(t *testing.T) {
	frag := &plan.StoreFragment{
		Store: store.Relational,
		Plan: &plan.Projection{
			Units: []plan.ProjectionUnit{{
				Alias:   "p",
				Fetches: []plan.FetchPlan{{Store: store.Relational, Props: []string{"firstName", "lastName"}}},
			}},
			Input: &plan.EntityScan{
				Alias: "p", Labels: []string{"Person"}, DataStore: store.Relational,
				Filter: []*plan.ConditionNode{
					{Type: plan.CondEq, Property: "id", Value: "42", DataType: "int", DataStore: store.Relational},
				},
			},
		},
	}

	spec := LowerFragment(frag)
	if spec.Table != "Person" {
		t.Errorf("Table = %q, want Person", spec.Table)
	}
	if len(spec.Projections) != 2 || len(spec.Aggs) != 0 || len(spec.Filters) != 1 {
		t.Fatalf("spec = %+v", spec)
	}

	sql, args := BuildRelationalSQL(spec)
	want := "SELECT `firstName` AS `p.firstName`, `lastName` AS `p.lastName` FROM `Person` WHERE `id` = ?"
	if sql != want {
		t.Errorf("SQL:\n got %q\nwant %q", sql, want)
	}
	if len(args) != 1 {
		t.Errorf("args = %v, want 1 element", args)
	}
}

// 集約フラグメント（scan + aggregate）→ GROUP BY 付き SELECT。
func TestLowerFragmentAggregate(t *testing.T) {
	frag := &plan.StoreFragment{
		Store: store.Relational,
		Plan: &plan.Aggregate{
			GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "gender", OutName: "gender"}},
			Aggs:      []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
			Input:     &plan.EntityScan{Alias: "p", Labels: []string{"Person"}, DataStore: store.Relational},
		},
	}

	spec := LowerFragment(frag)
	sql, _ := BuildRelationalSQL(spec)
	want := "SELECT `gender` AS `gender`, COUNT(*) AS `cnt` FROM `Person` GROUP BY `gender`"
	if sql != want {
		t.Errorf("SQL:\n got %q\nwant %q", sql, want)
	}
}
