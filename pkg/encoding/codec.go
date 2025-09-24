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

type NopCodec struct{}

func (NopCodec) Encode(s []byte) ([]byte, error) {
	return s, nil
}

func (NopCodec) Decode(s []byte) ([]byte, error) {
	return s, nil
}

type md5Encoder struct{}

func NewMD5Encoder() Encoder {
	return &md5Encoder{}
}

func (e *md5Encoder) Encode(s []byte) ([]byte, error) {
	bs := md5.Sum(s)
	return bs[:], nil
}

type sha256Encoder struct{}

func NewSHA256Encoder() Encoder {
	return &sha256Encoder{}
}

func (e *sha256Encoder) Encode(s []byte) ([]byte, error) {
	bs := sha256.Sum256(s)
	return bs[:], nil
}
