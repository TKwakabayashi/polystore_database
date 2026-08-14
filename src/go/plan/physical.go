package plan

import (
	"fmt"
	"strings"

	"polystore_database/src/go/store"
)

// EmitKind は StoreFragment が下流へ何を出力するかを表す（パイプライン上の出力種別）。
// 行指向/列指向とは無関係な直交軸である点に注意（それはエンジン内部表現の違い）。
type EmitKind int

const (
	// EmitResult は材料化・集約済みの最終行（Row）を出力する（tail まで委譲した場合）。
	EmitResult EmitKind = iota
	// EmitBindings は束縛 UUID の中間レコード（core.Record）を出力する
	//（record パイプラインの source として上位演算子へ流す）。
	EmitBindings
)

func (e EmitKind) String() string {
	if e == EmitBindings {
		return "bindings"
	}
	return "result"
}

// StoreFragment は「1 つのストアへ委譲する 1 クエリ」を表す物理演算子。
// 融合パス（planner/fusion.go）が生成し、各エンジンが実行する。
//
// 意味の源泉は常に Plan（委譲する論理サブプラン＝最小演算子の木）であり、ネイティブクエリは
// Plan を走査して生成（lowering）する。委譲できないエンジン／生成不能な構造では Plan を
// コーディネータでそのまま実行すればよい（結果は等価）ため、専用の fallback は持たない。
//
//   - Emits:    Bindings（束縛 UUID を流す source）/ Result（最終行）。
//   - Verbatim: Plan の既知の忠実な lowering（graph 全体委譲時の原 Cypher）。空なら Plan から生成する。
//   - Input:    下位フラグメント/統合からの境界入力（IN-list 等の供給元。葉ソースなら nil）。
type StoreFragment struct {
	Store store.Kind

	Plan  PlanNode // 委譲する論理サブプラン（意味の源泉。lowering の対象）
	Input PlanNode // 境界入力（葉ソースなら nil）

	Emits EmitKind

	// Verbatim は Plan に対する既知の忠実な lowering（graph 全体委譲の原 Cypher）。
	// 原文をそのまま発行することで Neo4j baseline と同一クエリになることを保証する。
	Verbatim string
	Params   map[string]string // Verbatim 用パラメータ（実行時に型付け）

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
	if f.Verbatim != "" {
		return fmt.Sprintf("StoreFragment[%s verbatim]: %s", f.Store, strings.Join(strings.Fields(f.Verbatim), " "))
	}
	sub := "∅"
	if f.Plan != nil {
		sub = f.Plan.String()
	}
	return fmt.Sprintf("StoreFragment[%s %s]: %s", f.Store, f.Emits, sub)
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
