package hystrix

import (
	"encoding/json"
	"fmt"
	"time"
)

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

type Base[T any] struct{ val T }

func (b Base[T]) Get() T   { return b.val }
func (b *Base[T]) Set(v T) { b.val = v }

type Mutable[T int | time.Duration] struct {
	Val       Base[T]
	Validator Base[func(T) bool]
	OnUpdate  Base[func(T)]
}

func (m Mutable[T]) Equal(other Mutable[T]) bool {
	return m.Val.Get() == other.Val.Get()
}

func (m Mutable[T]) String() string {
	return fmt.Sprintf("%v", m.Val.Get())
}

func (m *Mutable[T]) Update(v T) {
	if validator := m.Validator.Get(); validator != nil && !validator(v) {
		return
	}

	oldVal := m.Val
	m.Val.Set(v)
	if oldVal != m.Val && m.OnUpdate.Get() != nil {
		m.OnUpdate.Get()(v)
	}
}

func (m Mutable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Val.Get())
}

func (m *Mutable[T]) UnmarshalJSON(data []byte) error {
	var Val T
	err := json.Unmarshal(data, &Val)
	if err == nil {
		m.Val.Set(Val)
		return nil
	}

	return fmt.Errorf("unmarshal Mutable failed, expect: %v, got: %w", Val, err)
}

type Config struct {
	Timeout           Mutable[time.Duration] `json:"timeout"`            // ms
	KeepAliveInterval Mutable[time.Duration] `json:"keepalive_interval"` // ms
	KeepAliveProbes   Mutable[int]           `json:"keepalive_probes"`
	MaxConcurrency    Mutable[int]           `json:"max_concurrency"`
	MinQPSThreshold   Mutable[int]           `json:"min_qps_threshold"`
	FailureThreshold  Mutable[int]           `json:"failure_threshold"` // 0~100
}

func NewConfig() Config {
	return Config{
		Timeout: Mutable[time.Duration]{
			Val:       Base[time.Duration]{val: DefaultTimeout},
			Validator: Base[func(v time.Duration) bool]{val: func(v time.Duration) bool { return v >= 0 }},
		},
		KeepAliveInterval: Mutable[time.Duration]{
			Val:       Base[time.Duration]{val: DefaultKeepAliveInterval},
			Validator: Base[func(v time.Duration) bool]{val: func(v time.Duration) bool { return v >= time.Second }},
		},
		KeepAliveProbes: Mutable[int]{
			Val:       Base[int]{val: DefaultRecoveryProbes},
			Validator: Base[func(v int) bool]{val: func(v int) bool { return v >= 3 }},
		},
		MaxConcurrency: Mutable[int]{
			Val:       Base[int]{val: DefaultMaxConcurrency},
			Validator: Base[func(v int) bool]{val: func(v int) bool { return v >= 1 }},
		},
		MinQPSThreshold: Mutable[int]{
			Val:       Base[int]{val: DefaultMinQPSNum},
			Validator: Base[func(v int) bool]{val: func(v int) bool { return v >= 0 }},
		},
		FailureThreshold: Mutable[int]{
			Val:       Base[int]{val: DefaultFailureThreshold},
			Validator: Base[func(v int) bool]{val: func(v int) bool { return v >= 0 && v <= 100 }},
		},
	}
}
