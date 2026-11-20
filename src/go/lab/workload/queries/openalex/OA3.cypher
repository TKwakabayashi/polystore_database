MATCH (a:Author)-[:AUTHORED]->(w:Work)<-[:AUTHORED]-(b:Author)
WHERE w.year = $year AND w.cited_by_count >= $minCited
RETURN b.author_id, count(*) AS c
ORDER BY c DESC, b.author_id ASC
LIMIT 20
