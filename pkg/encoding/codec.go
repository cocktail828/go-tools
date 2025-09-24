package encoding

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"hash/crc32"
	"hash/crc64"
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

type crc32Encoder struct{}

func NewCRC32Encoder() Encoder {
	return &crc32Encoder{}
}

func (e *crc32Encoder) Encode(s []byte) ([]byte, error) {
	crc := crc32.ChecksumIEEE(s)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, crc)
	return buf, nil
}

type crc64Encoder struct{}

func NewCRC64Encoder() Encoder {
	return &crc64Encoder{}
}

func (e *crc64Encoder) Encode(s []byte) ([]byte, error) {
	crc := crc64.Checksum(s, crc64.MakeTable(crc64.ECMA))
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, crc)
	return buf, nil
}
