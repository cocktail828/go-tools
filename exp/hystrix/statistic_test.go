package hystrix

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFailPercent(t *testing.T) {
	s := &GetCircuit(t.Name()).statistic
	pct := 40
	for i := 0; i < 100; i++ {
		t := SuccessEvent
		if i < pct {
			t = ErrorEvent
		}
		s.Update(Event{eventType: t, stopAt: time.UnixMilli(0)})
	}

	assert.Equal(t, 40, s.FailRate(0))
}

func TestStatistic_Success(t *testing.T) {
	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, <-errChan)

	msec := time.Now().UnixMilli()
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.requests.QPS(msec, 1))
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.success.QPS(msec, 1))
	assert.Equal(t, float64(0), GetCircuit(t.Name()).statistic.failure.QPS(msec, 1))
}

func TestStatistic_Timeout(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{Timeout: 15 * time.Millisecond})

	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	assert.Equal(t, ErrTimeout, <-errChan)

	msec := time.Now().UnixMilli()
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.requests.QPS(msec, 1))
	assert.Equal(t, float64(0), GetCircuit(t.Name()).statistic.success.QPS(msec, 1))
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.failure.QPS(msec, 1))
}

func TestStatistic_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := GoC(ctx, t.Name(), func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, ErrCanceled, <-errChan)

	msec := time.Now().UnixMilli()
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.requests.QPS(msec, 1))
	assert.Equal(t, float64(0), GetCircuit(t.Name()).statistic.success.QPS(msec, 1))
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.failure.QPS(msec, 1))
}

func TestStatistic_CircuitOpen(t *testing.T) {
	GetCircuit(t.Name()).Manually(true)

	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, ErrCircuitOpen, <-errChan)

	msec := time.Now().UnixMilli()
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.requests.QPS(msec, 1))
	assert.Equal(t, float64(0), GetCircuit(t.Name()).statistic.success.QPS(msec, 1))
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.failure.QPS(msec, 1))
}

func TestStatistic_MaxConcurrency(t *testing.T) {
	ConfigureCommand(t.Name(), Setting{MaxConcurrency: 1})
	<-GetCircuit(t.Name()).tickets

	errChan := GoC(context.Background(), t.Name(), func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, ErrMaxConcurrency, <-errChan)

	msec := time.Now().UnixMilli()
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.requests.QPS(msec, 1))
	assert.Equal(t, float64(0), GetCircuit(t.Name()).statistic.success.QPS(msec, 1))
	assert.Equal(t, 7.8125, GetCircuit(t.Name()).statistic.failure.QPS(msec, 1))
}
