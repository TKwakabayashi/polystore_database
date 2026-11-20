# ldbc_snb_sf1

LDBC Social Network Benchmark, scale factor 1, from the official Neo4j dump (Person email and speaks are
lists). 2,997,352 nodes / 17,196,776 relationships. Dates are DATE, datetimes ZONED DATETIME, ids INTEGER.

The workloads are in `src/go/lab/workload/ldbc_snb.go` (queries in `queries/ldbc_snb/`). `stats.json` is
the fixed statistics the goldens use; `admin stats` writes the live ones to `datastore/runtime/`.
