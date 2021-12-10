package epochid

import (
	"math"

	"github.com/pkg/errors"
)

// BlockNumber extracts the activation block number from an epoch id.
func BlockNumber(epochID uint64) int64 {
	return int64(epochID >> 32) // take first 4 bytes
}

// SequenceNumber extracts the sequence number from an epoch id.
func SequenceNumber(epochID uint64) int64 {
	return int64(epochID & math.MaxUint32) // take last 4 bytes
}

func New(seq uint64, blk uint64) (uint64, error) {
	if seq>>32 != 0 {
		return 0, errors.Errorf("sequence number %d out of bounds", seq)
	}
	if blk>>32 != 0 {
		return 0, errors.Errorf("block number %d out of bounds", blk)
	}
	return blk<<32 | seq, nil
}
