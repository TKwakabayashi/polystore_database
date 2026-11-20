MATCH (q:Post)-[:LINKS_TO]->(m:Post)-[:LINKS_TO]->(r:Post)
WHERE q.score >= $minScore AND m.view_count >= $minViews
RETURN r.post_id, count(*) AS links
ORDER BY links DESC, r.post_id ASC
LIMIT 20
