package utils

import (
	"newsclip/backend/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 서명에 사용할 비밀키
var jwtSecretKey = []byte(config.GetEnv("JWT_SECRET_KEY"))

// JWT Claims 정의
type JwtClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Access Token 생성
func GenerateAccessToken(userID uint, username string) (string, error) {
	// Access Token 만료 시간 (예: 1시간)
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &JwtClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// Refresh Token 생성
func GenerateRefreshToken(userID uint) (string, time.Time, error) {
	// Refresh Token 만료 시간 (예: 7일)
	expirationTime := time.Now().Add(168 * time.Hour) // 24 * 7

	claims := &JwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}
