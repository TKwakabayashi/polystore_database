# yelp — schema & sizes

Yelp Open Dataset businesses/users/reviews. (:User)-[:REVIEWED]->(:Business). NOTE: not loaded (manual download required by Yelp's terms).

- **Manifest name**: `yelp`
- **Source**: yelp.com/dataset (manual)
- **License / use**: Yelp Dataset License (no redistribution)
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`
- **NOTE**: no exported CSV on disk; sizes below are unavailable (dataset not loaded).

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Business | business_id | 0 | 0 B | business_id:string, name:string, city:string, stars:double, review_count:int, categories:string[] |
| User | user_id | 0 | 0 B | user_id:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| REVIEWED | User → Business | 0 | 0 B | — |

**Totals**: 0 nodes, 0 relationships, ~0 B of CSV (property text dominates the store).
