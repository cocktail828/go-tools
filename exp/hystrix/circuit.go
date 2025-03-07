package hystrix

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cocktail828/go-tools/algo/rolling"
)

var (
	// DefaultTimeout is how long to wait for command to complete, in milliseconds
	DefaultTimeout = 1000 * time.Millisecond
	// DefaultMuteWindow is how long, in milliseconds, to wait after a circuitbreaker opens before testing for recovery
	DefaultMuteWindow = 5000 * time.Millisecond
	// DefaultRecoveryProbes is how many probes circuitbreaker will try, before recovery
	DefaultRecoveryProbes = 3
	// DefaultMaxConcurrency is how many commands of the same type can run at the same time
	DefaultMaxConcurrency = 10
	// DefaultMinRequestNum is the minimum number of requests needed before a circuitbreaker can be tripped due to health
	DefaultMinRequestNum = 20
	// DefaultErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	DefaultFailureThreshold = 20 // 0~100
)

type Setting struct {
	Timeout                    time.Duration `json:"timeout"`     // ms
	MuteWindow                 time.Duration `json:"mute_window"` // ms
	Probes                     int           `json:"probes"`
	MaxConcurrency             int           `json:"max_concurrency"`
	HealthCheckMinReqThreshold int           `json:"min_request_threshold"`
	OpenOnFailureThreshold     int           `json:"failure_threshold"` // 0~100
}

func (s *Setting) Normalize() {
	if s.Timeout <= 0 {
		s.Timeout = DefaultTimeout
	}

	if s.MuteWindow <= 0 {
		s.MuteWindow = DefaultMuteWindow
	}

	if s.Probes <= 0 {
		s.Probes = DefaultRecoveryProbes
	}

	if s.MaxConcurrency <= 0 {
		s.MaxConcurrency = DefaultMaxConcurrency
	}

	if s.HealthCheckMinReqThreshold <= 0 {
		s.HealthCheckMinReqThreshold = DefaultMinRequestNum
	}

	if s.OpenOnFailureThreshold <= 0 {
		s.OpenOnFailureThreshold = DefaultFailureThreshold
	}
}

func Configure(settings map[string]Setting) {
	for k, v := range settings {
		ConfigureCommand(k, v)
	}
}

// ConfigureCommand 方法用于配置单个断路器, may blocked
func ConfigureCommand(name string, setting Setting) {
	setting.Normalize()
	val, exist := circuitBreakers.LoadOrStore(name, newCircuitBreaker(name, setting))
	if exist { // overwrite old settings
		cb := val.(*CircuitBreaker)
		cb.updateSetting(setting)
	}
}

var (
	// name -> *CircuitBreaker
	circuitBreakers = &sync.Map{}
)

type CircuitBreaker struct {
	name        string
	setting     Setting
	lastTestAt  atomic.Int64  // nanoseconds
	isOpen      atomic.Bool   // circuit state
	isForceOpen bool          // manually trun on/off the circuit
	tickets     chan struct{} // for concurrency control
	statistic   Statistic
	recovery    Recovery
}

func GetCircuit(name string) *CircuitBreaker {
	s := Setting{}
	s.Normalize()
	val, _ := circuitBreakers.LoadOrStore(name, newCircuitBreaker(name, s))
	return val.(*CircuitBreaker)
}

func newCircuitBreaker(name string, s Setting) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:    name,
		setting: s,
		tickets: make(chan struct{}, s.MaxConcurrency),
		statistic: Statistic{
			requests: rolling.NewRolling(0, 128),
			success:  rolling.NewRolling(0, 128),
			failure:  rolling.NewRolling(0, 128),
		},
		recovery: Recovery{array: make([]bool, s.Probes)},
	}

	for i := 0; i < cb.setting.MaxConcurrency; i++ {
		cb.tickets <- struct{}{}
	}

	return cb
}

func (cb *CircuitBreaker) activeCount() int {
	return cb.setting.MaxConcurrency - len(cb.tickets)
}

// open or close the circuitbreaker manually
func (cb *CircuitBreaker) Manually(toggle bool) {
	cb.isForceOpen = toggle
}

func (cb *CircuitBreaker) IsOpen() bool {
	if cb.isForceOpen {
		return true
	}

	msec := nowFunc().UnixMilli()
	if cb.isOpen.Load() {
		// recovery?
		if cb.recovery.IsHealthy() {
			cb.setClose()
			return false
		}
		return true
	}

	// regardless of whether it succeeds or not, do not perform a health check when the current QPS is too low.
	if int(cb.statistic.QPS(msec)) < cb.setting.HealthCheckMinReqThreshold {
		return false
	}

	// too many failures, open the circuitbreaker
	if cb.statistic.FailRate(msec) >= cb.setting.OpenOnFailureThreshold {
		cb.setOpen()
		return true
	}

	return false
}

func (cb *CircuitBreaker) allowRequest() bool {
	return !cb.IsOpen() || cb.allowSingleTest()
}

// single test should be applied when circuit is opened
func (cb *CircuitBreaker) allowSingleTest() bool {
	if cb.isOpen.Load() {
		msec := nowFunc().UnixNano()
		lastTestAt := cb.lastTestAt.Load()
		if msec > lastTestAt+cb.setting.MuteWindow.Nanoseconds() {
			swapped := cb.lastTestAt.CompareAndSwap(lastTestAt, msec)
			if swapped {
				log.Printf("hystrix-go: allowing single test - %v\n", cb.name)
			}
			return swapped
		}
	}

	return false
}

func (cb *CircuitBreaker) setOpen() {
	if cb.isOpen.CompareAndSwap(false, true) {
		cb.lastTestAt.Store(nowFunc().UnixNano())
		cb.statistic.Reset()
		log.Printf("hystrix-go: opening circuit - %v\n", cb.name)
	}
}

func (cb *CircuitBreaker) setClose() {
	if cb.isOpen.CompareAndSwap(true, false) {
		cb.recovery.Reset()
		cb.statistic.Reset()
		log.Printf("hystrix-go: closing circuit - %v\n", cb.name)
	}
}

func (cb *CircuitBreaker) updateSetting(setting Setting) {
	for i := setting.MaxConcurrency; i < cb.setting.MaxConcurrency; i++ {
		// for narrow
		<-cb.tickets
	}

	for i := cb.setting.MaxConcurrency; i < setting.MaxConcurrency; i++ {
		// for expand
		cb.tickets <- struct{}{}
	}

	cb.setting.Timeout = setting.Timeout
	cb.setting.MuteWindow = setting.MuteWindow
	cb.setting.MaxConcurrency = setting.MaxConcurrency
	cb.setting.HealthCheckMinReqThreshold = setting.HealthCheckMinReqThreshold
	cb.setting.OpenOnFailureThreshold = setting.OpenOnFailureThreshold
}

func (cb *CircuitBreaker) feedback(eventType EventType, start, stop time.Time, runDuration time.Duration) {
	if cb.isOpen.Load() {
		cb.recovery.Update(eventType == SuccessEvent)
	}

	cb.statistic.Update(Event{
		eventType:        eventType,
		startAt:          start,
		stopAt:           stop,
		runDuration:      runDuration,
		concurrencyInUse: float64(cb.activeCount()) / float64(cb.setting.MaxConcurrency),
	})
}
