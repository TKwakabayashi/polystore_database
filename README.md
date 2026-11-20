# polystore_database

A research polystore that runs Cypher queries across heterogeneous databases: Neo4j, MongoDB, MySQL,
ClickHouse, Cassandra and LevelDB / RocksDB. It plans each query against a data catalog based on
U-Schema [1], which records where every property is placed, and sends one native query per store. Data
placement and the iteration model can be switched to compare performance.

---

## Architecture

```
Cypher
   │
   ▼
query/cypher     parse (ANTLR) → AST → a declarative logical tree (query/ir)
   │
   ▼
query/logical    pattern planning (start vertex, edge order, anchored splits; rules or statistics)
                 and logical rewrites
   │
   ▼
query/physical   placement (which store reads each property), merging consecutive operators of one
                 store into one store query, absorption, deferred fetches
   │
   ▼
query/exec       five iteration models over the same plan; coordinator operators
   │
   ▼
stores/<db>      per database: capabilities, compilation to the native query, execution,
                 its own profile (L4) and migration reads / writes
```

| Package | Role |
|---|---|
| `query/` | Cypher to result: `query.Plan` / `query.Run` / `query.RenderPlan` |
| `stores/` | The database boundary. `stores.DB` (capabilities, `Compile`, `Stream`), `Shipper`, `Profiler`, `Mover`; `stores/dbs` dispatches by kind and runs queries through a `Session` |
| `catalog/` | Knowledge about an instance: the mapping (placement, types, unique keys, uuid storage form) and statistics |
| `data/` | Maintenance that rewrites data: migrations between stores, consolidation back to Neo4j, setup |
| `lab/` | Experiments: the spec-driven runner (`lab run`), the operator benchmark, workload definitions |
| `base/` | Foundations: `id` (the only authority on UUIDs), `codec` (value encoding, KV keys), `kind` (database kinds and limits) |
| `settings/` | Default values of internal switches (read only by `cmd` and `lab`) |

Iteration models (`exec.Model`) all give the same result:

| Model | How rows flow |
|---|---|
| `bulk` | every operator's full output is materialized |
| `volcano` | pull, one input row per store round trip |
| `vectorized` | pull, `VectorSize` rows per round trip |
| `stream` | push, one goroutine per operator |
| `vecstream` | pull with an exchange of parallel workers |

`docs/EXTENDING.md` (local) describes the layers, the UUID invariants and how to add a database,
operator, iteration model or workload.

---

## Layout

```
polystore_database/
├── src/go/                 the system (Go module polystore_database/src/go)
│   ├── cmd/{query,admin,lab}/   entry points: one query / instance maintenance / experiments
│   ├── query/ stores/ catalog/ data/ lab/ base/ settings/ internal/
│   └── tests/              tests that need databases: db/ (the docker stack), throwaway/ (a throwaway environment)
├── config/                 config.json (copy config.example.json): connections and the dataset
├── datastore/              the data: docker/, datasets/<dataset>/ (static), runtime/ (the running dataset), shared/
├── tools/antlr/            the ANTLR jar (parser generation)
├── analysis/               scripts that read lab run results
├── docs/                   design documents (local)
└── results/                experiment results (local)
```

Unit tests and goldens sit next to the code (`*_test.go`, `testdata/`), as Go expects.

---

## Setup and running

Requirements: Docker and Go. Commands run from `src/go/` unless noted.

### 1. Start the databases (from `datastore/docker/`)

```bash
cd datastore/docker
docker compose --env-file ../datasets/ldbc_snb_sf1/dataset.env up -d
# restore Neo4j from a dump (first time, with neo4j stopped):
#   docker compose --env-file ../datasets/ldbc_snb_sf1/dataset.env --profile load run --rm neo4j-load
```

The ports are shared, so only one dataset can be up at a time. LevelDB / RocksDB are embedded (no
container). See `datastore/README.md`.

### 2. Configuration

Copy `config/config.example.json` to `config/config.json` and fill in the connections and the dataset.
Paths in it are relative to the config file. RocksDB needs a `-tags rocksdb` build (cgo; see
`src/go/stores/conn/rocksdb.go`).

### 3. Data setup (after starting or switching a dataset)

```bash
cd src/go
go run ./cmd/admin setup   # rebuild datastore/runtime/: the record, uuids and indexes on Neo4j, empty other stores
go run ./cmd/admin stats   # statistics for cost-based planning (again after changing data)
```

### 4. One query

```bash
go run ./cmd/query -workload IC11 [-target custom|neo4j|both] [-format rows|timing|detail]
```

### 5. Migrations

```bash
go run ./cmd/admin migrate -workload AGG5 -migmode graph_to_rdb   # move a workload's properties
go run ./cmd/admin consolidate [-exclude Label.prop]              # bring everything back to Neo4j
go run ./cmd/lab workflow -workload AGG5 -migmode graph_to_rdb    # Neo4j baseline, migration, this system
```

`-migmode` is `<from>_to_<to>`, with database names (`neo4j`, `mysql`, `mongo`, `clickhouse`,
`cassandra`, `leveldb`) or short names (`graph`, `rdb`, `doc`, `col`, `wide`, `kvs`).

### 6. Experiments

```bash
go run ./cmd/lab run -spec lab/specs/models.json
python3 ../../analysis/summarize.py ../../results/<folder>
```

A spec (fields: `Spec` in `src/go/lab/runner/spec.go`, examples in `src/go/lab/specs/`) runs workloads
× query versions × parameter cases × placements × iteration models × tunings. `"verify": true` compares
every result with Neo4j running the original Cypher; `"mode": "profile"` records spans per operator, with
store calls split into first row / receive / decode; `"store_profile": true` records each database's own
profile (Neo4j PROFILE, MySQL EXPLAIN ANALYZE, MongoDB explain, ClickHouse query_log, Cassandra tracing);
`"pprof": true` writes Go profiles. Specs with `"mode": "opbench"` run the per-operator benchmark
(`src/go/lab/opbench/specs/`). Results go to `results/<date>_<name>/`.

Internal behavior (iteration model, pushdown, vector size, ...) is switched in `src/go/settings/`, not on
the command line: flags only say what to run.

---

## Tests

```bash
cd src/go
go test ./...                       # unit and golden tests (no databases)
go test -race ./query/exec/         # every placement × every iteration model on an in-memory graph
# with the docker stack up:
go test -tags integration ./tests/db/ -run 'TestSmoke$' -v          # small queries × 6 placements vs Neo4j
go test -tags integration ./tests/db/ -run TestRoundtrip -v # copy migrations there and back
```

`tests/throwaway/` (`TestUUIDFormat`, `TestMove`: migrations that delete the source) needs a throwaway
environment; see the comments at the top of its files.

---

## References

[1] Carlos J. Fernández Candel, Diego Sevilla Ruiz, and Jesús J. García-Molina.
A unified metamodel for nosql and relational databases. *Information Systems*,
Vol. 104, p. 101898, 2022.
