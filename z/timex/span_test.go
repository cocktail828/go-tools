package timex

import (
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	r := NewTimeRecorder()
	for range 3 {
		time.Sleep(time.Millisecond * 100)
		t.Logf("duration=%v", r.Duration())
	}
	t.Logf("elapse=%v", r.Elapse())
}
