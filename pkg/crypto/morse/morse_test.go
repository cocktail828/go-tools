package morse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	input     string // 输入值
	separator string // 分隔符
	output    string // 期望值
}{
	{"", "", ""},
	{"1", "/", ".----"},
	{"F", "/", "..-."},
	{"dongle", "|", "-..|---|-.|--.|.-..|."},
	{"SOS", "/", ".../---/..."},
}

func TestEncode(t *testing.T) {
	for index, test := range tests {
		dst, err := Encode(reflectx.StringToBytes(test.input), test.separator)

		t.Run(fmt.Sprintf("test_%d", index), func(t *testing.T) {
			assert.Nil(t, err)
			assert.Equal(t, test.output, dst)
		})
	}
}

func TestDecode(t *testing.T) {
	for index, test := range tests {
		dst, err := Decode(reflectx.StringToBytes(test.output), test.separator)

		t.Run(fmt.Sprintf("test_%d", index), func(t *testing.T) {
			assert.Nil(t, err)
			assert.Equal(t, strings.ToLower(test.input), dst)
		})
	}
}

func TestError(t *testing.T) {
	_, err1 := Encode([]byte("hello world"), "/")
	assert.EqualValues(t, "can't contain spaces", err1.Error())

	_, err2 := Decode([]byte("hello world"), "/")
	assert.EqualValues(t, "unknown character: hello world", err2.Error())
}
