# stack_history — schema & sizes

Stack Exchange WITH full edit history. Adds per-post edit-revision chains (:Rev)-[:PREV_EDIT*] plus every edited text snapshot (PostHistory).

- **Manifest name**: `stackhistory`
- **Source**: archive.org/details/stackexchange
- **License / use**: CC BY-SA 4.0
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Post | post_id | 244,066 | 310.6 MB | post_id:string, post_type:int, score:int, view_count:int, creation_date:string, title:string, tags:string, body:string |
| User | user_id | 375,100 | 16.3 MB | user_id:string, display_name:string, reputation:int, creation_date:string |
| Rev | rev_id | 653,560 | 580.5 MB | rev_id:string, post_id:string, history_type:int, creation_date:string, comment:string, text:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| AUTHORED | User → Post | 235,574 | 3.3 MB | — |
| PARENT_OF | Post → Post | 178,628 | 2.6 MB | — |
| LINKS_TO | Post → Post | 13,562 | 205.1 KB | — |
| HAS_EDIT | Post → Rev | 653,560 | 10.0 MB | — |
| PREV_EDIT | Rev → Rev | 409,489 | 6.5 MB | — |
| MADE_EDIT | User → Rev | 586,439 | 8.5 MB | — |

**Totals**: 1,272,726 nodes, 2,077,252 relationships, ~938.5 MB of CSV (property text dominates the store).
