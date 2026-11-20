MATCH (c:Customer)-[:PLACED]->(o:Order)-[:CONTAINS]->(l:Lineitem)
WHERE c.mktsegment = $segment AND o.orderdate < $date AND l.shipdate > $date
RETURN o.orderkey, o.orderdate, sum(l.extendedprice) AS revenue
ORDER BY revenue DESC, o.orderkey ASC
LIMIT 10
