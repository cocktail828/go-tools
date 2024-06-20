package caller_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z/caller"
	"github.com/stretchr/testify/assert"
)

func TestCaller(t *testing.T) {
	assert.Equal(t, "github.com/cocktail828/go-tools/z/caller_test.TestCaller", caller.Current().Name())
}
