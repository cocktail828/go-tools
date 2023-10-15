package z_test

import (
	"testing"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	assert.Equal(t, true, z.EqualValues(1, 1))
	assert.Equal(t, true, z.EqualValues("a", "a"))
	assert.Equal(t, true, z.EqualValues([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, true, z.EqualValues(map[string]string{"a": "a", "b": "b"}, map[string]string{"b": "b", "a": "a"}))
	assert.Equal(t, false, z.EqualValues(1, 2))
	assert.Equal(t, false, z.EqualValues("a", "b"))
	assert.Equal(t, false, z.EqualValues([]string{"a", "b"}, []string{"b", "a"}))
	assert.Equal(t, false, z.EqualValues(map[string]string{"a": "a", "b": "b"}, map[string]string{"a": "a", "b": "a"}))
}
