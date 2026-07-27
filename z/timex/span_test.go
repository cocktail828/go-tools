package timex

import (
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	r := Recorder{}
	r.Reset()
	for range 3 {
		time.Sleep(time.Millisecond * 100)
		t.Logf("duration=%v", r.Duration())
	}
	t.Logf("elapse=%v", r.Elapse())
}

func TestBlockChecker(t *testing.T) {
	bc := NewBlockChecker(time.Millisecond*100, func(dur time.Duration) {
		t.Log("busy work timeout...", dur)
	})
	for range 5 {
		bc.Ping()
		time.Sleep(time.Millisecond * 50)
	}

	for range 5 {
		bc.Ping()
		time.Sleep(time.Millisecond * 500)
	}
}
