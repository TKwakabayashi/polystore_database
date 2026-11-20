MATCH (cn:Nation)<-[:CUST_IN]-(c:Customer)-[:PLACED]->(o:Order)-[:CONTAINS]->(l:Lineitem)
      -[:SUPPLIED_BY]->(s:Supplier)-[:SUPP_IN]->(sn:Nation)
WHERE cn.name = $custNation AND sn.name = $suppNation
      AND l.shipdate >= $from AND l.shipdate <= $to
RETURN count(*) AS lines, sum(l.extendedprice) AS revenue
