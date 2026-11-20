MATCH (p:Person {firstName: $name})-[:KNOWS*1..2]-(f:Person)
WITH DISTINCT f
MATCH (f)-[:IS_LOCATED_IN]->(c:City)
RETURN f.id, f.lastName, c.name
