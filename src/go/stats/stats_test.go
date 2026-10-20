package stats

import (
	"math"
	"testing"
)

// 小さなグラフ: Person 10 人、Post 100 件。HAS_CREATOR は Post → Person に 100 本、KNOWS は Person 間に 30 本。
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
		t.Errorf("ラベル無しは全ノード: %v", got)
	}
	if got := s.Count([]string{"City"}); got != 110 {
		t.Errorf("統計に無いラベルは絞らない: %v", got)
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
		{"person の投稿数", []string{"Person"}, []string{"Post"}, []string{"HAS_CREATOR"}, In, 10},
		{"post の作者", []string{"Post"}, []string{"Person"}, []string{"HAS_CREATOR"}, Out, 1},
		{"KNOWS 無向", []string{"Person"}, []string{"Person"}, []string{"KNOWS"}, Both, 6},
		{"ラベルの分からない端は関係を持つ最も狭いラベル", nil, nil, []string{"KNOWS"}, Out, 3},
		{"行き先のラベルの割合", []string{"Person"}, []string{"Post"}, []string{"KNOWS"}, Out, 0},
	}
	for _, c := range cases {
		if got := s.Degree(c.from, c.types, c.dir, c.to); !near(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
