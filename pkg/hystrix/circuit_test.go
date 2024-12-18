package hystrix

import (
	"sync"
	"testing"
	"time"

	"math/rand"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestGetCircuit(t *testing.T) {
	cb1 := getCircuit("foo")
	cb2 := getCircuit("foo")
	assert.Equal(t, cb1, cb2)
}

func TestReportEventOpenThenClose(t *testing.T) {
	Configure("", Config{ErrorPercentThreshold: 50})
	cb := getCircuit("")
	assert.NotNil(t, cb)
	assert.False(t, cb.IsOpen())

	cb.statistic = failingPercent(100)
	assert.False(t, cb.statistic.IsHealthy(time.Now()))

	assert.Equal(t, nil, cb.ReportEvent([]string{"success"}, time.Now(), 0))
}

func TestReportEventMultiThreaded(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	run := func() bool {
		// Make the circuit easily open and close intermittently.
		Configure("", Config{
			MaxConcurrentRequests:  1,
			ErrorPercentThreshold:  1,
			RequestVolumeThreshold: 1,
			SleepWindow:            10,
		})

		cb := getCircuit("")
		count := 5
		wg := &sync.WaitGroup{}
		wg.Add(count)
		c := make(chan bool, count)
		for i := 0; i < count; i++ {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Error(r)
						c <- false
					} else {
						wg.Done()
					}
				}()
				// randomized eventType to open/close circuit
				eventType := "rejected"
				if rand.Intn(3) == 1 {
					eventType = "success"
				}
				err := cb.ReportEvent([]string{eventType}, time.Now(), time.Second)
				if err != nil {
					t.Error(err)
				}
				time.Sleep(time.Millisecond)
				// cb.IsOpen() internally calls cb.setOpen()
				cb.IsOpen()
			}()
		}
		go func() {
			wg.Wait()
			c <- true
		}()
		return <-c
	}
	if err := quick.Check(run, nil); err != nil {
		t.Error(err)
	}
}
