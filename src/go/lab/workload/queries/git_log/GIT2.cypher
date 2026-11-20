MATCH p=(c:Commit)-[:PARENT*1..5]->(a:Commit)
WHERE c.files_changed >= $minFiles
      AND all(x IN nodes(p) WHERE x.deletions < $maxDeletions)
RETURN a.sha, count(*) AS paths
ORDER BY paths DESC, a.sha ASC
LIMIT 20
