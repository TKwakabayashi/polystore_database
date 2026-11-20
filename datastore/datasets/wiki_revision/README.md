# wiki_revision — schema & sizes

Wikipedia full page-edit history. Per-page revision chains (:Revision)-[:PREV*] (thousands deep); full wikitext of every revision + parsed image links.

- **Manifest name**: `wikirevision`
- **Source**: dumps.wikimedia.org
- **License / use**: CC BY-SA 4.0 + GFDL
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Page | page_id | 32,969 | 690.4 KB | page_id:string, title:string, namespace:int |
| Revision | rev_id | 2,500,000 | 17.5 GB | rev_id:string, page_id:string, timestamp:string, comment:string, minor:boolean, text_bytes:long, sha1:string, images:string[], text:string |
| Contributor | name | 314,623 | 6.4 MB | name:string, user_id:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| HAS_REVISION | Page → Revision | 2,500,000 | 31.9 MB | — |
| PREV | Revision → Revision | 2,467,025 | 36.6 MB | — |
| EDITED | Contributor → Revision | 2,499,487 | 49.2 MB | — |

**Totals**: 2,847,592 nodes, 7,466,512 relationships, ~17.6 GB of CSV (property text dominates the store).
