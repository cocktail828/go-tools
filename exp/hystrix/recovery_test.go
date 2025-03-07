package hystrix_test

import (
	"testing"

	"github.com/cocktail828/go-tools/exp/hystrix"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	r := hystrix.NewRecovery(3)
	for i := 0; i < 5; i++ {
		r.Update(true)
		if i < 2 {
			assert.EqualValues(t, false, r.IsHealthy())
		} else {
			assert.EqualValues(t, true, r.IsHealthy())
		}
	}
}
