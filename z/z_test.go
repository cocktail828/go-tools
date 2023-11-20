package z_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	type xxx struct {
		A int `env:"a,optional" default:"1"`
		B int `env:"b,required"`
		C int `default:"3"`
	}
	x := &xxx{}
	os.Setenv("b", "2")
	assert.Equal(t, nil, z.BindEnv(x))
	assert.Equal(t, 1, x.A)
	assert.Equal(t, 2, x.B)
	assert.Equal(t, 3, x.C)

	v := ""
	os.Setenv("v", "1")
	assert.Equal(t, nil, z.BindEnvVal(&v, "v"))
	assert.Equal(t, "1", v)
}
