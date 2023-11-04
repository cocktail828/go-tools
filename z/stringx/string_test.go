package stringx_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z/stringx"
	"github.com/stretchr/testify/assert"
)

func TestXxx(t *testing.T) {
	assert.Equal(t, true, stringx.Contains([]string{"a", "b"}, "a"))
	assert.Equal(t, false, stringx.Contains([]string{"b", "c"}, "a"))
	assert.Equal(t, true, stringx.Equal([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, false, stringx.Equal([]string{"a"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{"a"}, stringx.Overlap([]string{"a"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a"}, stringx.Overlap([]string{"a"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a"}, stringx.Elimate([]string{"a"}, []string{}))
	assert.Equal(t, []string{"a"}, stringx.Elimate([]string{"a", "b"}, []string{"b"}))
}
