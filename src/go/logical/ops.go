package logical

import (
	"fmt"
	"strings"
)

// Node は論理演算子。意味だけを持ち、スロット割り当てやストアは持たない（物理計画化で決める）。
type Node interface {
	Children() []Node
	// Describe は演算子 1 個の決定的な 1 行表示（子は含まない）。golden で使う。
	Describe() string
}

// Direction は From → To に対する関係の向き。
type Direction int

const (
	Out  Direction = iota // (From)-[]->(To)
	In                    // (From)<-[]-(To)
	Both                  // (From)-[]-(To)
)

// Reverse は From と To を入れ替えたときの向き。
func (d Direction) Reverse() Direction {
	switch d {
	case Out:
		return In
	case In:
		return Out
	default:
		return Both
	}
}

// Pattern は向きを含む関係パターン片（"-[r:T]->" 等）。
func (d Direction) Pattern(body string) string {
	switch d {
	case Out:
		return "-[" + body + "]->"
	case In:
		return "<-[" + body + "]-"
	default:
		return "-[" + body + "]-"
	}
}

// ObjKind はプロパティの持ち主の種類。
type ObjKind int

const (
	NodeObj ObjKind = iota
	RelObj
)

func (k ObjKind) String() string {
	if k == RelObj {
		return "rel"
	}
	return "node"
}

// HopRange は可変長のホップ範囲。Max < 0 は上限なし。Min は省略時 1。
type HopRange struct {
	Min int
	Max int
}

// Cypher は Cypher の範囲表記（`*1..3`, `*2..`, `*3`）。
func (h HopRange) Cypher() string {
	switch {
	case h.Max < 0:
		return fmt.Sprintf("*%d..", h.Min)
	case h.Min == h.Max:
		return fmt.Sprintf("*%d", h.Min)
	default:
		return fmt.Sprintf("*%d..%d", h.Min, h.Max)
	}
}

// Empty は範囲が空（どの長さも満たさない）か。
func (h HopRange) Empty() bool { return h.Max >= 0 && h.Min > h.Max }

// Constrain は「ホップ数 op k」を範囲へ交わらせる（op は = < <= > >=）。表現できない op（<>）は ok=false。
func (h HopRange) Constrain(op string, k int) (HopRange, bool) {
	lo, hi := h.Min, h.Max
	tighten := func(newLo, newHi int) {
		if newLo > lo {
			lo = newLo
		}
		if newHi >= 0 && (hi < 0 || newHi < hi) {
			hi = newHi
		}
	}
	switch op {
	case "=":
		tighten(k, k)
	case "<":
		tighten(lo, k-1)
	case "<=":
		tighten(lo, k)
	case ">":
		tighten(k+1, -1)
	case ">=":
		tighten(k, -1)
	default:
		return h, false
	}
	return HopRange{Min: lo, Max: hi}, true
}

// SegPart は可変長区間のどの要素を対象にするか。
type SegPart int

const (
	SegNodes SegPart = iota // 区間の中間ノード（端点を含まない）
	SegRels                 // 区間の全関係
)

func (p SegPart) String() string {
	if p == SegRels {
		return "rels"
	}
	return "nodes"
}

// SegPred は可変長区間の要素すべてが満たすべき条件 `all(Var IN part WHERE Pred)`。
// Pred の中で要素のプロパティは ElemProp{Var, …} で参照する。
type SegPred struct {
	Part SegPart
	Var  string
	Pred Expr
}

func (s SegPred) String() string {
	return fmt.Sprintf("all(%s IN %s WHERE %s)", s.Var, s.Part, s.Pred)
}

// ================================
// ストアに触れる演算子（配置の対象）
// ================================

// EntityScan はラベル Labels のエンティティを走査して Alias を束縛する（葉）。
type EntityScan struct {
	Alias  string
	Labels []string
}

// GetProperty は Alias のプロパティ Prop を値列 "Alias.Prop" として足す（左外部結合。無ければ null）。
// Labels / Kind は持ち主（mapping を引くキー）。
type GetProperty struct {
	Input  Node
	Alias  string
	Prop   string
	Kind   ObjKind
	Labels []string
}

// Key は出力する値列のキー。
func (g *GetProperty) Key() string { return g.Alias + "." + g.Prop }

// GetListProperty はリスト列 Source（パス列または区間列）の Part の各要素について Prop を取り、
// 値のリスト列 "Source.part.Prop" を足す。ElemLabels は要素のラベル候補（mapping を引くキー）。
type GetListProperty struct {
	Input      Node
	Source     string
	Part       SegPart
	Prop       string
	ElemLabels []string
}

// Key は出力するリスト値列のキー。
func (g *GetListProperty) Key() string { return ListPropKey(g.Source, g.Part, g.Prop) }

// ListPropKey は GetListProperty の出力キー。
func ListPropKey(source string, part SegPart, prop string) string {
	return source + "." + part.String() + "." + prop
}

