package chain

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecurse(t *testing.T) {
	slice := []string{}
	s := ChainInterceptors(
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "1 in")
			r, e := handler(ctx, in)
			slice = append(slice, "1 out")
			return r, e
		},
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "2 in")
			r, e := handler(ctx, in)
			slice = append(slice, "2 out")
			return r, e
		},
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "3 in")
			r, e := handler(ctx, in)
			slice = append(slice, "3 out")
			return r, e
		},
	)

	val, err := s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		slice = append(slice, "end")
		return 10, nil
	})
	assert.Equal(t, val, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"1 in", "2 in", "3 in", "end", "3 out", "2 out", "1 out"}, slice)
}

func TestFailure(t *testing.T) {
	slice := []string{}
	s := ChainInterceptors(
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "1 in")
			r, e := handler(ctx, in)
			slice = append(slice, "1 out")
			return r, e
		},
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "2 in")
			return nil, net.ErrClosed
		},
		func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			slice = append(slice, "3 in")
			r, e := handler(ctx, in)
			slice = append(slice, "3 out")
			return r, e
		},
	)

	_, err := s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		slice = append(slice, "end")
		return 10, nil
	})
	assert.EqualError(t, err, net.ErrClosed.Error())
	assert.EqualValues(t, []string{"1 in", "2 in", "1 out"}, slice)
}

func TestNilInter(t *testing.T) {
	s := ChainInterceptors()
	_, err := s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		return nil, nil
	})
	assert.NoError(t, err)
}
