MATCH (u:User)-[:AUTHORED]->(p:Post)
WHERE u.reputation >= $minReputation
RETURN u.user_id, count(p) AS posts, avg(p.score) AS avgScore, max(p.view_count) AS maxViews
ORDER BY posts DESC, u.user_id ASC
LIMIT 20
