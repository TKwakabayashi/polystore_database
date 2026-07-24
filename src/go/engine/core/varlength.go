package core

import "fmt"

// VarLengthRange は可変長リレーションのレンジリテラル（Cypher の `*`, `*1..3`, `*0..`,
// `*..3`, `*2`）を返す。max<0 は上限なし、min<0 は下限なしの番兵で、plan.VarLengthExpand.String()
// と同じ規約に従う。
//
// 以前は 3 エンジンが個別に `fmt.Sprintf("*%d..%d", min, max)` を使っており、`*0..`（min=0,
// max=-1 の上限なし）で `*0..-1` という不正 Cypher を生成していた（IS6 が SyntaxError）。
// 生成を 1 本化してこのバグを塞ぐ。
func VarLengthRange(min, max int) string {
	switch {
	case min < 0 && max < 0:
		return "*"
	case max < 0:
		return fmt.Sprintf("*%d..", min)
	case min < 0:
		return fmt.Sprintf("*..%d", max)
	case min == max:
		return fmt.Sprintf("*%d", min)
	default:
		return fmt.Sprintf("*%d..%d", min, max)
	}
}
