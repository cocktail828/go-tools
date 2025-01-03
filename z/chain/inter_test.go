package chain_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/cocktail828/go-tools/z/chain"
)

func TestRecurse(t *testing.T) {
	s := chain.ChainUnaryInterceptors([]chain.UnaryInterceptor[any]{
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("1 in")
			r, e := handler(ctx, in)
			fmt.Println("1 out")
			return r, e
		},
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("2 in")
			r, e := handler(ctx, in)
			fmt.Println("2 out")
			return r, e
		},
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("3 in")
			r, e := handler(ctx, in)
			fmt.Println("3 out")
			return r, e
		},
	})
	fmt.Println(s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		fmt.Println("end")
		return 10, nil
	}))
}

func TestFailure(t *testing.T) {
	s := chain.ChainUnaryInterceptors([]chain.UnaryInterceptor[any]{
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("1 in")
			r, e := handler(ctx, in)
			fmt.Println("1 out")
			return r, e
		},
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("2 in")
			return nil, net.ErrClosed
		},
		func(ctx context.Context, in any, handler chain.UnaryHandler[any]) (resp any, err error) {
			fmt.Println("3 in")
			r, e := handler(ctx, in)
			fmt.Println("3 out")
			return r, e
		},
	})
	fmt.Println(s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		fmt.Println("end")
		return nil, nil
	}))
}

func TestNilInter(t *testing.T) {
	s := chain.ChainUnaryInterceptors([]chain.UnaryInterceptor[any]{})
	fmt.Println(s(context.TODO(), nil, func(ctx context.Context, in any) (any, error) {
		fmt.Println("end")
		return nil, nil
	}))
}
