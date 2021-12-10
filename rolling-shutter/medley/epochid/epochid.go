package epochid

import (
	"math"
)

// BlockNumber extracts the activation block number from an epoch id.
func BlockNumber(epochID uint64) uint32 {
	return uint32(epochID >> 32) // take first 4 bytes
}

// SequenceNumber extracts the sequence number from an epoch id.
func SequenceNumber(epochID uint64) uint32 {
	return uint32(epochID & math.MaxUint32) // take last 4 bytes
}

func New(seq uint32, blk uint32) uint64 {
	return uint64(blk)<<32 | uint64(seq)
}
