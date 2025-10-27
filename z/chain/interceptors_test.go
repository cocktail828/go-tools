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
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "1")
			r, e := handler(ctx, in)
			slice = append(slice, "1")
			return r, e
		},
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "2")
			r, e := handler(ctx, in)
			slice = append(slice, "2")
			return r, e
		},
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "3")
			r, e := handler(ctx, in)
			slice = append(slice, "3")
			return r, e
		},
	)

	val, err := s(context.TODO(), "test", func(ctx context.Context, in string) (any, error) {
		slice = append(slice, "end")
		return 10, nil
	})
	assert.Equal(t, val, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"1", "2", "3", "end", "3", "2", "1"}, slice)
}

func TestFailure(t *testing.T) {
	slice := []string{}
	s := ChainInterceptors(
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "1")
			r, e := handler(ctx, in)
			slice = append(slice, "1")
			return r, e
		},
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "2")
			return nil, net.ErrClosed
		},
		func(ctx context.Context, in string, handler UnaryHandler[string]) (resp any, err error) {
			slice = append(slice, "3")
			r, e := handler(ctx, in)
			slice = append(slice, "3")
			return r, e
		},
	)

	_, err := s(context.TODO(), "test", func(ctx context.Context, in string) (any, error) {
		slice = append(slice, "end")
		return 10, nil
	})
	assert.EqualError(t, err, net.ErrClosed.Error())
	assert.EqualValues(t, []string{"1", "2", "1"}, slice)
}

func TestEmptyInterceptors(t *testing.T) {
	s := ChainInterceptors[string]()
	_, err := s(context.TODO(), "test", func(ctx context.Context, in string) (any, error) {
		return nil, nil
	})
	assert.NoError(t, err)
}
