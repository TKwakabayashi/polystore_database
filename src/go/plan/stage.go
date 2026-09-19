package plan

import (
	"fmt"
	"strings"
)

// stage.go は WITH 多段（Phase B2）の境界演算子。
//   - RowFilter: post-WITH の WHERE を Row 世界（実体化値）で評価する（式は Expr、評価は core.EvalExpr）。
//   - RowSource: 前段の Row 群を後段 record パイプラインの Record へ橋渡しする（ステージ境界の source）。
//
// ステージ連鎖の論理形（線形）:
//   Stage1(record)→Projection→[Aggregate/Sort/Limit] → [RowFilter] → RowSource → Stage2(record)→…→Return

// RowFilter は Row 世界の述語フィルタ（post-WITH WHERE）。Cond は Row 上で評価する任意の式。
type RowFilter struct {
	Input PlanNode
	Cond  Expr
}

func (f *RowFilter) Children() []PlanNode { return []PlanNode{f.Input} }
func (f *RowFilter) String() string {
	c := "None"
	if f.Cond != nil {
		c = f.Cond.String()
	}
	return fmt.Sprintf("RowFilter[%s]", c)
}

// RowProjCol は RowProject の 1 列。Expr が非 nil ならそれを入力 Row 上で評価して Out へ、
// nil なら入力 Row のキー Src の値を Out へ写す（単純な列選択/改名）。
type RowProjCol struct {
	Out  string
	Src  string
	Expr Expr // 計算列（WITH の CASE/算術など）。nil なら Src コピー。
}

// RowProject は Row 世界の列選択/改名。WITH ステージの出力を「WITH 項目の出力名」で揃える
// （束縛は alias、値は alias.prop、集約は OutName を入力キーとして読み、出力名へ写す）。
// これにより post-WITH WHERE（RowFilter）・RowSource・最終 RETURN が clean 名で参照できる。
type RowProject struct {
	Input PlanNode
	Cols  []RowProjCol
}

func (p *RowProject) Children() []PlanNode { return []PlanNode{p.Input} }
func (p *RowProject) String() string {
	var cs []string
	for _, c := range p.Cols {
		switch {
		case c.Expr != nil:
			cs = append(cs, c.Out+"←"+c.Expr.String())
		case c.Out == c.Src:
			cs = append(cs, c.Out)
		default:
			cs = append(cs, c.Out+"←"+c.Src)
		}
	}
	return fmt.Sprintf("RowProject[%s]", strings.Join(cs, ", "))
}

// RowSource は Row→Record 橋渡し（ステージ境界の source）。前段 Row の各行から、
//   - BindCols（WITH の束縛項目＝bare alias）の UUID を OutputSlot のスロットへ
//   - ValCols（WITH の値項目＝式/集約/prop の出力名）を Record.Vals へ
//
// 詰めて後段へ流す。Input は前段の Row 生成サブツリー（Projection/Aggregate/…、または RowFilter）。
type RowSource struct {
	Input       PlanNode
	BindCols    []string // Row キー（alias）→ スロット束縛（UUID）
	ValCols     []string // Row キー（出力名）→ Record.Vals
	OutputAlias []string
	OutputSlot  SlotTable
}

func (s *RowSource) Children() []PlanNode { return []PlanNode{s.Input} }
func (s *RowSource) String() string {
	return fmt.Sprintf("RowSource(binds:%s, vals:%s, slot:%v)",
		strings.Join(s.BindCols, ","), strings.Join(s.ValCols, ","), s.OutputSlot.VarToSlot)
}
