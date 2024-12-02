package chain_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/chain"
)

type nop struct{ name string }

func (n nop) Name() string { return n.name }
func (n nop) Execute(c *chain.Context) {
	fmt.Println("pre handle by", n.name)
	c.Next()
	fmt.Println("post handle by", n.name)
}

func TestChain(t *testing.T) {
	c := chain.Chain{}
	z.Must(c.Use(nop{"a"}, nop{"b"}, nop{"c"}))
	c.Handle(context.Background(), nil)
}
