package stats

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Collect は Neo4j の count store から統計を集める（ラベル数 × タイプ数 × 2 の count クエリ。どれも O(1)）。
func Collect(ctx context.Context, driver neo4j.DriverWithContext) (*Stats, error) {
	sess := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(ctx)

	count := func(q string) (int64, error) {
		res, err := sess.Run(ctx, q, nil)
		if err != nil {
			return 0, fmt.Errorf("stats: %s: %w", q, err)
		}
		rec, err := res.Single(ctx)
		if err != nil {
			return 0, fmt.Errorf("stats: %s: %w", q, err)
		}
		n, _ := rec.Values[0].(int64)
		return n, nil
	}
	names := func(q string) ([]string, error) {
		res, err := sess.Run(ctx, q, nil)
		if err != nil {
			return nil, fmt.Errorf("stats: %s: %w", q, err)
		}
		var out []string
		for res.Next(ctx) {
			if s, ok := res.Record().Values[0].(string); ok {
				out = append(out, s)
			}
		}
		return out, res.Err()
	}

	s := &Stats{Labels: map[string]int64{}, RelTypes: map[string]int64{},
		Out: map[string]map[string]int64{}, In: map[string]map[string]int64{}}
	var err error
	if s.Nodes, err = count("MATCH (n) RETURN count(n)"); err != nil {
		return nil, err
	}
	labels, err := names("CALL db.labels()")
	if err != nil {
		return nil, err
	}
	types, err := names("CALL db.relationshipTypes()")
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if s.RelTypes[t], err = count(fmt.Sprintf("MATCH ()-[r:`%s`]->() RETURN count(r)", t)); err != nil {
			return nil, err
		}
	}
	for _, l := range labels {
		if s.Labels[l], err = count(fmt.Sprintf("MATCH (n:`%s`) RETURN count(n)", l)); err != nil {
			return nil, err
		}
		s.Out[l], s.In[l] = map[string]int64{}, map[string]int64{}
		for _, t := range types {
			out, err := count(fmt.Sprintf("MATCH (:`%s`)-[r:`%s`]->() RETURN count(r)", l, t))
			if err != nil {
				return nil, err
			}
			in, err := count(fmt.Sprintf("MATCH (:`%s`)<-[r:`%s`]-() RETURN count(r)", l, t))
			if err != nil {
				return nil, err
			}
			if out > 0 {
				s.Out[l][t] = out
			}
			if in > 0 {
				s.In[l][t] = in
			}
		}
	}
	return s, nil
}
