package reflectx

import (
	"io"
	"net"
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
	assert.Equal(t, false, IsNil(net.AddrError{"test", "127.0.0.1"}))
}

func TestBytesToString(t *testing.T) {
	for _, s := range []string{"", "a", "ab", "abc"} {
		assert.Equal(t, s, BytesToString(StringToBytes(s)))
	}
}
