package utils

import "golang.org/x/crypto/bcrypt"

// 비밀번호를 해싱합니다.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
