package physical

import (
	"fmt"
	"strings"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/store"
)

// Node は物理演算子。
type Node interface {
	Children() []Node
	Describe() string
}

// Argument は委譲する論理部分木（StoreQuery.Ops）やコーディネータ演算子の葉に置く目印で、
// 「境界（下の物理演算子）から受け取る行」を表す。Cols は受け取る識別列（IN $ids で渡すキー）。
type Argument struct {
	Cols []string
}

func (*Argument) Children() []logical.Node { return nil }
func (a *Argument) Describe() string {
	return fmt.Sprintf("Argument(%s)", strings.Join(a.Cols, ", "))
}

// StoreQuery は 1 つのストアへ委譲する 1 本のクエリ。
//   - Ops:     委譲する論理部分木。葉は EntityScan（Input なし）か Argument（Input の行の Keys を IN $ids で受け取る）。
//   - Keys:    Input の行と結果を突き合わせる識別列。実行時は結果をキーで入力行へ結合する（入力の重複は保つ）。
//   - Outputs: ストアから返す列（Finalize が要求解析で決める）。
type StoreQuery struct {
	Store   store.Kind
	Ops     logical.Node
	Input   Node
	Keys    []string
	Outputs []string
}

// IsFetch は Ops が「Argument の上の GetProperty だけ」か（値の取得専用のクエリ）。
func (q *StoreQuery) IsFetch() bool {
	n := q.Ops
	seen := false
	for {
		switch x := n.(type) {
		case *logical.GetProperty:
			seen = true
			n = x.Input
			continue
		case *Argument:
			return seen
		}
		return false
	}
}

// Leaf は Input を持たない（ストア内で完結する）クエリか。
func (q *StoreQuery) Leaf() bool { return q.Input == nil }

func (q *StoreQuery) Children() []Node {
	if q.Input == nil {
		return nil
	}
	return []Node{q.Input}
}

func (q *StoreQuery) Describe() string {
	kind := "StoreQuery"
	if q.IsFetch() {
		kind = "Fetch"
	}
	keys := ""
	if len(q.Keys) > 0 {
		keys = " in=(" + strings.Join(q.Keys, ", ") + ")"
	}
	return fmt.Sprintf("%s[%s]%s out=(%s)", kind, q.Store, keys, strings.Join(q.Outputs, ", "))
}

// ListFetch はリスト列 Source の要素（全行の和集合）について Prop を取得し、値のリスト列 Key() を作る。
type ListFetch struct {
	Input  Node
	Store  store.Kind
	Source string
	Part   logical.SegPart
	Prop   string
	Labels []string
}

// Key は作るリスト値列のキー。
func (f *ListFetch) Key() string { return logical.ListPropKey(f.Source, f.Part, f.Prop) }

func (f *ListFetch) Children() []Node { return []Node{f.Input} }
func (f *ListFetch) Describe() string {
	return fmt.Sprintf("ListFetch[%s] %s from %s [elem:%s]", f.Store, f.Key(), f.Source, strings.Join(f.Labels, "|"))
}

// Coord はコーディネータで実行する論理演算子 1 個。Op の入力は Argument に差し替えてある。
type Coord struct {
	Op     logical.Node
	Inputs []Node
}

func (c *Coord) Children() []Node { return c.Inputs }
func (c *Coord) Describe() string { return "Coord " + c.Op.Describe() }

// Dump は物理計画を決定的に表示する。StoreQuery は委譲する論理部分木を `: ` 付きで展開する。
func Dump(n Node) string {
	var b strings.Builder
	var walk func(Node, int)
	walk = func(n Node, d int) {
		if n == nil {
			return
		}
		ind := strings.Repeat("  ", d)
		b.WriteString(ind + n.Describe() + "\n")
		if q, ok := n.(*StoreQuery); ok {
			for _, line := range strings.Split(strings.TrimRight(logical.Dump(q.Ops), "\n"), "\n") {
				b.WriteString(ind + "  : " + line + "\n")
			}
		}
		for _, c := range n.Children() {
			walk(c, d+1)
		}
	}
	walk(n, 0)
	return b.String()
}
