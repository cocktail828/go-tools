package hystrix

import (
	"math/rand"
	"sync"
	"testing"
	"testing/quick"
	"time"
)

func TestReportEventConcurrency(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	run := func() bool {
		// Make the circuit easily open and close intermittently.
		ConfigureCommand(t.Name(), Setting{
			MuteWindow:                 10,
			MaxConcurrency:             1,
			HealthCheckMinReqThreshold: 1,
			OpenOnFailureThreshold:     1,
		})

		cb := GetCircuit(t.Name())
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
				eventType := MaxConcurrencyEvent
				if rand.Intn(3) == 1 {
					eventType = SuccessEvent
				}
				cb.feedback(eventType, time.Now(), time.Now(), time.Second)
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
