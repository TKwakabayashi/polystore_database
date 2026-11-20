// Package stats holds the statistics the planner uses: node counts per label and relationship counts
// for (label)-[type]->() and (label)<-[type]-() (average degrees). They are O(1) values of Neo4j's
// count store; value distributions are not kept.
//
// `admin stats` collects them into stats.json next to the mapping (PathFor). Without the file the
// planner falls back to rules.
package stats

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Stats are counts over the whole graph. Out[L][T] is the number of (:L)-[:T]->(), In[L][T] of (:L)<-[:T]-().
type Stats struct {
	Nodes    int64                       `json:"nodes"`
	Labels   map[string]int64            `json:"labels"`
	RelTypes map[string]int64            `json:"relTypes"`
	Out      map[string]map[string]int64 `json:"out"`
	In       map[string]map[string]int64 `json:"in"`
}

// PathFor returns stats.json in the directory of mapping.json.
func PathFor(mappingPath string) string {
	return filepath.Join(filepath.Dir(mappingPath), "stats.json")
}

// Load reads the statistics; a missing file gives (nil, nil).
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

// Save writes the statistics.
func (s *Stats) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ---- estimates (used by the planner) ----

// Dir is a relationship's direction seen from From.
type Dir int

const (
	Out Dir = iota
	In
	Both
)

// Count bounds the nodes that have all the labels (the smallest label count); with no known label, all nodes.
func (s *Stats) Count(labels []string) float64 {
	n := float64(s.Nodes)
	for _, l := range labels {
		// A label missing from the statistics (added later, or a sub-label not in the data) does not narrow.
		if c, ok := s.Labels[l]; ok && float64(c) < n {
			n = float64(c)
		}
	}
	return clampMin(n)
}

// Degree is the average number of nodes with labels to reached from one from-node over relationships
// of types (all types if empty) in direction dir.
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

// oneWay is the degree for one type and direction: (from-side ends / from count) x (share of to-side ends that have the to labels).
func (s *Stats) oneWay(from []string, t string, src map[string]map[string]int64, to []string, dst map[string]map[string]int64) float64 {
	total := float64(s.RelTypes[t])
	if total == 0 {
		return 0
	}
	l, ok := s.narrowest(from)
	if !ok {
		// An end of unknown label is taken as the narrowest label having that relationship type (Person for KNOWS).
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

// narrowest is the known label with the fewest nodes.
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

// narrowestWith is the smallest label with src[l][t] > 0 (it has type t in this direction).
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
