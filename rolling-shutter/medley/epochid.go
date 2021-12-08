package medley

import "github.com/pkg/errors"

// ActivationBlockNumberFromEpochID extracts the activation block number from an epoch id.
func ActivationBlockNumberFromEpochID(epochID uint64) int64 {
	return int64(epochID >> (8 * 4)) // take first 4 bytes
}

// SequenceNumberFromEpochID extracts the sequence number from an epoch id.
func SequenceNumberFromEpochID(epochID uint64) int64 {
	return int64(epochID << (8 * 4) >> (8 * 4)) // take last 4 bytes
}

func EncodeEpochID(seq uint64, blk uint64) (uint64, error) {
	maxUint32 := uint64(^uint32(0))
	if seq > maxUint32 || blk > maxUint32 {
		return 0, errors.Errorf("cannot fit block number %d and sequence number %d into an epochID", blk, seq)
	}
	return (blk << (8 * 4)) + seq, nil
}
