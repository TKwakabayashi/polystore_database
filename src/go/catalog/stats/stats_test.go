package stats

import (
	"math"
	"testing"
)

// A small graph: 10 Person, 100 Post; 100 HAS_CREATOR from Post to Person, 30 KNOWS between Persons.
func toy() *Stats {
	return &Stats{
		Nodes:    110,
		Labels:   map[string]int64{"Person": 10, "Post": 100},
		RelTypes: map[string]int64{"HAS_CREATOR": 100, "KNOWS": 30},
		Out:      map[string]map[string]int64{"Post": {"HAS_CREATOR": 100}, "Person": {"KNOWS": 30}},
		In:       map[string]map[string]int64{"Person": {"HAS_CREATOR": 100, "KNOWS": 30}},
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCount(t *testing.T) {
	s := toy()
	if got := s.Count([]string{"Person"}); got != 10 {
		t.Errorf("Person: %v", got)
	}
	if got := s.Count(nil); got != 110 {
		t.Errorf("no label means every node: %v", got)
	}
	if got := s.Count([]string{"City"}); got != 110 {
		t.Errorf("a label without statistics does not narrow: %v", got)
	}
}

func TestDegree(t *testing.T) {
	s := toy()
	cases := []struct {
		name     string
		from, to []string
		types    []string
		dir      Dir
		want     float64
	}{
		{"posts per person", []string{"Person"}, []string{"Post"}, []string{"HAS_CREATOR"}, In, 10},
		{"author of a post", []string{"Post"}, []string{"Person"}, []string{"HAS_CREATOR"}, Out, 1},
		{"KNOWS undirected", []string{"Person"}, []string{"Person"}, []string{"KNOWS"}, Both, 6},
		{"an unlabeled end takes the narrowest label with the relationship", nil, nil, []string{"KNOWS"}, Out, 3},
		{"share of the destination label", []string{"Person"}, []string{"Post"}, []string{"KNOWS"}, Out, 0},
	}
	for _, c := range cases {
		if got := s.Degree(c.from, c.types, c.dir, c.to); !near(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
