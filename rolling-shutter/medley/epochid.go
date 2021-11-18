package medley

// ActivationBlockNumberFromEpochID extracts the activation block number from an epoch id.
func ActivationBlockNumberFromEpochID(epochID uint64) uint32 {
	return uint32(epochID >> (8 * 4)) // take first 4 bytes
}
