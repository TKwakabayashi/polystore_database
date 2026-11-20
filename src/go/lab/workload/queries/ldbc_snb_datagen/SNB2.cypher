MATCH t=(c:Comment)-[:REPLY_TO_COMMENT*1..4]->(root:Comment)
WHERE c.creation_date >= $since
      AND all(x IN nodes(t) WHERE x.length >= $minLength)
RETURN root.comment_id, count(*) AS replies
ORDER BY replies DESC, root.comment_id ASC
LIMIT 20
