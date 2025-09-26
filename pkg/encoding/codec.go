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

type MD5Encoder struct{}

func (e MD5Encoder) Encode(s []byte) ([]byte, error) {
	bs := md5.Sum(s)
	return bs[:], nil
}

type SHA256Encoder struct{}

func (e SHA256Encoder) Encode(s []byte) ([]byte, error) {
	bs := sha256.Sum256(s)
	return bs[:], nil
}

type CRC32Encoder struct{}

func (e CRC32Encoder) Encode(s []byte) ([]byte, error) {
	crc := crc32.ChecksumIEEE(s)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, crc)
	return buf, nil
}

type CRC64Encoder struct{}

func (e CRC64Encoder) Encode(s []byte) ([]byte, error) {
	crc := crc64.Checksum(s, crc64.MakeTable(crc64.ECMA))
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, crc)
	return buf, nil
}
