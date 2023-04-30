package fasttime

import (
	"time"
)

const (
	internalToUnixMilli int64 = 62135596800000
)

type TimeIntf interface {
	Minute() int
	Second() int
}

var _ TimeIntf = Time{}

type Time struct {
	hour        int
	minute      int
	second      int
	millisecond int   // the millisecond offset within the second, in the range [0, 1000].
	unixMilli   int64 // the number of milliseconds elapsed since January 1, 1970 UTC.
}

func From(t time.Time) Time {
	return fromUnixMilliSec(t.UnixMilli())
}

func Now() Time {
	return fromUnixMilliSec(time.Now().UnixMilli())
}

func (tt Time) String() string {
	return time.UnixMilli(tt.unixMilli).String()
}

// Add returns the time t+d.
func (tt Time) Add(d time.Duration) Time {
	return fromUnixMilliSec(tt.unixMilli + int64(d/time.Millisecond))
}

func fromUnixMilliSec(tmpUnixMilli int64) Time {
	tmpMilliSec := tmpUnixMilli + internalToUnixMilli
	tmpSec := tmpMilliSec / 1e3

	return Time{
		hour:        int(tmpSec / 3600 % 24),
		minute:      int(tmpSec % 3600 / 60),
		second:      int(tmpSec % 60),
		millisecond: int(tmpMilliSec % 1e3),
		unixMilli:   tmpUnixMilli,
	}
}

func (tt Time) Time() time.Time {
	return time.UnixMilli(tt.unixMilli)
}

func (tt Time) IsZero() bool {
	return tt.unixMilli == 0
}

func (tt Time) Since(t Time) time.Duration {
	return time.Millisecond * time.Duration(tt.unixMilli-t.unixMilli)
}

// After reports whether the time instant t is after u.
func (tt Time) After(t Time) bool {
	return tt.unixMilli > t.UnixMilli()
}

// Before reports whether the time instant t is before u.
func (tt Time) Before(t Time) bool {
	return tt.unixMilli < t.UnixMilli()
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (tt Time) Sub(u Time) time.Duration {
	return time.Millisecond * time.Duration(tt.unixMilli-u.UnixMilli())
}

// Hour returns the hour within the day specified by t, in the range [0, 23].
func (tt Time) Hour() int {
	return tt.hour
}

// Minute returns the minute offset within the hour specified by t, in the range [0, 59].
func (tt Time) Minute() int {
	return tt.minute
}

// Second returns the second offset within the minute specified by t, in the range [0, 59].
func (tt Time) Second() int {
	return tt.second
}

// MilliSecond returns the milliSecond offset within the second specified by t, in the range [0, 1000].
func (tt Time) MilliSecond() int {
	return tt.millisecond
}

// Nanosecond returns the nanosecond offset within the second specified by t, in the range [0, 999999999].
func (tt Time) Unix() int64 {
	return tt.unixMilli / 1e3
}

// UnixMilli returns t as a Unix time, the number of milliseconds elapsed since January 1, 1970 UTC.
// The result is undefined if the Unix time in milliseconds cannot be represented by an int64
// (a date more than 292 million years before or after 1970). The result does not depend on the location associated with t.
func (tt Time) UnixMilli() int64 {
	return tt.unixMilli
}
