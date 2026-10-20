// Package stats はクエリ処理系（刷新版）が計画に使う統計: ラベルごとのノード数と、
// (ラベル)-[タイプ]->() / (ラベル)<-[タイプ]-() の関係数（＝平均次数）。どれも Neo4j の count store から
// O(1) で取れる値だけで、プロパティの値の分布（異なり数など）は持たない。
//
// 収集は `-mode stats`（Collect → Save）。ファイルは mapping.json と同じディレクトリの stats.json
// （PathFor）。ファイルが無ければ計画は統計を使わない規則に戻る。
package stats

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Stats はグラフ全体の件数。Out[L][T] は (:L)-[:T]->() の数、In[L][T] は (:L)<-[:T]-() の数。
type Stats struct {
	Nodes    int64                       `json:"nodes"`
	Labels   map[string]int64            `json:"labels"`
	RelTypes map[string]int64            `json:"relTypes"`
	Out      map[string]map[string]int64 `json:"out"`
	In       map[string]map[string]int64 `json:"in"`
}

// PathFor は mapping.json と同じディレクトリの stats.json。
func PathFor(mappingPath string) string {
	return filepath.Join(filepath.Dir(mappingPath), "stats.json")
}

// Load は統計を読む。ファイルが無ければ (nil, nil)。
func Load(path string) (*Stats, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s Stats
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("stats: %s: %w", path, err)
	}
	return &s, nil
}

// Save は統計を書く。
func (s *Stats) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ---- 推定（計画が使う）----

// Dir は From から見た関係の向き。
type Dir int

const (
	Out Dir = iota
	In
	Both
)

// Count はラベルを全部持つノード数の上界（最も少ないラベルの件数）。既知のラベルが無ければ全ノード数。
func (s *Stats) Count(labels []string) float64 {
	n := float64(s.Nodes)
	for _, l := range labels {
		// 統計に無いラベル（収集後に増えた・データに無いサブラベル）は情報なしとして絞らない。
		if c, ok := s.Labels[l]; ok && float64(c) < n {
			n = float64(c)
		}
	}
	return clampMin(n)
}

// Degree は from のノード 1 件あたり、タイプ types（空なら全タイプ）の関係を向き dir にたどって
// to のラベルを持つノードへ着く平均件数。
func (s *Stats) Degree(from []string, types []string, dir Dir, to []string) float64 {
	if len(types) == 0 {
		for t := range s.RelTypes {
			types = append(types, t)
		}
	}
	var deg float64
	for _, t := range types {
		switch dir {
		case Out:
			deg += s.oneWay(from, t, s.Out, to, s.In)
		case In:
			deg += s.oneWay(from, t, s.In, to, s.Out)
		default:
			deg += s.oneWay(from, t, s.Out, to, s.In) + s.oneWay(from, t, s.In, to, s.Out)
		}
	}
	return deg
}

// oneWay は 1 タイプ・1 方向の次数: (from 側の端の数 / from の件数) × (to 側の端のうち to のラベルの割合)。
func (s *Stats) oneWay(from []string, t string, src map[string]map[string]int64, to []string, dst map[string]map[string]int64) float64 {
	total := float64(s.RelTypes[t])
	if total == 0 {
		return 0
	}
	l, ok := s.narrowest(from)
	if !ok {
		// ラベルの分からない端は、そのタイプの関係を持つ最も狭いラベルとみなす（KNOWS の端なら Person）。
		l, ok = s.narrowestWith(src, t)
	}
	deg := total / clampMin(float64(s.Nodes))
	if ok {
		deg = float64(src[l][t]) / clampMin(float64(s.Labels[l]))
	}
	if l, ok := s.narrowest(to); ok {
		deg *= float64(dst[l][t]) / total
	}
	return deg
}

// narrowest は件数が最も少ない既知のラベル。
func (s *Stats) narrowest(labels []string) (string, bool) {
	best, found := "", false
	for _, l := range labels {
		c, ok := s.Labels[l]
		if !ok {
			continue
		}
		if !found || c < s.Labels[best] || (c == s.Labels[best] && l < best) {
			best, found = l, true
		}
	}
	return best, found
}

// narrowestWith は、src[l][t] > 0（タイプ t の関係をこの向きに持つ）ラベルのうち件数が最も少ないもの。
func (s *Stats) narrowestWith(src map[string]map[string]int64, t string) (string, bool) {
	best, found := "", false
	for l, byType := range src {
		if byType[t] == 0 {
			continue
		}
		if !found || s.Labels[l] < s.Labels[best] || (s.Labels[l] == s.Labels[best] && l < best) {
			best, found = l, true
		}
	}
	return best, found
}

func clampMin(x float64) float64 {
	if x < 1 {
		return 1
	}
	return x
}
