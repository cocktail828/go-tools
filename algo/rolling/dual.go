// Package rolling implements a high-performance sliding window algorithm for metrics collection
package rolling

import (
	"fmt"
	"strings"

	"github.com/cocktail828/go-tools/algo/mathx"
	"github.com/cocktail828/go-tools/z/timex"
)

// DualRolling wraps a Rolling counter to track two separate values simultaneously
// It uses bitwise operations to merge two 32-bit integers into a single 64-bit counter
// This is useful for tracking pairs of related metrics like success/failure counts,
// request/response sizes, etc.

type DualRolling struct {
	r *Rolling // Underlying single rolling counter
}

// String returns a string representation of the DualRolling counter
// showing non-zero high and low counter values along with their timestamps
func (dr *DualRolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "DualRolling: num:%d\n", dr.r.numCounter)
	for i := int64(0); i < dr.r.numCounter; i++ {
		curr := &dr.r.counters[i]
		if n := curr.Load(); n > 0 {
			high, low := mathx.SplitInt64(n)
			fmt.Fprintf(sb, "\t%03d => %d %d\t%vns\n", i, high, low, curr.nanots.Load())
		}
	}

	return sb.String()
}

// IncrBy atomically increases both counter values for the current time
// v0: value to add to the high 32 bits
// v1: value to add to the low 32 bits
func (dr *DualRolling) IncrBy(v0, v1 int) {
	dr.r.incrBy(timex.UnixNano(), mathx.MergeInt32(int32(v0), int32(v1)))
}

// Count returns the total values for both counters and the number of valid windows
// within the specified number of time windows ending at current time
// Returns: (highValue, lowValue, validWindows)
func (dr *DualRolling) Count(num int) (int64, int64, int64) {
	cnt, win := dr.r.count(timex.UnixNano(), int64(num))
	high, low := mathx.SplitInt64(cnt)
	return int64(high), int64(low), win
}

// QPS calculates the queries per second for both counters based on
// the specified number of time windows ending at current time
// Returns: (highQPS, lowQPS)
func (dr *DualRolling) QPS(num int) (float64, float64) {
	cnt0, cnt1, win := dr.Count(num)
	return calcQPS(cnt0, win), calcQPS(cnt1, win)
}

// Reset atomically resets both counter values to zero for all windows
func (dr *DualRolling) Reset() {
	for i := range dr.r.counters {
		dr.r.counters[i].Store(0)
	}
}

// At creates a snapshot of the DualRolling counter at the specified timestamp
// The snapshot allows operations on the counter as if it were at that timestamp
func (dr *DualRolling) At(nsec int64) *dualSnapshot {
	return &dualSnapshot{dr.r, nsec}
}

// dualSnapshot represents a view of the DualRolling counter at a specific timestamp
// It allows performing operations on the counter at that frozen point in time

type dualSnapshot struct {
	ro *Rolling
	tm int64
}

// IncrBy increases both counter values at the snapshot's timestamp
// v0: value to add to the high 32 bits
// v1: value to add to the low 32 bits
func (s *dualSnapshot) IncrBy(v0, v1 int) {
	s.ro.incrBy(s.tm, mathx.MergeInt32(int32(v0), int32(v1)))
}

// Count returns the total values for both counters and the number of valid windows
// at the snapshot's timestamp
// Returns: (highValue, lowValue, validWindows)
func (s *dualSnapshot) Count(num int) (int64, int64, int64) {
	cnt, win := s.ro.count(s.tm, int64(num))
	high, low := mathx.SplitInt64(cnt)
	return int64(high), int64(low), win
}

// QPS calculates the queries per second for both counters at the snapshot's timestamp
// Returns: (highQPS, lowQPS)
func (s *dualSnapshot) QPS(num int) (float64, float64) {
	cnt0, cnt1, win := s.Count(num)
	return calcQPS(cnt0, win), calcQPS(cnt1, win)
}
