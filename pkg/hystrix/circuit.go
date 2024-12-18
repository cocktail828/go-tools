package hystrix

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/pkg/hystrix/event"
	"github.com/cocktail828/go-tools/pkg/semaphore"
	"github.com/cocktail828/go-tools/z/locker"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

const (
	// DefaultTimeout is how long to wait for command to complete, in milliseconds
	DefaultTimeout = 1000 * time.Millisecond
	// DefaultMaxConcurrent is how many commands of the same type can run at the same time
	DefaultMaxConcurrent = 10
	// DefaultQPSThreshold is the minimum number of requests needed before a c can be tripped due to health
	DefaultQPSThreshold = 20
	// DefaultProbeInterval is how long, in milliseconds, to wait after a c opens before testing for recovery
	DefaultProbeInterval = 100 * time.Millisecond
	// DefaultErrorPercentThreshold causes circuits to isOpened once the rolling measure of errors exceeds this percent of requests
	DefaultErrorPercentThreshold = 80
)

type Config struct {
	Timeout               time.Duration `json:"timeout"`
	ProbeInterval         time.Duration `json:"sleep_window"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestQPSThreshold   int           `json:"request_qps_threshold"`
	ErrorPercentThreshold int           `json:"error_percent_threshold"`
}

func (cfg *Config) Normalize() {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	if cfg.MaxConcurrentRequests == 0 {
		cfg.MaxConcurrentRequests = DefaultMaxConcurrent
	}

	if cfg.RequestQPSThreshold == 0 {
		cfg.RequestQPSThreshold = DefaultQPSThreshold
	}

	if cfg.ProbeInterval == 0 {
		cfg.ProbeInterval = DefaultProbeInterval
	}

	if cfg.ErrorPercentThreshold == 0 {
		cfg.ErrorPercentThreshold = DefaultErrorPercentThreshold
	}
}

// circuitBreaker is created for each ExecutorPool to track whether requests
// should be attempted, or rejected if the Health of the c is too low.
type circuitBreaker struct {
	Config
	Name      string
	isOpened  atomic.Bool
	forceOpen atomic.Bool
	ticker    *time.Ticker
	sem       *semaphore.Weighted
	statistic *statistic
}

var (
	sfg               singleflight.Group
	circuitBreakerMu  sync.RWMutex
	circuitBreakerMap = map[string]*circuitBreaker{}
)

// Configure applies settings for a circuit
func Configure(name string, cfg Config) {
	cfg.Normalize()
	cb := getCircuit(name)
	locker.WithLock(&circuitBreakerMu, func() {
		cb.Config = cfg
	})
}

// getCircuit returns the c for the given command and whether this call created it.
func getCircuit(name string) *circuitBreaker {
	var c *circuitBreaker
	var has bool
	locker.WithRLock(&circuitBreakerMu, func() {
		c, has = circuitBreakerMap[name]
	})

	if has {
		return c
	}

	val, _, _ := sfg.Do(name, func() (any, error) {
		v := newCircuitBreaker(name)
		locker.WithLock(&circuitBreakerMu, func() {
			circuitBreakerMap[name] = v
		})
		return v, nil
	})
	return val.(*circuitBreaker)
}

// newCircuitBreaker creates a circuitBreaker with associated Health
func newCircuitBreaker(name string) *circuitBreaker {
	cb := &circuitBreaker{
		Config: Config{
			Timeout:               DefaultTimeout,
			MaxConcurrentRequests: DefaultMaxConcurrent,
			RequestQPSThreshold:   DefaultQPSThreshold,
			ProbeInterval:         DefaultProbeInterval,
			ErrorPercentThreshold: DefaultErrorPercentThreshold,
		},
		Name:      name,
		sem:       semaphore.NewWeighted(int64(DefaultQPSThreshold)),
		statistic: newStatistic(name),
	}
	cb.ticker = time.NewTicker(cb.ProbeInterval)
	return cb
}

func (c *circuitBreaker) TryAcquire() bool {
	return c.sem.TryAcquire(1)
}

func (c *circuitBreaker) Acquire() {
	c.sem.Acquire(context.Background(), 1)
}

func (c *circuitBreaker) Release() {
	c.sem.Release(1)
}

func (c *circuitBreaker) ActiveCount() int {
	return int(c.sem.Occupied())
}

// Interactive allows manually causing the fallback logic for all instances
// of a given command.
func (c *circuitBreaker) Interactive(forceOpen bool) {
	c.forceOpen.Store(forceOpen)
}

// IsOpen is called before any Command execution to check whether or
// not it should be attempted. An "isOpened" c means it is disabled.
func (c *circuitBreaker) IsOpen() bool {
	if c.forceOpen.Load() || c.isOpened.Load() {
		return true
	}

	if int(c.statistic.Requests().QPS()) < c.RequestQPSThreshold {
		return false
	}

	// too many failures, isOpened the c
	if !c.statistic.IsHealthy(time.Now()) {
		c.SetOpen()
		return true
	}

	return false
}

// AllowRequest is checked before a command executes, ensuring that c state and metric health allow it.
// When the c is isOpened, this call will occasionally return true to measure whether the external service
// has recovered.
func (c *circuitBreaker) AllowRequest() bool {
	return !c.IsOpen() || c.allowSingleTest()
}

func (c *circuitBreaker) allowSingleTest() bool {
	select {
	case <-c.ticker.C:
		log.Printf("circuit: allowing single test to possibly close c %v", c.Name)
		return true
	default:
		return false
	}
}

func (c *circuitBreaker) SetOpen() {
	if c.isOpened.CompareAndSwap(false, true) {
		log.Printf("circuit: opening c %v", c.Name)
	}
}

func (c *circuitBreaker) SetClose() {
	if c.isOpened.CompareAndSwap(true, false) {
		log.Printf("circuit: closing c %v", c.Name)
	}
}

// ReportEvent records command statistic for tracking recent error rates and exposing data to the dashboard.
func (c *circuitBreaker) ReportEvent(events []event.Event, start time.Time, elapsed time.Duration) error {
	if len(events) == 0 {
		return errors.Errorf("no event types sent for statistic")
	}

	opened := c.isOpened.Load()
	if events[0] == event.Success && opened {
		c.SetClose()
	}

	select {
	case c.statistic.Chan <- temporary{
		Types:       events,
		Start:       start,
		Elapsed:     elapsed,
		Concurrency: c.ActiveCount(),
	}:
	default:
		return errors.Errorf("statistic channel (%v) is at capacity", c.Name)
	}

	return nil
}
