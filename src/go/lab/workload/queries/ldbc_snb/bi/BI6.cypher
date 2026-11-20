MATCH (tag:Tag {name: $tag})<-[:HAS_TAG]-(message1:Message)-[:HAS_CREATOR]->(person1:Person)
OPTIONAL MATCH (message1)<-[:LIKES]-(person2:Person)
OPTIONAL MATCH (person2)<-[:HAS_CREATOR]-(message2:Message)<-[like:LIKES]-(person3:Person)
RETURN
  person1.id,
  count(DISTINCT like) AS authorityScore
ORDER BY
  authorityScore DESC,
  person1.id ASC
LIMIT 100
