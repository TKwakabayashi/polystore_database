package setup

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// typeLabel adds a sub-label to nodes of base label Base according to their type property, matched
// case-insensitively against the sub-label name ('city' and 'City' both give City).
type typeLabel struct {
	Base   string
	Labels []string
}

// ldbcTypeLabels are the sub-labels the LDBC dump distinguishes by type instead of labels. Workloads use
// (c:City), (m:Post) and so on, so setup adds them.
var ldbcTypeLabels = []typeLabel{
	{Base: "Place", Labels: []string{"City", "Country", "Continent"}},
	{Base: "Message", Labels: []string{"Post", "Comment"}},
	{Base: "Organisation", Labels: []string{"Company", "University"}},
}

// applyTypeLabels adds the sub-labels per rules (only to nodes without them; idempotent) and warns with the
// count of base-label nodes that match no sub-label.
func applyTypeLabels(ctx context.Context, session neo4j.SessionWithContext, rules []typeLabel) error {
	for _, r := range rules {
		for _, l := range r.Labels {
			fmt.Printf("Labeling :%s by type as :%s...\n", r.Base, l)
			inner := fmt.Sprintf("MATCH (n:`%s`) WHERE toLower(n.type) = '%s' AND NOT n:`%s` RETURN n", r.Base, strings.ToLower(l), l)
			query := "CALL apoc.periodic.iterate($inner, $set, {batchSize: 10000, parallel: false}) " +
				"YIELD total, failedOperations, errorMessages RETURN total, failedOperations, errorMessages"
			rec, err := single(ctx, session, query, map[string]any{"inner": inner, "set": fmt.Sprintf("SET n:`%s`", l)})
			if err != nil {
				return fmt.Errorf("label :%s as :%s: %w", r.Base, l, err)
			}
			if failed, _ := rec.Get("failedOperations"); failed != int64(0) {
				msgs, _ := rec.Get("errorMessages")
				return fmt.Errorf("label :%s as :%s: %v failed: %v", r.Base, l, failed, msgs)
			}
			total, _ := rec.Get("total")
			fmt.Printf("  %v\n", total)
		}
		conds := make([]string, len(r.Labels))
		for i, l := range r.Labels {
			conds[i] = fmt.Sprintf("n:`%s`", l)
		}
		rec, err := single(ctx, session, fmt.Sprintf("MATCH (n:`%s`) WHERE NOT (%s) RETURN count(n) AS c, collect(DISTINCT n.type)[..5] AS types",
			r.Base, strings.Join(conds, " OR ")), nil)
		if err != nil {
			return fmt.Errorf("count unlabeled :%s: %w", r.Base, err)
		}
		if c, _ := rec.Get("c"); c != int64(0) {
			types, _ := rec.Get("types")
			fmt.Printf("warning: %[2]v of :%[1]s match none of the sublabels %[3]v (types, e.g. %[4]v)\n", r.Base, c, r.Labels, types)
		}
	}
	return nil
}

// single runs a one-row query in a write transaction and returns the row.
func single(ctx context.Context, session neo4j.SessionWithContext, query string, params map[string]any) (*neo4j.Record, error) {
	rec, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return res.Single(ctx)
	})
	if err != nil {
		return nil, err
	}
	return rec.(*neo4j.Record), nil
}
