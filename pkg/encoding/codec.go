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

func (e *md5Encoder) Encode(s []byte) ([]byte, error) {
	return md5.New().Sum(s), nil
}

type sha256Encoder struct{}

func NewSHA256Encoder() Encoder {
	return &sha256Encoder{}
}

func (e *sha256Encoder) Encode(s []byte) ([]byte, error) {
	return sha256.New().Sum(s), nil
}