// Expand は束縛済みの From から関係をたどり、Rel と To を束縛する（片側束縛）。
type Expand struct {
	Input    Node
	From     string
	Rel      string
	To       string
	Types    []string
	Dir      Direction
	ToLabels []string
}

// ExpandInto は From と To が束縛済みのとき、その間の関係 Rel を束縛する（両端束縛・循環を閉じる）。
type ExpandInto struct {
	Input Node
	From  string
	Rel   string
	To    string
	Types []string
	Dir   Direction
}

// VarLengthExpand は From から Range のホップ数で関係をたどり To を束縛する。経路ごとに 1 行。
// SegPreds は区間の要素に対する条件。EmitPath のとき区間の中身を区間列 SegColumn(Seg) として出す。
type VarLengthExpand struct {
	Input    Node
	From     string
	Seg      string
	To       string
	Types    []string
	Dir      Direction
	ToLabels []string
	Range    HopRange
	SegPreds []SegPred
	EmitPath bool
}

// SegColumn は可変長区間 seg の区間列キー。
func SegColumn(seg string) string { return "_seg_" + seg }

// ================================
// 計算だけの演算子
// ================================

// Filter は Pred が真の行だけを残す（3 値論理）。正規形では AND を分解した 1 項。
type Filter struct {
	Input Node
	Pred  Expr
}

// ProjItem は射影の 1 列。
type ProjItem struct {
	Name string
	Expr Expr
}

// Projection は列の選択・改名・式の評価（WITH / RETURN）。
type Projection struct {
	Input Node
	Items []ProjItem
}

// PathElem はパスの構成要素。Seg が真なら可変長区間（Var は区間名）、偽なら束縛変数（ノードまたは関係）。
type PathElem struct {
	Var string
	Seg bool
}

// ProjectPath は束縛変数と区間列からパス列 Path を組み立てる。
type ProjectPath struct {
	Input Node
	Path  string
	Elems []PathElem
}

// AggItem は集約 1 個。Arg が nil なら count(*)。
type AggItem struct {
	Name     string
	Fn       string
	Distinct bool
	Arg      Expr
}

// Aggregation は Keys でグループ化して Aggs を計算する（暗黙 GROUP BY）。
type Aggregation struct {
	Input Node
	Keys  []ProjItem
	Aggs  []AggItem
}

// SortKey は並べ替えキー。
type SortKey struct {
	Expr Expr
	Desc bool
}

// Sort は Keys の順に並べる。
type Sort struct {
	Input Node
	Keys  []SortKey
}

// Limit は先頭 Skip 行を飛ばして Count 行を残す。Count < 0 は無制限。
type Limit struct {
	Input Node
	Skip  int
	Count int
}

// Distinct は Cols（空なら全列）で重複を除く。
type Distinct struct {
	Input Node
	Cols  []string
}

// NodeHashJoin は 2 入力を識別列 Keys で結合する。Keys が空なら直積。
type NodeHashJoin struct {
	Left  Node
	Right Node
	Keys  []string
}

// RelUniqueFilter は同一 MATCH 内の関係（Rels の識別列と Segs の区間列に含まれる関係）が互いに異なる行だけを残す。
type RelUniqueFilter struct {
	Input Node
	Rels  []string
	Segs  []string
}

// ---- Children ----

func (*EntityScan) Children() []Node        { return nil }
func (n *GetProperty) Children() []Node     { return []Node{n.Input} }
func (n *GetListProperty) Children() []Node { return []Node{n.Input} }
func (n *Expand) Children() []Node          { return []Node{n.Input} }
func (n *ExpandInto) Children() []Node      { return []Node{n.Input} }
func (n *VarLengthExpand) Children() []Node { return []Node{n.Input} }
func (n *Filter) Children() []Node          { return []Node{n.Input} }
func (n *Projection) Children() []Node      { return []Node{n.Input} }
func (n *ProjectPath) Children() []Node     { return []Node{n.Input} }
func (n *Aggregation) Children() []Node     { return []Node{n.Input} }
func (n *Sort) Children() []Node            { return []Node{n.Input} }
func (n *Limit) Children() []Node           { return []Node{n.Input} }
func (n *Distinct) Children() []Node        { return []Node{n.Input} }
func (n *NodeHashJoin) Children() []Node    { return []Node{n.Left, n.Right} }
func (n *RelUniqueFilter) Children() []Node { return []Node{n.Input} }

// ---- Describe ----

func labelSuffix(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	return ":" + strings.Join(ls, "|")
}

func (n *EntityScan) Describe() string {
	return fmt.Sprintf("EntityScan(%s%s)", n.Alias, labelSuffix(n.Labels))
}

func (n *GetProperty) Describe() string {
	return fmt.Sprintf("GetProperty(%s) [%s%s]", n.Key(), n.Kind, labelSuffix(n.Labels))
}

