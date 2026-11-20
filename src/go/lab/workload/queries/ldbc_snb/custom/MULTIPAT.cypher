MATCH (p:Person {id: $personId})-[:KNOWS]-(friend:Person),
      (friend)-[:IS_LOCATED_IN]->(city:Place)
RETURN friend.id, friend.firstName, city.id
ORDER BY friend.id ASC
LIMIT 20
