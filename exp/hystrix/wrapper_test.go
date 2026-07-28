package hystrix

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGo(t *testing.T) {
	h := NewHystrix(NewConfig())

	assert.NoError(t, <-h.Go(t.Name(), func() error { return nil }))
	assert.Equal(t, net.ErrClosed, <-h.Go(t.Name(), func() error { return net.ErrClosed }))
}

func TestDo(t *testing.T) {
	h := NewHystrix(NewConfig())

	assert.NoError(t, h.Do(t.Name(), func() error { return nil }))
	assert.Equal(t, net.ErrClosed, h.Do(t.Name(), func() error { return net.ErrClosed }))
}

func TestDo_CircuitOpenViaTrigger(t *testing.T) {
	h := NewHystrix(NewConfig())
	h.Trigger(true)
	assert.Equal(t, ErrCircuitOpen, h.Do(t.Name(), func() error { return nil }))
}
