package logical

import "testing"

func TestHopRangeCypher(t *testing.T) {
	cases := []struct {
		h    HopRange
		want string
	}{
		{HopRange{1, -1}, "*1.."},
		{HopRange{0, -1}, "*0.."},
		{HopRange{2, 2}, "*2"},
		{HopRange{1, 3}, "*1..3"},
	}
	for _, c := range cases {
		if got := c.h.Cypher(); got != c.want {
			t.Errorf("%+v: got %s want %s", c.h, got, c.want)
		}
	}
}

func TestHopRangeConstrain(t *testing.T) {
	cases := []struct {
		h    HopRange
		op   string
		k    int
		want HopRange
		ok   bool
	}{
		{HopRange{1, 3}, "<=", 2, HopRange{1, 2}, true},
		{HopRange{1, 3}, "<", 2, HopRange{1, 1}, true},
		{HopRange{1, -1}, "<=", 4, HopRange{1, 4}, true},
		{HopRange{1, 3}, ">=", 2, HopRange{2, 3}, true},
		{HopRange{1, 3}, ">", 5, HopRange{6, 3}, true}, // 空範囲
		{HopRange{0, -1}, "=", 2, HopRange{2, 2}, true},
		{HopRange{1, 3}, "<>", 2, HopRange{1, 3}, false},
	}
	for _, c := range cases {
		got, ok := c.h.Constrain(c.op, c.k)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("%+v %s %d: got %+v,%v want %+v,%v", c.h, c.op, c.k, got, ok, c.want, c.ok)
		}
	}
	if !(HopRange{6, 3}).Empty() || (HopRange{1, -1}).Empty() {
		t.Error("Empty の判定が想定外")
	}
}

func TestDump(t *testing.T) {
	var n Node = &EntityScan{Alias: "p", Labels: []string{"Person"}}
	n = &GetProperty{Input: n, Alias: "p", Prop: "id", Kind: NodeObj, Labels: []string{"Person"}}
	n = &Filter{Input: n, Pred: &Bin{Op: "=", L: &Column{Key: "p.id"}, R: &Lit{Val: int64(1)}}}
	n = &VarLengthExpand{Input: n, From: "p", Seg: "s", To: "f", Types: []string{"KNOWS"}, Dir: Both,
		Range: HopRange{1, 2}, EmitPath: true,
		SegPreds: []SegPred{{Part: SegNodes, Var: "x", Pred: &Bin{Op: "=", L: &ElemProp{Var: "x", Prop: "gender"}, R: &Lit{Val: "male"}}}}}
	n = &GetListProperty{Input: n, Source: SegColumn("s"), Part: SegNodes, Prop: "gender", ElemLabels: []string{"Person"}}
	n = &Aggregation{Input: n, Keys: []ProjItem{{Name: "f", Expr: &Column{Key: "f"}}},
		Aggs: []AggItem{{Name: "c", Fn: "count"}}}
	n = &Limit{Input: &Sort{Input: n, Keys: []SortKey{{Expr: &Column{Key: "c"}, Desc: true}}}, Count: 3}

	want := `Limit[3]
  Sort[c DESC]
    Aggregation[keys: f | count(*) AS c]
      GetListProperty(_seg_s.nodes.gender) [elem:Person]
        VarLengthExpand((p)-[s:KNOWS*1..2]-(f)) seg[all(x IN nodes WHERE (x.gender = "male"))] emit _seg_s
          Filter((p.id = 1))
            GetProperty(p.id) [node:Person]
              EntityScan(p:Person)
`
	if got := Dump(n); got != want {
		t.Errorf("Dump\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestColumns(t *testing.T) {
	e := &Bin{Op: "AND",
		L: &Bin{Op: "<", L: &Column{Key: "a.id"}, R: &Column{Key: "b.id"}},
		R: &Quant{Kind: "all", Var: "x", List: &Call{Fn: "nodes", Args: []Expr{&Column{Key: "p"}}},
			Pred: &Bin{Op: "=", L: &ElemProp{Var: "x", Prop: "g"}, R: &Column{Key: "a.id"}}}}
	got := Columns(e)
	if len(got) != 3 || got[0] != "a.id" || got[1] != "b.id" || got[2] != "p" {
		t.Errorf("Columns: %v", got)
	}
}
