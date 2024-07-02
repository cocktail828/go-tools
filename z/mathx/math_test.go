package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmp(t *testing.T) {
	assert.Equal(t, 1, Min(1, 2))
	assert.Equal(t, 1, Max(1, 0))
}
