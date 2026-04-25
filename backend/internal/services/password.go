package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword 使用 Argon2id
func HashPassword(password string) (string, error) {
	// 1. 隨機 salt 16 bytes
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// 2. Argon2id 參數
	timeCost := uint32(3)       // 計算迭代次數
	memory := uint32(64 * 1024) // 64 MB
	threads := uint8(4)         // 並行度
	keyLen := uint32(32)        // hash 長度 32 bytes

	hash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, keyLen)

	// 3. 返回 base64 編碼
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// 存 DB 的字串可以包含參數
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, timeCost, threads, encodedSalt, encodedHash), nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}
	if version != argon2.Version {
		// 可選：直接 return false 或 warning
		return false, fmt.Errorf("unsupported argon2 version: %d", version)
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	// constant time 比對
	var diff byte
	for i := 0; i < len(hash); i++ {
		diff |= hash[i] ^ computed[i]
	}
	return diff == 0, nil
}
