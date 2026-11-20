# opbench — experiments per logical operator

Measures one logical operator at a time under varied conditions. An experiment is a JSON spec with
`"mode": "opbench"`, run by the experiment runner (cwd = `src/go`, the docker stack up):

```bash
go run ./cmd/lab run -spec lab/opbench/specs/quick.json
```

Results go to `results/<date>_<name>/` (`-out` changes the parent directory): `opbench.csv` (one row per
run) and `opbench.plans.txt` (the physical plan and native queries of each condition). `results/` is
local only (never add it to git). Standard output shows the median of each condition.

## How it works

For each condition it builds a logical plan with the target operator and only the smallest prefix that
makes its input:

```
Argument(a) ← N uuids sampled from ldbc_sf1 in uuid order (input=ids)   or EntityScan(a) (input=scan)
  → prefix (GetProperty of the values the target uses) → [Unwind([1]) when at=coordinator] → target
```

- The **root physical step** of the plan is the target (`target_ms` / `target_step`); `exec_ms` is the
  whole plan with its prefix. A step's time is **the sum over all its batches**, so in stream / vecstream,
  which run in parallel, it can exceed the wall-clock `exec_ms` (worker times add up). Compare models by
  `exec_ms`.
- `at=coordinator` puts `Unwind([1])` (always on the coordinator, rows unchanged) right before the target,
  so the target is not merged into or absorbed by a store query. `at=store` follows the physical planner
  (in the store when it can).
- `input=scan` starts from a scan of every vertex of the label. That is a leaf query without input, so a
  store absorbs the aggregations, sorts and distincts it can run (`ids` queries take the boundary's IN, so
  they absorb nothing).
- Conditions with different row counts use prefixes of the same sample (nested inputs compare well).
- `match_graph` says whether the result multiset equals the graph placement's (bulk, default settings).
  `false` means a column was not migrated (the non-graph store has no values) or a bug. A Limit without a
  sort gives `n/a`.
- Data is only read. A non-graph placement places just the target property on that store in the runtime
  mapping and uses values already migrated there.

## Spec file

(`//` comments are explanations here; JSON files cannot have them. Unknown fields are errors.)

```json
{
  "mode": "opbench",
  "name": "…",
  "warmup": 1,             // unmeasured runs per condition
  "repeat": 3,             // measured runs (default 3)
  "timeout": "60s",        // limit of one run (default 60s; past it the condition stops and the next starts)
  "experiments": [ { "op": "…", …, "sweep": { … } } ]
}
```

| op | required | optional | plan (a is the input vertex) |
|---|---|---|---|
| `EntityScan` | label | prop + cmp + value (scan condition) | `EntityScan(a)` [→ `GetProperty` → `Filter`] |
| `GetProperty` | label, prop | | `Argument(a)` → `GetProperty(a.prop)` |
| `Expand` | label, rel | dir, to | `Expand((a)-[r:rel]-(b:to))` |
| `VarLengthExpand` | label, rel | dir, to, min, max (max < 0: unbounded) | `VarLengthExpand((a)-[s:rel*min..max]-(b:to))` |
| `Filter` | label, prop, cmp, value | | `GetProperty` → `Filter(a.prop cmp value)` (`IN` takes an array) |
| `Aggregation` | label, agg | prop (group key), aggProp (aggregated property; empty: count(*)) | `GetProperty`… → `Aggregation` |
| `Sort` | label, prop | desc | `GetProperty` → `Sort` |
| `Limit` | label, count | prop | [`GetProperty`] → `Limit` |
| `Distinct` | label, prop | | `GetProperty` → `Projection(a.prop)` → `Distinct` |

Common optional fields: `name`, `sample` (a Cypher WHERE selecting the input on variable `n`, e.g.
`"n.id < 10000"`).

`sweep` axes (every combination runs; an empty axis is the one value in parentheses):

| axis | values | default |
|---|---|---|
| `placement` | graph / relational / columnar (ClickHouse) / widecolumn (Cassandra) / document / kvs (LevelDB): the store holding the target property | graph |
| `rows` | input rows (ignored with input=scan) | 1000 |
| `input` | ids / scan | ids |
| `at` | store / coordinator (Filter, Aggregation, Sort, Limit, Distinct only) | store |
| `model` | bulk / volcano / vectorized / stream / vecstream | bulk |
| `chunk` | ids per store query (0: the store's default) | settings |
| `parallelism` | concurrent chunks per store query | settings |
| `vector` | batch width | settings |
| `list` | name of the uuid list used for input=ids (below) | sampled |
| `order` | raw (list order) / sorted (by id, duplicates kept) | raw |

`Projection` (requires props; filters first when prop + cmp + value are set) takes several columns.
`role` (start / mid) is only copied to the CSV.

## Input uuid lists

`"listsFile": "lists/ldbc.json"` defines lists made by running the prefix of real queries (see `ListSpec`
in `list.go`). A list is saved to `<output>/inputs/<name>.csv` with its order and duplicates and only read
afterwards (delete the file to remake it). An input of N rows is the list's first N; a definition returning
column k uses the N smallest k in list order. Conditions whose list is shorter than N are skipped.

## CSV columns

`spec, experiment, role, op, prop, predicate, input, list, order, rows, in_distinct, placement, at, model, chunk, parallelism,
vector, rep, exec_ms, target_ms, target_step, round_trips, out_rows, out_rows_expanded, match_graph, steps, error`

- `out_rows` is the executor's row count; `out_rows_expanded` expands weighted rows (variable-length duplicates).
- `steps` lists every physical step (`operator:time ms:output rows`, ` | ` separated, leaf to root).
- A timeout or failure writes its reason to `error` and skips that condition's remaining repetitions.

## Adding an operator

Add a case to `validate` (required fields) and `barrierable` (whether at=coordinator is allowed) in
`spec.go` and to `BuildLogical` (the plan's shape) in `plan.go`, and one shape check to `TestPlans` in
`opbench_test.go`.
