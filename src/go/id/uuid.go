// Package id は UUID の唯一の権威。生成・境界変換・順序・プロパティ名を集約する。
//
// 内部表現は 17 バイト（種別 1 バイト N/E ＋ UUID 16 バイト）の固定長配列で、比較可能（map のキーにできる）。
// 文字列境界の表現は "N-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"（種別・ハイフン・正規形の UUID）。
// 外周コードは表現に依存しない（文字列化は String、バイトは Bytes、順序は Compare だけを使う）。
package id

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Kind は識別子の種別（ノードか関係か）。文字列表現の先頭文字と内部表現の先頭バイトに同じ値を使う。
type Kind byte

const (
	KindNode Kind = 'N'
	KindEdge Kind = 'E'
)

// Size は内部表現のバイト数。
const Size = 17

// TextSize は文字列表現の長さ（"N-" + 36 文字）。
const TextSize = 2 + 36

// UUID は内部表現。ゼロ値は空（識別子なし）。
type UUID [Size]byte

// PropName は uuid を表すプロパティ/カラム名の唯一の定義（string だが表現ではなくスキーマ識別子）。
const PropName = "uuid"

// New は種別 k の新しい UUID を採番する（唯一の生成口。v4 相当）。
func New(k Kind) UUID {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	raw[6] = (raw[6] & 0x0f) | 0x40 // version 4
	raw[8] = (raw[8] & 0x3f) | 0x80 // variant 10
	return Make(k, raw)
}

// Make は種別と 16 バイトから UUID を作る（決定的な識別子が要るテストや変換用）。
func Make(k Kind, raw [16]byte) UUID {
	var u UUID
	u[0] = byte(k)
	copy(u[1:], raw[:])
	return u
}

// Kind は種別（空なら 0）。
func (u UUID) Kind() Kind { return Kind(u[0]) }

func validKind(b byte) bool { return b == byte(KindNode) || b == byte(KindEdge) }

// Parse は文字列境界（Cypher 値・DB の文字列保存値・表示）→ UUID。空文字列は空 UUID。
func Parse(s string) (UUID, error) {
	if s == "" {
		return UUID{}, nil
	}
	if len(s) != TextSize || !validKind(s[0]) || s[1] != '-' {
		return UUID{}, fmt.Errorf("id: malformed uuid %q", s)
	}
	body := s[2:]
	if body[8] != '-' || body[13] != '-' || body[18] != '-' || body[23] != '-' {
		return UUID{}, fmt.Errorf("id: malformed uuid %q", s)
	}
	var hexs [32]byte
	n := 0
	for i := 0; i < len(body); i++ {
		if body[i] == '-' {
			continue
		}
		hexs[n] = body[i]
		n++
	}
	var u UUID
	u[0] = s[0]
	if _, err := hex.Decode(u[1:], hexs[:]); err != nil {
		return UUID{}, fmt.Errorf("id: malformed uuid %q: %w", s, err)
	}
	return u, nil
}

// FromBytes はバイト境界（バイナリ保存値・KVS キー）→ UUID。17 バイトの内部表現か、文字列表現のバイト列を受け付ける。
func FromBytes(b []byte) (UUID, error) {
	switch len(b) {
	case 0:
		return UUID{}, nil
	case Size:
		if !validKind(b[0]) {
			return UUID{}, fmt.Errorf("id: malformed binary uuid %x", b)
		}
		var u UUID
		copy(u[:], b)
		return u, nil
	}
	return Parse(string(b))
}

// FromAny は driver の native 値（SQL / CQL / Neo4j の interface{}）→ UUID。
// 文字列・バイト列（どちらの保存形式でも）・UUID を受け付け、解釈できない値は空 UUID を返す。
// ドライバ固有の型（bson の Binary など）は各ドライバ境界で []byte にしてから渡す。
func FromAny(v interface{}) UUID {
	var (
		u   UUID
		err error
	)
	switch x := v.(type) {
	case UUID:
		return x
	case string:
		if len(x) == Size { // バイナリの保存値を文字列として受け取った（SQL ドライバの []byte → string など）
			u, err = FromBytes([]byte(x))
		} else {
			u, err = Parse(x)
		}
	case []byte:
		u, err = FromBytes(x)
	default:
		return UUID{}
	}
	if err != nil {
		return UUID{}
	}
	return u
}

// String は UUID → 文字列表現（表示・文字列での保存値）。空なら "".
func (u UUID) String() string {
	if u.Empty() {
		return ""
	}
	var buf [TextSize]byte
	buf[0], buf[1] = u[0], '-'
	hex.Encode(buf[2:10], u[1:5])
	buf[10] = '-'
	hex.Encode(buf[11:15], u[5:7])
	buf[15] = '-'
	hex.Encode(buf[16:20], u[7:9])
	buf[20] = '-'
	hex.Encode(buf[21:25], u[9:11])
	buf[25] = '-'
	hex.Encode(buf[26:38], u[11:17])
	return string(buf[:])
}

// Bytes は UUID → 17 バイトの内部表現（バイナリでの保存値）。空なら nil。
func (u UUID) Bytes() []byte {
	if u.Empty() {
		return nil
	}
	b := make([]byte, Size)
	copy(b, u[:])
	return b
}

// Empty は空 UUID か。
func (u UUID) Empty() bool { return u == UUID{} }

// MarshalText / UnmarshalText は JSON などのテキスト形式で文字列表現を使う。
func (u UUID) MarshalText() ([]byte, error) { return []byte(u.String()), nil }

func (u *UUID) UnmarshalText(b []byte) error {
	v, err := Parse(string(b))
	if err != nil {
		return err
	}
	*u = v
	return nil
}

// Compare は順序の唯一の規約（内部表現のバイト順。文字列表現の辞書順とも一致する: 種別 E < N、
// 16 進の小文字は数値順に並ぶ）。
func Compare(a, b UUID) int { return bytes.Compare(a[:], b[:]) }

// Less は Compare(a, b) < 0。
func Less(a, b UUID) bool { return Compare(a, b) < 0 }
