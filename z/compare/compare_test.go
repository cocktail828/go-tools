package compare_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z/compare"
	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	assert.Equal(t, true, compare.EqualValues(1, 1))
	assert.Equal(t, true, compare.EqualValues("a", "a"))
	assert.Equal(t, true, compare.EqualValues([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, true, compare.EqualValues(map[string]string{"a": "a", "b": "b"}, map[string]string{"b": "b", "a": "a"}))
	assert.Equal(t, false, compare.EqualValues(1, 2))
	assert.Equal(t, false, compare.EqualValues("a", "b"))
	assert.Equal(t, false, compare.EqualValues([]string{"a", "b"}, []string{"b", "a"}))
	assert.Equal(t, false, compare.EqualValues(map[string]string{"a": "a", "b": "b"}, map[string]string{"a": "a", "b": "a"}))
}
