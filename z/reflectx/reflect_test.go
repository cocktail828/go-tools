package reflectx

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeCloser struct{}

func (FakeCloser) Close() error { return nil }

func TestIsNil(t *testing.T) {
	var k io.Closer = func() *FakeCloser {
		return nil
	}()
	assert.Equal(t, false, k == nil)
	assert.Equal(t, true, IsNil(k))
}
