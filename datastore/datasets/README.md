# datasets

One folder per dataset. Nothing here changes when a dataset is started; `admin setup` copies what it
needs into `../runtime/`.

| file | content |
|---|---|
| `dataset.env` | Docker settings (dataset name, image, database name, memory; `VOLUMES` keeps older container and volume names) |
| `mapping.json` | the schema template: every entity, relationship and property with its type, all placed on Neo4j |
| `stats.json` | planning statistics fixed for this dataset (only where goldens use them) |
| `README.md` | source, licence, schema and size |
| `dumps/` | Neo4j dumps (`<database>.dump`) and older generations in `dumps/_archive/` (not in git) |

Datasets of one schema at several scales (`ldbc_snb_datagen_sf1` … `_sf30`) each have their folder; their
`mapping.json` files are copies.
