package reflectx_test

import (
	"io"
	"testing"

	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/stretchr/testify/assert"
)

type FakeCloser struct{}

func (FakeCloser) Close() error { return nil }

func TestIsNil(t *testing.T) {
	var k io.Closer = func() *FakeCloser {
		return nil
	}()
	assert.Equal(t, false, k == nil)
	assert.Equal(t, true, reflectx.IsNil(k))
}

func TestEnv(t *testing.T) {
	type xxx struct {
		A int    `env:"a" default:"1"`
		B string `env:"b" default:"2" validate:"required"`
		C int    `default:"3"`
	}
	x := &xxx{}
	assert.Equal(t, nil, reflectx.BindEnv(x))
	assert.Equal(t, 1, x.A)
	assert.Equal(t, "2", x.B)
	assert.Equal(t, 3, x.C)
}
