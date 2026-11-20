MATCH e=(r:Rev)-[:PREV_EDIT*1..5]->(o:Rev)
WHERE r.creation_date >= $since
      AND none(x IN nodes(e) WHERE x.history_type = 8)
RETURN o.post_id, count(*) AS chains
ORDER BY chains DESC, o.post_id ASC
LIMIT 20
