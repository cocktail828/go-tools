package base64

import (
	"encoding/base64"
)

type Encoding struct{}

func NewCodec() *Encoding { return &Encoding{} }

func (enc *Encoding) Decode(s []byte) ([]byte, error) {
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	n, err := base64.StdEncoding.Decode(dbuf, []byte(s))
	return dbuf[:n], err
}

func (enc *Encoding) Encode(s []byte) ([]byte, error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
	base64.StdEncoding.Encode(buf, s)
	return buf, nil
}
