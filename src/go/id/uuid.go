// Package id は UUID の唯一の権威。生成・境界変換・順序・プロパティ名を集約する。
//
// 内部表現（現状 string）を将来 []byte / int64 へ変える際は、このパッケージ内の
// 境界変換群（Parse / FromBytes / FromAny / String / Bytes / Compare）だけを差し替えれば
// 済むように、外周コードは string 依存を持たない（順序は Compare を唯一の規約とする）。
package id

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// UUID は内部表現。将来 []byte / int64 へ変える予定。変更時はこのパッケージのみ触る。
type UUID string

// PropName は uuid を表すプロパティ/カラム名の唯一の定義（string だが表現ではなくスキーマ識別子）。
const PropName = "uuid"

// New は新しい UUID を採番する（表現非依存の唯一の生成口。v4 相当）。
func New() UUID {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return UUID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}

// Parse は文字列境界（Cypher 値・表示など）→ UUID。
func Parse(s string) (UUID, error) { return UUID(s), nil }

// FromBytes はバイト境界（LevelDB 複合キー等）→ UUID。
func FromBytes(b []byte) (UUID, error) { return UUID(b), nil }

// FromAny は driver の native 値（bson / SQL / CQL の interface{}）→ UUID。
// 未知の型は空 UUID を返す（旧 migrator.asUUID と同一の緩い契約）。
func FromAny(v interface{}) UUID {
	switch s := v.(type) {
	case string:
		return UUID(s)
	case []byte:
		return UUID(s)
	case UUID:
		return s
	default:
		return ""
	}
}

// String は UUID → 文字列境界（DB 保存値・表示）。表現の一部ではなく境界エンコード。
func (u UUID) String() string { return string(u) }

// Bytes は UUID → バイト（LevelDB 複合キー / シャード境界の原始表現）。
func (u UUID) Bytes() []byte { return []byte(u) }

func (u UUID) Empty() bool { return u == "" }

// Compare は順序の唯一の規約（永続化表現＝ Bytes() の順序）。
func Compare(a, b UUID) int { return strings.Compare(string(a), string(b)) }

// Less は Compare(a, b) < 0。
func Less(a, b UUID) bool { return Compare(a, b) < 0 }
