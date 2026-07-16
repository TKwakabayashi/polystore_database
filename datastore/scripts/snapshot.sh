#!/usr/bin/env bash
# 選択したデータセットの Neo4j を dump する。
#   scripts/snapshot.sh ldbc_sf1
#   scripts/snapshot.sh openalex
#
# 出力先は snapshots/<dataset>/neo4j/<db名>.dump（compose の復元側と同じ規約）。
# 既存の dump は --overwrite-destination で上書きされる点に注意。
# 取っておきたい世代は snapshots/_archive/<dataset>_<説明>.dump へ退避すること。
set -euo pipefail
cd "$(dirname "$0")/.."   # datastore/ を基準に動く

ID="${1:-ldbc_sf1}"
ENV_FILE="env/${ID}.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "unknown dataset: ${ID} (no ${ENV_FILE})" >&2
  echo "available: $(ls env/*.env | xargs -n1 basename | sed 's/\.env$//' | tr '\n' ' ')" >&2
  exit 1
fi

# dump するデータベース名は env に合わせる（dump ファイル名＝DB 名になる）。
DB="$(sed -n 's/^NEO4J_DATABASE=//p' "$ENV_FILE" | tail -1)"
DB="${DB:-neo4j}"

OUT="snapshots/${ID}/neo4j"
mkdir -p "$OUT"

dc() { docker compose --env-file "$ENV_FILE" "$@"; }

dc stop neo4j
dc run --rm \
  -v "$(pwd)/${OUT}:/dumps" \
  neo4j neo4j-admin database dump "$DB" --to-path=/dumps --overwrite-destination
dc start neo4j
echo "snapshot -> ${OUT}/${DB}.dump"
