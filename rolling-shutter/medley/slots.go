package medley

func BlockTimestampToSlot(blockTimestamp uint64, genesisSlotTimestamp uint64, secondsPerSlot uint64) uint64 {
	return (blockTimestamp - genesisSlotTimestamp) / secondsPerSlot
}

func SlotToEpoch(slot uint64, slotsPerEpoch uint64) uint64 {
	return slot / slotsPerEpoch
}

func SlotToTimestamp(slot uint64, genesisSlotTimestamp uint64, secondsPerSlot uint64) uint64 {
	return genesisSlotTimestamp + slot*secondsPerSlot
}
