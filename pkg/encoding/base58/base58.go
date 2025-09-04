// Package base58 implements base58 encoding
package base58

import (
	"errors"
)

var (
	ErrIllegalInput    = errors.New("base58: illegal base58 data")
	ErrInvalidAlphabet = errors.New("base58: the alphabet length must be 58")
)

var (
	encodeStd      = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	StdEncoding, _ = NewEncoding(encodeStd)
)

type Encoding struct {
	alphabet [58]byte
	index    [256]int
}

// NewCodec returns a new padded Encoding defined by the given alphabet,
// which must be a 58-byte string
func NewEncoding(alphabet string) (*Encoding, error) {
	e := new(Encoding)
	if len(alphabet) != 58 {
		return nil, ErrInvalidAlphabet
	}
	copy(e.alphabet[:], alphabet)

	for i := range e.index {
		e.index[i] = -1
	}
	for i, c := range alphabet {
		if c > 255 {
			return nil, errors.New("base58: alphabet contains invalid characters")
		}
		e.index[byte(c)] = i
	}
	return e, nil
}

func (enc *Encoding) Encode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	digits := make([]int, 0, len(src)*2)
	for _, b := range src {
		carry := int(b)
		for j := 0; j < len(digits) || carry > 0; j++ {
			if j < len(digits) {
				carry += digits[j] << 8
				digits[j] = carry % 58
			} else {
				digits = append(digits, carry%58)
			}
			carry = carry / 58
		}
	}

	var leadingZeros int
	for _, b := range src {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	dst := make([]byte, leadingZeros+len(digits))
	for i := 0; i < leadingZeros; i++ {
		dst[i] = enc.alphabet[0]
	}
	for i, d := range digits {
		dst[leadingZeros+len(digits)-1-i] = enc.alphabet[d]
	}
	return dst, nil
}

func (enc *Encoding) Decode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	for _, c := range src {
		if enc.index[c] == -1 {
			return nil, ErrIllegalInput
		}
	}

	bytes := make([]int, 0, len(src))
	for _, c := range src {
		if c == enc.alphabet[0] && len(bytes) == 0 {
			continue
		}
		carry := enc.index[c]
		for j := 0; j < len(bytes) || carry > 0; j++ {
			if j < len(bytes) {
				carry += bytes[j] * 58
				bytes[j] = carry & 0xff
			} else {
				bytes = append(bytes, carry&0xff)
			}
			carry = carry >> 8
		}
	}

	var leadingZeros int
	for _, c := range src {
		if c != enc.alphabet[0] {
			break
		}
		leadingZeros++
	}

	dst := make([]byte, leadingZeros+len(bytes))
	for i := 0; i < leadingZeros; i++ {
		dst[i] = 0
	}
	for i, b := range bytes {
		dst[leadingZeros+len(bytes)-1-i] = byte(b)
	}
	return dst, nil
}
