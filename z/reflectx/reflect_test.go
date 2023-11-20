package reflectx_test

import (
	"io"
	"os"
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
		A int `env:"a,optional" default:"1"`
		B int `env:"b,required"`
		C int `default:"3"`
	}
	x := &xxx{}
	os.Setenv("b", "2")
	assert.Equal(t, nil, reflectx.BindEnv(x))
	assert.Equal(t, 1, x.A)
	assert.Equal(t, 2, x.B)
	assert.Equal(t, 3, x.C)

	v := ""
	os.Setenv("v", "1")
	assert.Equal(t, nil, reflectx.BindEnvVal(&v, "v"))
	assert.Equal(t, "1", v)
}
