package hystrix

import "time"

var (
	// DefaultTimeout is how long to wait for command to complete, in milliseconds
	DefaultTimeout = 1000 * time.Millisecond
	// DefaultKeepAliveInterval is how long, in milliseconds, to wait after a circuitbreaker opens before testing for recovery
	DefaultKeepAliveInterval = 2000 * time.Millisecond
	// DefaultRecoveryProbes is how many probes circuitbreaker will try, before recovery
	DefaultRecoveryProbes = 5
	// DefaultMaxConcurrency is how many commands of the same type can run at the same time
	DefaultMaxConcurrency = 10
	// DefaultMinQPSNum is the minimum number of requests needed before a circuitbreaker can be tripped due to health
	DefaultMinQPSNum = 20
	// DefaultErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	DefaultFailureThreshold = 20 // 0~100
)

type Config struct {
	Timeout           time.Duration `json:"timeout"`            // ms
	KeepAliveInterval time.Duration `json:"keepalive_interval"` // ms
	KeepAliveProbes   int           `json:"keepalive_probes"`
	MaxConcurrency    int           `json:"max_concurrency"`
	MinQPSThreshold   int           `json:"min_qps_threshold"`
	FailureThreshold  int           `json:"failure_threshold"` // 0~100
}

func setter[T int | time.Duration](v T, defalt T) T {
	if v <= 0 {
		return defalt
	}
	return v
}

func (cfg *Config) Normalize() {
	cfg.Timeout = setter(0, DefaultTimeout)
	cfg.KeepAliveInterval = setter(0, DefaultKeepAliveInterval)
	cfg.KeepAliveProbes = setter(0, DefaultRecoveryProbes)
	cfg.MaxConcurrency = setter(0, DefaultMaxConcurrency)
	cfg.MinQPSThreshold = setter(0, DefaultMinQPSNum)
	cfg.FailureThreshold = setter(0, DefaultFailureThreshold)
}

func (cfg *Config) Update(config Config) {
	config.Normalize()
	cfg.Timeout = config.Timeout
	cfg.KeepAliveInterval = config.KeepAliveInterval
	cfg.KeepAliveProbes = config.KeepAliveProbes
	cfg.MaxConcurrency = config.MaxConcurrency
	cfg.MinQPSThreshold = config.MinQPSThreshold
	cfg.FailureThreshold = config.FailureThreshold
}
