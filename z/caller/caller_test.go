package caller_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z/caller"
	"github.com/stretchr/testify/assert"
)

func inner(t *testing.T) {
	bts := caller.Backtrace()
	assert.Equal(t, 2, len(bts))
	assert.Equal(t, "github.com/cocktail828/go-tools/z/caller_test.inner", bts[0].Name)
	assert.Equal(t, "github.com/cocktail828/go-tools/z/caller_test.TestBacktrace", bts[1].Name)
}

func TestBacktrace(t *testing.T) {
	inner(t)
}

func TestCaller(t *testing.T) {
	assert.Equal(t, "github.com/cocktail828/go-tools/z/caller_test.TestCaller", caller.Last().Name)
}

func BenchmarkCaller(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			caller.Backtrace()
		}
	})
}
