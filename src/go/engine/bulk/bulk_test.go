package bulk

import (
	"testing"

	"polystore_database/src/go/plan"
)

func TestCompareValues(t *testing.T) {
	if compareValues(1, 2) >= 0 {
		t.Errorf("compareValues(1,2) は負であるべき")
	}
	if compareValues("b", "a") <= 0 {
		t.Errorf("compareValues(\"b\",\"a\") は正であるべき")
	}
	if compareValues(nil, 1) != -1 {
		t.Errorf("compareValues(nil,1) は -1 であるべき")
	}
	if compareValues(nil, nil) != 0 {
		t.Errorf("compareValues(nil,nil) は 0 であるべき")
	}
}

func TestToFloat64(t *testing.T) {
	if toFloat64(int32(5)) != 5 || toFloat64(int64(7)) != 7 || toFloat64("3.5") != 3.5 {
		t.Errorf("toFloat64 の変換が不正")
	}
}

func TestShapeValueAggregate(t *testing.T) {
	item := plan.ReturnItem{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{OutName: "cnt"}}
	r := Row{"cnt": int64(42)}
	if shapeValue(item, r) != int64(42) {
		t.Errorf("shapeValue(aggregate) = %v, want 42", shapeValue(item, r))
	}
}

func TestBulkAggregateGroupCount(t *testing.T) {
	o := &plan.Aggregate{
		GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "country", OutName: "p.country"}},
		Aggs:      []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}}, // count(*)
	}
	in := []Row{
		{"p.country": "JP"},
		{"p.country": "JP"},
		{"p.country": "US"},
	}
	out := bulkAggregate(o, in)
	if len(out) != 2 {
		t.Fatalf("group 数 = %d, want 2", len(out))
	}
	got := map[string]int64{}
	for _, r := range out {
		c, _ := r["p.country"].(string)
		got[c], _ = r["cnt"].(int64)
	}
	if got["JP"] != 2 || got["US"] != 1 {
		t.Errorf("集約結果が不正: %v", got)
	}
}
