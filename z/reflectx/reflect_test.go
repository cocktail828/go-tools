package reflectx_test

import (
	"io"
	"testing"

	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/stretchr/testify/assert"
)

type FakeCloser struct{}

func (FakeCloser) Close() error { return nil }

func TestIsNil(t *testing.T) {
	var k io.Closer = func() *FakeCloser {
		return nil
	}()
	assert.Equal(t, false, k == nil)
	assert.Equal(t, true, reflectx.IsNil(k))
}
