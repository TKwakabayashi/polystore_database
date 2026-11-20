# openalex — schema & sizes

OpenAlex scholarly graph. (:Author)-[:AUTHORED]->(:Work) and citation chains (:Work)-[:CITES]->(:Work).

- **Manifest name**: `openalex`
- **Source**: OpenAlex API / S3 snapshot
- **License / use**: CC0
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Work | work_id | 21,914,490 | 2.3 GB | work_id:string, title:string, year:int, cited_by_count:int, venue:string |
| Author | author_id | 9,634,353 | 110.3 MB | author_id:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| AUTHORED | Author → Work | 33,549,633 | 766.3 MB | — |
| CITES | Work → Work | 163,773,009 | 3.6 GB | — |

**Totals**: 31,548,843 nodes, 197,322,642 relationships, ~6.8 GB of CSV (property text dominates the store).
