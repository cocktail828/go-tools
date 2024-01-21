package configor_test

import (
	"testing"

	"github.com/cocktail828/go-tools/configor"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

type Demo struct {
	A int    `env:"A" default:"10"`
	S string `required:"true"`
}

func TestConfigor(t *testing.T) {
	func() {
		d := Demo{}
		z.Must(configor.Load(&d, []byte(`A = 1
		S = '1111'`)))
		assert.Equal(t, Demo{1, "1111"}, d)
	}()

	func() {
		d := Demo{}
		z.Must(configor.Load(&d, []byte("S = '1111'")))
		assert.Equal(t, Demo{10, "1111"}, d)
	}()
}
