// Package native はクエリ処理系（刷新版）の段階 R4: 物理計画の StoreQuery / ListFetch を、
// 各ストアのネイティブクエリ（Cypher / SQL / CQL / Mongo pipeline / KVS の取得計画）へ変換する。
//
// 結果の列名は c0, c1, … に統一する（Mongo のフィールド名に "." を使えないため）。
// Query.Columns[i] が ci に対応する計画上の列キー（physical.StoreQuery.Outputs と同じ並び）。
// 入力（境界の識別列）は実行時に束ねる: InKeys が 1 つなら uuid のリスト、2 つ以上なら
// {key: uuid} の行のリストを、パラメータ InParam として渡す。
package native

import (
	"fmt"
	"sort"
	"strings"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/store"
)

// InParam は境界の識別列を渡すパラメータ名（Cypher の $in、SQL/CQL の展開位置）。
const InParam = "in"

// TypeFn はプロパティの型名を返す（mapping の type）。リテラルの型付けに使う。
type TypeFn func(kind logical.ObjKind, labels []string, prop string) string

// Query は 1 本のネイティブクエリ。
type Query struct {
	Store    store.Kind
	Columns  []string  // 結果の ci に対応する計画上の列キー
	ColTypes []string  // 各列の型名（プロパティ列のみ。結果の値の復元に使う）
	ColKinds []ColKind // 各列の種類（識別列 / 値列 / 区間・パス列）
	InKeys   []string  // 境界から受け取る識別列（空なら葉）

	// graph（Cypher）・relational（SQL）・columnar（CQL）
	Text   string
	Params map[string]any // Cypher の名前付きパラメータ
	Args   []any          // SQL / CQL の位置パラメータ（InArgAt の位置に入力 uuid を展開する）
	InArg  int            // Args 中で入力 uuid を差し込む位置（-1 なら無し）

	// document（Mongo）
	Collection string
	Pipeline   []Stage // 先頭の $match に、実行時 `uuid: {$in: ids}` を足す（InKeys があるとき）

	// kvs
	KVS *KVSQuery
}

// ColKind は結果の列の種類。識別列は uuid 文字列、区間・パス列は {nodes, rels} の uuid リストで返る。
type ColKind int

const (
	KindValue ColKind = iota
	KindIdent
	KindList
)

// Stage は Mongo の集約パイプラインの 1 段（演算子 → 中身）。表示用に順序付きの組で持つ。
type Stage struct {
	Op   string
	Body Doc
}

// Doc は順序付きの Mongo ドキュメント。
type Doc []Field

// Field は Doc の 1 要素。
type Field struct {
	Key string
	Val any
}

// KVSQuery は LevelDB からの取得計画。
//   - Scan: ラベルの全件（PrimaryEq があれば逆引きインデックスの前方一致）から uuid を得る
//   - InKeys があれば入力 uuid について点取得する
//   - Props を BuildEntityKey(Label, uuid, prop) で点取得し、Filters をメモリ評価する
type KVSQuery struct {
	Label     string
	Scan      bool
	PrimaryEq *KVSEq
	Props     []string
	Types     map[string]string
	Filters   []logical.Expr // 列キーは "alias.prop"
	Alias     string
}

// KVSEq は逆引きインデックスで引ける等値条件。
type KVSEq struct {
	Prop  string
	Value any
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ast.ErrUnsupported, fmt.Sprintf(format, args...))
}

// Lower は StoreQuery をネイティブクエリにする。
func Lower(q *physical.StoreQuery, types TypeFn) (*Query, error) {
	var (
		out *Query
		err error
	)
	switch q.Store {
	case store.Graph:
		out, err = lowerCypher(q, types)
	case store.Relational:
		out, err = lowerSQL(q, types, dialectMySQL)
	case store.Columnar:
		out, err = lowerSQL(q, types, dialectCQL)
	case store.Document:
		out, err = lowerMongo(q, types)
	case store.Kvs:
		out, err = lowerKVS(q, types)
	default:
		return nil, unsupported("store %s", q.Store)
	}
	if err != nil {
		return nil, err
	}
	own := owners(q.Ops)
	out.ColTypes = make([]string, len(out.Columns))
	for i, c := range out.Columns {
		out.ColTypes[i] = colType(c, own, types)
	}
	if len(out.ColKinds) != len(out.Columns) {
		return nil, unsupported("column kinds not set for %s query", q.Store)
	}
	return out, nil
}

