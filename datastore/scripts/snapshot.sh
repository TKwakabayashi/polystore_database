#!/usr/bin/env bash
# Dumps the Neo4j of a dataset.
#   scripts/snapshot.sh ldbc_snb_sf1
#   scripts/snapshot.sh openalex
#
# Output: datasets/<dataset>/dumps/<database>.dump (where compose's restore reads).
# An existing dump is overwritten (--overwrite-destination).
# Move generations worth keeping to datasets/<dataset>/dumps/_archive/ first.
set -euo pipefail
cd "$(dirname "$0")/.."   # run from datastore/

ID="${1:-ldbc_snb_sf1}"
ENV_FILE="datasets/${ID}/dataset.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "unknown dataset: ${ID} (no ${ENV_FILE})" >&2
  echo "available: $(ls -d datasets/*/ | xargs -n1 basename | tr '\n' ' ')" >&2
  exit 1
fi

# The database to dump follows the env (the dump file is named after the database).
DB="$(sed -n 's/^NEO4J_DATABASE=//p' "$ENV_FILE" | tail -1)"
DB="${DB:-neo4j}"

OUT="datasets/${ID}/dumps"
mkdir -p "$OUT"

dc() { docker compose -f docker/compose.yaml --env-file "$ENV_FILE" "$@"; }

dc stop neo4j
dc run --rm \
  -v "$(pwd)/${OUT}:/dumps" \
  neo4j neo4j-admin database dump "$DB" --to-path=/dumps --overwrite-destination
dc start neo4j
echo "snapshot -> ${OUT}/${DB}.dump"
