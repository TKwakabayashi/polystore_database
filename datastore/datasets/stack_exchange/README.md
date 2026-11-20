# stack_exchange — schema & sizes

Stack Exchange Q&A. (:User)-[:AUTHORED]->(:Post), question->answer (PARENT_OF), related links (LINKS_TO). Post body text kept.

- **Manifest name**: `stackexchange`
- **Source**: archive.org/details/stackexchange
- **License / use**: CC BY-SA 4.0
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Post | post_id | 244,066 | 306.7 MB | post_id:string, post_type:int, score:int, view_count:int, creation_date:string, tags:string, body:string |
| User | user_id | 375,100 | 16.3 MB | user_id:string, display_name:string, reputation:int, creation_date:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| AUTHORED | User → Post | 235,574 | 3.3 MB | — |
| PARENT_OF | Post → Post | 178,628 | 2.6 MB | — |
| LINKS_TO | Post → Post | 13,562 | 205.1 KB | — |

**Totals**: 619,166 nodes, 427,764 relationships, ~329.2 MB of CSV (property text dominates the store).
