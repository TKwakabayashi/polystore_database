package core

import (
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// stagebridge.go は WITH 多段（Phase B2）のステージ境界変換。RowSource が使う「前段 Row →
// 後段 Record」変換と、Projection が使う「carried Vals → Row」マージを全エンジンで共有する。

// BuildStageRecords は前段の Row 群を後段 record パイプラインの Record へ変換する。
//   - bindCols: WITH の束縛項目（bare alias）。Row[alias]（束縛 uuid 文字列）を outSlot のスロットへ。
//   - valCols:  WITH の値項目（式/集約/prop の出力名）。Row[name] を Record.Vals へ持ち越す。
func BuildStageRecords(rows []map[string]any, bindCols, valCols []string, outSlot plan.SlotTable) []Record {
	slotCount := len(outSlot.VarToSlot)
	out := make([]Record, 0, len(rows))
	for _, row := range rows {
		rec := Record{Slots: make([]id.UUID, slotCount)}
		for _, bc := range bindCols {
			idx, ok := outSlot.VarToSlot[bc]
			if !ok {
				continue
			}
			if v, ok := row[bc]; ok && v != nil {
				rec.Slots[idx] = id.FromAny(v)
			}
		}
		if len(valCols) > 0 {
			rec.Vals = make(map[string]any, len(valCols))
			for _, vc := range valCols {
				rec.Vals[vc] = row[vc]
			}
		}
		out = append(out, rec)
	}
	return out
}

// ProjectRow は WITH ステージ出力を clean 名へ揃える（RowProject 実行の共通ロジック）。
// cols の各 {Out, Src} について out[Out] = row[Src]。全エンジンの Row 世界で共有する。
func ProjectRow(row map[string]any, cols []plan.RowProjCol) map[string]any {
	out := make(map[string]any, len(cols))
	for _, c := range cols {
		if c.Expr != nil {
			v, err := EvalExpr(c.Expr, row)
			if err != nil {
				v = nil // 評価不能は null（保守的）
			}
			out[c.Out] = v
			continue
		}
		out[c.Out] = row[c.Src]
	}
	return out
}

// MergeVals は Record が持ち越す carried 値（Vals）を Row（wide row）へ書き込む。
// Projection が Row 生成時に呼び、WITH 由来のスカラ/計算列を下流へ渡す。
func MergeVals(row map[string]any, vals map[string]any) {
	for k, v := range vals {
		row[k] = v
	}
}
