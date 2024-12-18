package hystrix

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func failingPercent(p int) *statistic {
	m := newStatistic("")
	for i := 0; i < 100; i++ {
		t := "success"
		if i < p {
			t = "failure"
		}
		m.Chan <- temporary{Types: []string{t}}
	}

	// Updates needs to be flushed
	time.Sleep(100 * time.Millisecond)
	return m
}

func TestErrorPercent(t *testing.T) {
	m := failingPercent(40)
	now := time.Now()

	p := m.ErrorPercent(now)
	assert.Equal(t, 40, p, "ErrorPercent() should return 40")

	Configure("", Config{ErrorPercentThreshold: 39})
	assert.False(t, m.IsHealthy(now), "the metrics should be unhealthy")
}
