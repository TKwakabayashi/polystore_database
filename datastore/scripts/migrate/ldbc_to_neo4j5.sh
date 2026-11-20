#!/usr/bin/env bash
# Moves ldbc_snb_sf1's (then ldbc_sf1) Neo4j 2025 (Enterprise, block format) store to Neo4j 5 (Community).
#
# A store cannot be downgraded from 2025 to 5, and loading the dump does not work either,
# so this makes a logical copy: export typed CSV from 2025 and bulk import it again with
# Neo4j 5's neo4j-admin database import full.
#
# Done on 2026-07-16. It need not run again and is kept as a record of the steps
# (running it needs the 2025-enterprise block-format store again).
#
#   scripts/migrate/ldbc_to_neo4j5.sh export   # export CSV from 2025
#   scripts/migrate/ldbc_to_neo4j5.sh import   # bulk import into a temporary Neo4j 5 volume
#   scripts/migrate/ldbc_to_neo4j5.sh schema   # recreate constraints and indexes
#   scripts/migrate/ldbc_to_neo4j5.sh dump     # write a Neo4j 5 dump
#
# The CSV goes to the import volume (polystore_ldbc_sf1_neo4j_import).
# The headers carry types (:datetime/:date/:long), so creationDate does not turn back into a string.
set -euo pipefail
cd "$(dirname "$0")/../.."   # run from datastore/

ENV_FILE=datasets/ldbc_snb_sf1/dataset.env
SRC_CONTAINER=polystore-ldbc_sf1-neo4j-1
IMPORT_VOL=polystore_ldbc_sf1_neo4j_import
V5_VOL=polystore_ldbc_sf1_neo4j_data_v5
V5_IMAGE=neo4j:5
PW=password123

cyp() { docker exec "$SRC_CONTAINER" cypher-shell -u neo4j -p "$PW" --format plain "$@"; }

# Writes one file with apoc.export.csv.query. $1=cypher, $2=output file name
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
  # types with uuid + creationDate
  for t in HAS_TAG HAS_MEMBER IS_LOCATED_IN HAS_CREATOR LIKES REPLY_OF CONTAINER_OF HAS_INTEREST KNOWS HAS_MODERATOR; do
    export_csv "MATCH (a)-[r:${t}]->(b) RETURN id(a) AS \`:START_ID\`, id(b) AS \`:END_ID\`, \"${t}\" AS \`:TYPE\`, r.uuid AS uuid, toString(r.creationDate) AS \`creationDate:datetime\`" "rels_$(echo "$t" | tr 'A-Z' 'a-z').csv"
  done
  # types with uuid only
  for t in HAS_TYPE IS_PART_OF IS_SUBCLASS_OF; do
    export_csv "MATCH (a)-[r:${t}]->(b) RETURN id(a) AS \`:START_ID\`, id(b) AS \`:END_ID\`, \"${t}\" AS \`:TYPE\`, r.uuid AS uuid" "rels_$(echo "$t" | tr 'A-Z' 'a-z').csv"
  done
  # types with more properties
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
  # multiline-fields is on so newlines in content do not break rows.
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
  local out=datasets/ldbc_snb_sf1/dumps
  mkdir -p "$out" "${out}/_archive"
  # The existing 2025 dump cannot be read by 5, so move it to _archive before overwriting.
  if [[ -f "${out}/neo4j.dump" && ! -f "${out}/_archive/neo4j2025.dump" ]]; then
    mv "${out}/neo4j.dump" "${out}/_archive/neo4j2025.dump"
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
