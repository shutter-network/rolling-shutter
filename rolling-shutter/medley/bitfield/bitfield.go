package bitfield

// GetIndexes returns the sorted indexes that are set in the bitField.
func GetIndexes(bitField []byte) []int32 {
	var indexes []int32
	for m, b := range bitField {
		for i := 0; i < 8; i++ {
			if b&(1<<i) != 0 {
				indexes = append(indexes, int32(m*8+i))
			}
		}
	}
	return indexes
}

func MakeBitfieldFromIndex(indexes ...int32) []byte {
	var maxIndex int32 = 0
	for _, i := range indexes {
		if i > maxIndex {
			maxIndex = i
		}
	}

	bitfield := make([]byte, 1+maxIndex/8)
	for _, i := range indexes {
		bitfield[i/8] |= 1 << (i % 8)
	}

	return bitfield
}

func AddBitfields(bf1 []byte, bf2 []byte) []byte {
	if len(bf1) < len(bf2) {
		bf1, bf2 = bf2, bf1
	}
	out := make([]byte, len(bf1))
	for i, b := range bf1 {
		out[i] = b
		if i < len(bf2) {
			out[i] |= bf2[i]
		}
	}
	return out
}
