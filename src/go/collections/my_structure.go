package collections

import (
	"sort"
	"strings"
)

// ================================
// Set definition
// ================================

type Set map[string]struct{}

func (s Set) Insert(elem string) {
	s[elem] = struct{}{}
}

func (s Set) Remove(elem string) {
	delete(s, elem)
}

// ConvertSlice は要素をソート済みスライスで返す。
// 反復順を決定的にすることで、これを消費するスロット割当（planner.RefinePlan）が
// 実行ごとに安定し、論理プランが再現可能になる（golden プランテストの前提）。
func (s Set) ConvertSlice() []string {
	res := make([]string, 0, len(s))
	for k := range s {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (s Set) Print() string {
	result := []string{}
	for elem := range s {
		result = append(result, elem)
	}
	return strings.Join(result, ", ")
}
