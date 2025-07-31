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
	for i := 0; i < 4; i++ {
		e.Check(errNoop)
	}
	assert.False(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() + int64(time.Minute) })
	for i := 0; i < 6; i++ {
		e.Check(nil)
	}
	assert.True(t, e.Alive())
}

func TestPercentageEvaluater(t *testing.T) {
	e := NewPercentageEvaluater(.9, .95)
	assert.True(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() })
	for i := 0; i < 100; i++ {
		if i < 91 {
			e.Check(errNoop)
		} else {
			e.Check(nil)
		}
	}
	assert.False(t, e.Alive())

	for i := 0; i < 20000; i++ {
		e.Check(nil)
	}
	assert.True(t, e.Alive())

	timex.SetTime(func() int64 { return time.Now().UnixNano() + int64(time.Minute) })
	for i := 0; i < 100; i++ {
		if i < 3 {
			e.Check(errNoop)
		} else {
			e.Check(nil)
		}
	}
	assert.True(t, e.Alive())
}
