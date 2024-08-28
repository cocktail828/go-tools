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

func SHA256(src string, salt string) string {
	h := sha256.New()
	h.Write([]byte(src + salt))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
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

func Base64Decode(pwd string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

func Base64Encode(pwd string) string {
	return base64.StdEncoding.EncodeToString([]byte(pwd))
}

func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
