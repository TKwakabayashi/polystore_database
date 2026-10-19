// Package physical はクエリ処理系（刷新版）の物理計画化: 論理計画の各演算子にストアを割り当て、
// 同一ストアで連続する演算子を 1 本のネイティブクエリ（StoreQuery）へ統合し、
// 値は必要になった段階で取得する（docs/operator_plan.md §5）。統計は使わない。
package physical

import (
	"sort"

	"polystore_database/src/go/logical"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/store"
)

// Catalog は配置情報（mapping.json）を引く。ストアの選択は計画 1 回ごとに状態を持つ
// （複製された列は、同じ alias の他の列と同じストアを優先する）ので、計画ごとに NewCatalog する。
type Catalog struct {
	md     *schema.MappingDictionary
	chosen map[string]store.Kind // alias → 既に選んだストア
}

// NewCatalog は mapping 辞書から Catalog を作る。md が nil なら全列 graph とみなす。
func NewCatalog(md *schema.MappingDictionary) *Catalog {
	return &Catalog{md: md, chosen: map[string]store.Kind{}}
}

// EntityOf はノードのラベルを、それを含む entity 名にする（Post → Message のように、entity の
// labels に列挙されたサブラベルを解決する）。どの entity にも無ければ空。
func (c *Catalog) EntityOf(label string) string {
	if c.md == nil {
		return ""
	}
	if _, ok := c.md.Entities[label]; ok {
		return label
	}
	for _, name := range sortedEntities(c.md) {
		for _, l := range c.md.Entities[name].Labels {
			if l == label {
				return name
			}
		}
	}
	return ""
}

// Entities はラベルの並びを entity 名の並び（重複なし・解決できないものは元のまま）にする。
func (c *Catalog) Entities(labels []string) []string {
	var out []string
	for _, l := range labels {
		e := c.EntityOf(l)
		if e == "" {
			e = l
		}
		dup := false
		for _, x := range out {
			if x == e {
				dup = true
			}
		}
		if !dup {
			out = append(out, e)
		}
	}
	return out
}

// IsEntity はラベルがそれ自体 entity（非 graph ストアの表・コレクションの単位）か。
// サブラベル（Post など）は entity の表に区別が無いので、非 graph ストアでは走査できない。
func (c *Catalog) IsEntity(label string) bool {
	if c.md == nil {
		return false
	}
	_, ok := c.md.Entities[label]
	return ok
}

// databases は持ち主（ラベル/関係タイプの候補）とプロパティから、置かれているストア名の一覧を返す。
func (c *Catalog) databases(kind logical.ObjKind, labels []string, prop string) []string {
	if c.md == nil {
		return nil
	}
	if kind == logical.NodeObj {
		labels = c.Entities(labels)
	}
	for _, l := range labels {
		if kind == logical.RelObj {
			if r, ok := c.md.Relationships[l]; ok {
				if p, ok := r.Properties[prop]; ok {
					return p.Databases
				}
			}
			continue
		}
		if e, ok := c.md.Entities[l]; ok {
			if p, ok := e.Properties[prop]; ok {
				return p.Databases
			}
		}
	}
	return nil
}

// StoreOf はプロパティを読むストアを決める。mapping に無い列は graph（全データの源泉）。
func (c *Catalog) StoreOf(alias string, kind logical.ObjKind, labels []string, prop string) store.Kind {
	var cands []store.Kind
	for _, d := range c.databases(kind, labels, prop) {
		if k, err := store.ParseKind(d); err == nil {
			cands = append(cands, k)
		}
	}
	if len(cands) == 0 {
		return store.Graph
	}
	prev, hasPrev := c.chosen[alias]
	pick := choose(cands, prev, hasPrev)
	if !hasPrev && alias != "" {
		c.chosen[alias] = pick
	}
	return pick
}

// choose は複製の中から「同じ alias で既に選んだストア → graph → 先頭」の順で選ぶ。
func choose(cands []store.Kind, prev store.Kind, hasPrev bool) store.Kind {
	if hasPrev {
		for _, k := range cands {
			if k == prev {
				return k
			}
		}
	}
	for _, k := range cands {
		if k == store.Graph {
			return k
		}
	}
	return cands[0]
}

