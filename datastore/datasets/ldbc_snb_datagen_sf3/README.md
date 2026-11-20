# ldbc_snb_datagen — schema & sizes

LDBC SNB social network (BI, composite-merged-fk). KNOWS social graph + nested comment reply chains (REPLY_TO_COMMENT); Post/Comment carry content text, Post an imageFile. Built at scale factors sf1/sf3/sf10/sf30.

- **Manifest name**: `ldbc_snb`
- **Source**: ldbc/datagen-standalone (Spark)
- **License / use**: LDBC / Apache-2.0 generator, synthetic data
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Person | person_id | 170,654 | 31.7 MB | person_id:string, first_name:string, last_name:string, gender:string, birthday:string, browser:string, creation_date:string, location_ip:string, speaks:string, email:string |
| Forum | forum_id | 1,659,632 | 110.0 MB | forum_id:string, title:string, creation_date:string |
| Post | post_id | 21,651,342 | 2.3 GB | post_id:string, content:string, image_file:string, length:int, browser:string, creation_date:string |
| Comment | comment_id | 54,737,515 | 4.7 GB | comment_id:string, content:string, length:int, browser:string, creation_date:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| KNOWS | Person → Person | 5,524,302 | 150.0 MB | — |
| POSTED | Person → Post | 21,651,342 | 588.7 MB | — |
| COMMENTED | Person → Comment | 54,737,515 | 1.5 GB | — |
| REPLY_TO_POST | Comment → Post | 29,607,257 | 842.4 MB | — |
| REPLY_TO_COMMENT | Comment → Comment | 25,130,258 | 718.8 MB | — |
| IN_FORUM | Post → Forum | 21,651,342 | 607.9 MB | — |
| MEMBER_OF | Person → Forum | 90,198,118 | 2.4 GB | — |
| MODERATES | Person → Forum | 1,659,632 | 45.2 MB | — |

**Totals**: 78,219,143 nodes, 250,159,766 relationships, ~13.9 GB of CSV (property text dominates the store).

## Scale factors (dumps built)

| scale | nodes | relationships | dump |
|---|--:|--:|--:|
| sf1  | 2,971,786 | 8,904,937 | 254 MB |
| sf3  | 8,487,591 | 25,989,160 | 752 MB |
| sf10 | 27,205,783 | 85,647,152 | 2.5 GB |
| sf30 | 78,219,143 | 250,159,766 | 7.4 GB |

(The size table above is for the current export, **sf30**.)
