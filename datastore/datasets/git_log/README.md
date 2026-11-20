# git_log — schema & sizes

Git commit-history graph. Deep (:Commit)-[:PARENT*] ancestry chains; full commit message text + per-file numstat as the heavy/log properties.

- **Manifest name**: `gitlog`
- **Source**: git clone (default kubernetes/kubernetes)
- **License / use**: Apache-2.0 (default repo); metadata only
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Commit | sha | 160,059 | 59.9 MB | sha:string, author_name:string, author_email:string, authored_date:string, committer_name:string, committed_date:string, subject:string, message:string, files_changed:int, insertions:int, deletions:int |
| Author | email | 5,455 | 189.8 KB | email:string, name:string |
| File | path | 100,746 | 6.3 MB | path:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| PARENT | Commit → Commit | 225,451 | 17.6 MB | — |
| AUTHORED | Author → Commit | 160,059 | 10.4 MB | — |
| TOUCHED | Commit → File | 720,311 | 67.0 MB | insertions:int, deletions:int |

**Totals**: 266,260 nodes, 1,105,821 relationships, ~161.5 MB of CSV (property text dominates the store).
