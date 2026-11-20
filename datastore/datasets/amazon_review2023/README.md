# amazon_review2023 — schema & sizes

Amazon Reviews 2023 (multimodal). Users review Items; review text + rating live on the REVIEWED edge, item images kept as URLs.

- **Manifest name**: `amazon_reviews_2023`
- **Source**: HuggingFace McAuley-Lab/Amazon-Reviews-2023
- **License / use**: research use per dataset card
- **Placement**: all properties in `graph` (single-store Neo4j) — see `mapping.json`

## Node labels

| label | key | nodes | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| Item | parent_asin | 3,735,584 | 2.3 GB | parent_asin:string, title:string, price:double, average_rating:double, rating_number:int, main_category:string, description:string, image_urls:string[] |
| User | user_id | 23,235,852 | 642.6 MB | user_id:string |

## Relationships

| type | from → to | edges | on-disk (CSV) | properties |
|---|---|--:|--:|---|
| REVIEWED | User → Item | 66,737,952 | 16.7 GB | timestamp:long, rating:float, title:string, text:string, helpful_vote:int, verified_purchase:boolean |

**Totals**: 26,971,436 nodes, 66,737,952 relationships, ~19.6 GB of CSV (property text dominates the store).
