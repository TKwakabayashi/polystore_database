#!/usr/bin/env bash
# ldbc_sf1 の Neo4j 2025(Enterprise/block形式) ストアを Neo4j 5(Community) へ移行する。
#
# ストアは 2025 -> 5 へダウングレードできず、dump のロードでも不可。
# そのため「2025 から型付き CSV を書き出し、5 の neo4j-admin database import full
# でバルクインポートし直す」という論理コピーを行う。
#
# 2026-07-16 に実行済み。再実行の必要はないが、手順の記録として残してある
# （実行するには 2025-enterprise 製の block 形式ストアが再度必要）。
#
#   scripts/migrate/ldbc_to_neo4j5.sh export   # 2025 から CSV 書き出し
#   scripts/migrate/ldbc_to_neo4j5.sh import   # 5 の一時ボリュームへバルクインポート
#   scripts/migrate/ldbc_to_neo4j5.sh schema   # 制約・インデックスを再作成
#   scripts/migrate/ldbc_to_neo4j5.sh dump     # 5 形式の dump を出力
#
# CSV は import ボリューム(polystore_ldbc_sf1_neo4j_import)に置く。
# 型ヘッダ(:datetime/:date/:long)を付けるので creationDate は文字列に戻らない。
set -euo pipefail
cd "$(dirname "$0")/../.."   # datastore/ を基準に動く

ENV_FILE=env/ldbc_sf1.env
SRC_CONTAINER=polystore-ldbc_sf1-neo4j-1
IMPORT_VOL=polystore_ldbc_sf1_neo4j_import
V5_VOL=polystore_ldbc_sf1_neo4j_data_v5
V5_IMAGE=neo4j:5
PW=password123

cyp() { docker exec "$SRC_CONTAINER" cypher-shell -u neo4j -p "$PW" --format plain "$@"; }

