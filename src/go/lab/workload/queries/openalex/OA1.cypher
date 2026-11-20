MATCH (w:Work)-[:CITES]->(m:Work)-[:CITES]->(o:Work)
WHERE w.year = $year AND w.cited_by_count >= $minCited AND m.year >= $minYear
RETURN o.title, count(*) AS c
ORDER BY c DESC, o.title ASC
LIMIT 20
