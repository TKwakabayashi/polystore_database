package id

import (
	"math"
	"math/big"
)

// ShardBounds returns the [lo, hi) bounds of shard shardIdx/shardTotal on the first byte of the stored value
// (nil: unbounded). The value space is split evenly, so all shards cover the key space exactly (256 leading bytes
// in binary, 16 leading hex digits as text; more than 16 shards leave empty shards with text storage).
func ShardBounds(shardIdx, shardTotal int) (lo, hi []byte) {
	if shardTotal <= 1 {
		return nil, nil
	}
	at := func(k int) []byte { return []byte{byte(k * 256 / shardTotal)} }
	if !Binary() {
		at = func(k int) []byte { return []byte{hexDigits[k*16/shardTotal]} }
	}
	if shardIdx > 0 {
		lo = at(shardIdx)
	}
	if shardIdx < shardTotal-1 {
		hi = at(shardIdx + 1)
	}
	return lo, hi
}

const hexDigits = "0123456789abcdef"

// ShardTokenBounds splits Cassandra's token(uuid) space [MinInt64, MaxInt64] evenly and returns shard shardIdx's
// [lo, hi); hasHi=false is the last shard (unbounded). Tokens are a uniform hash, independent of the uuid form.
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
