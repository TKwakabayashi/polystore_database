package opbench

import (
	"encoding/json"
	"fmt"
	"math"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/schema"
)

// 計画で使う変数名。入力の頂点は a、展開の行き先は b。
const (
	inAlias  = "a"
	outAlias = "b"
	// barrierVar は at=coordinator のとき対象の直前に置く Unwind([1]) の列名。Unwind は常にコーディネータで
	// 実行されて行を変えない（1 要素のリスト）ので、その後ろの対象もストアへ統合・吸収されなくなる。
	barrierVar = "_barrier"
)

// BuildLogical は実験 e の条件 c の論理計画を組む: 葉（Argument(a) か EntityScan(a)）→ 前段 → 対象。
// 根が対象の演算子になる（物理計画でも根のステップが対象を含む）。
func BuildLogical(e *Experiment, c Condition) (logical.Node, error) {
	var leaf logical.Node = &logical.Argument{Cols: []string{inAlias}}
	if c.Input == "scan" {
		leaf = &logical.EntityScan{Alias: inAlias, Labels: []string{e.Label}}
	}
	get := func(in logical.Node, prop string) logical.Node {
		return &logical.GetProperty{Input: in, Alias: inAlias, Prop: prop, Kind: logical.NodeObj, Labels: []string{e.Label}}
	}
	col := func(prop string) logical.Expr { return &logical.Column{Key: inAlias + "." + prop} }
	barrier := func(in logical.Node) logical.Node {
		if c.At != "coordinator" {
			return in
		}
		return &logical.Unwind{Input: in, List: &logical.List{Elems: []logical.Expr{&logical.Lit{Val: int64(1)}}}, Var: barrierVar}
	}

	switch e.Op {
	case "EntityScan":
		if e.Cmp == "" {
			return leaf, nil
		}
		pred, err := e.predicate(col(e.Prop))
		if err != nil {
			return nil, err
		}
		return &logical.Filter{Input: get(leaf, e.Prop), Pred: pred}, nil

	case "GetProperty":
		return get(leaf, e.Prop), nil

	case "Expand":
		return &logical.Expand{Input: leaf, From: inAlias, Rel: "r", To: outAlias, Types: []string{e.Rel},
			Dir: direction(e.Dir), ToLabels: labels(e.To)}, nil

	case "VarLengthExpand":
		// max < 0 は上限なし（logical.HopRange と同じ）。
		return &logical.VarLengthExpand{Input: leaf, From: inAlias, Seg: "s", To: outAlias, Types: []string{e.Rel},
			Dir: direction(e.Dir), ToLabels: labels(e.To), Range: logical.HopRange{Min: e.Min, Max: e.Max}}, nil

	case "Filter":
		pred, err := e.predicate(col(e.Prop))
		if err != nil {
			return nil, err
		}
		return &logical.Filter{Input: barrier(get(leaf, e.Prop)), Pred: pred}, nil

	case "Aggregation":
		in := leaf
		agg := &logical.Aggregation{}
		if e.Prop != "" {
			in = get(in, e.Prop)
			agg.Keys = []logical.ProjItem{{Name: "key", Expr: col(e.Prop)}}
		}
		item := logical.AggItem{Name: "value", Fn: e.Agg}
		if e.AggProp != "" {
			if e.AggProp != e.Prop {
				in = get(in, e.AggProp)
			}
			item.Arg = col(e.AggProp)
		}
		agg.Aggs = []logical.AggItem{item}
		agg.Input = barrier(in)
		return agg, nil

	case "Sort":
		return &logical.Sort{Input: barrier(get(leaf, e.Prop)), Keys: []logical.SortKey{{Expr: col(e.Prop), Desc: e.Desc}}}, nil

	case "Limit":
		in := leaf
		if e.Prop != "" {
			in = get(in, e.Prop)
		}
		return &logical.Limit{Input: barrier(in), Count: e.Count}, nil

	case "Distinct":
		proj := &logical.Projection{Input: get(leaf, e.Prop), Items: []logical.ProjItem{{Name: "value", Expr: col(e.Prop)}}}
		return &logical.Distinct{Input: barrier(proj)}, nil
	}
	return nil, fmt.Errorf("opbench: 未対応の op %q", e.Op)
}

// predicate は「列 cmp 値」の述語（IN は値の配列）。
func (e *Experiment) predicate(left logical.Expr) (logical.Expr, error) {
	switch e.Cmp {
	case "=", "<>", "<", ">", "<=", ">=":
		return &logical.Bin{Op: e.Cmp, L: left, R: &logical.Lit{Val: jsonValue(e.Value)}}, nil
	case "IN":
		vs, ok := e.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("IN の value は配列: %v", e.Value)
		}
		l := &logical.List{}
		for _, v := range vs {
			l.Elems = append(l.Elems, &logical.Lit{Val: jsonValue(v)})
		}
		return &logical.Bin{Op: "IN", L: left, R: l}, nil
	}
	return nil, fmt.Errorf("未対応の cmp %q", e.Cmp)
}

// jsonValue は JSON の数を整数なら int64 にする（リテラルの型付けは物理計画が mapping の型で行う）。
func jsonValue(v any) any {
	if f, ok := v.(float64); ok && f == math.Trunc(f) && math.Abs(f) < 1<<53 {
		return int64(f)
	}
	return v
}

func direction(d string) logical.Direction {
	switch d {
	case "in":
		return logical.In
	case "both":
		return logical.Both
	}
	return logical.Out
}

func labels(l string) []string {
	if l == "" {
		return nil
	}
	return []string{l}
}

// Props は実験が使う（配置を動かす対象の）プロパティ。
func (e *Experiment) Props() []string {
	var out []string
	for _, p := range []string{e.Prop, e.AggProp} {
		if p != "" && !contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// PlaceMapping は base の写しで、実験のプロパティだけを store に置いた mapping を返す（graph なら base のまま）。
// 統計（Stats）は引き継ぐ。
func PlaceMapping(base *schema.MappingDictionary, e *Experiment, store string) (*schema.MappingDictionary, error) {
	if store == "graph" || len(e.Props()) == 0 {
		return base, nil
	}
	b, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	var md schema.MappingDictionary
	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}
	md.Stats = base.Stats
	ent := entityOf(&md, e.Label)
	if ent == nil {
		return nil, fmt.Errorf("mapping に %s のエンティティが無い", e.Label)
	}
	for _, p := range e.Props() {
		pm, ok := ent.Properties[p]
		if !ok {
			return nil, fmt.Errorf("mapping に %s.%s が無い", e.Label, p)
		}
		pm.Databases = []string{store}
	}
	return &md, nil
}

// entityOf はラベル（サブラベルを含む）のエンティティ（Post → Message）。
func entityOf(md *schema.MappingDictionary, label string) *schema.EntityMapping {
	if ent, ok := md.Entities[label]; ok {
		return ent
	}
	for _, ent := range md.Entities {
		if contains(ent.Labels, label) {
			return ent
		}
	}
	return nil
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
