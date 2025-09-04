package base64

import (
	"encoding/base64"
)

type Encoding struct{}

func NewCodec() *Encoding { return &Encoding{} }

func (enc *Encoding) Decode(bs []byte) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(string(bs))
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func (enc *Encoding) Encode(bs []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(bs)), nil
}
