package ir

import (
	"polystore_database/src/go/base/id"
	"testing"
)

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
		{HopRange{1, 3}, ">", 5, HopRange{6, 3}, true}, // empty range
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
		t.Error("unexpected Empty")
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

// An id set keeps the first appearance of each id.
func TestIDSet(t *testing.T) {
	a, b, c := id.Make([16]byte{1}), id.Make([16]byte{2}), id.Make([16]byte{3})
	s := NewIDSet([]id.UUID{b, a, b})
	if len(s.IDs) != 2 || s.IDs[0] != b || s.IDs[1] != a {
		t.Errorf("IDs = %v", s.IDs)
	}
	if !s.Has(a) || s.Has(c) {
		t.Error("Has")
	}
	var none *IDSet
	if none.Has(a) {
		t.Error("a nil set has nothing")
	}
}

// Transform copies the tree bottom-up and leaves the original unchanged; an element id set does not read a column.
func TestTransformInSet(t *testing.T) {
	orig := &Quant{Kind: "all", Var: "n", List: &Call{Fn: "nodes", Args: []Expr{&Column{Key: "p"}}},
		Pred: &Bin{Op: "AND", L: &Bin{Op: "=", L: &ElemProp{Var: "n", Prop: "g"}, R: &Lit{Val: "m"}}, R: &Lit{Val: true}}}
	before := orig.String()
	got := Transform(orig, func(e Expr) Expr {
		if b, ok := e.(*Bin); ok && b.Op == "=" {
			return &InSet{Var: b.L.(*ElemProp).Var, Elem: true, Set: "set0"}
		}
		return e
	})
	if orig.String() != before || HasSets(orig) {
		t.Errorf("original changed: %s", orig)
	}
	if !HasSets(got) || got.String() != "all(n IN nodes(p) WHERE (n IN $set0 AND true))" {
		t.Errorf("transformed: %s", got)
	}
	if cols := Columns(got); len(cols) != 1 || cols[0] != "p" {
		t.Errorf("Columns = %v", cols)
	}
	if cols := Columns(&InSet{Var: "a", Set: "set0"}); len(cols) != 1 || cols[0] != "a" {
		t.Errorf("Columns of a column id set = %v", cols)
	}
}
