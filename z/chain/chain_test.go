package chain_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/chain"
	"github.com/stretchr/testify/assert"
)

type nop struct{ name string }

func (n nop) Handle(c *chain.Context) {
	c.Set("t", n.name)
	c.Next()
	c.Set("t", n.name)
}

type anop struct{ name string }

func (n anop) Handle(c *chain.Context) {
	c.Set("t", n.name)
	c.Abort()
	c.Set("t", n.name)
}

type T struct {
	req   any
	array []any
}

func (t *T) Set(key string, value any)  { t.array = append(t.array, value); fmt.Println(value) }
func (t *T) Get(key string) (any, bool) { return t.array, true }
func (t *T) Meta() any                  { return t.req }

func TestChain(t *testing.T) {
	c := chain.Chain{}
	z.Must(c.Use(nop{"a"}, nop{"b"}, nop{"c"}))
	x := &T{}
	c.Serve(x)
	val, _ := x.Get("t")
	assert.EqualValues(t, []any{"a", "b", "c", "c", "b", "a"}, val)
}

func TestAbort(t *testing.T) {
	c := chain.Chain{}
	z.Must(c.Use(nop{"a"}, nop{"b"}, anop{"xx"}, nop{"c"}))
	x := &T{}
	c.Serve(x)
	val, _ := x.Get("t")
	assert.EqualValues(t, []any{"a", "b", "xx", "xx", "b", "a"}, val)
}
