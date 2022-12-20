package wallet

import "math"

func varIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

func varSliceSerializeSize(val []byte) int {
	return varIntSerializeSize(uint64(len(val))) + len(val)
}
