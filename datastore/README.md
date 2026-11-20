# datastore

Everything about the data: the Docker stack, each dataset's static files and dumps, the state of the
running dataset, and what is shared across datasets. Neo4j / MongoDB / MySQL / ClickHouse / Cassandra run
in Docker; LevelDB / RocksDB are embedded (their files are under runtime/).

## Layout

```
datastore/
├── docker/                    # every dataset: the stack's definition
│   ├── compose.yaml           #   run with --env-file ../datasets/<dataset>/dataset.env
│   └── .env                   #   the default without --env-file (ldbc_snb_sf1)
├── datasets/<dataset>/        # static, one folder per dataset (nothing changes when it starts)
│   ├── dataset.env            #   Docker settings: DATASET, VOLUMES (older container names), image, memory
│   ├── mapping.json           #   the schema template (every property on Neo4j)
│   ├── stats.json             #   fixed statistics (where goldens use them)
│   ├── README.md              #   source, licence, schema, size
│   └── dumps/                 #   <database>.dump and dumps/_archive/ (not in git)
├── runtime/                   # the running dataset only; admin setup rebuilds it (not in git)
│   ├── dataset                #   the dataset it was built for
│   ├── mapping.json           #   the record: where the data is now (migrations rewrite it)
│   ├── stats.json             #   planning statistics (admin stats)
│   └── leveldb/  rocksdb/     #   the embedded stores
├── shared/                    # kept across datasets
│   ├── uuid_leases.json       #   the uuid lease ledger (not in git)
│   └── uuid_leases.example.json
└── scripts/                   # snapshot.sh, migrate/ (one-off records)
```

## Usage

```bash
cd datastore/docker
# start / stop a dataset (only one at a time: the ports are shared)
docker compose --env-file ../datasets/ldbc_snb_sf1/dataset.env up -d
docker compose --env-file ../datasets/ldbc_snb_sf1/dataset.env stop

# restore a dump from datasets/<dataset>/dumps/ (with neo4j stopped)
docker compose --env-file ../datasets/openalex/dataset.env --profile load run --rm neo4j-load

# after starting or switching a dataset, rebuild runtime/ (from src/go)
go run ./cmd/admin setup

# take a dump of a dataset (from datastore/)
scripts/snapshot.sh ldbc_snb_sf1
```

`admin setup` takes the dataset from the running containers' compose project, rebuilds `runtime/` from the
dataset's `mapping.json`, assigns missing uuids (taking a lease from `shared/uuid_leases.json`), creates the
uuid indexes and empties the other stores (it also creates the Cassandra keyspace; ClickHouse's database is
created on connect).

Every dataset has its own containers and volumes (`polystore_<VOLUMES or DATASET>_neo4j_data`), so switching
needs no reload. `VOLUMES` keeps the names made before a dataset was renamed (`ldbc_snb_sf1` uses
`ldbc_sf1`).

## uuid leases

uuids are made from a lease number and a sequence (`id.Allocator`). Lease numbers are one sequence across
every dataset, so uuids never collide between datasets, and `shared/uuid_leases.json` records which dataset
took which lease. Keep it with the dumps: a dump with uuids carries the leases it used. Without the ledger
uuid assignment stops (create it from the example only for data without uuids).

## Adding a dataset

1. Create `datasets/<dataset>/` with `dataset.env` (copy one: `DATASET` / `NEO4J_DATABASE` / `NEO4J_IMAGE` /
   memory), `mapping.json` (every property on Neo4j) and a `README.md`.
2. Put the dump at `datasets/<dataset>/dumps/<database>.dump` (**the file name is the database name**:
   `neo4j-admin database load` looks for dumps that way).
3. Restore with `--profile load`, start with `up -d`, then `admin setup`.

## Notes

- **Neo4j stores are not compatible across versions and cannot be downgraded.** `NEO4J_IMAGE` must match the
  major version of the Neo4j that made the dump; moving versions needs a logical copy through CSV
  (`scripts/migrate/ldbc_to_neo4j5.sh`).
- **Block-format dumps are Enterprise only.** `datasets/ldbc_snb_sf1/dumps/_archive/neo4j2025.dump` is one
  and needs `neo4j:2025-enterprise`.
- **Neo4j refuses to start when heap + pagecache exceed the Docker VM's memory** (7.7G now; every env keeps
  heap 4G / pagecache 1G or less).