// LowerList は ListFetch を、要素のラベルごとの値取得クエリにする（結果の列は c0=uuid, c1=値）。
func LowerList(f *physical.ListFetch, types TypeFn) ([]*Query, error) {
	kind := logical.NodeObj
	if f.Part == logical.SegRels {
		kind = logical.RelObj
	}
	var out []*Query
	for _, label := range f.Labels {
		g := &logical.GetProperty{Input: &physical.Argument{Cols: []string{"_e"}}, Alias: "_e", Prop: f.Prop, Kind: kind, Labels: []string{label}}
		sq := &physical.StoreQuery{Store: f.Store, Ops: g, Keys: []string{"_e"}, Outputs: []string{"_e", g.Key()}}
		q, err := Lower(sq, types)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if len(out) == 0 {
		return nil, unsupported("list fetch without element labels")
	}
	return out, nil
}

// ================================
// 共通の補助
// ================================

// chain は単項演算子の連なりを葉から根の順に並べる（二項演算子があれば nil）。
func chain(n logical.Node) []logical.Node {
	var rev []logical.Node
	for n != nil {
		rev = append(rev, n)
		cs := n.Children()
		if len(cs) == 0 {
			break
		}
		if len(cs) > 1 {
			return nil
		}
		n = cs[0]
	}
	out := make([]logical.Node, len(rev))
	for i, x := range rev {
		out[len(rev)-1-i] = x
	}
	return out
}

// owner は alias の持ち主の種類とラベル（GetProperty から集める）。
type owner struct {
	kind   logical.ObjKind
	labels []string
}

func owners(n logical.Node) map[string]owner {
	out := map[string]owner{}
	var walk func(logical.Node)
	walk = func(n logical.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *logical.GetProperty:
			if _, ok := out[x.Alias]; !ok {
				out[x.Alias] = owner{kind: x.Kind, labels: x.Labels}
			}
		case *logical.EntityScan:
			if _, ok := out[x.Alias]; !ok {
				out[x.Alias] = owner{kind: logical.NodeObj, labels: x.Labels}
			}
		case *logical.Expand:
			out[x.Rel] = owner{kind: logical.RelObj, labels: x.Types}
		case *logical.ExpandInto:
			out[x.Rel] = owner{kind: logical.RelObj, labels: x.Types}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

// colType は値列 "alias.prop" の型名。
func colType(key string, own map[string]owner, types TypeFn) string {
	i := strings.LastIndex(key, ".")
	if i < 0 || types == nil {
		return ""
	}
	o, ok := own[key[:i]]
	if !ok {
		return ""
	}
	return types(o.kind, o.labels, key[i+1:])
}

// typed はリテラルを列の型へ変換する（失敗・型不明ならそのまま）。
func typed(v any, typ string) any {
	if typ == "" || v == nil {
		return v
	}
	out, err := codec.ConvertToNativeType(v, typ)
	if err != nil {
		return v
	}
	return out
}

// litType は比較・IN の相手が列なら、その型名を返す。
func litType(other logical.Expr, own map[string]owner, types TypeFn) string {
	if c, ok := other.(*logical.Column); ok {
		return colType(c.Key, own, types)
	}
	return ""
}

func colName(i int) string { return fmt.Sprintf("c%d", i) }

// Render はネイティブクエリを決定的に表示する（golden 用）。
func (q *Query) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] columns=(%s)", q.Store, strings.Join(q.Columns, ", "))
	if len(q.InKeys) > 0 {
		fmt.Fprintf(&b, " in=(%s)", strings.Join(q.InKeys, ", "))
	}
	b.WriteByte('\n')
	switch {
	case q.KVS != nil:
		k := q.KVS
		src := "ids"
		if k.Scan {
			src = "scan"
			if k.PrimaryEq != nil {
				src = fmt.Sprintf("index %s=%s", k.PrimaryEq.Prop, show(k.PrimaryEq.Value))
			}
		}
		fmt.Fprintf(&b, "  KVS %s %s get=(%s)", k.Label, src, strings.Join(k.Props, ", "))
		for _, f := range k.Filters {
			fmt.Fprintf(&b, " filter %s", f)
		}
		b.WriteByte('\n')
	case q.Store == store.Document:
		fmt.Fprintf(&b, "  db.%s.aggregate([\n", q.Collection)
		for _, s := range q.Pipeline {
			fmt.Fprintf(&b, "    {%s: %s}\n", s.Op, show(s.Body))
		}
		b.WriteString("  ])\n")
	default:
		fmt.Fprintf(&b, "  %s\n", q.Text)
		if len(q.Params) > 0 {
			keys := make([]string, 0, len(q.Params))
			for k := range q.Params {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			parts := make([]string, len(keys))
			for i, k := range keys {
				parts[i] = k + "=" + show(q.Params[k])
			}
			fmt.Fprintf(&b, "  params: %s\n", strings.Join(parts, ", "))
		}
		if len(q.Args) > 0 {
			parts := make([]string, len(q.Args))
			for i, a := range q.Args {
				parts[i] = show(a)
			}
			fmt.Fprintf(&b, "  args: %s\n", strings.Join(parts, ", "))
		}
	}
	return b.String()
}

// show は値を型付きで表示する（golden で型の取り違えを検出できるように）。
func show(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("%q", x)
	case Doc:
		parts := make([]string, len(x))
		for i, f := range x {
			parts[i] = f.Key + ": " + show(f.Val)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = show(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case inMarker:
		return "<in>"
	default:
		return fmt.Sprintf("%v(%T)", x, x)
	}
}

// inMarker は位置パラメータ・Mongo の中で「入力 uuid がここに入る」目印。
type inMarker struct{}

// InMarker は Args / Pipeline 中の入力 uuid の差し込み位置を表す値。
var InMarker = inMarker{}
