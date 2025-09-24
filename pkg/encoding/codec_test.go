package encoding

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	t.Run("md5", func(t *testing.T) {
		encoder := NewMD5Encoder()
		bs, _ := encoder.Encode([]byte("123"))
		assert.Equal(t, "202cb962ac59075b964b07152d234b70", hex.EncodeToString(bs))
	})

	t.Run("sha256", func(t *testing.T) {
		encoder := NewSHA256Encoder()
		bs, _ := encoder.Encode([]byte("123"))
		assert.Equal(t, "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", hex.EncodeToString(bs))
	})

	t.Run("crc32", func(t *testing.T) {
		encoder := NewCRC32Encoder()
		bs, _ := encoder.Encode([]byte("123"))
		assert.Equal(t, "884863d2", hex.EncodeToString(bs))
	})

	t.Run("crc64", func(t *testing.T) {
		encoder := NewCRC64Encoder()
		bs, _ := encoder.Encode([]byte("123"))
		assert.Equal(t, "30232844071cc561", hex.EncodeToString(bs))
	})
}
