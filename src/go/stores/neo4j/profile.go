package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/query/result"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

// Profile runs q under PROFILE: the plan tree with rows and db hits per operator.
func (DB) Profile(ctx context.Context, reg *conn.Registry, q *stores.Query, in []stores.Key) (*result.StoreProfile, error) {
	d, err := driver(reg)
	if err != nil {
		return nil, err
	}
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(context.WithoutCancel(ctx))
	text := q.TextBody().Text
	res, err := sess.Run(ctx, "PROFILE "+text, graphParams(q, in))
	if err != nil {
		return nil, fmt.Errorf("cypher %q: %w", text, err)
	}
	sum, err := res.Consume(ctx)
	if err != nil {
		return nil, fmt.Errorf("cypher %q: %w", text, err)
	}
	p := &result.StoreProfile{Store: "neo4j", Query: text, Time: sum.ResultAvailableAfter() + sum.ResultConsumedAfter()}
	if plan := sum.Profile(); plan != nil {
		p.Rows = plan.Records()
		p.Raw = profileTree(plan, &p.DBHits)
	}
	return p, nil
}

// profileTree turns a profiled plan into maps (for JSON) and sums the db hits.
func profileTree(n neo4j.ProfiledPlan, hits *int64) map[string]any {
	*hits += n.DbHits()
	m := map[string]any{"operator": n.Operator(), "rows": n.Records(), "db_hits": n.DbHits(),
		"identifiers": n.Identifiers(), "arguments": n.Arguments()}
	var kids []any
	for _, c := range n.Children() {
		kids = append(kids, profileTree(c, hits))
	}
	if len(kids) > 0 {
		m["children"] = kids
	}
	return m
}
