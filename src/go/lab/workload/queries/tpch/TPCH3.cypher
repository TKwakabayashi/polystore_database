MATCH (c:Customer {custkey: '1'})-[:PLACED]->(o:Order)-[:CONTAINS]->(l:Lineitem)
      -[:SUPPLIED_BY]->(s:Supplier)-[:SUPP_IN]->(n:Nation)
WHERE l.returnflag = $flag
RETURN n.name, count(DISTINCT s.suppkey) AS suppliers, sum(l.quantity) AS qty
ORDER BY qty DESC, n.name ASC
