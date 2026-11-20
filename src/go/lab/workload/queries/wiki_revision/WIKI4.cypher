MATCH (r:Revision)-[:PREV*1..3]->(o:Revision)<-[:EDITED]-(d:Contributor)
WHERE r.text_bytes >= $minBytes AND o.timestamp >= $since
RETURN d.name, count(*) AS edits
ORDER BY edits DESC, d.name ASC
LIMIT 20
