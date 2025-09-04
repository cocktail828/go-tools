// Package base100 implements base100 encoding, fork from https://github.com/stek29/base100
package base100

import "errors"

var (
	// ErrInvalidLength is returned when length of string being decoded is
	// not divisible by four
	ErrInvalidLength = errors.New("base100: len(data) should be divisible by 4")

	// ErrInvalidData is returned if data is not a valid base100 string
	ErrInvalidData = errors.New("base100: data is invalid")
)

type Encoding struct{}

func NewEncoding() *Encoding {
	return &Encoding{}
}

func (enc *Encoding) Encode(src []byte) ([]byte, error) {
	buf := make([]byte, len(src)*4)
	for i, v := range src {
		buf[i*4+0] = 0xf0
		buf[i*4+1] = 0x9f
		buf[i*4+2] = byte((uint16(v)+55)/64 + 0x8f)
		buf[i*4+3] = (v+55)%64 + 0x80
	}
	return buf, nil
}

func (enc *Encoding) Decode(src []byte) ([]byte, error) {
	if len(src)%4 != 0 {
		return nil, ErrInvalidLength
	}
	buf := make([]byte, len(src)/4)
	for i := 0; i != len(src); i += 4 {
		if src[i+0] != 0xf0 || src[i+1] != 0x9f {
			return nil, ErrInvalidData
		}
		buf[i/4] = (src[i+2]-0x8f)*64 + src[i+3] - 0x80 - 55
	}
	return buf, nil
}
