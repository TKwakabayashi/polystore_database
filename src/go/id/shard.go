package id

import (
	"math"
	"math/big"
)

// ShardBounds は shardIdx/shardTotal に対応する UUID の [lo, hi) 境界を返す。
// lo.Empty() は下限なし、hi.Empty() は上限なし。先頭バイト空間を等分するため、
// 全シャードで UUID 空間を過不足なく分割する（順序は Compare＝Bytes() の順序に依存）。
func ShardBounds(shardIdx, shardTotal int) (lo, hi UUID) {
	if shardTotal <= 1 {
		return "", ""
	}
	if shardIdx > 0 {
		lo = UUID([]byte{byte(shardIdx * 256 / shardTotal)})
	}
	if shardIdx < shardTotal-1 {
		hi = UUID([]byte{byte((shardIdx + 1) * 256 / shardTotal)})
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
