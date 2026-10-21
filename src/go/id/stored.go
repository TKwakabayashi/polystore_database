package id

import "polystore_database/src/go/settings"

// stored.go は保存形式（settings.UUIDStorage）に従う境界変換。ストアへ書く値・照合に渡す値・KVS のキー片は
// 必ずここを通す（表示やログは String を使う）。読み取りは FromAny / FromBytes がどちらの形式も受け付ける。

// Binary は保存形式がバイナリか。
func Binary() bool { return settings.UUIDStorage == settings.UUIDBinary }

// Stored は u の保存値（文字列形式なら string、バイナリなら 17 バイトの []byte）。空なら nil。
func Stored(u UUID) any {
	if u.Empty() {
		return nil
	}
	if Binary() {
		return u.Bytes()
	}
	return u.String()
}

// StoredSlice は us の保存値のスライス（[]string か [][]byte。ドライバへ IN のリストとして渡す型付きの値）。
func StoredSlice(us []UUID) any {
	if Binary() {
		out := make([][]byte, len(us))
		for i, u := range us {
			out[i] = u.Bytes()
		}
		return out
	}
	out := make([]string, len(us))
	for i, u := range us {
		out[i] = u.String()
	}
	return out
}

// StoredKey は KVS のキーに埋め込むバイト列（長さは常に StoredLen。区切り文字で分けずに位置で切り出す）。
func StoredKey(u UUID) []byte {
	if Binary() {
		return u.Bytes()
	}
	return []byte(u.String())
}

// StoredLen は保存値のバイト数（キーから位置で切り出すときに使う）。
func StoredLen() int {
	if Binary() {
		return Size
	}
	return TextSize
}

// StoredBound は ShardBounds の先頭バイトの境界を、保存値どうしで比べられる値にする（Mongo の範囲条件用）。
// バイナリは長さも比べる型（BinData）なので 17 バイトへ 0 で埋める。nil は nil のまま。
func StoredBound(prefix []byte) any {
	if prefix == nil {
		return nil
	}
	if Binary() {
		b := make([]byte, Size)
		copy(b, prefix)
		return b
	}
	return string(prefix)
}
