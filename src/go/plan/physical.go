package plan

import (
	"fmt"
	"strings"

	"polystore_database/src/go/store"
)

// StoreFragment は連続する同一ストアアクセスの演算子を 1 ネイティブクエリへ融合した物理演算子。
// 融合パス（planner/fusion.go）が生成し、各エンジンの op_fragment.go が実行する。
//   - RawQuery != "": graph 全体委譲の高速路。原 Cypher をそのまま発行し Neo4j baseline と一致させる。
//   - RawQuery == "": Ops 部分木を engine/core.LowerFragment でネイティブクエリへ翻訳する。
//   - Input != nil:   下位フラグメント/統合演算子からの境界入力（IN-list 等の供給元）。
type StoreFragment struct {
	Store store.Kind

	Ops   PlanNode // 融合した論理演算子の部分木（非 graph 全体委譲時に翻訳対象）
	Input PlanNode // 境界入力（葉ソースなら nil）

	RawQuery string            // graph 全体委譲(row-mode): 原 Cypher（$param 付き）
	Params   map[string]string // graph 全体委譲(row-mode): $param（実行時に型付け）

	// AsRecords=true は record-mode（部分融合）: record パイプライン（scan+filter+expand）を
	// 融合した source。各エンジンが Ops（graph record 部分木）から束縛 UUID を返す 1 本のクエリを
	// 生成し（core.BuildGraphRecordCypher）、結果を OutputSlot に従って Record slot へ格納する。
	// 融合実行未対応のエンジンや生成不能な構造では Ops を通常実行してフォールバックする（結果は等価）。
	AsRecords bool

	OutputAlias []string // 統合演算子が結線するための出力束縛
	OutputSlot  SlotTable
}

func (f *StoreFragment) Children() []PlanNode {
	if f.Input != nil {
		return []PlanNode{f.Input}
	}
	return nil
}

func (f *StoreFragment) String() string {
	if f.RawQuery != "" {
		return fmt.Sprintf("StoreFragment[%s raw]: %s", f.Store, strings.Join(strings.Fields(f.RawQuery), " "))
	}
	ops := "∅"
	if f.Ops != nil {
		ops = f.Ops.String()
	}
	return fmt.Sprintf("StoreFragment[%s]: %s", f.Store, ops)
}

// TailEntity は tail pushdown で一時テーブルへ staging する束縛エンティティ（alias）と、
// その永続テーブル（Label）を表す。実行時に uuid をキーに永続テーブルへ JOIN してプロパティを引く。
type TailEntity struct {
	Alias string   // 束縛エイリアス（例 "author"）
	Table string   // 永続テーブル/ラベル（例 "Person"）。JOIN 先
	Props []string // この alias から tail が参照するプロパティ（SELECT/GROUP BY 用）
}

// TailPushdown は「traversal で集めた中間 UUID を単一の非 graph ストアの一時テーブルへロードし、
// RETURN 句 tail（Projection/Aggregate/GroupBy/Sort/Limit）をそのストアのネイティブエンジンで
// 実行する」物理演算子。融合パス（planner）が settings.TailPushdown 有効時に、
// tail 参照プロパティが単一ストアに解決でき能力も満たす場合にのみ生成する。
//
// 実行対応エンジン（現状 bulk のみ）は Input を実行して中間 Record（束縛 UUID）を得てから
// 一時テーブル＋SQL で tail を計算する。未対応エンジンは Fallback（元 coordinator tail）を通常実行する
// （結果は等価）。これにより「tail を engine で計算」vs「ネイティブ実行」の A/B を同一プランで比較できる。
type TailPushdown struct {
	Store store.Kind

	Input    PlanNode // record source（graph 融合フラグメント。束縛 UUID を返す）
	Fallback PlanNode // 未対応エンジン用: 元 coordinator tail（Return→…→record pipeline）

	InputSlot SlotTable // alias → Input 出力 Record のスロット番号

	Entities   []TailEntity    // staging するエンティティ（uuid 列）＋ JOIN 先テーブル
	Return     []ReturnItem    // 出力列（SELECT 別名＝ReturnItem.Name）
	GroupKeys  []GroupKey      // GROUP BY キー（staging エンティティのプロパティ）
	Aggs       []AggregateItem // 集約式
	OrderItems []OrderItem     // ORDER BY（出力別名で並べる）
	Limit      int             // LIMIT（0 以下で無し）
}

