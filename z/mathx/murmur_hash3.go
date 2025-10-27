package mathx

import (
	"encoding/binary"
)

// MurmurHash3_32 calculates the 32-bit MurmurHash3 hash value for the given input data.
// Parameters:
//
//	data: input data (byte slice)
//	seed: hash seed (used to customize hash results, different seeds can generate different hash values)
//
// Returns: 32-bit hash value
func MurmurHash3_32(data []byte, seed uint32) uint32 {
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
		r1 uint32 = 15
		r2 uint32 = 13
		m  uint32 = 5
		n  uint32 = 0xe6546b64
	)

	hash := seed
	length := uint32(len(data))
	blocks := length / 4 // process 4 bytes per block

	for i := uint32(0); i < blocks; i++ {
		k := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])

		k *= c1
		k = (k << r1) | (k >> (32 - r1)) // rotate left r1 bits
		k *= c2

		hash ^= k
		hash = (hash << r2) | (hash >> (32 - r2)) // rotate left r2 bits
		hash = hash*m + n
	}

	// handle remaining bytes
	remaining := data[blocks*4:]
	var k1 uint32
	switch len(remaining) {
	case 3:
		k1 ^= uint32(remaining[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(remaining[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(remaining[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		hash ^= k1
	}

	hash ^= length
	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> 13
	hash *= 0xc2b2ae35
	hash ^= hash >> 16

	return hash
}

// MurmurHash3_64 calculates the 64-bit MurmurHash3 hash value for the given input data.
// Parameters:
//
//	data: input data (byte slice)
//	seed: hash seed (used to customize hash results, different seeds can generate different hash values)
//
// Returns: 64-bit hash value
func MurmurHash3_64(data []byte, seed uint64) uint64 {
	const (
		c1 uint64 = 0x87c37b91114253d5
		c2 uint64 = 0x4cf5ad432745937f
		r1 uint64 = 31
		r2 uint64 = 27
		r3 uint64 = 33
		m  uint64 = 5
		n  uint64 = 0x52dce729
	)

	hash := seed
	length := uint64(len(data))
	blocks := length / 8 // process 8 bytes per block

	// process 8 bytes per block
	for i := uint64(0); i < blocks; i++ {
		// read 8 bytes from byte slice (little-endian)
		k := binary.LittleEndian.Uint64(data[i*8 : (i+1)*8])

		// 混合操作
		k *= c1
		k = (k << r1) | (k >> (64 - r1)) // rotate left r1 bits
		k *= c2

		hash ^= k
		hash = (hash << r2) | (hash >> (64 - r2)) // rotate left r2 bits
		hash = hash*m + n
	}

	// handle remaining bytes
	remaining := data[blocks*8:]
	var k1 uint64
	switch len(remaining) {
	case 7:
		k1 ^= uint64(remaining[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(remaining[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(remaining[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(remaining[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(remaining[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(remaining[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(remaining[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (64 - r1))
		k1 *= c2
		hash ^= k1
	}

	// finalize hash
	hash ^= length
	hash ^= hash >> r3
	hash *= c1
	hash ^= hash >> r2
	hash *= c2
	hash ^= hash >> r3

	return hash
}