func (n *GetListProperty) Describe() string {
	return fmt.Sprintf("GetListProperty(%s) [elem%s]", n.Key(), labelSuffix(n.ElemLabels))
}

func (n *Expand) Describe() string {
	return fmt.Sprintf("Expand((%s)%s(%s%s))", n.From, n.Dir.Pattern(n.Rel+labelSuffix(n.Types)), n.To, labelSuffix(n.ToLabels))
}

func (n *ExpandInto) Describe() string {
	return fmt.Sprintf("ExpandInto((%s)%s(%s))", n.From, n.Dir.Pattern(n.Rel+labelSuffix(n.Types)), n.To)
}

func (n *VarLengthExpand) Describe() string {
	s := fmt.Sprintf("VarLengthExpand((%s)%s(%s%s))", n.From,
		n.Dir.Pattern(n.Seg+labelSuffix(n.Types)+n.Range.Cypher()), n.To, labelSuffix(n.ToLabels))
	if len(n.SegPreds) > 0 {
		parts := make([]string, len(n.SegPreds))
		for i, p := range n.SegPreds {
			parts[i] = p.String()
		}
		s += " seg[" + strings.Join(parts, " AND ") + "]"
	}
	if n.EmitPath {
		s += " emit " + SegColumn(n.Seg)
	}
	return s
}

func (n *Filter) Describe() string { return fmt.Sprintf("Filter(%s)", n.Pred) }

func projItems(items []ProjItem) string {
	parts := make([]string, len(items))
	for i, it := range items {
		if c, ok := it.Expr.(*Column); ok && c.Key == it.Name {
			parts[i] = it.Name
		} else {
			parts[i] = fmt.Sprintf("%s AS %s", it.Expr, it.Name)
		}
	}
	return strings.Join(parts, ", ")
}

func (n *Projection) Describe() string { return fmt.Sprintf("Projection[%s]", projItems(n.Items)) }

func (n *ProjectPath) Describe() string {
	parts := make([]string, len(n.Elems))
	for i, e := range n.Elems {
		if e.Seg {
			parts[i] = SegColumn(e.Var)
		} else {
			parts[i] = e.Var
		}
	}
	return fmt.Sprintf("ProjectPath(%s = [%s])", n.Path, strings.Join(parts, ", "))
}

func (a AggItem) String() string {
	arg := "*"
	if a.Arg != nil {
		arg = a.Arg.String()
	}
	if a.Distinct {
		arg = "DISTINCT " + arg
	}
	return fmt.Sprintf("%s(%s) AS %s", a.Fn, arg, a.Name)
}

func (n *Aggregation) Describe() string {
	keys := "∅"
	if len(n.Keys) > 0 {
		keys = projItems(n.Keys)
	}
	aggs := make([]string, len(n.Aggs))
	for i, a := range n.Aggs {
		aggs[i] = a.String()
	}
	return fmt.Sprintf("Aggregation[keys: %s | %s]", keys, strings.Join(aggs, ", "))
}

func (n *Sort) Describe() string {
	parts := make([]string, len(n.Keys))
	for i, k := range n.Keys {
		dir := "ASC"
		if k.Desc {
			dir = "DESC"
		}
		parts[i] = k.Expr.String() + " " + dir
	}
	return fmt.Sprintf("Sort[%s]", strings.Join(parts, ", "))
}

func (n *Limit) Describe() string {
	if n.Skip > 0 {
		return fmt.Sprintf("Limit[skip %d, %d]", n.Skip, n.Count)
	}
	return fmt.Sprintf("Limit[%d]", n.Count)
}

func (n *Distinct) Describe() string {
	if len(n.Cols) == 0 {
		return "Distinct"
	}
	return fmt.Sprintf("Distinct[%s]", strings.Join(n.Cols, ", "))
}

func (n *NodeHashJoin) Describe() string {
	if len(n.Keys) == 0 {
		return "NodeHashJoin[cartesian]"
	}
	return fmt.Sprintf("NodeHashJoin[%s]", strings.Join(n.Keys, ", "))
}

func (n *RelUniqueFilter) Describe() string {
	parts := append([]string{}, n.Rels...)
	for _, s := range n.Segs {
		parts = append(parts, SegColumn(s))
	}
	return fmt.Sprintf("RelUniqueFilter[%s]", strings.Join(parts, ", "))
}

// Dump は計画木を決定的に表示する（1 行 1 演算子、子はインデント）。
func Dump(n Node) string {
	var b strings.Builder
	var walk func(Node, int)
	walk = func(n Node, d int) {
		if n == nil {
			return
		}
		b.WriteString(strings.Repeat("  ", d))
		b.WriteString(n.Describe())
		b.WriteByte('\n')
		for _, c := range n.Children() {
			walk(c, d+1)
		}
	}
	walk(n, 0)
	return b.String()
}
