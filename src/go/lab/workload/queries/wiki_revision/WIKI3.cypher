MATCH (c:Contributor)-[:EDITED]->(r:Revision)<-[:HAS_REVISION]-(p:Page)
WHERE r.timestamp >= $since AND r.text_bytes >= $minBytes
RETURN c.name, count(DISTINCT p.page_id) AS pages
ORDER BY pages DESC, c.name ASC
LIMIT 20
