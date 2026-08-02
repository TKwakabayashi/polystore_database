package core

import "polystore_database/src/go/id"

// Record は record パイプライン（EntityScan/Filter/Expand）を流れる中間結果の 1 行。
// Slots はスロット表インデックス順に束縛された uuid の並び。stream/bulk 両エンジンが
// この型を共有する（各エンジンの Record はこれへの型エイリアス）。表現は id.UUID に
// 閉じており、ドライバ境界（access 層）でのみ文字列/バイトへ変換する。
type Record struct {
	Slots []id.UUID
}
