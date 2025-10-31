package utils

import (
	"errors"
	"newsclip/backend/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 서명에 사용할 비밀키
var jwtSecretKey = []byte(config.GetEnv("JWT_SECRET_KEY"))

// === [수정] JWT Claims 정의 ===
type JwtClaims struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"` // [추가] Nickname 필드
	jwt.RegisteredClaims
}

// === [수정] Access Token 생성 (username -> nickname) ===
func GenerateAccessToken(userID uint, nickname string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &JwtClaims{
		UserID:   userID,
		Nickname: nickname, // [추가] Nickname 값 할당
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
	expirationTime := time.Now().Add(168 * time.Hour) // 7일

	claims := &JwtClaims{
		UserID: userID,
		// Nickname은 Refresh 토큰에 필수는 아님
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

// [추가] Token 검증 (auth_middleware.go가 사용)
func ValidateToken(tokenString string) (*JwtClaims, error) {
	claims := &JwtClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("유효하지 않은 토큰")
	}

	return claims, nil
}