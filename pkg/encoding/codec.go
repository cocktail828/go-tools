package encoding

import (
	"crypto/md5"
	"crypto/sha256"
)

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode([]byte) ([]byte, error)
}

type Decoder interface {
	Decode([]byte) ([]byte, error)
}

type md5Encoder struct{}

func NewMD5Encoder() Encoder {
	return &md5Encoder{}
}

func (e *md5Encoder) Encode(bs []byte) ([]byte, error) {
	h := md5.New()
	h.Write(bs)
	return h.Sum(nil), nil
}

type sha256Encoder struct{}

func NewSHA256Encoder() Encoder {
	return &sha256Encoder{}
}

func (e *sha256Encoder) Encode(bs []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(bs)
	return h.Sum(nil), nil
}
