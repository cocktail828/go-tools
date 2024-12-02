package chain

import (
	"context"
)

// UnaryInfo consists of various information about a unary RPC on
// server side. All per-rpc information may be mutated by the interceptor.
type UnaryInfo struct {
	// Server is the service implementation the user provides. This is read-only.
	Server any
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
}

type UnaryHandler func(ctx context.Context, req any) (any, error)
type UnaryInterceptor func(ctx context.Context, req any, info *UnaryInfo, handler UnaryHandler) (resp any, err error)

func getChainUnaryHandler(interceptors []UnaryInterceptor, curr int, info *UnaryInfo, finalHandler UnaryHandler) UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, req any) (any, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

func ChainUnaryInterceptors(interceptors []UnaryInterceptor) UnaryInterceptor {
	if len(interceptors) == 0 {
		return func(ctx context.Context, req any, info *UnaryInfo, handler UnaryHandler) (resp any, err error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req any, info *UnaryInfo, handler UnaryHandler) (any, error) {
		return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
	}
}
