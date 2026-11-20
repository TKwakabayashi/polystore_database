MATCH (m:Message {id: $messageId })
RETURN
    m.creationDate as messageCreationDate,
    coalesce(m.content, m.imageFile) as messageContent
