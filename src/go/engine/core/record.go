package core

import "polystore_database/src/go/id"

// Record は record パイプライン（EntityScan/Filter/Expand）を流れる中間結果の 1 行。
// Slots はスロット表インデックス順に束縛された uuid の並び。stream/bulk 両エンジンが
// この型を共有する（各エンジンの Record はこれへの型エイリアス）。表現は id.UUID に
// 閉じており、ドライバ境界（access 層）でのみ文字列/バイトへ変換する。
//
// Vals は WITH 多段（Phase B2）で前段が持ち越すスカラ/計算値（列名→値）。単段クエリでは nil
// （ゼロオーバーヘッド・既存挙動不変）。carried 値は行の系譜上で不変なので、Expand/Filter 等の
// record コピー時に read-only で共有参照を通す（RowSource が設定・Projection が Row へマージ）。
type Record struct {
	Slots []id.UUID
	Vals  map[string]any
}
