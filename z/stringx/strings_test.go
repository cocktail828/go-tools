package stringx_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z/stringx"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.True(t, stringx.EqualValues([]string{"a", "c"}, stringx.Unique([]string{"a", "c", "c"})))
	assert.True(t, stringx.Subset([]string{"a", "c"}, []string{"a", "c", "d"}))
	assert.False(t, stringx.Subset([]string{"a", "e"}, []string{"a", "c", "d"}))
}
