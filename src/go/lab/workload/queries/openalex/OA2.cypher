MATCH c=(w:Work)-[:CITES*1..3]->(o:Work)
WHERE w.cited_by_count >= $minCited
      AND all(x IN nodes(c) WHERE x.year >= $minYear)
RETURN o.venue, count(*) AS paths
ORDER BY paths DESC, o.venue ASC
LIMIT 20
