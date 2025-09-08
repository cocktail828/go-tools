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
		Start: func(_ context.Context) error {
			time.Sleep(time.Millisecond * 10)
			return nil
		},
	}
	assert.NoError(t, gs.GoContext(context.Background()))
}

func TestGS_Fail(t *testing.T) {
	gs := Graceful{
		Start: func(_ context.Context) error {
			time.Sleep(time.Millisecond * 10)
			return net.ErrClosed
		},
	}
	assert.ErrorIs(t, gs.GoContext(context.Background()), net.ErrClosed)
}

func TestGS_ParentCancel(t *testing.T) {
	gs := Graceful{
		Start: func(_ context.Context) error {
			time.Sleep(time.Millisecond * 200)
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, gs.GoContext(ctx), context.Canceled)
}
