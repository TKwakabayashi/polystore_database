package id

import (
	"math"
	"math/big"
)

// ShardBounds は shardIdx/shardTotal に対応する、保存値（文字列表現・バイナリどちらでも）の先頭 1 バイトでの
// [lo, hi) 境界を返す。nil は下限なし / 上限なし。先頭バイト空間を等分するので、全シャードでキー空間を
// 過不足なく分割する（保存値のバイト順に依存。文字列表現では先頭が種別文字なので偏る）。
func ShardBounds(shardIdx, shardTotal int) (lo, hi []byte) {
	if shardTotal <= 1 {
		return nil, nil
	}
	if shardIdx > 0 {
		lo = []byte{byte(shardIdx * 256 / shardTotal)}
	}
	if shardIdx < shardTotal-1 {
		hi = []byte{byte((shardIdx + 1) * 256 / shardTotal)}
	}
	return lo, hi
}

// ShardTokenBounds は Cassandra の token(uuid) 空間 [MinInt64, MaxInt64] を等分し、
// shardIdx の [lo, hi) を返す。hasHi=false は最終シャード（上限なし）を表す。
// token 空間は表現非依存（一様分布ハッシュ）なので UUID 内部表現の変更に影響されない。
func ShardTokenBounds(shardIdx, shardTotal int) (lo, hi int64, hasHi bool) {
	if shardTotal <= 1 {
		return math.MinInt64, 0, false
	}
	min := big.NewInt(math.MinInt64)
	span := new(big.Int).Sub(big.NewInt(math.MaxInt64), min) // 2^64 - 1
	total := big.NewInt(int64(shardTotal))
	at := func(k int) int64 {
		v := new(big.Int).Mul(span, big.NewInt(int64(k)))
		v.Div(v, total)
		v.Add(v, min)
		return v.Int64()
	}
	lo = at(shardIdx)
	if shardIdx == shardTotal-1 {
		return lo, 0, false
	}
	return lo, at(shardIdx + 1), true
}
