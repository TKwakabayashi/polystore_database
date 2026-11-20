# imdb — schema & sizes

IMDb non-commercial titles/people. (:Person)-[:WORKED_ON]->(:Title); ~91M edges.

- **Manifest name**: `imdb`
- **Source**: datasets.imdbws.com
- **License / use**: IMDb non-commercial terms
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Title | tconst | 12,655,827 | 745.8 MB | tconst:string, primary_title:string, title_type:string, start_year:int, runtime:int, genres:string[], average_rating:double, num_votes:int |
| Person | nconst | 7,138,948 | 270.5 MB | nconst:string, primary_name:string, birth_year:int, professions:string[] |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| WORKED_ON | Person → Title | 91,176,426 | 1.8 GB | — |

**Totals**: 19,794,775 nodes, 91,176,426 relationships, ~2.7 GB of CSV (property text dominates the store).
