package mathx

// Next2Power rounds x up to the next power of 2, if it's not already one.
func Next2Power(x int64) int64 {
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

func Min[T int | uint | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](min, val T) T {
	if val <= min {
		return val
	}
	return min
}

func Max[T int | uint | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](max, val T) T {
	if val >= max {
		return val
	}
	return max
}
