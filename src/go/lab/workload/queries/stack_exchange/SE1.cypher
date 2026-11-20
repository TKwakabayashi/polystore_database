MATCH (q:Post)-[:PARENT_OF]->(a:Post)<-[:AUTHORED]-(u:User)
WHERE q.view_count >= $minViews AND a.score >= $minScore AND u.reputation >= $minReputation
RETURN u.display_name, count(*) AS answers
ORDER BY answers DESC, u.display_name ASC
LIMIT 20
