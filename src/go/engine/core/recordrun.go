package core

import "polystore_database/src/go/plan"

// recordrun.go は「record ラン（隣接同一ストアの scan/filter 連鎖）」を 1 クエリへ畳む lowering。
//
// graph ランは traversal を含むため専用の BuildGraphRecordCypher が担う。ここが扱うのは
// traversal を含まないラン（scan＋同一 alias のフィルタ）で、条件をマージした合成 EntityScan を返す。
// 各エンジンは既存の Scan<Store> 実装をそのまま呼べばよく、ストアごとの新しいクエリビルダは要らない。

// MergeRecordRun は StoreFragment.Plan のランが「EntityScan ＋ 同一 alias・同一ストアの Filter 群」
// なら、全条件をマージした合成 EntityScan を返す。畳めない形なら ok=false（呼び出し側は Plan を通常実行）。
//
// 畳める条件:
//   - ランが EntityScan から始まる（境界入力を持たない source ラン）。
//   - Filter は全て EntityScan と同じ alias・同じストア（別 alias は中間結果の絞り込みで単純合成できない）。
//   - 実際に 2 演算子以上ある（1 演算子なら既存の scan 実行と往復数が同じなので畳む意味がない）。
func MergeRecordRun(f *plan.StoreFragment) (*plan.EntityScan, bool) {
	var chain []plan.PlanNode
	for n := f.Plan; n != nil; {
		chain = append(chain, n)
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	if len(chain) < 2 {
		return nil, false
	}

	scan, ok := chain[len(chain)-1].(*plan.EntityScan)
	if !ok || scan.DataStore != f.Store {
		return nil, false
	}

	merged := *scan // 浅いコピー（元プランを壊さない）
	merged.Filter = append([]*plan.ConditionNode{}, scan.Filter...)

	// 葉→根で Filter を吸収する（scan 自身は除く）。
	for i := len(chain) - 2; i >= 0; i-- {
		fl, ok := chain[i].(*plan.Filter)
		if !ok {
			return nil, false // traversal 等が混ざる → 対象外
		}
		if fl.DataStore != f.Store || fl.Alias != scan.Alias {
			return nil, false
		}
		merged.Filter = append(merged.Filter, fl.Filter...)
		if len(fl.Labels) > 0 {
			merged.Labels = fl.Labels
		}
	}

	// 出力スロットはランの最上位演算子のものに合わせる（下流の期待と一致させる）。
	merged.OutputSlot = f.OutputSlot
	if top, ok := chain[0].(*plan.Filter); ok {
		merged.OutputAlias = top.OutputAlias
	}
	return &merged, true
}
