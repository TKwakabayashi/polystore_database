package id

import (
	"fmt"
	"sync/atomic"
)

// alloc.go allocates uuids that are unique by construction (random v4 is unique only probabilistically).
//
// An Allocator receives a lease number once at startup from a shared counter (a local file per dataset,
// conn.LeaseUUIDAllocator) and then only counts a sequence number in the process. (lease, seq) are laid out
// in 64 bits, mixed by a bijection on uint64 and packed into a 16-byte UUIDv8, so different pairs always give
// different values (nodes and relationships share the sequence, so the value needs no kind). Version 8 never
// equals a random v4 either. Mixing spreads the leading bytes, which sequential numbers would bunch (ShardBounds
// would split unevenly). Running out of lease or seq is an error for now.

const (
	// LeaseBits is the width of a lease number (about 16 million startups).
	LeaseBits = 24
	// SeqBits is the width of a sequence number per lease (about 1.1 trillion ids).
	SeqBits = 64 - LeaseBits
)

// Allocator allocates uuids under one lease number. Safe for concurrent use.
type Allocator struct {
	lease uint64
	seq   atomic.Uint64 // the next sequence number
}

// NewAllocator makes an allocator for lease, which must be unique per allocator (taken from the shared counter).
func NewAllocator(lease uint64) (*Allocator, error) {
	if lease >= 1<<LeaseBits {
		return nil, fmt.Errorf("id: uuid lease %d exceeds %d bits", lease, LeaseBits)
	}
	return &Allocator{lease: lease}, nil
}

// Lease returns the lease number.
func (a *Allocator) Lease() uint64 { return a.lease }

// Next allocates the next UUID (v8), for nodes and relationships alike.
func (a *Allocator) Next() (UUID, error) {
	s := a.seq.Add(1) - 1
	if s >= 1<<SeqBits {
		return UUID{}, fmt.Errorf("id: uuid lease %d exhausted (%d bits)", a.lease, SeqBits)
	}
	return Make(pack(mix(a.lease<<SeqBits | s))), nil
}

// mix is a bijection on uint64 (splitmix64's finalizer: right xorshifts and odd multiplications are both invertible).
func mix(z uint64) uint64 {
	z ^= z >> 30
	z *= 0xbf58476d1ce4e5b9
	z ^= z >> 27
	z *= 0x94d049bb133111eb
	z ^= z >> 31
	return z
}

// pack puts 64 bits into the free bits of a UUIDv8, high bits first (injective): the top 48 bits in bytes 0-5,
// the low 4 bits of byte 6 (version 8 above), byte 7, and the low 4 bits of byte 8 (variant 10 above, then two
// zero bits). Bytes 9-15 are 0 (reserved).
func pack(v uint64) [16]byte {
	var b [16]byte
	for i := 0; i < 6; i++ {
		b[i] = byte(v >> (56 - 8*i))
	}
	b[6] = 0x80 | byte(v>>12)&0x0f
	b[7] = byte(v >> 4)
	b[8] = 0x80 | byte(v)&0x0f
	return b
}
