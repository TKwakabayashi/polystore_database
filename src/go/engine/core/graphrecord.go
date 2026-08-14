package core

import (
	"fmt"
	"strings"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// BuildGraphRecordCypher は graph の record パイプライン部分木（EntityScan + graph Filter +
// Expand/VarLengthExpand）を、束縛 UUID を返す 1 本の Cypher に融合する（部分融合 P3b）。
// 生成: MATCH <path> WHERE <label 条件 AND graph プロパティ filter> RETURN a.uuid AS a, ...
//
// 返り値の params は $valN → 値。outAliases は下流 Projection が必要とする alias 群
// （各 alias の uuid を返す）。構造が非対応（非 graph ノード・非線形 traversal・エイリアス欠落）
// の場合は空文字を返す（呼び出し側＝fusion が安全にコーディネータ木へフォールバックする）。
//
// 規約は既存 access_graph.go（ScanGraphBulk / bulkFilterGraph）に一致させる:
//   - ラベルは WHERE 側に (alias:L1 OR alias:L2) で置く（インライン :L1:L2 は AND になるため不可）。
//   - 比較演算子は core.SQLOp（全 6 種・Cypher 互換）。
//   - 値は codec.ConvertToNativeType で型付けして param 化。
func BuildGraphRecordCypher(sub plan.PlanNode, outAliases []string) (string, map[string]interface{}) {
	// 根→葉へ集めてから葉→根（＝traversal の順）で処理する。
	var chain []plan.PlanNode
	for n := sub; n != nil; {
		chain = append(chain, n)
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}

	var pattern strings.Builder
	placed := map[string]bool{}   // パターンに登場済みの alias
	incoming := map[string]bool{} // 下位フラグメント（境界）から束縛が供給される alias
	last := ""                    // 直近に置いたノード alias（線形連結の検証用）
	var whereParts []string
	params := map[string]interface{}{}
	pi := 0

	// placeIncoming は境界から供給される alias をパターン起点として置き、uuid IN 条件を付ける。
	// 実際の uuid 配列は実行時にエンジンが IncomingParam(alias) へ束縛する。
	placeIncoming := func(alias string) bool {
		if last != "" || !incoming[alias] {
			return false
		}
		pattern.WriteString("(" + alias + ")")
		placed[alias] = true
		last = alias
		whereParts = append(whereParts, fmt.Sprintf("%s.uuid IN $%s", alias, IncomingParam(alias)))
		return true
	}

	addLabelCond := func(alias string, labels []string) {
		if len(labels) == 0 {
			return
		}
		var ors []string
		for _, l := range labels {
			ors = append(ors, fmt.Sprintf("%s:%s", alias, l))
		}
		whereParts = append(whereParts, "("+strings.Join(ors, " OR ")+")")
	}
	addFilters := func(conds []*plan.ConditionNode) {
		for _, c := range conds {
			if c == nil {
				continue
			}
			p := fmt.Sprintf("val%d", pi)
			pi++
			whereParts = append(whereParts, fmt.Sprintf("%s.%s %s $%s", c.Alias, c.Property, SQLOp(c.Type), p))
			params[p], _ = codec.ConvertToNativeType(c.Value, c.DataType)
		}
	}

	// 葉→根で処理。
	for i := len(chain) - 1; i >= 0; i-- {
		switch op := chain[i].(type) {
		case *plan.EntityScan:
			if op.DataStore != store.Graph {
				return "", nil
			}
			pattern.WriteString("(" + op.Alias + ")")
			placed[op.Alias] = true
			last = op.Alias
			addLabelCond(op.Alias, op.Labels)
			addFilters(op.Filter)
		case *plan.StoreFragment:
			// 境界: 下位ラン（別ストア）が供給する束縛。ここでは起点だけ控え、
			// 最初にこの alias を使う演算子でパターンへ置く。
			if op.Emits != plan.EmitBindings {
				return "", nil
			}
			for a := range op.OutputSlot.VarToSlot {
				incoming[a] = true
			}
		case *plan.Filter:
			if op.DataStore != store.Graph {
				return "", nil // 非 graph filter があれば部分融合の対象外
			}
			placeIncoming(op.Alias)
			addLabelCond(op.Alias, op.Labels)
			addFilters(op.Filter)
		case *plan.Expand:
			placeIncoming(op.SourceEntity)
			if op.SourceEntity != last {
				return "", nil // 非線形（分岐）traversal は非対応 → フォールバック
			}
			pattern.WriteString(edgePattern(op.Alias, op.RelLabel, op.Dir, ""))
			pattern.WriteString("(" + op.TargetEntity + ")")
			placed[op.TargetEntity] = true
			if op.Alias != "" {
				placed[op.Alias] = true // 関係にも uuid があり projection 対象になり得る
			}
			last = op.TargetEntity
			addLabelCond(op.TargetEntity, op.TargetLabels)
		case *plan.VarLengthExpand:
			placeIncoming(op.SourceEntity)
			if op.SourceEntity != last {
				return "", nil
			}
			pattern.WriteString(edgePattern(op.Alias, op.RelLabel, op.Dir, varLenRange(op.MinHops, op.MaxHops)))
			pattern.WriteString("(" + op.TargetEntity + ")")
			placed[op.TargetEntity] = true
			if op.Alias != "" {
				placed[op.Alias] = true
			}
			last = op.TargetEntity
			addLabelCond(op.TargetEntity, op.TargetLabels)
		default:
			return "", nil // 想定外ノード → フォールバック
		}
	}

	// RETURN する alias が全てパターンに登場していること。
	var rets []string
	for _, a := range outAliases {
		if a == "" {
			continue
		}
		if !placed[a] {
			return "", nil
		}
		rets = append(rets, fmt.Sprintf("%s.uuid AS %s", a, a))
	}
	if len(rets) == 0 {
		return "", nil
	}

	q := "MATCH " + pattern.String()
	if len(whereParts) > 0 {
		q += " WHERE " + strings.Join(whereParts, " AND ")
	}
	q += " RETURN " + strings.Join(rets, ", ")
	return q, params
}

// IncomingParam は境界（下位フラグメント）から供給される束縛 uuid 配列を渡す Cypher パラメータ名。
// 生成側（BuildGraphRecordCypher）と実行側（各エンジン）で同じ規約を使う。
func IncomingParam(alias string) string { return "in_" + alias }

// BindIncoming は境界フラグメントが供給した中間 Record 群から alias ごとの uuid 配列を作り、
// 上位ランのクエリパラメータへ束ねる（重複は除去する）。全エンジンで共通の結線規約。
func BindIncoming(params map[string]interface{}, b *plan.StoreFragment, recs []Record) {
	for alias, idx := range b.OutputSlot.VarToSlot {
		seen := make(map[string]struct{}, len(recs))
		ids := make([]string, 0, len(recs))
		for _, r := range recs {
			if idx >= len(r.Slots) {
				continue
			}
			u := r.Slots[idx]
			if u.Empty() {
				continue
			}
			s := u.String()
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			ids = append(ids, s)
		}
		params[IncomingParam(alias)] = ids
	}
}

// BoundaryFragment は Plan の連鎖に現れる下位フラグメント（境界入力）を返す（無ければ nil）。
// 一般セグメンタが作るラン境界を実行側が見つけるために使う。
func BoundaryFragment(p plan.PlanNode) *plan.StoreFragment {
	for n := p; n != nil; {
		if f, ok := n.(*plan.StoreFragment); ok && f.Emits == plan.EmitBindings {
			return f
		}
		ch := n.Children()
		if len(ch) == 0 {
			return nil
		}
		n = ch[0]
	}
	return nil
}

// edgePattern は方向付きの関係パターン片を返す（rangeLit は可変長時のみ）。
func edgePattern(alias, relLabel string, dir plan.Direction, rangeLit string) string {
	rel := alias
	if relLabel != "" {
		rel = alias + ":" + relLabel
	}
	body := "[" + rel + rangeLit + "]"
	switch dir {
	case plan.Incoming:
		return "<-" + body + "-"
	case plan.Bidirectional:
		return "-" + body + "-"
	default: // Outgoing
		return "-" + body + "->"
	}
}

// varLenRange は可変長ホップのレンジ表記（*、*min.. 等）を返す。VarLengthExpand.String() と整合。
func varLenRange(minH, maxH int) string {
	switch {
	case minH < 0 && maxH < 0:
		return "*"
	case maxH < 0:
		return fmt.Sprintf("*%d..", minH)
	case minH < 0:
		return fmt.Sprintf("*..%d", maxH)
	case minH == maxH:
		return fmt.Sprintf("*%d", minH)
	default:
		return fmt.Sprintf("*%d..%d", minH, maxH)
	}
}
