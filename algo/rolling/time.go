package rolling

import "time"

var (
	// for debug
	unixNano = func() int64 { return time.Now().UnixNano() }
)

func SetTime(f func() int64)    { unixNano = f }
func round(tm, gap int64) int64 { return (tm / gap) * gap }
