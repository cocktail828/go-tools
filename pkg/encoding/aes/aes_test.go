package aes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAES(t *testing.T) {
	codec := NewEncoding([]byte("1234567890123456"))
	for _, s := range []string{"hello world", "hello world 123"} {
		bs, err := codec.Encode([]byte(s))
		assert.NoError(t, err)

		bs, err = codec.Decode(bs)
		assert.NoError(t, err)
		if s != string(bs) {
			t.Fatalf("expect %s, but got %s", s, string(bs))
		}
	}
}
