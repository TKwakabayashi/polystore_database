MATCH (fo:Forum)<-[:IN_FORUM]-(m:Post)<-[:REPLY_TO_POST]-(c:Comment)
WHERE fo.creation_date >= $since AND c.length >= $minLength
RETURN fo.title, count(*) AS comments, avg(m.length) AS avgPostLength
ORDER BY comments DESC, fo.title ASC
LIMIT 20
