package utils

import "golang.org/x/crypto/bcrypt"

// 비밀번호를 해싱합니다.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// 해시된 비밀번호와 원본 비밀번호를 비교합니다.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // err가 nil이면 비밀번호 일치
}
