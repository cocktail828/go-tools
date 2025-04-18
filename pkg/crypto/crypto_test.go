package crypto

import (
	"crypto/rand"
	"testing"
)

func TestBase64Codec(t *testing.T) {
	codec := Base64Codec{}
	testData := []byte("hello, world!")

	encoded, err := codec.Encode(testData)
	if err != nil {
		t.Fatalf("Base64 Encode failed: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Base64 Decode failed: %v", err)
	}

	if string(decoded) != string(testData) {
		t.Fatalf("Base64 Decode mismatch: got %q, want %q", decoded, testData)
	}
}

func TestAESCodec(t *testing.T) {
	// 生成随机 AES Key（16, 24, 32 字节对应 AES-128, AES-192, AES-256）
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	codec := AESCodec{Key: key}
	testData := []byte("hello, world!")

	encoded, err := codec.Encode(testData)
	if err != nil {
		t.Fatalf("AES Encode failed: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("AES Decode failed: %v", err)
	}

	if string(decoded) != string(testData) {
		t.Fatalf("AES Decode mismatch: got %q, want %q", decoded, testData)
	}
}

// TestAESCodec_InvalidKey 测试无效 AES Key
func TestAESCodec_InvalidKey(t *testing.T) {
	// 无效 Key（长度不为 16, 24, 32）
	invalidKey := []byte("invalid-key")
	codec := AESCodec{Key: invalidKey}
	testData := []byte("hello, world!")

	_, err := codec.Encode(testData)
	if err == nil {
		t.Fatal("AES Encode should fail with invalid key")
	}
}

// TestBase64Codec_InvalidInput 测试无效 Base64 输入
func TestBase64Codec_InvalidInput(t *testing.T) {
	codec := Base64Codec{}
	invalidBase64 := []byte("not-a-valid-base64!")

	_, err := codec.Decode(invalidBase64)
	if err == nil {
		t.Fatal("Base64 Decode should fail with invalid input")
	}
}
