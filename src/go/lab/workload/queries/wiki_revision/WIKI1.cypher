MATCH h=(r:Revision)-[:PREV*1..10]->(o:Revision)
WHERE r.text_bytes >= $minBytes
      AND all(x IN nodes(h) WHERE x.text_bytes >= $minChainBytes)
RETURN o.page_id, count(*) AS chains
ORDER BY chains DESC, o.page_id ASC
LIMIT 20
