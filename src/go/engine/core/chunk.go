package core

import (
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// chunk.go は「1 クエリへ何件の識別子を載せるか」の決定を 1 か所に集約する。
//
// 定数の役割分担:
//   - store.LimitsOf(k)          : ストアの物理的制約（超えると壊れる）＋運用上の既定チャンク。
//   - settings.MaterializeChunkSize: 実測スイープ用の全体上書き（0 = ストア既定）。
//
// 実効値は必ず物理上限で clamp するため、上書きノブを大きくしても壊れない。

// ChunkSize は store k へ 1 クエリで載せる識別子数の実効上限を返す（0 = 分割しない）。
func ChunkSize(k store.Kind) int {
	lim := store.LimitsOf(k)

	want := lim.DefaultChunk
	if settings.MaterializeChunkSize > 0 {
		want = settings.MaterializeChunkSize
	}
	// 物理上限で必ず clamp（want==0＝無分割 も上限があれば上限まで下げる）。
	if lim.MaxInList > 0 && (want <= 0 || want > lim.MaxInList) {
		want = lim.MaxInList
	}
	if want < 0 {
		return 0
	}
	return want
}

// ForEachIDChunk は ids を store k の実効チャンクサイズで分割し、各塊に fn を適用する。
// チャンクサイズ 0（分割不要）のときは 1 回だけ全件を渡す。fn がエラーを返したら中断する。
func ForEachIDChunk(k store.Kind, ids []string, fn func(chunk []string) error) error {
	if len(ids) == 0 {
		return nil
	}
	size := ChunkSize(k)
	if size <= 0 || len(ids) <= size {
		return fn(ids)
	}
	for i := 0; i < len(ids); i += size {
		end := i + size
		if end > len(ids) {
			end = len(ids)
		}
		if err := fn(ids[i:end]); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIDChunkParams は ids を store k のチャンクへ分割し、params[key] に各チャンクを
// セットして fn を呼ぶ。Neo4j/Mongo のように「クエリ文は固定でパラメータだけ差し替える」
// 経路で、呼び出し側の変更を最小にするためのヘルパ。
func ForEachIDChunkParams(k store.Kind, ids []string, params map[string]interface{}, key string, fn func() error) error {
	return ForEachIDChunk(k, ids, func(chunk []string) error {
		params[key] = chunk
		return fn()
	})
}
