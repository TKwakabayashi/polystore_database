package setup

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ensureUUIDIndexes creates a uuid uniqueness constraint for every Neo4j label and relationship type
// (idempotent). Neo4j property indexes are per label and type, so there is no common label: queries (the store
// boundary) always give variables their labels and types. The names match those migrations create
// (<lowercase name>_uuid_unique). An index on the same property cannot coexist, so an old _uuid_idx is dropped first.
func ensureUUIDIndexes(ctx context.Context, session neo4j.SessionWithContext) error {
	labels, err := tokens(ctx, session, "CALL db.labels() YIELD label RETURN label AS t")
	if err != nil {
		return fmt.Errorf("failed to fetch labels: %w", err)
	}
	types, err := tokens(ctx, session, "CALL db.relationshipTypes() YIELD relationshipType RETURN relationshipType AS t")
	if err != nil {
		return fmt.Errorf("failed to fetch relationship types: %w", err)
	}
	var stmts []string
	for _, l := range labels {
		stmts = append(stmts, fmt.Sprintf("CREATE CONSTRAINT `%s_uuid_unique` IF NOT EXISTS FOR (n:`%s`) REQUIRE n.uuid IS UNIQUE",
			strings.ToLower(l), l))
	}
	for _, t := range types {
		stmts = append(stmts, fmt.Sprintf("DROP INDEX `%s_uuid_idx` IF EXISTS", strings.ToLower(t)),
			fmt.Sprintf("CREATE CONSTRAINT `%s_uuid_unique` IF NOT EXISTS FOR ()-[r:`%s`]-() REQUIRE r.uuid IS UNIQUE",
				strings.ToLower(t), t))
	}
	fmt.Printf("Creating uuid constraints/indexes (%d labels, %d relationship types)...\n", len(labels), len(types))
	for _, stmt := range stmts {
		if _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			res, err := tx.Run(ctx, stmt, nil)
			if err != nil {
				return nil, err
			}
			return res.Consume(ctx)
		}); err != nil {
			return fmt.Errorf("%s: %w", stmt, err)
		}
	}
	return nil
}

// tokens returns the single column t of the results as strings.
func tokens(ctx context.Context, session neo4j.SessionWithContext, query string) ([]string, error) {
	res, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	var out []string
	for res.Next(ctx) {
		if t, ok := res.Record().Get("t"); ok {
			if s, ok := t.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out, res.Err()
}
