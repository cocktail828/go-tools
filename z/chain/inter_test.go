package chain_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/cocktail828/go-tools/z/chain"
)

func TestXxx(t *testing.T) {
	s := chain.ChainUnaryInterceptors([]chain.UnaryInterceptor{
		func(ctx context.Context, req any, info chain.UnaryInfo, handler chain.UnaryHandler) (resp any, err error) {
			fmt.Println("1 in")
			r, e := handler(ctx, req)
			fmt.Println("1 out")
			return r, e
		},
		func(ctx context.Context, req any, info chain.UnaryInfo, handler chain.UnaryHandler) (resp any, err error) {
			fmt.Println("2 in")
			return nil, net.ErrClosed
			r, e := handler(ctx, req)
			fmt.Println("2 out")
			return r, e
		},
		func(ctx context.Context, req any, info chain.UnaryInfo, handler chain.UnaryHandler) (resp any, err error) {
			fmt.Println("3 in")
			r, e := handler(ctx, req)
			fmt.Println("3 out")
			return r, e
		},
	})
	s(context.TODO(), nil, chain.UnaryInfo{FullMethod: "xxx"}, func(ctx context.Context, req any) (any, error) {
		fmt.Println("end")
		return nil, nil
	})
}
