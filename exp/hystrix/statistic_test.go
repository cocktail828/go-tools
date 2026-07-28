package hystrix

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/cocktail828/go-tools/z/timex"
	"github.com/stretchr/testify/assert"
)

func TestStatistic_String(t *testing.T) {
	closed := Statistic{Concurrency: 2, FailRate: 12.5, QPS: 3.2, StateDuration: 100, IsOpen: false}
	s := closed.String()
	assert.True(t, strings.Contains(s, "State: closed"))
	assert.True(t, strings.Contains(s, "Concurrency: 2"))
	assert.True(t, strings.Contains(s, "12.50%"))

	open := Statistic{IsOpen: true}
	assert.True(t, strings.Contains(open.String(), "State: open"))
}

func TestStatistic_Values(t *testing.T) {
	timex.SetTime(func() int64 { return 0 })
	defer timex.ResetTime()

	h := NewHystrix(NewConfig())
	h.MinQPSThreshold.Update(0)

	// one success, one failure => 50% fail rate
	assert.NoError(t, h.DoC(context.Background(), t.Name(),
		func(ctx context.Context) error { return nil }))
	assert.Equal(t, net.ErrClosed, h.DoC(context.Background(), t.Name(),
		func(ctx context.Context) error { return net.ErrClosed }))

	st := h.Statistic()
	assert.EqualValues(t, 0, st.Concurrency)
	assert.InDelta(t, 50.0, st.FailRate, 0.01)
	assert.False(t, st.IsOpen)
}