// ElemLabels は可変長区間の要素のラベル候補。中間ノードは関係タイプの targetEntities から、
// 関係は関係タイプそのもの。
func (c *Catalog) ElemLabels(types []string, part logical.SegPart) []string {
	if part == logical.SegRels {
		return append([]string{}, types...)
	}
	if c.md == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range sortedEntities(c.md) {
		for _, t := range types {
			if ref, ok := c.md.Entities[e].References[t]; ok {
				for _, target := range ref.TargetEntities {
					if !seen[target] {
						seen[target] = true
						out = append(out, target)
					}
				}
			}
		}
	}
	return out
}

func sortedEntities(md *schema.MappingDictionary) []string {
	out := make([]string, 0, len(md.Entities))
	for k := range md.Entities {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ================================
// ストアの能力（ネイティブに表現できるもの）
// ================================

// CanExpress はストア k が述語 e をネイティブに評価できるか。
func CanExpress(k store.Kind, e logical.Expr) bool {
	ok := true
	logical.Walk(e, func(x logical.Expr) {
		switch v := x.(type) {
		case *logical.Column, *logical.Lit:
		case *logical.List:
		case *logical.Bin:
			switch v.Op {
			case "=", "<>", "<", ">", "<=", ">=", "IN", "AND":
			case "OR":
				if k == store.Columnar || k == store.Kvs {
					ok = false
				}
			default: // 算術
				if k != store.Graph && k != store.Relational {
					ok = false
				}
			}
		case *logical.Un:
			switch v.Op {
			case "IS NULL", "IS NOT NULL":
				if k == store.Columnar {
					ok = false
				}
			case "NOT":
				if k == store.Columnar || k == store.Kvs {
					ok = false
				}
			default:
				if k != store.Graph && k != store.Relational {
					ok = false
				}
			}
		default: // Call / Case / Quant / ElemProp
			if k != store.Graph {
				ok = false
			}
		}
	})
	return ok
}

// plainCols は式がすべて列参照か（非 graph ストアの射影・集約・並べ替えはこれに限る）。
func plainCols(es ...logical.Expr) bool {
	for _, e := range es {
		if e == nil {
			continue
		}
		if _, ok := e.(*logical.Column); !ok {
			return false
		}
	}
	return true
}

// CanAbsorb はストア k が行全体に対する演算子 op をネイティブに実行できるか（葉の StoreQuery に限って使う）。
func CanAbsorb(k store.Kind, op logical.Node) bool {
	if k == store.Graph {
		return true
	}
	switch x := op.(type) {
	case *logical.Projection:
		if k == store.Kvs {
			return false
		}
		for _, it := range x.Items {
			if !plainCols(it.Expr) {
				return false
			}
		}
		return true
	case *logical.Aggregation:
		switch k {
		case store.Kvs:
			return false
		case store.Columnar:
			if len(x.Keys) > 0 {
				return false
			}
		}
		for _, it := range x.Keys {
			if !plainCols(it.Expr) {
				return false
			}
		}
		for _, a := range x.Aggs {
			if a.Distinct && k == store.Columnar {
				return false
			}
			if !plainCols(a.Arg) {
				return false
			}
		}
		return true
	case *logical.Sort:
		if k != store.Relational && k != store.Document {
			return false
		}
		for _, sk := range x.Keys {
			if !plainCols(sk.Expr) {
				return false
			}
		}
		return true
	case *logical.Limit, *logical.Distinct:
		return k == store.Relational || k == store.Document
	}
	return false
}

// TypeOf はプロパティの型名（mapping の type。不明なら空）。ネイティブクエリのリテラルの型付けに使う。
func (c *Catalog) TypeOf(kind logical.ObjKind, labels []string, prop string) string {
	if c.md == nil {
		return ""
	}
	if kind == logical.NodeObj {
		labels = c.Entities(labels)
	}
	for _, l := range labels {
		if kind == logical.RelObj {
			if r, ok := c.md.Relationships[l]; ok {
				if p, ok := r.Properties[prop]; ok {
					return p.Type
				}
			}
			continue
		}
		if e, ok := c.md.Entities[l]; ok {
			if p, ok := e.Properties[prop]; ok {
				return p.Type
			}
		}
	}
	return ""
}
