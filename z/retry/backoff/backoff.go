package backoff

import (
	"math"
	"math/rand"
)

type BackOff interface {
	Next(attempts int) int64
}

// Noop is a DelayType which do not delay
type Noop struct{}

func (d *Noop) Next(_ int) int64 {
	return 0
}

// Random is a DelayType which keeps delay the same through all iterations
type Random struct {
	Range int64 // default 100
	Min   int64
}

func (d *Random) Next(_ int) int64 {
	if d.Range == 0 {
		d.Range = 100
	}
	return rand.Int63()%d.Range + d.Min
}

// Fixed is a DelayType which keeps delay the same through all iterations
type Fixed struct {
	Value int64
}

func (d *Fixed) Next(_ int) int64 {
	return d.Value
}

// Exponential is a DelayType which increases delay between consecutive retries
type Exponential struct {
	Max int64 // default 30000
	Min int64
}

func (d *Exponential) Next(n int) int64 {
	if d.Max < 30000 {
		d.Max = 30000
	}
	trange := d.Max
	if v := int64(math.Pow(2, float64(n+1))); v <= d.Max {
		trange = v
	}
	return rand.Int63()%trange + d.Min
}
