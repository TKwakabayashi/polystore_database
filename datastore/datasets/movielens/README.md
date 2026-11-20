# movielens — schema & sizes

MovieLens ratings. (:User)-[:RATED]->(:Movie); rating+timestamp on the edge.

- **Manifest name**: `movielens`
- **Source**: files.grouplens.org (ml-25m / ml-latest-small)
- **License / use**: GroupLens research terms
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Movie | movie_id | 62,423 | 3.1 MB | movie_id:string, title:string, year:int, genres:string[] |
| User | user_id | 162,541 | 1.1 MB | user_id:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| RATED | User → Movie | 25,000,095 | 646.8 MB | rating:float, timestamp:long |

**Totals**: 224,964 nodes, 25,000,095 relationships, ~651.1 MB of CSV (property text dominates the store).
