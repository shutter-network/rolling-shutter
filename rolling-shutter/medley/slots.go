package medley

func BlockTimestampToSlot(blockTimestamp uint64, genesisSlotTimestamp uint64, secondsPerSlot uint64) uint64 {
	return (blockTimestamp - genesisSlotTimestamp) / secondsPerSlot
}
