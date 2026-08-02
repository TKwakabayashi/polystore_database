package storage

import "polystore_database/src/go/store"

// StoreKind はデータストア種別。語彙は store.Kind に一本化したので、ここは互換のための
// 型エイリアス＋定数の再エクスポート（migrator / bench から storage.StoreKind / storage.Graph
// 等で参照され続けるため残す。将来 store.Kind へ直接移行して削除可能）。
type StoreKind = store.Kind

const (
	Graph      = store.Graph
	Columnar   = store.Columnar
	Relational = store.Relational
	Document   = store.Document
	Kvs        = store.Kvs
)

// ParseStoreKind は store.ParseKind への委譲。
func ParseStoreKind(s string) (StoreKind, error) { return store.ParseKind(s) }
