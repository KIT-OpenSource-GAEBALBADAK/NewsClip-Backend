package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// 32바이트(64글자) 랜덤 16진수 문자열 생성
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
