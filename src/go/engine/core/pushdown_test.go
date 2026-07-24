package core

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"polystore_database/src/go/plan"
)

// sampleGroupPushdown は「GROUP BY o.type して count(*)、id>=100 で絞り、cnt DESC、LIMIT 5」を
// 表す StorePushdown。SQL/CQL は Items を、Mongo は GroupKeys/Aggs を見るため両方を埋める。
func sampleGroupPushdown() *plan.StorePushdown {
	return &plan.StorePushdown{
		Table: "Organisation",
		Items: []plan.ReturnItem{
			{Name: "o.type", Props: []string{"type"}},
			{Name: "cnt", IsAggregate: true, Agg: &plan.AggregateItem{Func: plan.AggCount, OutName: "cnt"}},
		},
		Filters: []*plan.ConditionNode{
			{Property: "id", Type: plan.CondGreaterEq, Value: "100", DataType: "long"},
		},
		GroupKeys:  []plan.GroupKey{{Alias: "o", Prop: "type", OutName: "o.type"}},
		Aggs:       []plan.AggregateItem{{Func: plan.AggCount, OutName: "cnt"}},
		OrderItems: []plan.OrderItem{{Key: "cnt", Direction: plan.OrderDesc}},
		Limit:      5,
	}
}

func TestBuildRelationalSQL(t *testing.T) {
	q, args := BuildRelationalSQL(sampleGroupPushdown())
	want := "SELECT `type` AS `o.type`, COUNT(*) AS `cnt` FROM `Organisation` WHERE `id` >= ? GROUP BY `type` ORDER BY `cnt` DESC LIMIT 5"
	if q != want {
		t.Errorf("SQL mismatch:\n got: %s\nwant: %s", q, want)
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if v, ok := args[0].(int64); !ok || v != 100 {
		t.Errorf("arg[0] = %#v, want int64(100)", args[0])
	}
}

func TestBuildColumnarCQL(t *testing.T) {
	q, args := BuildColumnarCQL(sampleGroupPushdown())
	// CQL は集約項目のみ選択し（Items index 1 → c1）、末尾に ALLOW FILTERING を付ける。
	want := `SELECT count(*) AS c1 FROM "Organisation" WHERE "id" >= ? ALLOW FILTERING`
	if q != want {
		t.Errorf("CQL mismatch:\n got: %s\nwant: %s", q, want)
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if v, ok := args[0].(int64); !ok || v != 100 {
		t.Errorf("arg[0] = %#v, want int64(100)", args[0])
	}
}

// stageKeys はパイプライン各ステージの先頭キー（$match/$group/$sort/$limit 等）を返す。
func stageKeys(p mongo.Pipeline) []string {
	out := make([]string, 0, len(p))
	for _, stage := range p {
		if len(stage) > 0 {
			out = append(out, stage[0].Key)
		}
	}
	return out
}

func TestBuildMongoPipelineStages(t *testing.T) {
	p := BuildMongoPipeline(sampleGroupPushdown())
	got := stageKeys(p)
	want := []string{"$match", "$group", "$sort", "$limit"}
	if len(got) != len(want) {
		t.Fatalf("stage count = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stage[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

// DISTINCT 集約は $group で $addToSet し、後段の $addFields で $size を取る。
func TestBuildMongoPipelineDistinctAddsFields(t *testing.T) {
	o := &plan.StorePushdown{
		Table:     "Message",
		GroupKeys: []plan.GroupKey{{Alias: "p", Prop: "id", OutName: "p.id"}},
		Aggs:      []plan.AggregateItem{{Func: plan.AggCount, Prop: "id", Distinct: true, OutName: "cnt"}},
	}
	p := BuildMongoPipeline(o)
	got := stageKeys(p)
	want := []string{"$group", "$addFields"}
	if len(got) != len(want) {
		t.Fatalf("stage keys = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stage[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestTypeParams(t *testing.T) {
	out := TypeParams(map[string]string{
		"num":  "5",
		"date": "2011-06-16T00:00:00Z",
		"str":  "Germany",
	})
	if n, ok := out["num"].(int); !ok || n != 5 {
		t.Errorf("num = %#v, want int 5", out["num"])
	}
	if _, ok := out["date"].(time.Time); !ok {
		t.Errorf("date = %#v, want time.Time", out["date"])
	}
	if s, ok := out["str"].(string); !ok || s != "Germany" {
		t.Errorf("str = %#v, want string Germany", out["str"])
	}
}

func TestCoerceScalar(t *testing.T) {
	if v := CoerceScalar([]byte("42")); v != int64(42) {
		t.Errorf("CoerceScalar([]byte(42)) = %#v, want int64(42)", v)
	}
	if v := CoerceScalar([]byte("3.5")); v != 3.5 {
		t.Errorf("CoerceScalar([]byte(3.5)) = %#v, want 3.5", v)
	}
	if v := CoerceScalar([]byte("hi")); v != "hi" {
		t.Errorf("CoerceScalar([]byte(hi)) = %#v, want \"hi\"", v)
	}
	if v := CoerceScalar(int64(7)); v != int64(7) {
		t.Errorf("CoerceScalar(int64(7)) = %#v, want int64(7)", v)
	}
}
