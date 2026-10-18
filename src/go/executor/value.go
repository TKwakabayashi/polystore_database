// Package executor はクエリ処理系（刷新版）の実行器（R5: bulk）。物理計画を葉から全件実体化して
// 評価する。ストアへのアクセスは Backend に閉じ、ここでは
//   - StoreQuery: ネイティブクエリを発行し、入力行と識別列で結合する（入力の重複は保つ）
//   - ListFetch:  リスト列の要素（全行の和集合）の値を取得して、値のリスト列を作る
//   - Coord:      コーディネータの演算子（Filter / Projection / Aggregation / Sort / Limit / …）
//
// を行う。設計: docs/operator_plan.md §5。
package executor

import (
	"fmt"
	"math"
	"strings"
	"time"

	"polystore_database/src/go/id"
)

// Row は 1 行。列キーは logical の規約（識別列 "a" は id.UUID、値列 "a.p"、区間・パス列は *ListVal、
// リスト値列は []any）。
type Row map[string]any

// ListVal は区間列・パス列の値（ノードと関係の識別子のリスト）。区間列の Nodes は中間ノードのみ。
type ListVal struct {
	Nodes []id.UUID
	Rels  []id.UUID
}

func (r Row) clone() Row {
	out := make(Row, len(r))
	for k, v := range r {
		out[k] = v
	}
	return out
}

// normalize はストアのドライバが返す値を比較可能な共通型へそろえる
// （整数 → int64、浮動小数 → float64、[]byte → string、時刻 → UTC）。
func normalize(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return float64(x)
	case []byte:
		return string(x)
	case time.Time:
		return x.UTC()
	}
	return v
}

// ---- 比較（Cypher の意味）----

// typeRank は異なる型どうしの並べ替え順（ORDER BY 用）。null は常に最後。
func typeRank(v any) int {
	switch v.(type) {
	case bool:
		return 1
	case int64, float64:
		return 2
	case string:
		return 3
	case time.Time:
		return 4
	case id.UUID:
		return 5
	case nil:
		return 9
	}
	return 6
}

// compare は a と b を比べる。ok=false は比較不能（型が違う・null を含む）＝ 3 値論理の null。
func compare(a, b any) (int, bool) {
	if a == nil || b == nil {
		return 0, false
	}
	switch x := a.(type) {
	case int64:
		switch y := b.(type) {
		case int64:
			return cmpOrdered(x, y), true
		case float64:
			return cmpOrdered(float64(x), y), true
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return cmpOrdered(x, float64(y)), true
		case float64:
			return cmpOrdered(x, y), true
		}
	case string:
		if y, ok := b.(string); ok {
			return strings.Compare(x, y), true
		}
	case bool:
		if y, ok := b.(bool); ok {
			switch {
			case x == y:
				return 0, true
			case !x:
				return -1, true
			}
			return 1, true
		}
	case time.Time:
		if y, ok := b.(time.Time); ok {
			return x.Compare(y), true
		}
	case id.UUID:
		if y, ok := b.(id.UUID); ok {
			return id.Compare(x, y), true
		}
	}
	return 0, false
}

func cmpOrdered[T int64 | float64](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// orderCompare は ORDER BY の比較（全順序。null は最後、型が違えば型の順）。
func orderCompare(a, b any) int {
	if c, ok := compare(a, b); ok {
		return c
	}
	ra, rb := typeRank(a), typeRank(b)
	if ra != rb {
		return ra - rb
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

// equalValues は等値（DISTINCT・グループ化用。null は null と等しいとみなす）。
func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	c, ok := compare(a, b)
	return ok && c == 0
}

// keyString は値の組を、グループ化・重複除去・結合のキーにする文字列。
func keyString(vals ...any) string {
	var b strings.Builder
	for _, v := range vals {
		writeKey(&b, v)
		b.WriteByte(0)
	}
	return b.String()
}

func writeKey(b *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("n")
	case int64:
		fmt.Fprintf(b, "f%v", float64(x)) // 1 と 1.0 を同じキーに
	case float64:
		if math.IsNaN(x) {
			b.WriteString("nan")
		} else {
			fmt.Fprintf(b, "f%v", x)
		}
	case string:
		b.WriteString("s" + x)
	case id.UUID:
		b.WriteString("u" + x.String())
	case time.Time:
		b.WriteString("t" + x.UTC().Format(time.RFC3339Nano))
	case *ListVal:
		b.WriteString("L")
		for _, n := range x.Nodes {
			b.WriteString(n.String() + ",")
		}
		b.WriteString("|")
		for _, r := range x.Rels {
			b.WriteString(r.String() + ",")
		}
	case []any:
		b.WriteString("[")
		for _, e := range x {
			writeKey(b, e)
			b.WriteString(",")
		}
		b.WriteString("]")
	default:
		fmt.Fprintf(b, "?%v", x)
	}
}
