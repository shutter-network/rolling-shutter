package medley

import "encoding/binary"

func Uint64EpochIDToBytes(epochID uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, epochID)
	return b
}

func BytesEpochIDToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
