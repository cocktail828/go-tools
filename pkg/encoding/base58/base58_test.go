package base58

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	input  string // 输入值
	output string // 期望值
}{
	{"", ""},
	{"hello world", "StV1DL6CwTryKyV"},
}

func TestEncode(t *testing.T) {
	for _, test := range tests {
		bs, err := StdEncoding.Encode([]byte(test.input))
		assert.NoError(t, err)
		assert.Equal(t, test.output, string(bs))
	}
}

func TestDecode(t *testing.T) {
	for _, test := range tests {
		bs, err := StdEncoding.Decode([]byte(test.output))
		assert.NoError(t, err)
		assert.Equal(t, test.input, string(bs))
	}
}

func TestCodec(t *testing.T) {
	for i := 0; i < 1000; i++ {
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		input := make([]byte, 100)
		rand.Read(input)

		bs, err := StdEncoding.Encode(input)
		assert.NoError(t, err)

		bs, err = StdEncoding.Decode(bs)
		assert.NoError(t, err)
		if !bytes.Equal(bs, input) {
			t.Fatalf("expect %s, but got %s", input, string(bs))
		}
	}
}
