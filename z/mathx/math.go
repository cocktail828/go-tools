package mathx

import "strings"

// Next2Power rounds x up to the next power of 2, if it's not already one.
func Next2Power(x int64) int64 {
	if x >= 0 {
		return next2Power(x)
	}

	if y := -next2Power(-x); x == y {
		return y
	} else {
		return y / 2
	}
}

// x >= 0
func next2Power(x int64) int64 {
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

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func ToBase62(n int64) string {
	if n == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	for n > 0 {
		result = append(result, base62Chars[n%62])
		n /= 62
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func FromBase62(s string) int64 {
	var result int64
	for _, c := range s {
		result = result*62 + int64(strings.Index(base62Chars, string(c)))
	}
	return result
}