func (t *TailPushdown) Children() []PlanNode {
	if t.Input != nil {
		return []PlanNode{t.Input}
	}
	return nil
}

func (t *TailPushdown) String() string {
	return fmt.Sprintf("TailPushdown[%s] entities=%d groups=%d aggs=%d limit=%d",
		t.Store, len(t.Entities), len(t.GroupKeys), len(t.Aggs), t.Limit)
}

// IntegrateKeyKind は統合演算子の結合キー種別。
type IntegrateKeyKind int

const (
	KeyID    IntegrateKeyKind = iota // 束縛 UUID をキーにしたプロパティ材料化
	KeyValue                         // 非 ID 値による cross-store join
)

// ColumnRef は alias.prop 参照（Prop 空なら alias の束縛 identity）。
type ColumnRef struct {
	Alias string
	Prop  string
}

func (c ColumnRef) String() string {
	if c.Prop == "" {
		return c.Alias
	}
	return c.Alias + "." + c.Prop
}

// IntegrateKey は各入力側のどの列で結ぶかを表す（Refs は Integrate.Inputs と同じ並び）。
// 計画段階では結合の意味だけを持ち、物理アルゴリズムは保持しない。
type IntegrateKey struct {
	Kind IntegrateKeyKind
	Refs []ColumnRef
}

// Integrate はストア境界をまたぐ結果統合の物理演算子。
// 物理アルゴリズム（hash / nested-loop / batched-IN）は持たず、エンジンが実行時に
// 件数等のヒューリスティック（settings.IntegrateRowThreshold）で選択する。
// 計画段階では結合キーと必要カラムのみ宣言する（複数カラムの最適統合は計画時に不明なため）。
type Integrate struct {
	Inputs []PlanNode
	Keys   []IntegrateKey
	Needed []ColumnRef

	InputSlot  SlotTable
	OutputSlot SlotTable
}

// StorePushdownFromFragment は StoreFragment を実行時に既存の StorePushdown 実行経路へ
// ブリッジするための変換。graph 全体委譲は RawQuery をそのまま、非 graph は Ops 部分木
// （Return を含む論理木）を走査して Table/Filters/GroupKeys/Aggs/OrderItems/Limit/Items を
// 復元する。ネイティブクエリ生成は各エンジンの実績ある run*Pushdown をそのまま再利用する。
//   - P5 で FragmentSpec ベースの生成へ一本化し、StorePushdown ごと退役する予定の橋渡し。
func StorePushdownFromFragment(f *StoreFragment) *StorePushdown {
	sp := &StorePushdown{Store: f.Store}
	if f.RawQuery != "" {
		sp.Query = f.RawQuery
		sp.Params = f.Params
		return sp
	}
	for n := f.Ops; n != nil; {
		switch op := n.(type) {
		case *Return:
			sp.Items = op.Items
		case *Limit:
			sp.Limit = op.Count
		case *Sort:
			sp.OrderItems = append(sp.OrderItems, op.OrderItems...)
		case *Aggregate:
			sp.Aggs = append(sp.Aggs, op.Aggs...)
			sp.GroupKeys = append(sp.GroupKeys, op.GroupKeys...)
		case *Filter:
			sp.Filters = append(sp.Filters, op.Filter...)
		case *EntityScan:
			if sp.Table == "" && len(op.Labels) > 0 {
				sp.Table = op.Labels[0]
			}
			sp.Filters = append(sp.Filters, op.Filter...)
		}
		ch := n.Children()
		if len(ch) == 0 {
			break
		}
		n = ch[0]
	}
	return sp
}

func (i *Integrate) Children() []PlanNode { return i.Inputs }

func (i *Integrate) String() string {
	var keys []string
	for _, k := range i.Keys {
		var refs []string
		for _, r := range k.Refs {
			refs = append(refs, r.String())
		}
		kind := "id"
		if k.Kind == KeyValue {
			kind = "val"
		}
		keys = append(keys, fmt.Sprintf("%s(%s)", kind, strings.Join(refs, "=")))
	}
	var needed []string
	for _, n := range i.Needed {
		needed = append(needed, n.String())
	}
	return fmt.Sprintf("Integrate[keys: %s | needs: %s]",
		strings.Join(keys, ", "), strings.Join(needed, ", "))
}
