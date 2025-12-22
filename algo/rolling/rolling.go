// Package rolling implements a high-performance sliding window algorithm for metrics collection
// such as QPS calculation, request counting, etc.
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

// Rolling implements a sliding window counter with fixed precision
// It uses multiple atomic counters to track events over time
// and provides efficient QPS calculation

type Rolling struct {
	numCounter int64 // Number of counters (power of 2)
	counters   []struct {
		atomic.Int64              // Counter value
		nanots       atomic.Int64 // Timestamp in nanoseconds
	}
}

// NewRolling creates a new sliding window counter instance with 128ms precision
// num: number of counters, will be rounded up to the next power of 2
// Minimum number of counters is _ROLLING_MIN_COUNTER (128)
func NewRolling(num int) *Rolling {
	num = max(int(mathx.Next2Power(int64(num))), _ROLLING_MIN_COUNTER)
	return &Rolling{
		numCounter: int64(num),
		counters: make([]struct {
			atomic.Int64
			nanots atomic.Int64
		}, num),
	}
}

// String returns a string representation of the Rolling counter
// showing non-zero counter values and their timestamps
func (r *Rolling) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Rolling: num:%d\n", r.numCounter)
	for i := int64(0); i < r.numCounter; i++ {
		curr := &r.counters[i]
		if n := curr.Load(); n > 0 {
			fmt.Fprintf(sb, "\t%03d => %d\t%vns\n", i, n, curr.nanots.Load())
		}
	}

	return sb.String()
}

// indexByTime calculates the counter index for a given timestamp
// using bitwise operations for efficient lookup
func (r *Rolling) indexByTime(nsec int64) int64 {
	return (nsec / _ROLLING_PRECISION) & (r.numCounter - 1)
}

// incrBy atomically increases the counter value for the given timestamp
// It handles counter reset when the timestamp has moved to a new window
func (r *Rolling) incrBy(nsec, n int64) {
	pos := r.indexByTime(nsec)
	floor := mathx.Floor(nsec, _ROLLING_PRECISION)

	curr := &r.counters[pos]
	for old := curr.nanots.Load(); old < floor; {
		if curr.nanots.CompareAndSwap(old, nsec) {
			curr.Store(0) // Reset counter for new time window
			break
		}
	}
	curr.Add(n)
}

// count calculates the total number of events and valid windows within
// the specified number of time windows ending at the given timestamp
// It ignores expired counters
func (r *Rolling) count(nsec, num int64) (int64, int64) {
	if num > r.numCounter {
		num = r.numCounter
	}

	var cnt, win int64
	edge := r.indexByTime(nsec)
	old := mathx.Floor(nsec, _ROLLING_PRECISION) - _ROLLING_PRECISION*(num-1)

	for i := int64(0); i < num; i++ {
		indexByTime := (edge - i + r.numCounter) & (r.numCounter - 1)
		c := &r.counters[indexByTime]

		// Check whether the counter is expired
		if c.nanots.Load() >= old {
			win++
			cnt += c.Load()
		}
	}
	return cnt, win
}

// calcQPS calculates the queries per second based on total count and valid windows
func calcQPS(cnt, win int64) float64 {
	if win == 0 {
		return 0
	}
	return float64(cnt) * 1e3 / float64(win) / _ROLLING_WINSIZE
}

// IncrBy atomically increases the counter value for the current time
func (r *Rolling) IncrBy(n int) {
	r.incrBy(timex.UnixNano(), int64(n))
}

// Count returns the total number of events and valid windows within
// the specified number of time windows ending at current time
func (r *Rolling) Count(num int) (int64, int64) {
	cnt, win := r.count(timex.UnixNano(), int64(num))
	return cnt, win
}

// QPS calculates the queries per second for the current time
// based on the specified number of time windows
func (r *Rolling) QPS(num int) float64 {
	cnt, win := r.count(timex.UnixNano(), int64(num))
	return calcQPS(cnt, win)
}

// Reset atomically resets all counter values to zero
func (r *Rolling) Reset() {
	for i := range r.counters {
		r.counters[i].Store(0)
	}
}

// Dual creates a new DualRolling instance wrapping this Rolling counter
// DualRolling provides functionality for tracking both success and failure events
func (r *Rolling) Dual() *DualRolling {
	return &DualRolling{r}
}

// At creates a snapshot of the Rolling counter at the specified timestamp
// The snapshot allows operations on the counter as if it were at that timestamp
func (r *Rolling) At(nsec int64) *snapshot {
	return &snapshot{r, nsec}
}

// snapshot represents a view of the Rolling counter at a specific timestamp
// It allows performing operations on the counter at that frozen point in time

type snapshot struct {
	ro *Rolling
	tm int64
}

// IncrBy increases the counter value at the snapshot's timestamp
func (s *snapshot) IncrBy(n int) {
	s.ro.incrBy(s.tm, int64(n))
}

// Count returns the total events and valid windows at the snapshot's timestamp
func (s *snapshot) Count(num int) (int64, int64) {
	cnt, win := s.ro.count(s.tm, int64(num))
	return cnt, win
}

// QPS calculates the queries per second at the snapshot's timestamp
func (s *snapshot) QPS(num int) float64 {
	cnt, win := s.ro.count(s.tm, int64(num))
	return calcQPS(cnt, win)
}
