package chain

import (
	"context"
)

type UnaryHandler[T any] func(ctx context.Context, in T) (any, error)
type UnaryInterceptor[T any] func(ctx context.Context, in T, handler UnaryHandler[T]) (resp any, err error)

func getChainUnaryHandler[T any](interceptors []UnaryInterceptor[T], curr int, finalHandler UnaryHandler[T]) UnaryHandler[T] {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, in T) (any, error) {
		return interceptors[curr+1](ctx, in, getChainUnaryHandler(interceptors, curr+1, finalHandler))
	}
}

func ChainInterceptors[T any](interceptors ...UnaryInterceptor[T]) UnaryInterceptor[T] {
	if len(interceptors) == 0 {
		return func(ctx context.Context, in T, handler UnaryHandler[T]) (resp any, err error) {
			return handler(ctx, in)
		}
	}

	return func(ctx context.Context, in T, handler UnaryHandler[T]) (any, error) {
		return interceptors[0](ctx, in, getChainUnaryHandler(interceptors, 0, handler))
	}
}
