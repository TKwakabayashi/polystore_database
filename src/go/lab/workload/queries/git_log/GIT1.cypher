MATCH (c:Commit)-[:PARENT*1..3]->(a:Commit)
WHERE c.files_changed >= $minFiles
RETURN c.sha, a.sha, a.subject
ORDER BY a.committed_date DESC, a.sha ASC
LIMIT 20