# apoc.export.csv.query で 1 ファイル書き出す。$1=cypher, $2=出力ファイル名
export_csv() {
  local q="$1" f="$2"
  local esc=${q//\'/\\\'}
  echo "  -> $f"
  cyp "CALL apoc.export.csv.query('${esc}', '${f}', {}) YIELD file RETURN file;" >/dev/null
}

do_export() {
  echo "== export nodes"
  export_csv 'MATCH (n:Comment) RETURN id(n) AS `:ID`, "Comment;Message;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, toString(n.creationDate) AS `creationDate:datetime`, n.browserUsed AS browserUsed, n.content AS content, n.imageFile AS imageFile, n.length AS `length:long`, n.locationIP AS locationIP' nodes_comment.csv
  export_csv 'MATCH (n:Post) RETURN id(n) AS `:ID`, "Post;Message;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, toString(n.creationDate) AS `creationDate:datetime`, n.browserUsed AS browserUsed, n.content AS content, n.imageFile AS imageFile, n.language AS language, n.length AS `length:long`, n.locationIP AS locationIP' nodes_post.csv
  export_csv 'MATCH (n:Forum) RETURN id(n) AS `:ID`, "Forum;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, toString(n.creationDate) AS `creationDate:datetime`, n.title AS title' nodes_forum.csv
  export_csv 'MATCH (n:Person) RETURN id(n) AS `:ID`, "Person;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, toString(n.creationDate) AS `creationDate:datetime`, toString(n.birthday) AS `birthday:date`, n.browserUsed AS browserUsed, n.email AS email, n.firstName AS firstName, n.gender AS gender, n.lastName AS lastName, n.locationIP AS locationIP, n.speaks AS speaks' nodes_person.csv
  export_csv 'MATCH (n:Tag) RETURN id(n) AS `:ID`, "Tag;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, n.name AS name, n.url AS url' nodes_tag.csv
  export_csv 'MATCH (n:TagClass) RETURN id(n) AS `:ID`, "TagClass;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, n.name AS name, n.url AS url' nodes_tagclass.csv
  export_csv 'MATCH (n:Organisation) RETURN id(n) AS `:ID`, "Organisation;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, n.name AS name, n.type AS type, n.url AS url' nodes_organisation.csv
  export_csv 'MATCH (n:Place) RETURN id(n) AS `:ID`, "Place;Entity" AS `:LABEL`, n.id AS `id:long`, n.uuid AS uuid, n.name AS name, n.type AS type, n.url AS url' nodes_place.csv

  echo "== export relationships"
  # uuid + creationDate を持つ型
  for t in HAS_TAG HAS_MEMBER IS_LOCATED_IN HAS_CREATOR LIKES REPLY_OF CONTAINER_OF HAS_INTEREST KNOWS HAS_MODERATOR; do
    export_csv "MATCH (a)-[r:${t}]->(b) RETURN id(a) AS \`:START_ID\`, id(b) AS \`:END_ID\`, \"${t}\" AS \`:TYPE\`, r.uuid AS uuid, toString(r.creationDate) AS \`creationDate:datetime\`" "rels_$(echo "$t" | tr 'A-Z' 'a-z').csv"
  done
  # uuid のみの型
  for t in HAS_TYPE IS_PART_OF IS_SUBCLASS_OF; do
    export_csv "MATCH (a)-[r:${t}]->(b) RETURN id(a) AS \`:START_ID\`, id(b) AS \`:END_ID\`, \"${t}\" AS \`:TYPE\`, r.uuid AS uuid" "rels_$(echo "$t" | tr 'A-Z' 'a-z').csv"
  done
  # 追加プロパティを持つ型
  export_csv 'MATCH (a)-[r:WORK_AT]->(b) RETURN id(a) AS `:START_ID`, id(b) AS `:END_ID`, "WORK_AT" AS `:TYPE`, r.uuid AS uuid, toString(r.creationDate) AS `creationDate:datetime`, r.workFrom AS `workFrom:long`' rels_work_at.csv
  export_csv 'MATCH (a)-[r:STUDY_AT]->(b) RETURN id(a) AS `:START_ID`, id(b) AS `:END_ID`, "STUDY_AT" AS `:TYPE`, r.uuid AS uuid, toString(r.creationDate) AS `creationDate:datetime`, r.classYear AS `classYear:long`' rels_study_at.csv
  echo "export done"
}

do_import() {
  local args=()
  for f in nodes_comment nodes_post nodes_forum nodes_person nodes_tag nodes_tagclass nodes_organisation nodes_place; do
    args+=(--nodes="/import/${f}.csv")
  done
  for f in rels_has_tag rels_has_member rels_is_located_in rels_has_creator rels_likes rels_reply_of \
           rels_container_of rels_has_interest rels_knows rels_has_moderator rels_has_type \
           rels_is_part_of rels_is_subclass_of rels_work_at rels_study_at; do
    args+=(--relationships="/import/${f}.csv")
  done
  docker volume create "$V5_VOL" >/dev/null
  # content に改行が含まれても壊れないよう multiline-fields を有効化する。
  docker run --rm \
    -v "${V5_VOL}":/data -v "${IMPORT_VOL}":/import:ro \
    "$V5_IMAGE" neo4j-admin database import full neo4j \
      "${args[@]}" \
      --multiline-fields=true \
      --overwrite-destination=true
}

do_schema() {
  local c=polystore-ldbc_sf1-neo4j5-tmp
  docker exec "$c" cypher-shell -u neo4j -p "$PW" "CREATE CONSTRAINT node_uuid_unique IF NOT EXISTS FOR (n:Entity) REQUIRE n.uuid IS UNIQUE;"
  for t in CONTAINER_OF HAS_CREATOR HAS_INTEREST HAS_MEMBER HAS_MODERATOR HAS_TAG HAS_TYPE \
           IS_LOCATED_IN IS_PART_OF IS_SUBCLASS_OF KNOWS LIKES REPLY_OF STUDY_AT WORK_AT; do
    local name="$(echo "$t" | tr 'A-Z' 'a-z')_uuid_unique"
    docker exec "$c" cypher-shell -u neo4j -p "$PW" \
      "CREATE CONSTRAINT ${name} IF NOT EXISTS FOR ()-[r:${t}]-() REQUIRE r.uuid IS UNIQUE;"
  done
}

do_dump() {
  local out=snapshots/ldbc_sf1/neo4j
  mkdir -p "$out" snapshots/_archive
  # 2025 製の既存 dump は 5 では読めないので _archive へ退避してから上書きする。
  if [[ -f "${out}/neo4j.dump" && ! -f snapshots/_archive/ldbc_sf1_neo4j2025.dump ]]; then
    mv "${out}/neo4j.dump" snapshots/_archive/ldbc_sf1_neo4j2025.dump
  fi
  docker run --rm -v "${V5_VOL}":/data -v "$(pwd)/${out}":/dumps \
    "$V5_IMAGE" neo4j-admin database dump neo4j --to-path=/dumps --overwrite-destination
}

case "${1:-}" in
  export) do_export ;;
  import) do_import ;;
  schema) do_schema ;;
  dump)   do_dump ;;
  *) echo "usage: $0 {export|import|schema|dump}" >&2; exit 1 ;;
esac
