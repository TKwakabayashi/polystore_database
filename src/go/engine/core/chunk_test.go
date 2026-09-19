package core

import (
	"testing"

	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// ストア既定チャンクが使われること（settings 上書き無しの既定状態）。
func TestChunkSizeUsesStoreDefault(t *testing.T) {
	prev := settings.MaterializeChunkSize
	settings.MaterializeChunkSize = 0
	defer func() { settings.MaterializeChunkSize = prev }()

	cases := map[store.Kind]int{
		store.Relational: 1000,
		store.Columnar:   500,
		store.Graph:      50000,
		store.Document:   50000,
		store.Kvs:        0, // IN の概念が無い
	}
	for k, want := range cases {
		if got := ChunkSize(k); got != want {
			t.Errorf("ChunkSize(%s) = %d, want %d", k, got, want)
		}
	}
}

// settings の上書きが効くこと。ただし物理上限を超える指定は clamp されること。
func TestChunkSizeOverrideAndClamp(t *testing.T) {
	prev := settings.MaterializeChunkSize
	defer func() { settings.MaterializeChunkSize = prev }()

	// 上限より小さい上書きはそのまま通る。
	settings.MaterializeChunkSize = 200
	if got := ChunkSize(store.Relational); got != 200 {
		t.Errorf("上書き: ChunkSize(relational) = %d, want 200", got)
	}
	if got := ChunkSize(store.Graph); got != 200 {
		t.Errorf("上書き: ChunkSize(graph) = %d, want 200", got)
	}

	// MySQL のプレースホルダ上限(65535)を超える指定は clamp される。
	settings.MaterializeChunkSize = 1000000
	if got := ChunkSize(store.Relational); got != 65535 {
		t.Errorf("clamp: ChunkSize(relational) = %d, want 65535", got)
	}
	// Mongo も BSON 由来の上限で clamp される。
	if got := ChunkSize(store.Document); got != 200000 {
		t.Errorf("clamp: ChunkSize(document) = %d, want 200000", got)
	}
	// 硬い上限が無いストアは上書き値のまま。
	if got := ChunkSize(store.Graph); got != 1000000 {
		t.Errorf("clamp なし: ChunkSize(graph) = %d, want 1000000", got)
	}
}

// ForEachIDChunk が漏れ・重複なく全件を分割して渡すこと。
func TestForEachIDChunkCoversAllIDs(t *testing.T) {
	prev := settings.MaterializeChunkSize
	settings.MaterializeChunkSize = 3
	defer func() { settings.MaterializeChunkSize = prev }()

	ids := []string{"a", "b", "c", "d", "e", "f", "g"}
	var seen []string
	chunks := 0
	err := ForEachIDChunk(store.Relational, ids, func(chunk []string) error {
		chunks++
		if len(chunk) > 3 {
			t.Errorf("チャンクサイズ超過: %d", len(chunk))
		}
		seen = append(seen, chunk...)
		return nil
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if chunks != 3 { // 3+3+1
		t.Errorf("chunks = %d, want 3", chunks)
	}
	if len(seen) != len(ids) {
		t.Fatalf("渡された件数 = %d, want %d", len(seen), len(ids))
	}
	for i := range ids {
		if seen[i] != ids[i] {
			t.Errorf("順序が保たれていない: seen[%d]=%s want %s", i, seen[i], ids[i])
		}
	}
}

// 分割不要（チャンクサイズ 0 = kvs）なら 1 回だけ全件渡す。
func TestForEachIDChunkNoSplit(t *testing.T) {
	prev := settings.MaterializeChunkSize
	settings.MaterializeChunkSize = 0
	defer func() { settings.MaterializeChunkSize = prev }()

	ids := []string{"a", "b", "c"}
	calls := 0
	_ = ForEachIDChunk(store.Kvs, ids, func(chunk []string) error {
		calls++
		if len(chunk) != 3 {
			t.Errorf("chunk = %v, want 全件", chunk)
		}
		return nil
	})
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}
