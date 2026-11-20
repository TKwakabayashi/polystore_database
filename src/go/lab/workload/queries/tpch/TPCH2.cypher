MATCH (c:Customer)-[:PLACED]->(o:Order)-[:CONTAINS]->(l:Lineitem)-[:OF_PART]->(p:Part)
WHERE c.mktsegment = $segment AND o.orderdate >= $from AND o.orderdate < $to
      AND l.discount >= $minDiscount
RETURN p.brand, count(*) AS lines, sum(l.extendedprice) AS revenue
ORDER BY revenue DESC, p.brand ASC
LIMIT 10
