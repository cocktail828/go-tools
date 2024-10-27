package nacs_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestConfigor(t *testing.T) {
	data, err := os.ReadFile("configor.go")
	z.Must(err)

	t.Run("none-config", func(t *testing.T) {
		cfgor, err := nacs.NewConfigor("", "", "", true)
		z.Must(err)

		// assert.EqualValues(t, data, cfgor.GetRawCfg()) // panic
		assert.Nil(t, cfgor.GetByName("configor.go"))
	})

	t.Run("one-config", func(t *testing.T) {
		cfgor, err := nacs.NewConfigor("", "", "", true, "configor.go")
		z.Must(err)

		assert.EqualValues(t, data, cfgor.GetRawCfg())
		assert.EqualValues(t, data, cfgor.GetByName("configor.go"))
	})

	t.Run("more-config", func(t *testing.T) {
		cfgor, err := nacs.NewConfigor("", "", "", true, "configor.go", "configor_test.go")
		z.Must(err)

		// assert.EqualValues(t, data, cfgor.GetRawCfg()) // panic
		assert.EqualValues(t, data, cfgor.GetByName("configor.go"))
	})
}
