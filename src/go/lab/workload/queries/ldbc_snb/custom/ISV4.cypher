MATCH (m:Message)
WHERE m.creationDate >= $start AND m.creationDate < $end
MATCH (m)-[:REPLY_OF*0..]->(post:Post)
WHERE post.language = $language
RETURN m.id, m.length, post.id, post.creationDate
