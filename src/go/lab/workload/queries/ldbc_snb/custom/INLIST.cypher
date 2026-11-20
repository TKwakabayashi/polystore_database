MATCH (o:Organisation)
WHERE o.type IN ["Company", "University"]
RETURN o.id, o.type
ORDER BY o.id ASC
LIMIT 20
