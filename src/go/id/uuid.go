package id

type UUID string

// ParseUUID は境界（DB保存値・Cypher値などの外部表現）→ UUID。
func ParseUUID(s string) (UUID, error) { return UUID(s), nil }

// String は UUID → 外部表現（DB保存値・表示）。
func (u UUID) String() string { return string(u) }

func (u UUID) Empty() bool { return u == "" }
