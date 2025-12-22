package mathx

// Next2Power rounds x up to the next power of 2, if it's not already one.
func Next2Power(x int64) int64 {
	if x < 0 {
		panic("x cannot be a negative value")
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}

// NumOfOnes counts the number of set bits (ones) in the binary representation of x.
// This version supports negative numbers.
func NumOfOnes(x int64) int {
	ones := int64(0)
	if x < 0 {
		x = -x
	}

	for ; x != 0; x >>= 1 {
		ones += x & 0x1
	}
	return int(ones)
}

// Floor rounds x down to the nearest multiple of base.
// If base is 0, it returns 0.
func Floor(x, base int64) int64 { return x - x%base }

// Ceil rounds x up to the nearest multiple of base.
// If base is 0, it returns 0.
func Ceil(x, base int64) int64 {
	if x%base == 0 {
		return x
	}
	return x + (base - x%base)
}

// MergeInt32 merges two int32 values into a single int64 value.
func MergeInt32(high, low int32) int64 {
	return (int64(high) << 32) | int64(low)
}

// SplitInt64 splits an int64 value into two int32 values.
func SplitInt64(n int64) (high, low int32) {
	low = int32(n & 0xFFFFFFFF)
	high = int32((n >> 32) & 0xFFFFFFFF)
	return
}
