package storage

import "fmt"

type StoreKind int

const (
	Graph StoreKind = iota
	Columnar
	Relational
	Document
	Kvs
)

var storeKindNames = map[StoreKind]string{
	Graph:      "graph",
	Document:   "document",
	Kvs:        "kvs",
	Relational: "relational",
	Columnar:   "columnar",
}

func (k StoreKind) String() string {
	if s, ok := storeKindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("StoreKind(%d)", int(k))
}

func ParseStoreKind(s string) (StoreKind, error) {
	for k, name := range storeKindNames {
		if name == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown store kind: %q", s)
}
