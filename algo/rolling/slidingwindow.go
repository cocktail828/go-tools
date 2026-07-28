package rolling

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cocktail828/go-tools/algo/mathx"
	"github.com/cocktail828/go-tools/z/timex"
)

const (
	_ROLLING_MIN_COUNTER = 128                    // Minimum number of counters
	_ROLLING_WINSIZE     = 128                    // Window size in milliseconds
	_ROLLING_PRECISION   = _ROLLING_WINSIZE * 1e6 // 128ms precision in nanoseconds
)

// SlidingWindow implements a sliding window counter with fixed precision
// It uses multiple atomic counters to track events over time
// and provides efficient QPS calculation
type SlidingWindow struct {
	numCounter int64 // Number of counters (power of 2)
	counters   []struct {
		atomic.Int64              // Counter value
		nanots       atomic.Int64 // Timestamp in nanoseconds
	}
}

// NewSlidingWindow creates a new sliding window counter instance with 128ms precision
// n: number of counters, will be rounded up to the next power of 2
// Non-positive values are clamped to the minimum number of counters
// (_ROLLING_MIN_COUNTER, 128) so the constructor never panics.
func NewSlidingWindow(n int) *SlidingWindow {
	if n < _ROLLING_MIN_COUNTER {
		n = _ROLLING_MIN_COUNTER
	}
	n = max(int(mathx.Next2Power(int64(n))), _ROLLING_MIN_COUNTER)
	return &SlidingWindow{
		numCounter: int64(n),
		counters: make([]struct {
			atomic.Int64
			nanots atomic.Int64
		}, n),
	}
}

// String returns a string representation of the SlidingWindow counter
// showing non-zero counter values and their timestamps
func (sw *SlidingWindow) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "SlidingWindow: n:%d\n", sw.numCounter)
	for i := int64(0); i < sw.numCounter; i++ {
		curr := &sw.counters[i]
		if n := curr.Load(); n > 0 {
			fmt.Fprintf(sb, "\t%03d => %d\t%vns\n", i, n, curr.nanots.Load())
		}
	}

	return sb.String()
}

// indexByTime calculates the counter index for a given timestamp
// using bitwise operations for efficient lookup
func (sw *SlidingWindow) indexByTime(nsec int64) int64 {
	return (nsec / _ROLLING_PRECISION) & (sw.numCounter - 1)
}

func (sw *SlidingWindow) IncrBy(v int) {
	sw.incrBy(timex.UnixNano(), v)
}

// IncrBy atomically increases the counter value for the given timestamp
// It handles counter reset when the timestamp has moved to a new window
func (sw *SlidingWindow) incrBy(nsec int64, v int) {
	n := int64(v)
	pos := sw.indexByTime(nsec)
	floor := mathx.Floor(nsec, _ROLLING_PRECISION)
	curr := &sw.counters[pos]

	// Try to update the timestamp to the current window's floor value
	// If successful, reset the counter as we've moved to a new window
	for {
		old := curr.nanots.Load()

		// If we're still in the same window, just add to the counter
		if old == floor {
			curr.Add(n)
			return
		}

		// If we're in a new window (old < floor), try to claim it.
		if old < floor {
			// Snapshot the stale value left over from the previous window so
			// we can cancel it out once we win the slot.
			stale := curr.Load()
			if curr.nanots.CompareAndSwap(old, floor) {
				// Successfully claimed this window. Subtract the stale value
				// and add ours in a single atomic Add instead of Store, so any
				// concurrent Add landing on the freshly-claimed window is
				// preserved rather than clobbered.
				curr.Add(n - stale)
				return
			}
			// CAS failed, someone else updated it, retry
			continue
		}

		// old > floor shouldn't happen in normal operation (time going backwards)
		// In this case, we cannot safely update the counter as the timestamp
		// belongs to a future window. Ignore this increment to maintain correctness.
		return
	}
}

// estimate calculates the total number of events and valid windows within
// the specified number of time windows ending at the given timestamp
// It ignores expired counters
func (sw *SlidingWindow) estimate(nsec, n int64) (int64, int64) {
	// Nothing to look back over.
	if n <= 0 {
		return 0, 0
	}

	// Only the most recent sw.numCounter windows physically exist in the ring
	// buffer. Any window requested beyond that capacity has no backing slot, so
	// it contributes 0 to the total and is not a valid window. We iterate at
	// most sw.numCounter slots to avoid wrapping around and double-counting,
	// while the returned win reflects only the windows that actually hold data.
	scan := min(n, sw.numCounter)

	var cnt, win int64
	edge := sw.indexByTime(nsec)
	currentFloor := mathx.Floor(nsec, _ROLLING_PRECISION)
	// Look-back is bounded by scan: windows older than scan buckets ago either
	// were never requested (n < capacity) or no longer have a backing slot
	// (n > capacity, already overwritten in the ring), so they contribute 0.
	oldestFloor := currentFloor - _ROLLING_PRECISION*(scan-1)

	for i := int64(0); i < scan; i++ {
		indexByTime := (edge - i + sw.numCounter) & (sw.numCounter - 1)
		c := &sw.counters[indexByTime]

		// Check whether the counter belongs to a valid time window
		ts := c.nanots.Load()
		if ts >= oldestFloor && ts <= currentFloor {
			win++
			cnt += c.Load()
		}
	}
	return cnt, win
}

// Reset atomically resets all counter values and timestamps to zero
func (sw *SlidingWindow) Reset() {
	for i := range sw.counters {
		sw.counters[i].Store(0)
		sw.counters[i].nanots.Store(0)
	}
}

// Estimate returns the total events and valid windows at the Snapshot's timestamp
func (sw *SlidingWindow) Estimate(n int) (int64, int64) {
	cnt, win := sw.estimate(timex.UnixNano(), int64(n))
	return cnt, win
}

// QPS calculates the queries per second at the Snapshot's timestamp
func (sw *SlidingWindow) QPS(n int) float64 {
	cnt, win := sw.estimate(timex.UnixNano(), int64(n))
	if win == 0 {
		return 0
	}
	return float64(cnt) * 1e3 / float64(win) / _ROLLING_WINSIZE
}

// At creates a Snapshot of the SlidingWindow counter at the specified timestamp
// The Snapshot allows operations on the counter as if it were at that timestamp
func (sw *SlidingWindow) At(nsec int64) *Snapshot {
	return &Snapshot{sw, nsec}
}

// Snapshot represents a view of the SlidingWindow counter at a specific timestamp
// It allows performing operations on the counter at that frozen point in time
type Snapshot struct {
	sw *SlidingWindow
	tm int64
}

func (s *Snapshot) IncrBy(v int) {
	s.sw.incrBy(s.tm, v)
}

// Estimate returns the total events and valid windows at the Snapshot's timestamp
func (s *Snapshot) Estimate(n int) (int64, int64) {
	cnt, win := s.sw.estimate(s.tm, int64(n))
	return cnt, win
}

// QPS calculates the queries per second at the Snapshot's timestamp
func (s *Snapshot) QPS(n int) float64 {
	cnt, win := s.sw.estimate(s.tm, int64(n))
	if win == 0 {
		return 0
	}
	return float64(cnt) * 1e3 / float64(win) / _ROLLING_WINSIZE
}
