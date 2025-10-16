package chain

import (
	"context"
)

type UnaryHandler func(ctx context.Context, in any) (any, error)
type UnaryInterceptor func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error)

func getChainUnaryHandler(interceptors []UnaryInterceptor, curr int, finalHandler UnaryHandler) UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, in any) (any, error) {
		return interceptors[curr+1](ctx, in, getChainUnaryHandler(interceptors, curr+1, finalHandler))
	}
}

func ChainUnaryInterceptors(interceptors ...UnaryInterceptor) UnaryInterceptor {
	if len(interceptors) == 0 {
		return func(ctx context.Context, in any, handler UnaryHandler) (resp any, err error) {
			return handler(ctx, in)
		}
	}

	return func(ctx context.Context, in any, handler UnaryHandler) (any, error) {
		return interceptors[0](ctx, in, getChainUnaryHandler(interceptors, 0, handler))
	}
}
