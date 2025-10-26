package healthy

import (
	"errors"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

var errNoop = errors.New("noop error")

func TestCounterEvaluater(t *testing.T) {
	e := NewCounterEvaluater(3, 5)
	assert.True(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() })
	for range 4 {
		e.Check(errNoop)
	}
	assert.False(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() + int64(time.Minute) })
	for range 6 {
		e.Check(nil)
	}
	assert.True(t, e.Alive())
}

func TestPercentageEvaluater(t *testing.T) {
	e := NewPercentageEvaluater(.9, .95)
	assert.True(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() })
	for i := range 100 {
		if i < 91 {
			e.Check(errNoop)
		} else {
			e.Check(nil)
		}
	}
	assert.False(t, e.Alive())

	for range 20000 {
		e.Check(nil)
	}
	assert.True(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() + int64(time.Minute) })
	for i := range 100 {
		if i < 3 {
			e.Check(errNoop)
		} else {
			e.Check(nil)
		}
	}
	assert.True(t, e.Alive())
}
