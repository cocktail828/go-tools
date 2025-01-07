package cm4

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z/mathx"
)

// CM4 is a Count-Min sketch implementation with 4-bit counters, heavily
// based on Damian Gryski'cm CM4 [1].
//
// [1]: https://github.com/dgryski/go-tinylfu/blob/master/cm4.go
type CM4 struct {
	rows [cmDepth]cmRow
	seed [cmDepth]uint64
	mask uint64
	mu   sync.RWMutex // 添加读写锁以支持并发访问
}

const (
	// cmDepth is the number of counter copies to store (think of it as rows).
	cmDepth = 4
)

func NewCM4(numCounters int64) *CM4 {
	if numCounters == 0 {
		panic("CM4: bad numCounters")
	}
	// Get the next power of 2 for better cache performance.
	numCounters = mathx.Next2Power(numCounters)
	cm4 := &CM4{mask: uint64(numCounters - 1)}
	// Initialize rows of counters and seeds.
	// Cryptographic precision not needed
	source := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	for i := 0; i < cmDepth; i++ {
		cm4.seed[i] = source.Uint64()
		cm4.rows[i] = newCmRow(numCounters)
	}
	return cm4
}

// Increment increments the count(ers) for the specified key.
func (cm *CM4) Increment(hashed uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i := range cm.rows {
		cm.rows[i].increment((hashed ^ cm.seed[i]) & cm.mask)
	}
}

// Estimate returns the value of the specified key.
func (cm *CM4) Estimate(hashed uint64) int64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	min := byte(255)
	for i := range cm.rows {
		val := cm.rows[i].get((hashed ^ cm.seed[i]) & cm.mask)
		if val < min {
			min = val
		}
	}
	return int64(min)
}

// Reset halves all counter values.
func (cm *CM4) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, r := range cm.rows {
		r.reset()
	}
}

// Clear zeroes all counters.
func (cm *CM4) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, r := range cm.rows {
		r.clear()
	}
}

// cmRow is a row of bytes, with each byte holding two counters.
type cmRow []byte

func newCmRow(numCounters int64) cmRow {
	return make(cmRow, numCounters/2)
}

// get returns the value of the counter at index n.
func (r cmRow) get(n uint64) byte {
	return (r[n/2] >> ((n & 1) * 4)) & 0x0f
}

// increment increments the counter at index n.
func (r cmRow) increment(n uint64) {
	i := n / 2               // Index of the counter.
	cm := (n & 1) * 4        // Shift distance (even 0, odd 4).
	v := (r[i] >> cm) & 0x0f // Counter value.

	// Only increment if not max value (overflow wrap is bad for LFU).
	if v < 15 {
		r[i] += 1 << cm
	}
}

// reset halves all counter values.
func (r cmRow) reset() {
	for i := range r {
		// Halve each counter (two counters per byte).
		r[i] = (r[i] >> 1) & 0x77
	}
}

// clear zeroes all counters.
func (r cmRow) clear() {
	// Zero each counter.
	for i := range r {
		r[i] = 0
	}
}

// string returns a string representation of the row.
func (r cmRow) String() string {
	cm := ""
	for i := uint64(0); i < uint64(len(r)*2); i++ {
		cm += fmt.Sprintf("%02d ", (r[(i/2)]>>((i&1)*4))&0x0f)
	}
	cm = cm[:len(cm)-1]
	return cm
}
