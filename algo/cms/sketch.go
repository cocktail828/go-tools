package cms

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cocktail828/go-tools/z/mathx"
)

// CMSketch is a Count-Min sketch implementation with 4-bit counters, heavily
// based on Damian Gryski's CM4 [1].
//
// [1]: https://github.com/dgryski/go-tinylfu/blob/master/cm4.go
type CMSketch struct {
	rows [cmDepth]cmRow
	seed [cmDepth]uint64
	mask uint64
}

const (
	// cmDepth is the number of counter copies to store (think of it as rows).
	cmDepth = 4
)

func NewCMSketch(numCounters int64) *CMSketch {
	if numCounters == 0 {
		panic("CMSketch: bad numCounters")
	}
	// Get the next power of 2 for better cache performance.
	numCounters = mathx.Next2Power(numCounters)
	sketch := &CMSketch{mask: uint64(numCounters - 1)}
	// Initialize rows of counters and seeds.
	// Cryptographic precision not needed
	source := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	for i := 0; i < cmDepth; i++ {
		sketch.seed[i] = source.Uint64()
		sketch.rows[i] = newCmRow(numCounters)
	}
	return sketch
}

// Increment increments the count(ers) for the specified key.
func (s *CMSketch) Increment(hashed uint64) {
	for i := range s.rows {
		s.rows[i].increment((hashed ^ s.seed[i]) & s.mask)
	}
}

// Estimate returns the value of the specified key.
func (s *CMSketch) Estimate(hashed uint64) int64 {
	min := byte(255)
	for i := range s.rows {
		val := s.rows[i].get((hashed ^ s.seed[i]) & s.mask)
		if val < min {
			min = val
		}
	}
	return int64(min)
}

// Reset halves all counter values.
func (s *CMSketch) Reset() {
	for _, r := range s.rows {
		r.reset()
	}
}

// Clear zeroes all counters.
func (s *CMSketch) Clear() {
	for _, r := range s.rows {
		r.clear()
	}
}

// cmRow is a row of bytes, with each byte holding two counters.
type cmRow []byte

func newCmRow(numCounters int64) cmRow {
	return make(cmRow, numCounters/2)
}

func (r cmRow) get(n uint64) byte {
	return (r[n/2] >> ((n & 1) * 4)) & 0x0f
}

func (r cmRow) increment(n uint64) {
	// Index of the counter.
	i := n / 2
	// Shift distance (even 0, odd 4).
	s := (n & 1) * 4
	// Counter value.
	v := (r[i] >> s) & 0x0f
	// Only increment if not max value (overflow wrap is bad for LFU).
	if v < 15 {
		r[i] += 1 << s
	}
}

func (r cmRow) reset() {
	// Halve each counter.
	for i := range r {
		r[i] = (r[i] >> 1) & 0x77
	}
}

func (r cmRow) clear() {
	// Zero each counter.
	for i := range r {
		r[i] = 0
	}
}

func (r cmRow) string() string {
	s := ""
	for i := uint64(0); i < uint64(len(r)*2); i++ {
		s += fmt.Sprintf("%02d ", (r[(i/2)]>>((i&1)*4))&0x0f)
	}
	s = s[:len(s)-1]
	return s
}
