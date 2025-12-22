package cm4

import "github.com/cocktail828/go-tools/algo/mathx"

// CM4 is a Count-Min sketch implementation with 4-bit counters, heavily
// based on Damian Gryski'cm CM4 [1].
//
// cm4 is a small conservative-update count-min sketch implementation with 4-bit counters
//
// [1]: https://github.com/dgryski/go-tinylfu/blob/master/cm4.go
type CM4 struct {
	s    [depth]nvec
	mask uint32
}

const depth = 4

func NewCM4(w int) *CM4 {
	if w < 1 {
		panic("cm4: bad width")
	}

	// use 4 counters per item per level, for a total of 16 counters or 8 bytes per item, matching the TinyLFU paper.
	w32 := uint32(mathx.Next2Power(int64(w) * 4))
	c := CM4{
		mask: w32 - 1,
	}

	for i := 0; i < depth; i++ {
		c.s[i] = newNvec(int(w32))
	}

	return &c
}

func (c *CM4) Add(keyh uint64) {
	// The loop unrolling prevents this function from being inlined, but it still results in a slight overall speedup.
	c.s[3].inc(c.counterOffset(keyh, 3))
	c.s[2].inc(c.counterOffset(keyh, 2))
	c.s[1].inc(c.counterOffset(keyh, 1))
	c.s[0].inc(c.counterOffset(keyh, 0))
}

// a hash func
func (c *CM4) counterOffset(keyh uint64, level int) uint32 {
	// counterOffset gets inlined and the compiler removes the duplicated computations of h1 and h2, so there is no
	// benefit to accepting h1 and h2 as arguments.
	h1, h2 := uint32(keyh), uint32(keyh>>32)
	return (h1 + uint32(level)*h2) & c.mask
}

func (c *CM4) Estimate(keyh uint64) byte {
	var minVal byte = 255
	minVal = min(c.s[3].get(c.counterOffset(keyh, 3)), minVal)
	minVal = min(c.s[2].get(c.counterOffset(keyh, 2)), minVal)
	minVal = min(c.s[1].get(c.counterOffset(keyh, 1)), minVal)
	minVal = min(c.s[0].get(c.counterOffset(keyh, 0)), minVal)
	return minVal
}

func (c *CM4) Reset() {
	// There is no point in unrolling this loop, the cost is dominated by nvec.reset, which is O(n)
	for _, n := range c.s {
		n.reset()
	}
}

// nybble vector
type nvec []byte

func newNvec(w int) nvec {
	return make(nvec, w/2)
}

func (n nvec) get(i uint32) byte {
	// Ugly, but as a single expression so the compiler will inline it :/
	return byte(n[i/2]>>((i&1)*4)) & 0x0f
}

func (n nvec) inc(i uint32) {
	idx := i / 2
	shift := (i & 1) * 4
	v := (n[idx] >> shift) & 0x0f
	if v < 15 {
		n[idx] += 1 << shift
	}
}

func (n nvec) reset() {
	for i := range n {
		n[i] = (n[i] >> 1) & 0x77
	}
}
