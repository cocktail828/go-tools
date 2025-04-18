package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/bcrypt"
)

func SHA256(src string) string {
	h := sha256.New()
	h.Write([]byte(src))
	data := h.Sum(nil)
	return hex.EncodeToString(data)
}

func MD5(str string) string {
	h := md5.New()
	data := h.Sum([]byte(str))
	return hex.EncodeToString(data)
}

// PasswordEncrypt encrypt password
func PasswordEncrypt(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

type Codec interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type Base64Codec struct{}

func (Base64Codec) Decode(bs []byte) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(string(bs))
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func (Base64Codec) Encode(bs []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(bs)), nil
}

type AESCodec struct {
	Key []byte
}

func (c AESCodec) Encode(bs []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(bs))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], bs)

	return ciphertext, nil
}

func (c AESCodec) Decode(bs []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	iv := bs[:aes.BlockSize]
	bs = bs[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(bs, bs)

	return bs, nil
}
