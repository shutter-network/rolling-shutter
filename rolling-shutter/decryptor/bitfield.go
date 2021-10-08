package decryptor

import "math"

func getIndexes(bitField []byte) []int32 {
	var indexes []int32
	for m, b := range bitField {
		for i := 7; i >= 0; i-- {
			threshold := uint8(math.Pow(2, float64(i)))
			if b >= threshold {
				b -= threshold
				indexes = append(indexes, int32(m*8+i))
			}
		}
	}
	return indexes
}

func makeBitfieldFromArray(indexes []int32) []byte {
	out := []byte{}
	for _, i := range indexes {
		out = addBitfields(out, makeBitfieldFromIndex(i))
	}
	return out
}

func makeBitfieldFromIndex(index int32) []byte {
	bitfield := make([]byte, index/8)
	bit := math.Pow(2, float64(index%8))
	return append(bitfield, uint8(int64(bit)))
}

func addBitfields(bf1 []byte, bf2 []byte) []byte {
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
