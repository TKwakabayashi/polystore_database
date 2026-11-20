MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)
      <-[:REPLY_OF]-(comment:Comment)-[:HAS_CREATOR]->(author:Person)
RETURN author.firstName, author.lastName, count(comment) AS replyCount
ORDER BY replyCount DESC, author.lastName ASC
LIMIT 20
