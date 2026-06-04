package eventcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	kdfIterations = 100_000
	keySize       = 32 // AES-256
	saltSize      = 16
)

var ErrWrongPassword = errors.New("wrong password or corrupted data")

// Encrypt 使用 AES-256-GCM + PBKDF2-SHA256 加密 plaintext。
// 回傳 kdfSalt、cipherNonce、data 均為 base64 Standard 編碼。
func Encrypt(plaintext []byte, password string) (kdfSalt, cipherNonce, data string, err error) {
	salt := make([]byte, saltSize)
	if _, err = io.ReadFull(rand.Reader, salt); err != nil {
		return "", "", "", fmt.Errorf("generate salt: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, kdfIterations, keySize, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	kdfSalt = base64.StdEncoding.EncodeToString(salt)
	cipherNonce = base64.StdEncoding.EncodeToString(nonce)
	data = base64.StdEncoding.EncodeToString(ciphertext)
	return kdfSalt, cipherNonce, data, nil
}

// Decrypt 解密由 Encrypt 產生的資料。密碼錯誤或資料損毀時回傳 ErrWrongPassword。
func Decrypt(kdfSalt, cipherNonce, data, password string) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(kdfSalt)
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(cipherNonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, kdfIterations, keySize, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrWrongPassword
	}
	return plaintext, nil
}
