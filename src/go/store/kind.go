// Package store はデータストア種別の語彙（Kind enum）を提供する葉パッケージ。
//
// plan（IR）と storage（接続実装）と各エンジンが同じ語彙を共有できるよう、
// 依存を持たない最小パッケージに enum を置く。旧 storage.StoreKind と旧 plan.DataStore
// の二重定義（文字列ブリッジ）を store.Kind へ一本化した。
package store

import "fmt"

// Kind はデータストアの種別。
type Kind int

const (
	Graph Kind = iota
	Columnar
	Relational
	Document
	Kvs
)

var kindNames = map[Kind]string{
	Graph:      "graph",
	Document:   "document",
	Kvs:        "kvs",
	Relational: "relational",
	Columnar:   "columnar",
}

func (k Kind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// ParseKind は文字列表現（"graph" 等）を Kind に変換する。
func ParseKind(s string) (Kind, error) {
	for k, name := range kindNames {
		if name == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown store kind: %q", s)
}
