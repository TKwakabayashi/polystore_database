package collections

import "strings"

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

func (s Set) ConvertSlice() []string {
	res := make([]string, 0, len(s))
	for k := range s {
		res = append(res, k)
	}
	return res
}

func (s Set) Print() string {
	result := []string{}
	for elem := range s {
		result = append(result, elem)
	}
	return strings.Join(result, ", ")
}
