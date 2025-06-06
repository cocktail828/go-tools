package runnable

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGS_Success(t *testing.T) {
	gs := Graceful{
		Start: func() error {
			time.Sleep(time.Millisecond * 10)
			return nil
		},
	}
	assert.NoError(t, gs.Launch(context.Background()))
}

func TestGS_Fail(t *testing.T) {
	gs := Graceful{
		Start: func() error {
			time.Sleep(time.Millisecond * 10)
			return net.ErrClosed
		},
	}
	assert.Error(t, net.ErrClosed, gs.Launch(context.Background()))
}

func TestGS_ParentTimeout(t *testing.T) {
	gs := Graceful{
		Start: func() error {
			time.Sleep(time.Second)
			return nil
		},
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
	assert.Error(t, context.DeadlineExceeded, gs.Launch(ctx))
}

func TestGS_ParentCancel(t *testing.T) {
	gs := Graceful{
		Start: func() error {
			time.Sleep(time.Second)
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Error(t, context.Canceled, gs.Launch(ctx))
}
